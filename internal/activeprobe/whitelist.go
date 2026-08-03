// Package activeprobe, s3intel'in AKTIF mod (gerçek AWS S3 API çağrıları
// yapan) tek kod yoludur. Bu dosya, tüm aktif tarama işlemlerinin önündeki
// tek kapı bekçisidir: config/whitelist.yaml içinde listelenmeyen hiçbir
// bucket'a probe.go içinden erişilemez.
package activeprobe

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// defaultWhitelistPath, whitelist.yaml'ın proje kökünden çalıştırıldığında
// bulunacağı varsayılan konumdur.
const defaultWhitelistPath = "config/whitelist.yaml"

// whitelistPathOverride, testlerin (ve gerektiğinde CLI'nin) farklı bir
// whitelist dosyası göstermesine izin verir.
var whitelistPathOverride string

// SetWhitelistPath, whitelist dosyasının okunacağı yolu değiştirir. Boş
// string verilirse defaultWhitelistPath'e geri döner.
func SetWhitelistPath(path string) {
	whitelistPathOverride = path
}

func whitelistPath() string {
	if whitelistPathOverride != "" {
		return whitelistPathOverride
	}
	return defaultWhitelistPath
}

type whitelistConfig struct {
	AllowedBuckets []string `yaml:"allowed_buckets"`
}

func loadWhitelist() (map[string]bool, error) {
	path := whitelistPath()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("whitelist dosyası okunamadı (%s): %w", path, err)
	}

	var cfg whitelistConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("whitelist dosyası ayrıştırılamadı (%s): %w", path, err)
	}

	allowed := make(map[string]bool, len(cfg.AllowedBuckets))
	for _, bucket := range cfg.AllowedBuckets {
		allowed[bucket] = true
	}
	return allowed, nil
}

// IsAllowed, verilen bucket adının config/whitelist.yaml içindeki
// allowed_buckets listesinde olup olmadığını döner. Whitelist dosyası hiç
// okunamazsa (yok, bozuk, izin hatası vb.) güvenli taraf tercih edilir ve
// false döner — yani whitelist doğrulanamıyorsa hiçbir bucket'a izin verilmez.
func IsAllowed(bucketName string) bool {
	allowed, err := loadWhitelist()
	if err != nil {
		return false
	}
	return allowed[bucketName]
}
