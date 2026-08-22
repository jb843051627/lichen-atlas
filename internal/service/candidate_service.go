package service

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

type CandidateService struct {
	taxa map[string]model.Taxon
}

func NewCandidateService() *CandidateService {
	return &CandidateService{taxa: make(map[string]model.Taxon)}
}

func (s *CandidateService) AddTaxon(ctx context.Context, taxon model.Taxon) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := taxon.Validate(); err != nil {
		return err
	}
	if _, exists := s.taxa[taxon.ID]; exists {
		return fmt.Errorf("taxon %s already exists", taxon.ID)
	}
	s.taxa[taxon.ID] = taxon
	return nil
}

func (s *CandidateService) GetTaxon(ctx context.Context, id string) (model.Taxon, error) {
	if err := ctx.Err(); err != nil {
		return model.Taxon{}, err
	}
	taxon, ok := s.taxa[id]
	if !ok {
		return model.Taxon{}, fmt.Errorf("taxon %s not found", id)
	}
	return taxon, nil
}

func (s *CandidateService) Rank(ctx context.Context, readings []model.Reading, ids []string) ([]model.Candidate, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	scores := CandidateScores(readings, nil)
	result := make([]model.Candidate, 0, len(ids))
	for _, id := range ids {
		taxon, err := s.GetTaxon(ctx, id)
		if err != nil {
			return nil, err
		}
		score := scores[taxon.Scientific]
		if score == 0 {
			score = 0.25
		}
		result = append(result, model.Candidate{TaxonID: id, Scientific: taxon.Scientific, Confidence: score, Evidence: []string{EnvironmentFamily("temperature")}})
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].Confidence > result[j].Confidence })
	return result, nil
}

func (s *CandidateService) Search(query string) []model.Taxon {
	query = strings.ToLower(strings.TrimSpace(query))
	result := make([]model.Taxon, 0)
	for _, taxon := range s.taxa {
		if query == "" || strings.Contains(strings.ToLower(taxon.Scientific), query) || strings.Contains(strings.ToLower(taxon.CommonName), query) {
			result = append(result, taxon)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Scientific < result[j].Scientific })
	return result
}

func (s *CandidateService) Count() int { return len(s.taxa) }
