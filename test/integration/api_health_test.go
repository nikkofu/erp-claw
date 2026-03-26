package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/router"
)

func TestHealthRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewTestContainer()
	h := router.New(router.WithContainer(container))

	cases := []struct {
		path   string
		status string
	}{
		{path: "/api/platform/v1/health/livez", status: "live"},
		{path: "/api/platform/v1/health/readyz", status: "ready"},
	}

	for _, tc := range cases {
		t.Run(tc.status, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req.Header.Set("X-Tenant-ID", "tenant-health")
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", rec.Code)
			}

			var resp struct {
				Data map[string]string `json:"data"`
				Meta map[string]string `json:"meta"`
			}
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("decode response: %v", err)
			}

			if resp.Data["status"] != tc.status {
				t.Fatalf("unexpected status: want=%s got=%s", tc.status, resp.Data["status"])
			}
			if resp.Meta["request_id"] == "" {
				t.Fatal("meta.request_id is empty")
			}
		})
	}
}

func TestHealthRoutesLive(t *testing.T) {
	if os.Getenv("ERP_CLAW_LIVE_SMOKE") != "1" {
		t.Skip("live smoke disabled")
	}

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/api/platform/v1/health/livez", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("X-Tenant-ID", "tenant-live-smoke")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body struct {
		Data map[string]string `json:"data"`
		Meta map[string]string `json:"meta"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.Data["status"] != "live" {
		t.Fatalf("unexpected live smoke status: %q", body.Data["status"])
	}
	if body.Meta["request_id"] == "" {
		t.Fatal("meta.request_id is empty")
	}
}
