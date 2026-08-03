package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"s3intel/internal/classifier"
	"s3intel/internal/passiveintel"
	"s3intel/internal/riskengine"
)

var passiveKeyword string

var passiveCmd = &cobra.Command{
	Use:   "passive",
	Short: "grayhatwarfare üzerinden PASIF S3 istihbaratı toplar",
	Long: `passive komutu, SADECE grayhatwarfare API'sine (ya da GHW_API_KEY
ayarlı değilse mock/örnek veriye) HTTP sorgusu atar. internal/passiveintel
paketi hiçbir zaman gerçek bir S3 bucket'ına bağlanmaz ve hiçbir AWS SDK
import etmez.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if passiveKeyword == "" {
			return fmt.Errorf("--keyword zorunludur")
		}
		if verbose {
			fmt.Printf("[passive] grayhatwarfare '%s' anahtar kelimesiyle sorgulanıyor...\n", passiveKeyword)
		}

		searchResults, err := passiveintel.Search(context.Background(), passiveKeyword)
		if err != nil {
			return err
		}

		items := make([]riskengine.Item, 0, len(searchResults))
		for _, r := range searchResults {
			cat, baseRisk := classifier.Classify(r.Filename)
			items = append(items, riskengine.Item{
				Key:      r.FullPath,
				Source:   "passive",
				Category: cat,
				BaseRisk: baseRisk,
				Secrets:  nil, // pasif modda dosya içeriği hiç indirilmez, secretscan çalışmaz
			})
		}

		return writeResults(riskengine.ScoreAll(items))
	},
}

func init() {
	passiveCmd.Flags().StringVar(&passiveKeyword, "keyword", "", "grayhatwarfare arama anahtar kelimesi (zorunlu)")
	rootCmd.AddCommand(passiveCmd)
}
