package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"s3intel/internal/scanjobs"
)

var activeBucket string

var activeCmd = &cobra.Command{
	Use:   "active",
	Short: "Whitelist'teki kendi test bucket'ına karşı AKTIF keşif yapar",
	Long: `active komutu, SADECE config/whitelist.yaml içinde listelenen bucket'lara
karşı gerçek AWS S3 API çağrıları yapar (internal/activeprobe paketi).
Whitelist'te olmayan bir bucket adı verilirse hiçbir AWS çağrısı yapılmadan
hata ile durur.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if activeBucket == "" {
			return fmt.Errorf("--bucket zorunludur")
		}
		if verbose {
			fmt.Printf("[active] %s bucket'ı için whitelist kontrol ediliyor...\n", activeBucket)
		}

		results, err := scanjobs.RunActive(context.Background(), activeBucket)
		if err != nil {
			return err
		}

		return writeResults(results)
	},
}

func init() {
	activeCmd.Flags().StringVar(&activeBucket, "bucket", "", "whitelist'te olması gereken hedef bucket adı (zorunlu)")
	rootCmd.AddCommand(activeCmd)
}
