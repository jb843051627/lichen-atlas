package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

type SitePolicy struct {
	AllowedStatus []string
	MinElevation  float64
	MaxElevation  float64
	RequiresNote  bool
}

func (p SitePolicy) Validate(site model.Site) error {
	if site.ElevationM < p.MinElevation || site.ElevationM > p.MaxElevation {
		return fmt.Errorf("site elevation is outside policy")
	}
	if len(p.AllowedStatus) > 0 {
		allowed := false
		for _, status := range p.AllowedStatus {
			if status == site.Status {
				allowed = true
			}
		}
		if !allowed {
			return fmt.Errorf("site status %s is not allowed", site.Status)
		}
	}
	return nil
}

func SiteSlug(name string) string {
	parts := strings.Fields(strings.ToLower(name))
	return strings.Join(parts, "-")
}

func SortSites(sites []model.Site) []model.Site {
	result := append([]model.Site(nil), sites...)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Region == result[j].Region {
			return result[i].Name < result[j].Name
		}
		return result[i].Region < result[j].Region
	})
	return result
}

func SiteIsAccessible(ctx context.Context, site model.Site, now time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if site.Status != "open" {
		return fmt.Errorf("site %s is not open", site.ID)
	}
	if now.IsZero() {
		return fmt.Errorf("access time is missing")
	}
	return nil
}

func SiteStatusRank(status string) int {
	switch status {
	case "open":
		return 3
	case "restricted":
		return 2
	case "retired":
		return 1
	default:
		return 0
	}
}

func GroupSitesByRegion(sites []model.Site) map[string][]model.Site {
	result := make(map[string][]model.Site)
	for _, site := range sites {
		result[site.Region] = append(result[site.Region], site)
	}
	for region := range result {
		result[region] = SortSites(result[region])
	}
	return result
}
