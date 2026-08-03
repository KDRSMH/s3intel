package reporter

import (
	"encoding/csv"
	"io"
	"strconv"

	"s3intel/internal/riskengine"
)

// WriteCSV, risk sonuçlarını CSV formatında yazar.
func WriteCSV(w io.Writer, results []riskengine.Result) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	header := []string{"score", "category", "source", "secret_count", "key"}
	if err := cw.Write(header); err != nil {
		return err
	}

	for _, r := range results {
		row := []string{
			strconv.Itoa(r.Score),
			string(r.Category),
			r.Source,
			strconv.Itoa(len(r.Secrets)),
			r.Key,
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}

	return cw.Error()
}
