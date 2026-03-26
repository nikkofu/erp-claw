package integration

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/router"
)

func TestAdminReceivableFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	created := postJSONData(t, h, "/api/admin/v1/receivables", map[string]any{
		"external_ref": "SO-001",
	})
	receivableID := stringField(t, created, "id")
	if got := stringField(t, created, "status"); got != "open" {
		t.Fatalf("expected open receivable status, got %s", got)
	}

	detail := getJSONData(t, h, "/api/admin/v1/receivables/"+receivableID)
	if got := stringField(t, detail, "id"); got != receivableID {
		t.Fatalf("expected receivable id %s, got %s", receivableID, got)
	}
	if got := stringField(t, detail, "external_ref"); got != "SO-001" {
		t.Fatalf("expected external_ref SO-001, got %s", got)
	}

	list := getJSONArrayData(t, h, "/api/admin/v1/receivables")
	if len(list) != 1 {
		t.Fatalf("expected 1 receivable in list, got %d", len(list))
	}
	if got := stringField(t, list[0], "id"); got != receivableID {
		t.Fatalf("expected receivable list id %s, got %s", receivableID, got)
	}
}

func TestAdminReceivableCreateInvalidReturnsBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	env := doJSON(t, h, http.MethodPost, "/api/admin/v1/receivables", map[string]any{
		"external_ref": "   ",
	}, http.StatusBadRequest)
	if env.Meta["request_id"] == "" {
		t.Fatal("expected request_id metadata in bad request response")
	}
}
