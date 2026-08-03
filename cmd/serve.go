package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"s3intel/internal/webui"
)

var servePort int

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Tarayıcıda çalışan basit s3intel arayüzünü başlatır",
	Long: `serve komutu, aktif/pasif taramayı tarayıcıdan çalıştırabileceğin basit bir
yerel web arayüzü açar. SADECE 127.0.0.1 üzerinden erişilebilir; dışarıya
açık bir servis değildir. Arka planda aynı internal/scanjobs katmanı
üzerinden, aynı whitelist kuralıyla çalışır — active/passive CLI komutlarıyla
tamamen aynı mantığı kullanır, sadece görselleştirmesi farklıdır.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		addr := fmt.Sprintf("127.0.0.1:%d", servePort)
		return webui.Serve(addr)
	},
}

func init() {
	serveCmd.Flags().IntVar(&servePort, "port", 8080, "web arayüzünün dinleyeceği port (sadece 127.0.0.1'e bağlanır)")
	rootCmd.AddCommand(serveCmd)
}
