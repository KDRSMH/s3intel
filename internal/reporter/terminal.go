// Package reporter, riskengine tarafından üretilen skorlanmış sonuçları
// terminal (renkli tablo), JSON ve CSV formatlarında yazar.
package reporter

import (
	"fmt"
	"io"
	"text/tabwriter"

	"s3intel/internal/riskengine"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31;1m"
	colorYellow = "\033[33;1m"
	colorGreen  = "\033[32;1m"
)

func riskColor(score int) string {
	switch {
	case score >= 70:
		return colorRed
	case score >= 40:
		return colorYellow
	default:
		return colorGreen
	}
}

// WriteTerminal, risk sonuçlarını (zaten azalan sırada) renkli, hizalı bir
// tablo olarak yazar.
func WriteTerminal(w io.Writer, results []riskengine.Result) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)

	fmt.Fprintln(tw, "SKOR\tKATEGORİ\tKAYNAK\tSECRET\tANAHTAR/DOSYA")
	for _, r := range results {
		color := riskColor(r.Score)
		fmt.Fprintf(tw, "%s%3d%s\t%s\t%s\t%d\t%s\n",
			color, r.Score, colorReset, r.Category, r.Source, len(r.Secrets), r.Key)
	}

	return tw.Flush()
}
