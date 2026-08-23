package service

import (
	"sort"
	"strings"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

type CatalogTaxon struct {
	Scientific string
	Common     string
	Rank       string
}

var catalogTaxa = []CatalogTaxon{
	{Scientific: "Usnea longissima", Common: "beard lichen", Rank: "species"},
	{Scientific: "Lobaria pulmonaria", Common: "lung lichen", Rank: "species"},
	{Scientific: "Cladonia rangiferina", Common: "reindeer lichen", Rank: "species"},
	{Scientific: "Xanthoria parietina", Common: "sunburst lichen", Rank: "species"},
	{Scientific: "Ramalina menziesii", Common: "lace lichen", Rank: "species"},
	{Scientific: "Evernia prunastri", Common: "oakmoss", Rank: "species"},
	{Scientific: "Parmelia sulcata", Common: "hammered shield lichen", Rank: "species"},
	{Scientific: "Peltigera canina", Common: "dog lichen", Rank: "species"},
	{Scientific: "Stereocaulon paschale", Common: "snow lichen", Rank: "species"},
	{Scientific: "Cetraria islandica", Common: "Iceland moss", Rank: "species"},
	{Scientific: "Bryoria fuscescens", Common: "horsehair lichen", Rank: "species"},
	{Scientific: "Alectoria sarmentosa", Common: "witches hair", Rank: "species"},
	{Scientific: "Hypogymnia physodes", Common: "powdered ruffle lichen", Rank: "species"},
	{Scientific: "Lecanora conizaeoides", Common: "conifer rim lichen", Rank: "species"},
	{Scientific: "Rinodina sophodes", Common: "rim lichen", Rank: "species"},
	{Scientific: "Nephroma parile", Common: "heath shield lichen", Rank: "species"},
	{Scientific: "Peltula euploca", Common: "button lichen", Rank: "species"},
	{Scientific: "Lecidea lapicida", Common: "stone disk lichen", Rank: "species"},
	{Scientific: "Rhizocarpon geographicum", Common: "map lichen", Rank: "species"},
	{Scientific: "Acarospora fuscata", Common: "rock cobblestone lichen", Rank: "species"},
	{Scientific: "Caloplaca saxicola", Common: "orange rock lichen", Rank: "species"},
	{Scientific: "Pertusaria amara", Common: "bitter wart lichen", Rank: "species"},
	{Scientific: "Graphis scripta", Common: "script lichen", Rank: "species"},
	{Scientific: "Arthonia radiata", Common: "starburst lichen", Rank: "species"},
	{Scientific: "Lichenomphalia umbellifera", Common: "lichen agaric", Rank: "species"},
	{Scientific: "Dictyonema glabratum", Common: "web lichen", Rank: "species"},
	{Scientific: "Lichen", Common: "lichen", Rank: "genus"},
	{Scientific: "Lobariaceae", Common: "lung lichen family", Rank: "family"},
	{Scientific: "Parmeliaceae", Common: "shield lichen family", Rank: "family"},
	{Scientific: "Cladoniaceae", Common: "cup lichen family", Rank: "family"},
}

func CatalogSearch(query string) []CatalogTaxon {
	query = strings.ToLower(strings.TrimSpace(query))
	result := make([]CatalogTaxon, 0)
	for _, taxon := range catalogTaxa {
		if query == "" || strings.Contains(strings.ToLower(taxon.Scientific), query) || strings.Contains(strings.ToLower(taxon.Common), query) {
			result = append(result, taxon)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Scientific < result[j].Scientific })
	return result
}

func CatalogRanks() []string {
	seen := make(map[string]struct{})
	for _, taxon := range catalogTaxa {
		seen[taxon.Rank] = struct{}{}
	}
	result := make([]string, 0, len(seen))
	for rank := range seen {
		result = append(result, rank)
	}
	sort.Strings(result)
	return result
}

func CatalogByRank(rank string) []CatalogTaxon {
	result := make([]CatalogTaxon, 0)
	for _, taxon := range catalogTaxa {
		if rank == "" || strings.EqualFold(taxon.Rank, rank) {
			result = append(result, taxon)
		}
	}
	return result
}

func CompatibleConfidence(rank string, confidence float64) bool {
	if confidence < 0 || confidence > 1 {
		return false
	}
	if strings.EqualFold(rank, "species") {
		return confidence >= 0.55
	}
	return confidence <= 0.90
}

func CatalogContains(scientific string) bool {
	for _, taxon := range catalogTaxa {
		if strings.EqualFold(taxon.Scientific, scientific) {
			return true
		}
	}
	return false
}

func CandidateScores(readings []model.Reading, candidates []CatalogTaxon) map[string]float64 {
	result := make(map[string]float64, len(candidates))
	coverage := EnvironmentalCoverage(readings)
	for _, candidate := range candidates {
		score := 0.20
		if candidate.Rank == "species" {
			score += 0.20
		}
		if coverage["moisture"] > 0 {
			score += 0.10
		}
		if coverage["thermal"] > 0 {
			score += 0.10
		}
		if coverage["exposure"] > 0 {
			score += 0.10
		}
		result[candidate.Scientific] = score
	}
	return result
}

func CandidateTaxon(value model.Taxon) CatalogTaxon {
	return CatalogTaxon{Scientific: value.Scientific, Common: value.CommonName, Rank: value.Rank}
}
