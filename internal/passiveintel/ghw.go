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
	reqURL := fmt.Sprintf("%s?keywords=%s&access_token=%s",
		h.endpoint, url.QueryEscape(keyword), url.QueryEscape(h.apiKey))

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

func (m *mockSource) Search(ctx context.Context, keyword string) ([]SearchResult, error) {
	sample := []SearchResult{
		{
			Bucket:   "example-public-bucket",
			Filename: fmt.Sprintf("%s-config.env", keyword),
			FullPath: fmt.Sprintf("backups/%s-config.env", keyword),
			URL:      fmt.Sprintf("https://example-public-bucket.s3.amazonaws.com/backups/%s-config.env", keyword),
			Size:     1523,
			Region:   "us-east-1",
		},
		{
			Bucket:   "example-public-bucket",
			Filename: fmt.Sprintf("%s-database.sql", keyword),
			FullPath: fmt.Sprintf("dumps/%s-database.sql", keyword),
			URL:      fmt.Sprintf("https://example-public-bucket.s3.amazonaws.com/dumps/%s-database.sql", keyword),
			Size:     984213,
			Region:   "us-east-1",
		},
		{
			Bucket:   "another-leaky-bucket",
			Filename: fmt.Sprintf("%s.bak", keyword),
			FullPath: fmt.Sprintf("old/%s.bak", keyword),
			URL:      fmt.Sprintf("https://another-leaky-bucket.s3.amazonaws.com/old/%s.bak", keyword),
			Size:     20481,
			Region:   "eu-west-1",
		},
	}
	return sample, nil
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

// Search, NewSource() ile seçilen kaynağı (gerçek API ya da mock) verilen
// anahtar kelimeyle sorgular.
func Search(ctx context.Context, keyword string) ([]SearchResult, error) {
	return NewSource().Search(ctx, keyword)
}
