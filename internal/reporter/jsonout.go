package reporter

import (
	"encoding/json"
	"io"

	"s3intel/internal/riskengine"
)

// WriteJSON, risk sonuçlarını girintili (indented) JSON olarak yazar.
func WriteJSON(w io.Writer, results []riskengine.Result) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}
