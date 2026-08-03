package secretscan

import (
	"math"
	"regexp"
)

// entropyThreshold, bir string'in "muhtemel rastgele secret/token" olarak
// işaretlenmesi için gereken minimum Shannon entropi değeridir (bit/karakter).
const entropyThreshold = 4.5

// tokenPattern, entropy analizine aday olacak, en az 20 karakterlik
// alfasayısal/özel karakterli parçaları yakalar.
var tokenPattern = regexp.MustCompile(`[A-Za-z0-9+/=_\-]{20,}`)

// ShannonEntropy, bir string'in Shannon entropisini (bit/karakter) hesaplar.
// Yüksek entropi (~>4.5 bit/karakter), karakterlerin rastgele dağıldığını —
// dolayısıyla muhtemel bir token/secret olduğunu — gösterir. Düz metin
// (İngilizce/Türkçe kelimeler, tekrarlayan karakterler) genelde çok daha
// düşük entropiye sahiptir.
func ShannonEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}

	freq := make(map[rune]int)
	for _, r := range s {
		freq[r]++
	}

	length := float64(len(s))
	var entropy float64
	for _, count := range freq {
		p := float64(count) / length
		entropy -= p * math.Log2(p)
	}
	return entropy
}

// scanEntropy, regex tarafından zaten yakalanmış aralıkları atlayarak,
// entropy eşiğinin üzerindeki string'leri düşük güvenli bulgu olarak
// işaretler. Böylece regex'e yakalanmayan ama rastgele görünen (olası
// token/secret) string'ler de kaçırılmaz.
func scanEntropy(content string, covered []offsetRange) []Finding {
	var findings []Finding

	for _, loc := range tokenPattern.FindAllStringIndex(content, -1) {
		start, end := loc[0], loc[1]
		if overlapsAny(start, end, covered) {
			continue
		}

		token := content[start:end]
		if ShannonEntropy(token) >= entropyThreshold {
			findings = append(findings, Finding{
				Type:       "high_entropy_string",
				Match:      maskMiddle(token),
				Confidence: ConfidenceLow,
				Offset:     start,
			})
		}
	}

	return findings
}
