package riskengine

import (
	"testing"

	"s3intel/internal/classifier"
	"s3intel/internal/secretscan"
)

func TestScoreAll_SortsDescendingAndCapsAt100(t *testing.T) {
	items := []Item{
		{Key: "low.txt", Source: "passive", Category: classifier.CategoryOther, BaseRisk: 10},
		{
			Key:      "high.env",
			Source:   "active",
			Category: classifier.CategoryEnvConfig,
			BaseRisk: 70,
			Secrets: []secretscan.Finding{
				{Type: "aws_access_key", Confidence: secretscan.ConfidenceHigh},
				{Type: "generic_api_key", Confidence: secretscan.ConfidenceMedium},
			},
		},
		{
			Key:      "medium.bak",
			Source:   "passive",
			Category: classifier.CategoryBackup,
			BaseRisk: 55,
		},
	}

	results := ScoreAll(items)

	if len(results) != 3 {
		t.Fatalf("3 sonuç bekleniyordu, %d geldi", len(results))
	}
	if results[0].Key != "high.env" {
		t.Errorf("en yüksek riskli sonuç ilk sırada olmalı, ilk sırada %q var", results[0].Key)
	}
	if results[0].Score != maxScore {
		t.Errorf("70 + 40 + 20 = 130, %d'e sıkıştırılmalıydı, aldık %d", maxScore, results[0].Score)
	}
	for i := 1; i < len(results); i++ {
		if results[i-1].Score < results[i].Score {
			t.Fatalf("sonuçlar azalan sırada değil: %+v", results)
		}
	}
}

func TestScoreAll_NoSecretsUsesOnlyBaseRisk(t *testing.T) {
	items := []Item{
		{Key: "readme.txt", Source: "passive", Category: classifier.CategoryOther, BaseRisk: 10},
	}
	results := ScoreAll(items)
	if results[0].Score != 10 {
		t.Errorf("secret bulunmayan öğede skor sadece temel risk olmalı, beklenen 10, aldık %d", results[0].Score)
	}
}
