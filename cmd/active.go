package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"s3intel/internal/activeprobe"
	"s3intel/internal/riskengine"
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

		findings, err := activeprobe.EnumerateBucket(context.Background(), activeBucket)
		if err != nil {
			return err
		}

		items := make([]riskengine.Item, 0, len(findings))
		for _, f := range findings {
			items = append(items, riskengine.Item{
				Key:      f.Key,
				Source:   "active",
				Category: f.Category,
				BaseRisk: f.BaseRisk,
				Secrets:  f.SecretFindings,
			})
		}
		results := riskengine.ScoreAll(items)
		return writeResults(results)
	},
}

func init() {
	activeCmd.Flags().StringVar(&activeBucket, "bucket", "", "whitelist'te olması gereken hedef bucket adı (zorunlu)")
	rootCmd.AddCommand(activeCmd)
}
