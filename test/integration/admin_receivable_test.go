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

func TestAdminReceivableListSupportsStatusSortAndPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	createReceivable := func(externalRef string) string {
		resp := postJSONData(t, h, "/api/admin/v1/receivables", map[string]any{
			"external_ref": externalRef,
		})
		return stringField(t, resp, "id")
	}

	receivableA := createReceivable("SO-RCV-001")
	receivableB := createReceivable("SO-RCV-002")
	receivableC := createReceivable("SO-RCV-003")

	page1 := doJSONForArray(t, h, http.MethodGet, "/api/admin/v1/receivables?sort=id_asc&page=1&page_size=2", nil, http.StatusOK).Data
	if len(page1) != 2 {
		t.Fatalf("expected 2 receivable bills in page1, got %d", len(page1))
	}
	if got := stringField(t, page1[0], "id"); got != receivableA {
		t.Fatalf("expected page1 first receivable id %s, got %s", receivableA, got)
	}
	if got := stringField(t, page1[1], "id"); got != receivableB {
		t.Fatalf("expected page1 second receivable id %s, got %s", receivableB, got)
	}

	page2 := doJSONForArray(t, h, http.MethodGet, "/api/admin/v1/receivables?sort=id_asc&page=2&page_size=2", nil, http.StatusOK).Data
	if len(page2) != 1 {
		t.Fatalf("expected 1 receivable bill in page2, got %d", len(page2))
	}
	if got := stringField(t, page2[0], "id"); got != receivableC {
		t.Fatalf("expected page2 receivable id %s, got %s", receivableC, got)
	}

	openBills := doJSONForArray(t, h, http.MethodGet, "/api/admin/v1/receivables?status=open", nil, http.StatusOK).Data
	if len(openBills) != 3 {
		t.Fatalf("expected 3 open receivable bills, got %d", len(openBills))
	}
}

func TestAdminReceivableListRejectsInvalidQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	cases := []string{
		"/api/admin/v1/receivables?status=closed",
		"/api/admin/v1/receivables?sort=unknown",
		"/api/admin/v1/receivables?page=0",
		"/api/admin/v1/receivables?page_size=0",
	}
	for _, path := range cases {
		env := doJSON(t, h, http.MethodGet, path, nil, http.StatusBadRequest)
		if env.Meta["request_id"] == "" {
			t.Fatalf("expected request_id metadata in bad request response for %s", path)
		}
	}
}
