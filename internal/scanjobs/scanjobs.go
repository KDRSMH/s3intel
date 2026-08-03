// Package scanjobs, CLI (cmd) ve web arayüzü (internal/webui) tarafından
// ortaklaşa kullanılan iş mantığını barındırır: aktif/pasif taramayı çalıştırıp
// riskengine ile skorlanmış nihai sonucu üretir.
//
// Bu paket, activeprobe ve passiveintel'i BİRLİKTE kullanan tek üst-seviye
// orkestrasyon katmanıdır. Bu durum aktif/pasif izolasyonunu bozmaz: RunActive
// SADECE activeprobe.EnumerateBucket'ı çağırır (whitelist kontrolü orada
// uygulanır), RunPassive SADECE passiveintel.Search'ü çağırır — ikisi asla
// birbirinin verisine/ağ çağrısına karışmaz, sadece ortak riskengine.Result
// tipine dönüştürülürler.
package scanjobs

import (
	"context"

	"s3intel/internal/activeprobe"
	"s3intel/internal/classifier"
	"s3intel/internal/passiveintel"
	"s3intel/internal/riskengine"
)

// RunActive, whitelist'teki bir bucket'ı tarayıp skorlanmış sonuçları döner.
// Whitelist reddi durumunda hata, activeprobe.EnumerateBucket'tan olduğu gibi
// yukarı taşınır.
func RunActive(ctx context.Context, bucket string) ([]riskengine.Result, error) {
	findings, err := activeprobe.EnumerateBucket(ctx, bucket)
	if err != nil {
		return nil, err
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
	return riskengine.ScoreAll(items), nil
}

// RunPassive, grayhatwarfare (ya da GHW_API_KEY ayarlı değilse mock) üzerinden
// bir anahtar kelimeyi arayıp skorlanmış sonuçları döner. İçerik hiç
// indirilmediği için Secrets her zaman boştur.
func RunPassive(ctx context.Context, keyword string) ([]riskengine.Result, error) {
	return RunPassiveWithKey(ctx, keyword, "")
}

// RunPassiveWithKey, verilen API key ile grayhatwarfare üzerinden anahtar
// kelimeyi arayıp skorlanmış sonuçları döner. apiKey boşsa ortam değişkenine
// veya mock'a geri döner.
func RunPassiveWithKey(ctx context.Context, keyword, apiKey string) ([]riskengine.Result, error) {
	searchResults, err := passiveintel.SearchWithKey(ctx, keyword, apiKey)
	if err != nil {
		return nil, err
	}

	items := make([]riskengine.Item, 0, len(searchResults))
	for _, r := range searchResults {
		cat, baseRisk := classifier.Classify(r.Filename)
		items = append(items, riskengine.Item{
			Key:      r.FullPath,
			Source:   "passive",
			Category: cat,
			BaseRisk: baseRisk,
			Secrets:  nil,
		})
	}
	return riskengine.ScoreAll(items), nil
}
