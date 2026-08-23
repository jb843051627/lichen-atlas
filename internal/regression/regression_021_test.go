package regression

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jb843051627/lichen-atlas/internal/clock"
	"github.com/jb843051627/lichen-atlas/internal/handler"
	"github.com/jb843051627/lichen-atlas/internal/model"
	"github.com/jb843051627/lichen-atlas/internal/service"
	"github.com/jb843051627/lichen-atlas/internal/store"
)

var fixtureTime = time.Date(2026, 8, 22, 8, 0, 0, 0, time.UTC)

func fixture(t *testing.T) (*store.Store, *service.Application, model.Sample) {
	t.Helper()
	db, err := store.Open(t.TempDir() + "/survey.db")
	if err != nil {
		t.Fatal(err)
	}
	app := service.NewApplication(db, clock.Fixed{Value: fixtureTime})
	t.Cleanup(func() {
		defer func() { _ = recover() }()
		app.Close()
		_ = db.Close()
	})
	site := model.Site{ID: "site-001", Name: "North Ridge", Region: "north", ElevationM: 1800, Status: "open"}
	if err := app.CreateSite(context.Background(), site); err != nil {
		t.Fatal(err)
	}
	sample := model.Sample{ID: "sample-001", SiteID: site.ID, Collector: "field-team", Condition: "fresh", CollectedAt: fixtureTime}
	if err := app.CreateSample(context.Background(), sample); err != nil {
		t.Fatal(err)
	}
	return db, app, sample
}

func addReading(t *testing.T, app *service.Application, sampleID, id, kind string, value float64, at time.Time) {
	t.Helper()
	reading := model.Reading{ID: id, SampleID: sampleID, Kind: kind, Value: value, Unit: map[string]string{"temperature": "C", "humidity": "%", "light": "lux"}[kind], RecordedAt: at}
	if err := app.AddReading(context.Background(), reading); err != nil {
		t.Fatal(err)
	}
}

func addThreeReadings(t *testing.T, app *service.Application, sampleID string) {
	addReading(t, app, sampleID, "reading-1", "temperature", 4, fixtureTime.Add(time.Minute))
	addReading(t, app, sampleID, "reading-2", "humidity", 60, fixtureTime.Add(2*time.Minute))
	addReading(t, app, sampleID, "reading-3", "light", 120, fixtureTime.Add(3*time.Minute))
	if err := app.RefreshSampleReadiness(context.Background(), sampleID); err != nil {
		if !errors.Is(err, store.ErrConflict) {
			t.Fatal(err)
		}
	}
}

func addAcceptedIdentification(t *testing.T, app *service.Application, sample model.Sample) {
	t.Helper()
	if err := app.CreateTaxon(context.Background(), model.Taxon{ID: "taxon-001", Scientific: "Usnea longissima", Rank: "species"}); err != nil {
		t.Fatal(err)
	}
	addThreeReadings(t, app, sample.ID)
	identification := model.Identification{ID: "ident-001", SampleID: sample.ID, TaxonID: "taxon-001", Reviewer: "reviewer", Confidence: 0.8}
	if err := app.SubmitIdentification(context.Background(), identification); err != nil {
		t.Fatal(err)
	}
	if err := app.ReviewIdentification(context.Background(), model.Review{ID: "review-001", SampleID: sample.ID, Reviewer: "reviewer", Decision: "approve"}); err != nil {
		t.Fatal(err)
	}
}

func TestBug01_MissingSiteErrorReachesHTTPBoundary(t *testing.T) {
	_, app, _ := fixture(t)
	body := `{"ID":"missing-sample","SiteID":"absent-site","Collector":"field-team","CollectedAt":"2026-08-22T08:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/api/samples", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.NewServer(app).Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("missing site returned HTTP %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestBug02_MissingSampleKeepsNotFoundError(t *testing.T) {
	_, app, _ := fixture(t)
	_, err := app.GetSample(context.Background(), "absent-sample")
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("GetSample error = %v, want store.ErrNotFound", err)
	}
}

func TestBug03_ReadingCacheDoesNotExposeBackingSlice(t *testing.T) {
	db, app, sample := fixture(t)
	addReading(t, app, sample.ID, "reading-cache", "temperature", 4, fixtureTime.Add(time.Minute))
	values, err := db.ListReadings(context.Background(), sample.ID)
	if err != nil {
		t.Fatal(err)
	}
	values[0].Value = 99
	again, err := db.ListReadings(context.Background(), sample.ID)
	if err != nil {
		t.Fatal(err)
	}
	if again[0].Value != 4 {
		t.Fatalf("cached reading changed through caller slice: %.2f", again[0].Value)
	}
}

func TestBug04_TransectBufferIsSafeForConcurrentSnapshots(t *testing.T) {
	buffer := service.NewTransectBuffer()
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			buffer.Put(service.TransectPoint{Code: fmt.Sprintf("P-%02d", i), DistanceM: float64(i), SampleIDs: []string{"sample"}})
		}(i)
		go func() {
			defer wg.Done()
			_ = buffer.Snapshot()
		}()
	}
	wg.Wait()
}

func TestBug05_ReportHonorsCanceledContext(t *testing.T) {
	_, app, _ := fixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := app.BuildSiteReport(ctx, "site-001"); !errors.Is(err, context.Canceled) {
		t.Fatalf("BuildSiteReport error = %v, want context.Canceled", err)
	}
}

func TestBug06_TransactionRollsBackCallbackError(t *testing.T) {
	db, _, sample := fixture(t)
	wantErr := errors.New("stop archive")
	err := db.WithTx(context.Background(), func(tx *sql.Tx) error {
		_, err := tx.Exec(`INSERT INTO events(id,sample_id,event_type,payload,created_at) VALUES(?,?,?,?,?)`, "rollback-event", sample.ID, "field.note", "temporary", fixtureTime.Format(time.RFC3339Nano))
		if err != nil {
			return err
		}
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("WithTx error = %v, want callback error", err)
	}
	count, err := db.CountEvents(context.Background(), sample.ID)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("rolled-back event count = %d, want the original sample.created event only", count)
	}
}

func TestBug07_StateUpdateRejectsStaleFromState(t *testing.T) {
	db, _, sample := fixture(t)
	if err := db.UpdateSampleState(context.Background(), sample.ID, model.SampleDraft, model.SampleMeasured, fixtureTime.Format(time.RFC3339Nano)); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateSampleState(context.Background(), sample.ID, model.SampleDraft, model.SampleIdentified, fixtureTime.Format(time.RFC3339Nano)); !errors.Is(err, store.ErrConflict) {
		t.Fatalf("stale state update error = %v, want conflict", err)
	}
}

func TestBug08_DraftSampleCannotBeIdentified(t *testing.T) {
	_, app, sample := fixture(t)
	if err := app.CreateTaxon(context.Background(), model.Taxon{ID: "taxon-001", Scientific: "Usnea longissima", Rank: "species"}); err != nil {
		t.Fatal(err)
	}
	err := app.SubmitIdentification(context.Background(), model.Identification{ID: "ident-draft", SampleID: sample.ID, TaxonID: "taxon-001", Reviewer: "reviewer", Confidence: 0.7})
	if err == nil {
		t.Fatal("draft sample accepted an identification")
	}
}

func TestBug09_ReviewCannotBeSubmittedTwice(t *testing.T) {
	_, app, sample := fixture(t)
	addAcceptedIdentification(t, app, sample)
	err := app.ReviewIdentification(context.Background(), model.Review{ID: "review-002", SampleID: sample.ID, Reviewer: "reviewer-2", Decision: "approve"})
	if err == nil {
		t.Fatal("second review unexpectedly succeeded")
	}
}

func TestBug10_ArchiveOperatorMustBeIndependent(t *testing.T) {
	_, app, sample := fixture(t)
	addAcceptedIdentification(t, app, sample)
	err := app.ArchiveSample(context.Background(), model.ArchiveRecord{ID: "archive-001", SampleID: sample.ID, BoxCode: "BOX-1", SealedBy: "reviewer"})
	if err == nil {
		t.Fatal("reviewer was allowed to seal the same sample")
	}
}

func TestBug11_TaskCanOnlyBeClaimedOnce(t *testing.T) {
	db, _, sample := fixture(t)
	task := model.Task{ID: "task-001", SampleID: sample.ID, Kind: "identify", State: "queued", AvailableAt: fixtureTime}
	if err := db.EnqueueTask(context.Background(), task); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ClaimNextTask(context.Background(), fixtureTime); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ClaimNextTask(context.Background(), fixtureTime); err == nil {
		t.Fatal("running task was claimed a second time")
	}
}

func TestBug12_RunTaskReturnsProcessingError(t *testing.T) {
	db, app, sample := fixture(t)
	if err := app.QueueIdentification(context.Background(), model.Task{ID: "task-missing", SampleID: sample.ID, Kind: "identify"}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB().ExecContext(context.Background(), "PRAGMA foreign_keys = OFF"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB().ExecContext(context.Background(), "DELETE FROM samples WHERE id=?", sample.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB().ExecContext(context.Background(), "PRAGMA foreign_keys = ON"); err != nil {
		t.Fatal(err)
	}
	if err := app.RunNextTask(context.Background()); err == nil {
		t.Fatal("RunNextTask swallowed missing sample error")
	}
}

func TestBug13_FailedTaskPersistsFailedState(t *testing.T) {
	db, _, sample := fixture(t)
	if err := db.EnqueueTask(context.Background(), model.Task{ID: "task-failed", SampleID: sample.ID, Kind: "identify", State: "queued", AvailableAt: fixtureTime}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ClaimNextTask(context.Background(), fixtureTime); err != nil {
		t.Fatal(err)
	}
	if err := db.CompleteTask(context.Background(), "task-failed", fixtureTime, errors.New("instrument offline")); err != nil {
		t.Fatal(err)
	}
	tasks, err := db.ListTasks(context.Background(), sample.ID)
	if err != nil {
		t.Fatal(err)
	}
	if tasks[0].State != "failed" {
		t.Fatalf("task state = %s, want failed", tasks[0].State)
	}
}

func TestBug14_SampleCreationKeepsEventPayload(t *testing.T) {
	db, _, sample := fixture(t)
	events, err := db.ListEvents(context.Background(), sample.ID, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Payload == "" {
		t.Fatalf("sample creation event payload = %#v, want non-empty payload", events)
	}
}

func TestBug15_ReportAllowsSampleWithoutArchive(t *testing.T) {
	_, app, sample := fixture(t)
	report, err := app.BuildSiteReport(context.Background(), sample.SiteID)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Samples) != 1 || report.Samples[0].Archive != nil {
		t.Fatalf("unexpected report archive state: %#v", report.Samples)
	}
}

func TestBug16_ListReadingsReportsMissingSample(t *testing.T) {
	_, app, _ := fixture(t)
	if _, err := app.ListReadings(context.Background(), "absent-sample"); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("ListReadings error = %v, want not found", err)
	}
}

func TestBug17_EnvironmentRangeIsCheckedAtServiceBoundary(t *testing.T) {
	_, app, sample := fixture(t)
	err := app.AddReading(context.Background(), model.Reading{ID: "bad-humidity", SampleID: sample.ID, Kind: "humidity", Value: 150, Unit: "%", RecordedAt: fixtureTime})
	if err == nil {
		t.Fatal("out-of-range humidity was accepted")
	}
}

func TestBug18_DuplicateReadingTimestampIsRejected(t *testing.T) {
	_, app, sample := fixture(t)
	at := fixtureTime.Add(time.Minute)
	addReading(t, app, sample.ID, "duplicate-1", "temperature", 4, at)
	if err := app.AddReading(context.Background(), model.Reading{ID: "duplicate-2", SampleID: sample.ID, Kind: "temperature", Value: 5, Unit: "C", RecordedAt: at}); err == nil {
		t.Fatal("duplicate reading timestamp was accepted")
	}
}

func TestBug19_CollectorCannotReviewOwnIdentification(t *testing.T) {
	_, app, sample := fixture(t)
	addThreeReadings(t, app, sample.ID)
	if err := app.CreateTaxon(context.Background(), model.Taxon{ID: "taxon-001", Scientific: "Usnea longissima", Rank: "species"}); err != nil {
		t.Fatal(err)
	}
	if err := app.SubmitIdentification(context.Background(), model.Identification{ID: "ident-001", SampleID: sample.ID, TaxonID: "taxon-001", Reviewer: "field-team", Confidence: 0.7}); err == nil {
		t.Fatal("collector identity was accepted as reviewer")
	}
}

func TestBug20_ArchiveFailureIsAtomic(t *testing.T) {
	db, app, sample := fixture(t)
	addAcceptedIdentification(t, app, sample)
	if err := db.AppendEvent(context.Background(), model.Event{ID: "event-archive-" + sample.ID, SampleID: sample.ID, EventType: "field.note", Payload: "occupied", CreatedAt: fixtureTime}); err != nil {
		t.Fatal(err)
	}
	if err := app.ArchiveSample(context.Background(), model.ArchiveRecord{ID: "archive-001", SampleID: sample.ID, BoxCode: "BOX-1", SealedBy: "archiver"}); err == nil {
		t.Fatal("archive unexpectedly succeeded with occupied event id")
	}
	current, err := app.GetSample(context.Background(), sample.ID)
	if err != nil {
		t.Fatal(err)
	}
	if current.Status != model.SampleIdentified {
		t.Fatalf("sample status after failed archive = %s, want identified", current.Status)
	}
}

func TestBug21_ApplicationCloseIsIdempotent(t *testing.T) {
	_, app, _ := fixture(t)
	panicked := false
	func() {
		defer func() {
			if recover() != nil {
				panicked = true
			}
		}()
		app.Close()
		app.Close()
	}()
	if panicked {
		t.Fatal("closing application twice panicked")
	}
}

func TestBug22_TransectSnapshotCopiesSampleIDs(t *testing.T) {
	buffer := service.NewTransectBuffer()
	buffer.Put(service.TransectPoint{Code: "P-1", DistanceM: 1, SampleIDs: []string{"sample-1"}})
	snapshot := buffer.Snapshot()
	snapshot[0].SampleIDs[0] = "changed"
	again := buffer.Snapshot()
	if again[0].SampleIDs[0] != "sample-1" {
		t.Fatal("transect snapshot exposed nested slice")
	}
}

func TestBug23_TransectCollectionHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := service.CollectTransect(ctx, service.Transect{ID: "T-1", SiteID: "site-001", Name: "ridge", Points: []service.TransectPoint{{Code: "P-1", DistanceM: 1}}}, func(context.Context, service.TransectPoint) ([]model.Reading, error) { return nil, nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("CollectTransect error = %v, want context.Canceled", err)
	}
}

func TestBug24_SurveyDataSurvivesDatabaseRestart(t *testing.T) {
	path := t.TempDir() + "/restart.db"
	db, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	site := model.Site{ID: "restart-site", Name: "Restart Ridge", Region: "north", ElevationM: 1000, Status: "open", CreatedAt: fixtureTime, UpdatedAt: fixtureTime}
	if err := db.CreateSite(context.Background(), site); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	got, err := db.GetSite(context.Background(), site.ID)
	if err != nil || got.Name != site.Name {
		t.Fatalf("reopened site = %#v, err=%v", got, err)
	}
}

func TestBug25_RetiringSiteTwiceFails(t *testing.T) {
	_, app, _ := fixture(t)
	if err := app.RetireSite(context.Background(), "site-001"); err != nil {
		t.Fatal(err)
	}
	if err := app.RetireSite(context.Background(), "site-001"); err == nil {
		t.Fatal("retired site was retired a second time")
	}
}

func TestBug26_SampleCreationRollsBackWhenEventFails(t *testing.T) {
	db, app, _ := fixture(t)
	if err := db.AppendEvent(context.Background(), model.Event{ID: "event-atomic-sample", SampleID: "sample-001", EventType: "field.note", Payload: "occupied", CreatedAt: fixtureTime}); err != nil {
		t.Fatal(err)
	}
	err := app.CreateSample(context.Background(), model.Sample{ID: "atomic-sample", SiteID: "site-001", Collector: "field-team", Condition: "fresh", CollectedAt: fixtureTime})
	if err == nil {
		t.Fatal("sample creation unexpectedly succeeded")
	}
	if _, err := app.GetSample(context.Background(), "atomic-sample"); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("sample remained after failed event: %v", err)
	}
}

func TestBug27_ExportTimestampUsesRequestedTimezone(t *testing.T) {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	value := service.ExportTimestamp(fixtureTime, service.ExportPolicy{Timezone: location})
	if !strings.Contains(value, "+08:00") {
		t.Fatalf("export timestamp = %s, want +08:00", value)
	}
}

func TestBug28_ReportEnvelopeHonorsLimit(t *testing.T) {
	sites := []model.Site{{ID: "site-1", Name: "A"}, {ID: "site-2", Name: "B"}}
	envelope, err := service.BuildReportEnvelope(context.Background(), sites, service.ReportFilter{Limit: 1}, func(context.Context, string) (model.SiteReport, error) { return model.SiteReport{}, nil })
	if err != nil {
		t.Fatal(err)
	}
	if len(envelope.Reports) != 1 {
		t.Fatalf("report count = %d, want 1", len(envelope.Reports))
	}
}

func TestBug29_ReviewDecisionUsesLatestTimestamp(t *testing.T) {
	reviews := []model.Review{{ID: "r2", Decision: "reject", CreatedAt: fixtureTime.Add(2 * time.Hour)}, {ID: "r1", Decision: "approve", CreatedAt: fixtureTime}}
	if got := service.ReviewDecision(reviews); got != "reject" {
		t.Fatalf("ReviewDecision = %s, want reject", got)
	}
}

func TestBug30_ArchivedSampleCannotReceiveReading(t *testing.T) {
	db, app, sample := fixture(t)
	if err := db.UpdateSampleState(context.Background(), sample.ID, model.SampleDraft, model.SampleArchived, fixtureTime.Format(time.RFC3339Nano)); err != nil {
		t.Fatal(err)
	}
	if err := app.AddReading(context.Background(), model.Reading{ID: "late-reading", SampleID: sample.ID, Kind: "temperature", Value: 4, Unit: "C", RecordedAt: fixtureTime}); err == nil {
		t.Fatal("archived sample accepted a new field reading")
	}
}
