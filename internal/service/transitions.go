package service

import "github.com/jb843051627/lichen-atlas/internal/model"

func CanTransition(from, to string) bool {
	switch from {
	case model.SampleDraft:
		return to == model.SampleMeasured
	case model.SampleMeasured:
		return to == model.SampleIdentified
	case model.SampleIdentified:
		return to == model.SampleArchived
	default:
		return false
	}
}

func StateLabel(state string) string {
	switch state {
	case model.SampleDraft:
		return "field draft"
	case model.SampleMeasured:
		return "measurements complete"
	case model.SampleIdentified:
		return "identification reviewed"
	case model.SampleArchived:
		return "archived specimen"
	default:
		return "unknown state"
	}
}
