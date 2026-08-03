package webui

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleIndex_EmptyFormRendersWithoutResults(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handleIndex(rec, req)

	if rec.Code != 200 {
		t.Fatalf("200 bekleniyordu, %d geldi", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "<form") {
		t.Error("sayfa formu içermeli")
	}
	if strings.Contains(body, "<table>") {
		t.Error("target verilmeden sonuç tablosu gösterilmemeli")
	}
}

func TestHandleIndex_PassiveModeRendersResults(t *testing.T) {
	t.Setenv("GHW_API_KEY", "")

	req := httptest.NewRequest("GET", "/?mode=passive&target=backup", nil)
	rec := httptest.NewRecorder()
	handleIndex(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "<table>") {
		t.Errorf("sonuç tablosu bekleniyordu: %s", body)
	}
	if !strings.Contains(body, "backup") {
		t.Errorf("sonuçlar aranan kelimeyi içermeli: %s", body)
	}
}

func TestHandleIndex_ActiveModeShowsWhitelistError(t *testing.T) {
	req := httptest.NewRequest("GET", "/?mode=active&target=definitely-not-whitelisted", nil)
	rec := httptest.NewRecorder()
	handleIndex(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "class=\"error\"") {
		t.Errorf("whitelist reddi hata kutusu olarak gösterilmeliydi: %s", body)
	}
	if !strings.Contains(body, "whitelist") {
		t.Errorf("hata mesajı whitelist reddini belirtmeli: %s", body)
	}
}
