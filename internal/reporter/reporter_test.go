package reporter

import (
	"bytes"
	"strings"
	"testing"

	"s3intel/internal/classifier"
	"s3intel/internal/riskengine"
)

func sampleResults() []riskengine.Result {
	return []riskengine.Result{
		{
			Item: riskengine.Item{
				Key:      "config/.env",
				Source:   "active",
				Category: classifier.CategoryEnvConfig,
			},
			Score: 90,
		},
		{
			Item: riskengine.Item{
				Key:      "readme.txt",
				Source:   "passive",
				Category: classifier.CategoryOther,
			},
			Score: 10,
		},
	}
}

func TestWriteTerminal(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteTerminal(&buf, sampleResults()); err != nil {
		t.Fatalf("beklenmeyen hata: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "config/.env") {
		t.Errorf("terminal çıktısı dosya adını içermeli: %s", out)
	}
	if !strings.Contains(out, "90") {
		t.Errorf("terminal çıktısı skoru içermeli: %s", out)
	}
}

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteJSON(&buf, sampleResults()); err != nil {
		t.Fatalf("beklenmeyen hata: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"Key": "config/.env"`) {
		t.Errorf("JSON çıktısı beklenen alanı içermeli: %s", out)
	}
}

func TestWriteCSV(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteCSV(&buf, sampleResults()); err != nil {
		t.Fatalf("beklenmeyen hata: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "score,category,source,secret_count,key") {
		t.Errorf("CSV başlık satırı beklenmedik: %s", out)
	}
	if !strings.Contains(out, "config/.env") {
		t.Errorf("CSV çıktısı dosya adını içermeli: %s", out)
	}
}
