package routes

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// newTestRouter builds the full route tree without a database. Tests only
// exercise paths that return before any DB access (routing, auth gating,
// request validation), so config.DB being nil is fine.
func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	SetupRoutes(r)
	return r
}

func do(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	var rdr *strings.Reader
	if body != "" {
		rdr = strings.NewReader(body)
	} else {
		rdr = strings.NewReader("")
	}
	req := httptest.NewRequest(method, path, rdr)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestHealthAndDocs(t *testing.T) {
	r := newTestRouter()
	if w := do(r, http.MethodGet, "/api/health", ""); w.Code != http.StatusOK {
		t.Errorf("/api/health = %d, want 200", w.Code)
	}
	if w := do(r, http.MethodGet, "/api/docs", ""); w.Code != http.StatusOK {
		t.Errorf("/api/docs = %d, want 200", w.Code)
	}
}

func TestProtectedRouteRequiresAuth(t *testing.T) {
	r := newTestRouter()
	for _, path := range []string{"/api/cart", "/api/orders", "/api/wishlist"} {
		if w := do(r, http.MethodGet, path, ""); w.Code != http.StatusUnauthorized {
			t.Errorf("GET %s without token = %d, want 401", path, w.Code)
		}
	}
}

func TestLoginValidation(t *testing.T) {
	r := newTestRouter()
	// Missing required fields must be rejected before any DB lookup.
	if w := do(r, http.MethodPost, "/api/login", "{}"); w.Code != http.StatusUnprocessableEntity {
		t.Errorf("POST /api/login {} = %d, want 422", w.Code)
	}
}

func TestRegisterValidation(t *testing.T) {
	r := newTestRouter()
	body := `{"name":"x","email":"not-an-email","password":"short"}`
	if w := do(r, http.MethodPost, "/api/register", body); w.Code != http.StatusUnprocessableEntity {
		t.Errorf("POST /api/register invalid = %d, want 422", w.Code)
	}
}

func TestWebhookDisabledWithoutSecret(t *testing.T) {
	os.Unsetenv("WEBHOOK_SECRET")
	r := newTestRouter()
	body := `{"order_id":1,"status":"succeeded"}`
	if w := do(r, http.MethodPost, "/api/payments/webhook", body); w.Code != http.StatusServiceUnavailable {
		t.Errorf("webhook without secret = %d, want 503", w.Code)
	}
}

func TestUnknownRoute(t *testing.T) {
	r := newTestRouter()
	if w := do(r, http.MethodGet, "/api/does-not-exist", ""); w.Code != http.StatusNotFound {
		t.Errorf("unknown route = %d, want 404", w.Code)
	}
}
