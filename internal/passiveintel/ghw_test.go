package passiveintel

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestMockSource_ReturnsSampleData(t *testing.T) {
	src := newMockSource()
	results, err := src.Search(context.Background(), "backup")
	if err != nil {
		t.Fatalf("beklenmeyen hata: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("mock kaynak en az bir sonuç dönmeliydi")
	}
	found := false
	for _, r := range results {
		if strings.Contains(r.Filename, "backup") {
			found = true
		}
	}
	if !found {
		t.Errorf("sonuçlar aranan anahtar kelimeyi (backup) içermeliydi: %+v", results)
	}
}

func TestNewSource_UsesMockWhenNoAPIKey(t *testing.T) {
	t.Setenv("GHW_API_KEY", "")
	os.Unsetenv("GHW_API_KEY")

	src := NewSource()
	if _, ok := src.(*mockSource); !ok {
		t.Fatalf("GHW_API_KEY yokken mockSource dönmeliydi, %T döndü", src)
	}
}

func TestNewSource_UsesHTTPWhenAPIKeySet(t *testing.T) {
	t.Setenv("GHW_API_KEY", "fake-test-key")

	src := NewSource()
	if _, ok := src.(*httpSource); !ok {
		t.Fatalf("GHW_API_KEY ayarlıyken httpSource dönmeliydi, %T döndü", src)
	}
}

// TestHTTPSource_ParsesResponse, gerçek grayhatwarfare API'sine hiç
// dokunmadan, yerel bir httptest.Server üzerinden HTTP yanıt ayrıştırmanın
// doğru çalıştığını kanıtlar.
func TestHTTPSource_ParsesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"files":[{"bucket":"b1","filename":"f1.env","fullpath":"x/f1.env","url":"https://b1.s3.amazonaws.com/x/f1.env","size":100,"region":"us-east-1"}]}`))
	}))
	defer server.Close()

	src := &httpSource{apiKey: "fake", endpoint: server.URL, httpClient: server.Client()}
	results, err := src.Search(context.Background(), "test")
	if err != nil {
		t.Fatalf("beklenmeyen hata: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("1 sonuç bekleniyordu, %d geldi", len(results))
	}
	if results[0].Bucket != "b1" || results[0].Filename != "f1.env" {
		t.Errorf("ayrıştırılan sonuç beklenmedik: %+v", results[0])
	}
}

func TestHTTPSource_NonOKStatusReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	src := &httpSource{apiKey: "bad-key", endpoint: server.URL, httpClient: server.Client()}
	_, err := src.Search(context.Background(), "test")
	if err == nil {
		t.Fatal("401 yanıtı için hata bekleniyordu, nil döndü")
	}
}
