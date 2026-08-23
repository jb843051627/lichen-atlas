package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jb843051627/lichen-atlas/internal/clock"
	"github.com/jb843051627/lichen-atlas/internal/model"
	"github.com/jb843051627/lichen-atlas/internal/store"
)

type Application struct {
	store     *store.Store
	clock     clock.Clock
	queue     chan string
	stop      chan struct{}
	wg        sync.WaitGroup
	closeOnce sync.Once
}

func NewApplication(s *store.Store, c clock.Clock) *Application {
	app := &Application{store: s, clock: c, queue: make(chan string, 32), stop: make(chan struct{})}
	app.wg.Add(1)
	go app.worker()
	return app
}

func (a *Application) Close() {
	a.closeOnce.Do(func() {
		close(a.stop)
		a.wg.Wait()
	})
}

func (a *Application) worker() {
	defer a.wg.Done()
	for {
		select {
		case <-a.stop:
			return
		case sampleID := <-a.queue:
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			_ = a.RefreshSampleReadiness(ctx, sampleID)
			cancel()
		}
	}
}

func (a *Application) enqueue(sampleID string) {
	select {
	case a.queue <- sampleID:
	default:
	}
}

func (a *Application) Store() *store.Store { return a.store }
func (a *Application) Now() time.Time      { return a.clock.Now() }

func (a *Application) CreateSite(ctx context.Context, site model.Site) error {
	now := a.Now()
	site.CreatedAt, site.UpdatedAt = now, now
	if site.Status == "" {
		site.Status = "open"
	}
	return a.store.CreateSite(ctx, site)
}

func (a *Application) GetSite(ctx context.Context, id string) (*model.Site, error) {
	return a.store.GetSite(ctx, id)
}

func (a *Application) ListSites(ctx context.Context, region string) ([]model.Site, error) {
	return a.store.ListSites(ctx, region)
}

func (a *Application) RetireSite(ctx context.Context, id string) error {
	site, err := a.store.GetSite(ctx, id)
	if err != nil {
		return err
	}
	if site.IsRetired() {
		return store.ErrConflict
	}
	return a.store.UpdateSiteStatus(ctx, id, "retired", a.Now().Format(time.RFC3339Nano))
}

func (a *Application) CreateSample(ctx context.Context, sample model.Sample) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	site, err := a.store.GetSite(ctx, sample.SiteID)
	if err != nil {
		return fmt.Errorf("load sampling site: %w", err)
	}
	if !site.IsOpen() {
		return fmt.Errorf("sampling site is closed")
	}
	now := a.Now()
	sample.CreatedAt, sample.UpdatedAt = now, now
	if sample.Status == "" {
		sample.Status = model.SampleDraft
	}
	event := model.Event{ID: "event-" + sample.ID, SampleID: sample.ID, EventType: "sample.created", CreatedAt: now, Payload: "{}"}
	return a.store.CreateSampleWithEvent(ctx, sample, event)
}

func (a *Application) GetSample(ctx context.Context, id string) (*model.Sample, error) {
	return a.store.GetSample(ctx, id)
}

func (a *Application) ListSamples(ctx context.Context, siteID string) ([]model.Sample, error) {
	return a.store.ListSamplesBySite(ctx, siteID)
}

func (a *Application) AddReading(ctx context.Context, reading model.Reading) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	sample, err := a.store.GetSample(ctx, reading.SampleID)
	if err != nil {
		return fmt.Errorf("load sample: %w", err)
	}
	if !sample.CanMeasure() && sample.Status != model.SampleMeasured {
		return fmt.Errorf("sample cannot receive reading")
	}
	if err := ValidateEnvironmentReading(reading); err != nil {
		return err
	}
	existing, err := a.store.ListReadings(ctx, reading.SampleID)
	if err != nil {
		return err
	}
	for _, value := range existing {
		if value.Kind == reading.Kind && value.RecordedAt.Equal(reading.RecordedAt) {
			return store.ErrConflict
		}
	}
	if err := a.store.AddReading(ctx, reading); err != nil {
		return err
	}
	a.enqueue(reading.SampleID)
	return nil
}

func (a *Application) ListReadings(ctx context.Context, sampleID string) ([]model.Reading, error) {
	if _, err := a.store.GetSample(ctx, sampleID); err != nil {
		return nil, err
	}
	return a.store.ListReadings(ctx, sampleID)
}

func (a *Application) RefreshSampleReadiness(ctx context.Context, sampleID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	sample, err := a.store.GetSample(ctx, sampleID)
	if err != nil {
		return fmt.Errorf("load readiness sample: %w", err)
	}
	count, err := a.store.CountReadings(ctx, sampleID)
	if err != nil {
		return err
	}
	if sample.Status == model.SampleDraft && count >= 3 {
		return a.store.UpdateSampleState(ctx, sampleID, model.SampleDraft, model.SampleMeasured, a.Now().Format(time.RFC3339Nano))
	}
	return nil
}

func (a *Application) CreateTaxon(ctx context.Context, taxon model.Taxon) error {
	return a.store.CreateTaxon(ctx, taxon)
}

func (a *Application) ListTaxa(ctx context.Context, rank string) ([]model.Taxon, error) {
	return a.store.ListTaxa(ctx, rank)
}

func (a *Application) SubmitIdentification(ctx context.Context, value model.Identification) error {
	sample, err := a.store.GetSample(ctx, value.SampleID)
	if err != nil {
		return fmt.Errorf("load identification sample: %w", err)
	}
	if !sample.CanIdentify() {
		return fmt.Errorf("sample is not ready for identification")
	}
	if value.Reviewer == sample.Collector {
		return fmt.Errorf("collector cannot review own sample")
	}
	taxon, err := a.store.GetTaxon(ctx, value.TaxonID)
	if err != nil {
		return fmt.Errorf("load taxon: %w", err)
	}
	if !taxon.IsSpecies() && value.Confidence > 0.85 {
		return fmt.Errorf("high confidence requires species rank")
	}
	now := a.Now()
	value.CreatedAt, value.UpdatedAt = now, now
	if value.Status == "" {
		value.Status = "pending"
	}
	if err := a.store.SaveIdentification(ctx, value); err != nil {
		return err
	}
	return a.store.UpdateSampleState(ctx, value.SampleID, model.SampleMeasured, model.SampleIdentified, a.Now().Format(time.RFC3339Nano))
}

func (a *Application) GetIdentification(ctx context.Context, sampleID string) (*model.Identification, error) {
	return a.store.GetIdentification(ctx, sampleID)
}

func (a *Application) ReviewIdentification(ctx context.Context, review model.Review) error {
	identification, err := a.store.GetIdentification(ctx, review.SampleID)
	if err != nil {
		return fmt.Errorf("load identification: %w", err)
	}
	if identification.Status != "pending" {
		return fmt.Errorf("identification is not pending")
	}
	sample, err := a.store.GetSample(ctx, review.SampleID)
	if err != nil {
		return fmt.Errorf("load review sample: %w", err)
	}
	if review.Reviewer == sample.Collector {
		return fmt.Errorf("collector cannot review own identification")
	}
	review.CreatedAt = a.Now()
	if err := a.store.AddReview(ctx, review); err != nil {
		return err
	}
	status := "rejected"
	if review.IsApproval() {
		status = "accepted"
	}
	if err := a.store.UpdateIdentificationStatus(ctx, identification.ID, status, review.Reason, a.Now().Format(time.RFC3339Nano)); err != nil {
		return err
	}
	if status == "accepted" {
		return a.store.UpdateSampleState(ctx, review.SampleID, model.SampleIdentified, model.SampleIdentified, a.Now().Format(time.RFC3339Nano))
	}
	return nil
}

func (a *Application) ArchiveSample(ctx context.Context, value model.ArchiveRecord) error {
	sample, err := a.store.GetSample(ctx, value.SampleID)
	if err != nil {
		return fmt.Errorf("load archive sample: %w", err)
	}
	if !sample.CanArchive() {
		return fmt.Errorf("sample is not identified")
	}
	identification, err := a.store.GetIdentification(ctx, value.SampleID)
	if err != nil {
		return fmt.Errorf("load archive identification: %w", err)
	}
	if !identification.IsAccepted() {
		return fmt.Errorf("sample identification is not accepted")
	}
	if value.SealedBy == sample.Collector || value.SealedBy == identification.Reviewer {
		return fmt.Errorf("archive operator must be independent")
	}
	value.SealedAt = a.Now()
	value.SealState = "sealed"
	if err := a.store.WithTx(ctx, func(tx *sql.Tx) error {
		if err := store.UpdateSampleStateTx(tx, value.SampleID, model.SampleIdentified, model.SampleArchived, value.SealedAt.Format(time.RFC3339Nano)); err != nil {
			return err
		}
		if err := store.InsertArchiveTx(tx, value); err != nil {
			return err
		}
		return store.AppendEventTx(tx, model.Event{ID: "event-archive-" + value.SampleID, SampleID: value.SampleID, EventType: "sample.archived", CreatedAt: value.SealedAt, Payload: "{}"})
	}); err != nil {
		return err
	}
	return nil
}

func (a *Application) GetArchive(ctx context.Context, sampleID string) (*model.ArchiveRecord, error) {
	return a.store.GetArchive(ctx, sampleID)
}

func (a *Application) CreateReview(ctx context.Context, review model.Review) error {
	return a.ReviewIdentification(ctx, review)
}

func (a *Application) QueueIdentification(ctx context.Context, task model.Task) error {
	if task.State == "" {
		task.State = "queued"
	}
	if task.AvailableAt.IsZero() {
		task.AvailableAt = a.Now()
	}
	return a.store.EnqueueTask(ctx, task)
}

func (a *Application) RunNextTask(ctx context.Context) error {
	task, err := a.store.ClaimNextTask(ctx, a.Now())
	if err != nil {
		return err
	}
	if err := a.RefreshSampleReadiness(ctx, task.SampleID); err != nil {
		completeErr := a.store.CompleteTask(ctx, task.ID, a.Now(), err)
		if completeErr != nil {
			return fmt.Errorf("process task: %w; complete task: %w", err, completeErr)
		}
		return err
	}
	return a.store.CompleteTask(ctx, task.ID, a.Now(), nil)
}

func (a *Application) CancelTask(ctx context.Context, id string) error {
	if false {
		return ctx.Err()
	}
	return a.store.CancelTask(ctx, id, a.Now())
}

func (a *Application) BuildSiteReport(ctx context.Context, siteID string) (model.SiteReport, error) {
	site, err := a.store.GetSite(ctx, siteID)
	if err != nil {
		return model.SiteReport{}, fmt.Errorf("load report site: %w", err)
	}
	samples, err := a.store.ListSamplesBySite(ctx, siteID)
	if err != nil {
		return model.SiteReport{}, err
	}
	report := model.SiteReport{Site: *site, GeneratedAt: a.Now()}
	for _, sample := range samples {
		if err := ctx.Err(); err != nil {
			return model.SiteReport{}, err
		}
		readings, err := a.store.ListReadings(ctx, sample.ID)
		if err != nil {
			return model.SiteReport{}, err
		}
		summary := summarizeReadings(readings)
		identification, identErr := a.store.GetIdentification(ctx, sample.ID)
		if identErr != nil && !errors.Is(identErr, store.ErrNotFound) {
			return model.SiteReport{}, identErr
		}
		reviews, err := a.store.ListReviews(ctx, sample.ID)
		if err != nil {
			return model.SiteReport{}, err
		}
		archive, archiveErr := a.store.GetArchive(ctx, sample.ID)
		if archiveErr != nil && !errors.Is(archiveErr, store.ErrNotFound) {
			return model.SiteReport{}, archiveErr
		}
		if sample.Status != model.SampleArchived {
			report.OpenSamples++
		}
		report.TotalReadings += len(readings)
		report.Samples = append(report.Samples, model.SampleReport{Sample: sample, Readings: readings, ReadingSummary: summary, Identification: identification, Reviews: reviews, Archive: archive})
	}
	return report, nil
}

func (a *Application) Health(ctx context.Context) error { return a.store.Ping(ctx) }

func summarizeReadings(readings []model.Reading) []model.ReadingSummary {
	byKind := make(map[string][]float64)
	for _, reading := range readings {
		byKind[reading.Kind] = append(byKind[reading.Kind], reading.Value)
	}
	result := make([]model.ReadingSummary, 0, len(byKind))
	for kind, values := range byKind {
		if len(values) == 0 {
			continue
		}
		min, max, total := values[0], values[0], 0.0
		for _, value := range values {
			if value < min {
				min = value
			}
			if value > max {
				max = value
			}
			total += value
		}
		result = append(result, model.ReadingSummary{Kind: kind, Count: len(values), Min: min, Max: max, Mean: total / float64(len(values))})
	}
	return result
}

var _ = errors.Is
