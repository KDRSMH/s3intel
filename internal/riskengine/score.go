// Package riskengine, classifier (dosya adı bazlı temel risk) ve secretscan
// (içerikte bulunan secret'ların confidence'ı) sonuçlarını birleştirip
// 0-100 aralığında nihai bir risk skoru üretir ve sonuçları bu skora göre
// azalan sırada sıralar.
package riskengine

import (
	"sort"

	"s3intel/internal/classifier"
	"s3intel/internal/secretscan"
)

// Item, risk puanlaması için toplanan ham girdidir. Aktif modda Secrets
// dolu olabilir (içerik indirilip tarandığı için); pasif modda Secrets her
// zaman boştur çünkü dosya içeriği hiç indirilmez.
type Item struct {
	Key      string
	Source   string // "active" | "passive"
	Category classifier.Category
	BaseRisk int
	Secrets  []secretscan.Finding
}

// Result, bir Item'ın nihai 0-100 risk skoruyla birleştirilmiş hâlidir.
type Result struct {
	Item
	Score int
}

const (
	confidenceHighPoints   = 40
	confidenceMediumPoints = 20
	confidenceLowPoints    = 10
	maxScore               = 100
	minScore               = 0
)

// score, classifier'dan gelen temel riski + bulunan her secret'ın
// confidence'ına göre ek puanları toplayıp 0-100 aralığına sıkıştırır.
func score(item Item) int {
	total := item.BaseRisk
	for _, f := range item.Secrets {
		switch f.Confidence {
		case secretscan.ConfidenceHigh:
			total += confidenceHighPoints
		case secretscan.ConfidenceMedium:
			total += confidenceMediumPoints
		case secretscan.ConfidenceLow:
			total += confidenceLowPoints
		}
	}
	if total > maxScore {
		total = maxScore
	}
	if total < minScore {
		total = minScore
	}
	return total
}

// ScoreAll, verilen Item listesini puanlar ve sonucu risk skoruna göre
// AZALAN sırada döner (en riskli en üstte).
func ScoreAll(items []Item) []Result {
	results := make([]Result, 0, len(items))
	for _, item := range items {
		results = append(results, Result{Item: item, Score: score(item)})
	}
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	return results
}
