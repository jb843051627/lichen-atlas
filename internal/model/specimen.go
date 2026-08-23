package model

type Specimen struct {
	SampleID string
	Material string
	MassG    float64
	Drying   string
}

func (s Specimen) IsReadyForBox() bool {
	return s.Material != "" && s.MassG > 0 && s.Drying == "complete"
}
