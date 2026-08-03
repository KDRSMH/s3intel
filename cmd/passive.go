package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"s3intel/internal/scanjobs"
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

		results, err := scanjobs.RunPassive(context.Background(), passiveKeyword)
		if err != nil {
			return err
		}

		return writeResults(results)
	},
}

func init() {
	passiveCmd.Flags().StringVar(&passiveKeyword, "keyword", "", "grayhatwarfare arama anahtar kelimesi (zorunlu)")
	rootCmd.AddCommand(passiveCmd)
}
