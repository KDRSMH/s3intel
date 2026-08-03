// Package passiveintel, s3intel'in PASİF mod (grayhatwarfare API'sine HTTP
// sorgusu atan) tek kod yoludur. Bu paket hiçbir zaman bir AWS SDK import
// ETMEZ ve hiçbir zaman gerçek bir S3 bucket'ına bağlanmaz — sadece
// grayhatwarfare tarafından zaten indekslenmiş, üçüncü taraf verisini okur.
package passiveintel

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// SearchResult, grayhatwarfare API'sinden (ya da mock kaynaktan) dönen tek
// bir dosya kaydıdır.
type SearchResult struct {
	Bucket   string `json:"bucket"`
	Filename string `json:"filename"`
	FullPath string `json:"fullpath"`
	URL      string `json:"url"`
	Size     int64  `json:"size"`
	Region   string `json:"region"`
}

// Source, pasif istihbarat verisinin kaynağını soyutlayan arayüzdür.
// httpSource gerçek grayhatwarfare API'sine HTTP GET atar; mockSource ise
// API key olmadan da aracın test/demo edilebilmesi için sahte veri döner.
type Source interface {
	Search(ctx context.Context, keyword string) ([]SearchResult, error)
}

// defaultGHWEndpoint, grayhatwarfare "buckets" arama API'sinin uç noktasıdır.
const defaultGHWEndpoint = "https://buckets.grayhatwarfare.com/api/v1/files"

type httpSource struct {
	apiKey     string
	endpoint   string
	httpClient *http.Client
}

func newHTTPSource(apiKey string) *httpSource {
	return &httpSource{
		apiKey:     apiKey,
		endpoint:   defaultGHWEndpoint,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// Search, grayhatwarfare API'sine SADECE burada gerçek bir HTTP GET isteği
// gönderir. Bu satırların dışında bu pakette başka hiçbir ağ çağrısı yoktur.
func (h *httpSource) Search(ctx context.Context, keyword string) ([]SearchResult, error) {
	// GHW v1 API formatı: /api/v1/files/{keyword}?access_token={key}
	// Keyword URL path'inde olmalı, query parametresi olarak değil.
	reqURL := fmt.Sprintf("%s/%s?access_token=%s",
		strings.TrimRight(h.endpoint, "/"), url.PathEscape(keyword), url.QueryEscape(h.apiKey))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("grayhatwarfare isteği oluşturulamadı: %w", err)
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("grayhatwarfare API isteği başarısız: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("grayhatwarfare API beklenmeyen durum kodu döndürdü: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("grayhatwarfare yanıtı okunamadı: %w", err)
	}

	var parsed struct {
		Files []SearchResult `json:"files"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("grayhatwarfare yanıtı ayrıştırılamadı: %w", err)
	}
	return parsed.Files, nil
}

// mockSource, GHW_API_KEY ayarlı olmadığında kullanılan sahte veri
// kaynağıdır. Hiçbir ağ isteği göndermez, dolayısıyla API key olmadan da
// aracın uçtan uca test edilmesine izin verir.
type mockSource struct{}

func newMockSource() *mockSource { return &mockSource{} }

// mockPool, farklı anahtar kelimelere farklı sonuçlar döndürebilmek için
// geniş bir sahte dosya havuzudur. Her kayıt gerçekçi bir sızıntı senaryosunu
// temsil eder.
var mockPool = []SearchResult{
	// env / config dosyaları
	{Bucket: "corp-staging-assets", Filename: ".env", FullPath: ".env", URL: "https://corp-staging-assets.s3.amazonaws.com/.env", Size: 842, Region: "us-east-1"},
	{Bucket: "webapp-deploy-2024", Filename: "app.env", FullPath: "config/app.env", URL: "https://webapp-deploy-2024.s3.amazonaws.com/config/app.env", Size: 1203, Region: "us-east-1"},
	{Bucket: "devops-shared", Filename: "production.env", FullPath: "envs/production.env", URL: "https://devops-shared.s3.amazonaws.com/envs/production.env", Size: 2510, Region: "eu-west-1"},
	{Bucket: "startup-infra", Filename: "docker-compose.env", FullPath: "deploy/docker-compose.env", URL: "https://startup-infra.s3.amazonaws.com/deploy/docker-compose.env", Size: 675, Region: "us-west-2"},
	// config dosyaları
	{Bucket: "internal-tools-bkt", Filename: "config.yaml", FullPath: "config/config.yaml", URL: "https://internal-tools-bkt.s3.amazonaws.com/config/config.yaml", Size: 3890, Region: "us-east-1"},
	{Bucket: "devops-shared", Filename: "nginx.conf", FullPath: "server/nginx.conf", URL: "https://devops-shared.s3.amazonaws.com/server/nginx.conf", Size: 4200, Region: "eu-west-1"},
	{Bucket: "corp-staging-assets", Filename: "application.properties", FullPath: "spring/application.properties", URL: "https://corp-staging-assets.s3.amazonaws.com/spring/application.properties", Size: 1870, Region: "us-east-1"},
	// sql / database dosyaları
	{Bucket: "analytics-data-lake", Filename: "users_dump.sql", FullPath: "backups/users_dump.sql", URL: "https://analytics-data-lake.s3.amazonaws.com/backups/users_dump.sql", Size: 52428800, Region: "us-east-1"},
	{Bucket: "legacy-app-storage", Filename: "orders.sql", FullPath: "db/orders.sql", URL: "https://legacy-app-storage.s3.amazonaws.com/db/orders.sql", Size: 10485760, Region: "eu-central-1"},
	{Bucket: "startup-infra", Filename: "production.db", FullPath: "data/production.db", URL: "https://startup-infra.s3.amazonaws.com/data/production.db", Size: 104857600, Region: "us-west-2"},
	{Bucket: "webapp-deploy-2024", Filename: "schema.sqlite", FullPath: "database/schema.sqlite", URL: "https://webapp-deploy-2024.s3.amazonaws.com/database/schema.sqlite", Size: 2097152, Region: "us-east-1"},
	{Bucket: "analytics-data-lake", Filename: "customers.dump", FullPath: "exports/customers.dump", URL: "https://analytics-data-lake.s3.amazonaws.com/exports/customers.dump", Size: 8388608, Region: "us-east-1"},
	// private key dosyaları
	{Bucket: "devops-shared", Filename: "id_rsa", FullPath: ".ssh/id_rsa", URL: "https://devops-shared.s3.amazonaws.com/.ssh/id_rsa", Size: 3243, Region: "eu-west-1"},
	{Bucket: "corp-staging-assets", Filename: "server.pem", FullPath: "certs/server.pem", URL: "https://corp-staging-assets.s3.amazonaws.com/certs/server.pem", Size: 1704, Region: "us-east-1"},
	{Bucket: "internal-tools-bkt", Filename: "private.key", FullPath: "ssl/private.key", URL: "https://internal-tools-bkt.s3.amazonaws.com/ssl/private.key", Size: 1675, Region: "us-east-1"},
	{Bucket: "startup-infra", Filename: "id_ed25519", FullPath: "keys/id_ed25519", URL: "https://startup-infra.s3.amazonaws.com/keys/id_ed25519", Size: 464, Region: "us-west-2"},
	// credential dosyaları
	{Bucket: "devops-shared", Filename: "credentials", FullPath: ".aws/credentials", URL: "https://devops-shared.s3.amazonaws.com/.aws/credentials", Size: 290, Region: "eu-west-1"},
	{Bucket: "legacy-app-storage", Filename: "credential-store.json", FullPath: "auth/credential-store.json", URL: "https://legacy-app-storage.s3.amazonaws.com/auth/credential-store.json", Size: 1580, Region: "eu-central-1"},
	// backup dosyaları
	{Bucket: "legacy-app-storage", Filename: "site-backup.bak", FullPath: "backups/site-backup.bak", URL: "https://legacy-app-storage.s3.amazonaws.com/backups/site-backup.bak", Size: 209715200, Region: "eu-central-1"},
	{Bucket: "corp-staging-assets", Filename: "db-backup-2024.old", FullPath: "archive/db-backup-2024.old", URL: "https://corp-staging-assets.s3.amazonaws.com/archive/db-backup-2024.old", Size: 524288000, Region: "us-east-1"},
	{Bucket: "analytics-data-lake", Filename: "full-backup.tar.gz", FullPath: "disaster-recovery/full-backup.tar.gz", URL: "https://analytics-data-lake.s3.amazonaws.com/disaster-recovery/full-backup.tar.gz", Size: 1073741824, Region: "us-east-1"},
	// archive dosyaları
	{Bucket: "internal-tools-bkt", Filename: "source-code.zip", FullPath: "releases/source-code.zip", URL: "https://internal-tools-bkt.s3.amazonaws.com/releases/source-code.zip", Size: 15728640, Region: "us-east-1"},
	{Bucket: "legacy-app-storage", Filename: "logs-2024.tar.gz", FullPath: "logs/logs-2024.tar.gz", URL: "https://legacy-app-storage.s3.amazonaws.com/logs/logs-2024.tar.gz", Size: 31457280, Region: "eu-central-1"},
	// diğer dosyalar
	{Bucket: "webapp-deploy-2024", Filename: "passwords.txt", FullPath: "notes/passwords.txt", URL: "https://webapp-deploy-2024.s3.amazonaws.com/notes/passwords.txt", Size: 512, Region: "us-east-1"},
	{Bucket: "internal-tools-bkt", Filename: "api-tokens.csv", FullPath: "admin/api-tokens.csv", URL: "https://internal-tools-bkt.s3.amazonaws.com/admin/api-tokens.csv", Size: 2048, Region: "us-east-1"},
	{Bucket: "startup-infra", Filename: "secret-vars.json", FullPath: "ci/secret-vars.json", URL: "https://startup-infra.s3.amazonaws.com/ci/secret-vars.json", Size: 934, Region: "us-west-2"},
}

func (m *mockSource) Search(ctx context.Context, keyword string) ([]SearchResult, error) {
	lower := strings.ToLower(keyword)

	// Keyword ile eşleşen sonuçları filtrele
	var matched []SearchResult
	for _, r := range mockPool {
		entry := strings.ToLower(r.Filename + " " + r.FullPath + " " + r.Bucket)
		if strings.Contains(entry, lower) {
			matched = append(matched, r)
		}
	}

	// Eşleşme yoksa tüm havuzdan ilk 5'ini döndür (genel arama simülasyonu)
	if len(matched) == 0 {
		limit := 5
		if len(mockPool) < limit {
			limit = len(mockPool)
		}
		matched = mockPool[:limit]
	}

	return matched, nil
}

// NewSource, GHW_API_KEY ortam değişkeni ayarlıysa gerçek HTTP kaynağını,
// ayarlı değilse mock (sahte veri) kaynağını döner. Böylece kullanıcı gerçek
// bir API key olmadan da aracı deneyebilir.
func NewSource() Source {
	apiKey := os.Getenv("GHW_API_KEY")
	if apiKey == "" {
		return newMockSource()
	}
	return newHTTPSource(apiKey)
}

// NewSourceWithKey, verilen API key'i kullanarak kaynak seçer. Eğer apiKey
// boşsa GHW_API_KEY ortam değişkenine geri döner; o da boşsa mock kaynağı
// kullanır. Web arayüzünden çalışma zamanında API key girilmesini sağlar.
func NewSourceWithKey(apiKey string) Source {
	if apiKey == "" {
		return NewSource()
	}
	return newHTTPSource(apiKey)
}

// Search, NewSource() ile seçilen kaynağı (gerçek API ya da mock) verilen
// anahtar kelimeyle sorgular.
func Search(ctx context.Context, keyword string) ([]SearchResult, error) {
	return NewSource().Search(ctx, keyword)
}

// SearchWithKey, verilen API key ile kaynağı seçerek arama yapar. apiKey
// boşsa ortam değişkenine/mock'a geri döner.
func SearchWithKey(ctx context.Context, keyword, apiKey string) ([]SearchResult, error) {
	return NewSourceWithKey(apiKey).Search(ctx, keyword)
}
