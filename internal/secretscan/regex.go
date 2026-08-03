// Package secretscan, indirilen dosya İÇERİĞİNİ (sadece aktif modda,
// activeprobe tarafından indirilen nesnelerde) tarayıp olası sızmış
// secret/token'ları bulur. Regex (yüksek/orta güven) ve Shannon entropy
// (düşük güven) yaklaşımlarını birleştirir.
package secretscan

import (
	"regexp"
	"strings"
)

// Confidence, bir bulgunun ne kadar güvenilir olduğunu belirtir.
type Confidence string

const (
	ConfidenceHigh   Confidence = "high"
	ConfidenceMedium Confidence = "medium"
	ConfidenceLow    Confidence = "low"
)

// Finding, içerikte bulunan tek bir olası secret/sır.
type Finding struct {
	Type       string
	Match      string // ortası maskelenmiş eşleşme (raporda gerçek sır sızdırılmasın diye)
	Confidence Confidence
	Offset     int
}

type regexRule struct {
	name       string
	pattern    *regexp.Regexp
	confidence Confidence
}

// regexRules, kesin desen eşleşmesiyle yüksek/orta güvenli bulgu üreten
// kurallardır: AWS access key, AWS secret key (bağlamsal), private key
// header, generic api_key= ve password= desenleri.
var regexRules = []regexRule{
	{
		name:       "aws_access_key",
		pattern:    regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
		confidence: ConfidenceHigh,
	},
	{
		name:       "aws_secret_key",
		pattern:    regexp.MustCompile(`(?i)aws_secret_access_key\s*[:=]\s*['"]?[A-Za-z0-9/+]{40}['"]?`),
		confidence: ConfidenceHigh,
	},
	{
		name:       "private_key",
		pattern:    regexp.MustCompile(`-----BEGIN[A-Z ]*PRIVATE KEY-----`),
		confidence: ConfidenceHigh,
	},
	{
		name:       "generic_api_key",
		pattern:    regexp.MustCompile(`(?i)api[_-]?key\s*[:=]\s*['"]?[A-Za-z0-9_\-]{16,}['"]?`),
		confidence: ConfidenceMedium,
	},
	{
		name:       "password",
		pattern:    regexp.MustCompile(`(?i)password\s*[:=]\s*['"]?\S{6,}['"]?`),
		confidence: ConfidenceMedium,
	},
}

type offsetRange struct {
	start, end int
}

func overlapsAny(start, end int, ranges []offsetRange) bool {
	for _, r := range ranges {
		if start < r.end && end > r.start {
			return true
		}
	}
	return false
}

// maskMiddle, eşleşen değerin tamamını rapor/logda göstermek yerine ortasını
// maskeler — bir sızıntı tarama aracı, bulduğu gerçek sırları kendisi de
// gereksiz yere açığa çıkarmamalı.
func maskMiddle(s string) string {
	if len(s) <= 8 {
		return strings.Repeat("*", len(s))
	}
	return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
}

func scanRegex(content string) ([]Finding, []offsetRange) {
	var findings []Finding
	var covered []offsetRange

	for _, rule := range regexRules {
		for _, loc := range rule.pattern.FindAllStringIndex(content, -1) {
			start, end := loc[0], loc[1]
			findings = append(findings, Finding{
				Type:       rule.name,
				Match:      maskMiddle(content[start:end]),
				Confidence: rule.confidence,
				Offset:     start,
			})
			covered = append(covered, offsetRange{start: start, end: end})
		}
	}
	return findings, covered
}

// Scan, verilen metin içeriğini önce regex kurallarıyla (yüksek/orta güven),
// ardından regex'e yakalanmayan bölümlerde entropy analiziyle (düşük güven)
// tarar ve tüm bulguları döner.
func Scan(content string) []Finding {
	findings, covered := scanRegex(content)
	findings = append(findings, scanEntropy(content, covered)...)
	return findings
}
