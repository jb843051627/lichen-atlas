package service

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

// EnvironmentRule describes an instrument reading that can be compared across sites.
type EnvironmentRule struct {
	Kind   string
	Unit   string
	Min    float64
	Max    float64
	Family string
}

var environmentRules = []EnvironmentRule{
	{Kind: "temperature", Unit: "C", Min: -25, Max: 25, Family: "thermal"},
	{Kind: "humidity", Unit: "%", Min: 0, Max: 100, Family: "moisture"},
	{Kind: "light", Unit: "lux", Min: 0, Max: 180000, Family: "exposure"},
	{Kind: "substrate", Unit: "pH", Min: 2, Max: 10, Family: "substrate"},
	{Kind: "wind", Unit: "m/s", Min: 0, Max: 70, Family: "exposure"},
	{Kind: "surface_water", Unit: "mm", Min: 0, Max: 500, Family: "moisture"},
	{Kind: "canopy", Unit: "%", Min: 0, Max: 100, Family: "exposure"},
	{Kind: "bark_age", Unit: "year", Min: 0, Max: 400, Family: "substrate"},
	{Kind: "rock_moisture", Unit: "%", Min: 0, Max: 100, Family: "moisture"},
	{Kind: "soil_depth", Unit: "mm", Min: 0, Max: 2000, Family: "substrate"},
	{Kind: "uv_index", Unit: "index", Min: 0, Max: 20, Family: "exposure"},
	{Kind: "snow_cover", Unit: "cm", Min: 0, Max: 1200, Family: "thermal"},
	{Kind: "surface_temp", Unit: "C", Min: -40, Max: 60, Family: "thermal"},
	{Kind: "dew_point", Unit: "C", Min: -50, Max: 40, Family: "moisture"},
	{Kind: "co2", Unit: "ppm", Min: 200, Max: 2000, Family: "air"},
	{Kind: "nitrate", Unit: "mg/L", Min: 0, Max: 100, Family: "substrate"},
	{Kind: "conductivity", Unit: "uS/cm", Min: 0, Max: 5000, Family: "substrate"},
	{Kind: "wind_gust", Unit: "m/s", Min: 0, Max: 90, Family: "exposure"},
	{Kind: "rain_24h", Unit: "mm", Min: 0, Max: 600, Family: "moisture"},
	{Kind: "leaf_wetness", Unit: "%", Min: 0, Max: 100, Family: "moisture"},
	{Kind: "shade_index", Unit: "index", Min: 0, Max: 1, Family: "exposure"},
	{Kind: "roughness", Unit: "index", Min: 0, Max: 1, Family: "substrate"},
	{Kind: "altitude_pressure", Unit: "hPa", Min: 250, Max: 1100, Family: "air"},
	{Kind: "oxygen", Unit: "%", Min: 0, Max: 100, Family: "air"},
	{Kind: "dust", Unit: "ug/m3", Min: 0, Max: 1000, Family: "air"},
	{Kind: "soil_temperature", Unit: "C", Min: -30, Max: 50, Family: "thermal"},
	{Kind: "water_temperature", Unit: "C", Min: -5, Max: 40, Family: "thermal"},
	{Kind: "phosphate", Unit: "mg/L", Min: 0, Max: 50, Family: "substrate"},
	{Kind: "chloride", Unit: "mg/L", Min: 0, Max: 10000, Family: "substrate"},
	{Kind: "iron", Unit: "mg/L", Min: 0, Max: 500, Family: "substrate"},
	{Kind: "calcium", Unit: "mg/L", Min: 0, Max: 2000, Family: "substrate"},
	{Kind: "magnesium", Unit: "mg/L", Min: 0, Max: 1000, Family: "substrate"},
	{Kind: "surface_angle", Unit: "degree", Min: 0, Max: 90, Family: "exposure"},
	{Kind: "aspect", Unit: "degree", Min: 0, Max: 360, Family: "exposure"},
	{Kind: "rock_temperature", Unit: "C", Min: -40, Max: 80, Family: "thermal"},
	{Kind: "cloud_cover", Unit: "%", Min: 0, Max: 100, Family: "exposure"},
	{Kind: "visibility", Unit: "km", Min: 0, Max: 100, Family: "air"},
	{Kind: "barometric_trend", Unit: "hPa/h", Min: -30, Max: 30, Family: "air"},
}

func FindEnvironmentRule(kind string) (EnvironmentRule, bool) {
	for _, rule := range environmentRules {
		if strings.EqualFold(rule.Kind, strings.TrimSpace(kind)) {
			return rule, true
		}
	}
	return EnvironmentRule{}, false
}

func ValidateEnvironmentReading(reading model.Reading) error {
	rule, ok := FindEnvironmentRule(reading.Kind)
	if !ok {
		return fmt.Errorf("unknown environmental reading %q", reading.Kind)
	}
	if !strings.EqualFold(rule.Unit, reading.Unit) {
		return fmt.Errorf("unit %s does not match %s", reading.Unit, rule.Unit)
	}
	if math.IsNaN(reading.Value) || math.IsInf(reading.Value, 0) {
		return fmt.Errorf("reading %s is not finite", reading.Kind)
	}
	if reading.Value < rule.Min || reading.Value > rule.Max {
		return fmt.Errorf("reading %s is outside %s range", reading.Kind, rule.Family)
	}
	return nil
}

func EnvironmentFamily(kind string) string {
	rule, ok := FindEnvironmentRule(kind)
	if !ok {
		return "unknown"
	}
	return rule.Family
}

func EnvironmentKinds() []string {
	result := make([]string, 0, len(environmentRules))
	for _, rule := range environmentRules {
		result = append(result, rule.Kind)
	}
	sort.Strings(result)
	return result
}

func EnvironmentalCoverage(readings []model.Reading) map[string]int {
	result := make(map[string]int)
	for _, reading := range readings {
		if reading.IsEnvironmental() {
			result[EnvironmentFamily(reading.Kind)]++
		}
	}
	return result
}

func CompleteEnvironmentalSet(readings []model.Reading) bool {
	families := EnvironmentalCoverage(readings)
	return families["thermal"] > 0 && families["moisture"] > 0 && families["exposure"] > 0
}

func FamilyBalance(readings []model.Reading) float64 {
	coverage := EnvironmentalCoverage(readings)
	if len(coverage) == 0 {
		return 0
	}
	min, max := 0, 0
	for _, count := range coverage {
		if min == 0 || count < min {
			min = count
		}
		if count > max {
			max = count
		}
	}
	if max == 0 {
		return 0
	}
	return float64(min) / float64(max)
}
