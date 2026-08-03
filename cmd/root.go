// Package cmd, s3intel'in cobra tabanlı CLI komut ağacını tanımlar.
// Bu paket kendisi ne AWS SDK ne de grayhatwarfare HTTP çağrısı yapar —
// sadece internal/activeprobe (active.go) ve internal/passiveintel
// (passive.go) paketlerini çağırıp sonuçları internal/reporter ile yazdırır.
package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"s3intel/internal/reporter"
	"s3intel/internal/riskengine"
)

var (
	outputFormat string
	outputFile   string
	verbose      bool
)

var rootCmd = &cobra.Command{
	Use:   "s3intel",
	Short: "S3 keşif ve istihbarat aracı (aktif + pasif mod)",
	Long: `s3intel, iki kesin ayrılmış modda çalışan bir AWS S3 keşif ve istihbarat aracıdır:

  active  — SADECE config/whitelist.yaml içindeki kendi test bucket'larına karşı
            gerçek AWS S3 API çağrıları yapar (internal/activeprobe paketi).
  passive — SADECE grayhatwarfare API'sine (veya API key yoksa mock veriye) sorgu
            atar, hiçbir zaman gerçek bir S3 bucket'ına bağlanmaz
            (internal/passiveintel paketi).

Bu iki mod paket seviyesinde fiziksel olarak ayrılmıştır: internal/passiveintel
hiçbir AWS SDK import etmez, internal/activeprobe hiçbir HTTP/grayhatwarfare
import etmez. Ayrıntı için docs/architecture.md dosyasına bakın.`,
}

// Execute, cobra kök komutunu çalıştırır.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&outputFormat, "output", "terminal", "çıktı formatı: terminal, json, csv")
	rootCmd.PersistentFlags().StringVar(&outputFile, "output-file", "", "çıktının yazılacağı dosya (boşsa stdout)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "ayrıntılı günlük çıktısı")
}

// writeResults, --output flag'ine göre uygun reporter fonksiyonunu seçip
// sonuçları stdout'a ya da --output-file ile verilen dosyaya yazar.
func writeResults(results []riskengine.Result) error {
	var w io.Writer = os.Stdout
	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("çıktı dosyası oluşturulamadı: %w", err)
		}
		defer f.Close()
		w = f
	}

	switch outputFormat {
	case "terminal":
		return reporter.WriteTerminal(w, results)
	case "json":
		return reporter.WriteJSON(w, results)
	case "csv":
		return reporter.WriteCSV(w, results)
	default:
		return fmt.Errorf("bilinmeyen output formatı: %s (terminal, json, csv olabilir)", outputFormat)
	}
}
