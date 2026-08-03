package scanjobs

import (
	"context"
	"strings"
	"testing"
)

func TestRunActive_RejectsNonWhitelistedBucket(t *testing.T) {
	_, err := RunActive(context.Background(), "definitely-not-a-real-whitelisted-bucket-name")
	if err == nil {
		t.Fatal("whitelist dışı bucket için hata bekleniyordu")
	}
	if !strings.Contains(err.Error(), "whitelist") {
		t.Errorf("hata whitelist reddini belirtmeli, aldık: %v", err)
	}
}

func TestRunPassive_UsesMockAndReturnsScoredResults(t *testing.T) {
	t.Setenv("GHW_API_KEY", "")

	results, err := RunPassive(context.Background(), "backup")
	if err != nil {
		t.Fatalf("beklenmeyen hata: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("en az bir sonuç bekleniyordu")
	}
	for i := 1; i < len(results); i++ {
		if results[i-1].Score < results[i].Score {
			t.Fatalf("sonuçlar azalan sırada olmalı: %+v", results)
		}
	}
}
