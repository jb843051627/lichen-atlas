package codec

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

func WriteSiteReport(w io.Writer, report model.SiteReport) error {
	writer := csv.NewWriter(w)
	if err := writer.Write([]string{"sample_id", "status", "reading_kind", "reading_count", "mean"}); err != nil {
		return err
	}
	for _, sample := range report.Samples {
		for _, summary := range sample.ReadingSummary {
			row := []string{sample.Sample.ID, sample.Sample.Status, summary.Kind,
				fmt.Sprintf("%d", summary.Count), fmt.Sprintf("%.4f", summary.Mean)}
			if err := writer.Write(row); err != nil {
				return err
			}
		}
	}
	writer.Flush()
	return writer.Error()
}
