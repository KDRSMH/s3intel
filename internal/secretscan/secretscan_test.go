package secretscan

import (
	"os"
	"testing"
)

func TestShannonEntropy(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantLow bool
	}{
		{"tekrarlayan tek karakter", "aaaaaaaaaaaaaaaaaaaa", true},
		{"düz kelime", "passwordpasswordpassword", true},
		{"rastgele görünen karışık string", "aZ9$kQ2!mN7&pL4#xR8@zT1^", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := ShannonEntropy(tc.input)
			if tc.wantLow && e >= entropyThreshold {
				t.Errorf("düşük entropi bekleniyordu (< %.1f), aldık %.2f", entropyThreshold, e)
			}
			if !tc.wantLow && e < entropyThreshold {
				t.Errorf("yüksek entropi bekleniyordu (>= %.1f), aldık %.2f", entropyThreshold, e)
			}
		})
	}
}

func TestScan_RegexPatterns(t *testing.T) {
	cases := []struct {
		name       string
		content    string
		wantType   string
		wantConfid Confidence
	}{
		{"AWS access key", "aws_access_key_id = AKIAABCDEFGHIJKLMNOP", "aws_access_key", ConfidenceHigh},
		{"AWS secret key", "aws_secret_access_key = wJalrXUtnFEMIfakeK7MDENGbPxRfiCYFAKEKEYXYZ1", "aws_secret_key", ConfidenceHigh},
		{"private key header", "-----BEGIN RSA PRIVATE KEY-----\nMIIFAKE\n-----END RSA PRIVATE KEY-----", "private_key", ConfidenceHigh},
		{"password satırı", "password=SuperSecretPassword123!", "password", ConfidenceMedium},
		{"generic api key", "api_key=sk_test_FAKE1234567890abcdefFAKE", "generic_api_key", ConfidenceMedium},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			findings := Scan(tc.content)
			var match *Finding
			for i := range findings {
				if findings[i].Type == tc.wantType {
					match = &findings[i]
					break
				}
			}
			if match == nil {
				t.Fatalf("%q tipinde bulgu bekleniyordu, bulunamadı: %+v", tc.wantType, findings)
			}
			if match.Confidence != tc.wantConfid {
				t.Errorf("beklenen confidence %v, alınan %v", tc.wantConfid, match.Confidence)
			}
		})
	}
}

func TestScan_EntropyCatchesUnmatchedRandomToken(t *testing.T) {
	content := "random_token: 9fQ8sD6fG4hJ2kL0qWeRtYuIoPaSdFgHjKlZxCvBnM3"
	findings := Scan(content)

	found := false
	for _, f := range findings {
		if f.Type == "high_entropy_string" && f.Confidence == ConfidenceLow {
			found = true
		}
	}
	if !found {
		t.Errorf("regex'e yakalanmayan rastgele token entropy ile işaretlenmeliydi: %+v", findings)
	}
}

func TestScan_DoesNotDoubleCountRegexMatchesAsEntropy(t *testing.T) {
	content := "aws_access_key_id = AKIAABCDEFGHIJKLMNOP"
	findings := Scan(content)

	entropyCount := 0
	for _, f := range findings {
		if f.Type == "high_entropy_string" {
			entropyCount++
		}
	}
	if entropyCount != 0 {
		t.Errorf("regex tarafından zaten yakalanan aralık entropy tarafından tekrar işaretlenmemeliydi, %d kez işaretlendi", entropyCount)
	}
}

// TestScan_GroundTruthSampleFile, testdata/sample_secrets.txt (ground-truth
// örnek veri) içinde beklenen TÜM secret türlerinin gerçekten bulunduğunu
// doğrular. Bu, aracın bilinen bir ortamdaki doğruluğunun kanıtıdır.
func TestScan_GroundTruthSampleFile(t *testing.T) {
	data, err := os.ReadFile("../../testdata/sample_secrets.txt")
	if err != nil {
		t.Fatalf("testdata/sample_secrets.txt okunamadı: %v", err)
	}

	findings := Scan(string(data))

	wantTypes := map[string]bool{
		"aws_access_key":  false,
		"aws_secret_key":  false,
		"password":        false,
		"generic_api_key": false,
		"private_key":     false,
	}
	for _, f := range findings {
		if _, ok := wantTypes[f.Type]; ok {
			wantTypes[f.Type] = true
		}
	}
	for typ, seen := range wantTypes {
		if !seen {
			t.Errorf("ground-truth dosyasında %q türünde bir bulgu bekleniyordu ama bulunamadı", typ)
		}
	}

	hasHighEntropy := false
	for _, f := range findings {
		if f.Type == "high_entropy_string" {
			hasHighEntropy = true
		}
	}
	if !hasHighEntropy {
		t.Error("ground-truth dosyasında en az bir yüksek entropili string bekleniyordu")
	}

	if len(findings) < 5 {
		t.Errorf("en az 5 bulgu bekleniyordu, %d geldi: %+v", len(findings), findings)
	}
}
