package gingotenancy

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	tenantctx "github.com/DarkInno/gotenancy/core/context"
	"github.com/DarkInno/gotenancy/core/resolver"
	"github.com/DarkInno/gotenancy/core/store"
	"github.com/DarkInno/gotenancy/core/types"

	"github.com/gin-gonic/gin"
)

func TestTenantMiddlewareRejectsInactiveTenantByDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)
	backing := store.NewMemoryStore()
	active := types.Tenant{ID: "tenant-a", Name: "Tenant A", Status: types.TenantStatusActive}
	if err := backing.Create(context.Background(), active); err != nil {
		t.Fatalf("store.Create(active) error = %v", err)
	}
	if err := backing.Create(context.Background(), types.Tenant{ID: "tenant-b", Name: "Tenant B", Status: types.TenantStatusSuspended}); err != nil {
		t.Fatalf("store.Create(suspended) error = %v", err)
	}

	router := gin.New()
	router.Use(TenantMiddleware(resolver.NewComposite(resolver.NewHeaderContrib("", types.TenantIDStrategyString)), backing))
	router.GET("/ok", func(c *gin.Context) {
		tenant, ok := tenantctx.FromContext(c.Request.Context())
		if !ok {
			t.Fatal("tenant missing from request context")
		}
		c.JSON(http.StatusOK, gin.H{"tenant": tenant.ID.String()})
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/ok", nil)
	request.Header.Set(resolver.DefaultHeaderName, "tenant-a")
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("active tenant status = %d, body = %s", recorder.Code, recorder.Body.String())
	}

	recorder = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/ok", nil)
	request.Header.Set(resolver.DefaultHeaderName, "tenant-b")
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("suspended tenant status = %d, want 403", recorder.Code)
	}
}

func TestTenantMiddlewareRejectsMissingAndUnknownTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(TenantMiddleware(resolver.NewComposite(resolver.NewHeaderContrib("", types.TenantIDStrategyString)), store.NewMemoryStore()))
	router.GET("/ok", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/ok", nil))
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("missing tenant status = %d, want 401", recorder.Code)
	}

	recorder = httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/ok", nil)
	request.Header.Set(resolver.DefaultHeaderName, "missing")
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("unknown tenant status = %d, want 403", recorder.Code)
	}
}

func TestHostGuardMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/host", HostGuardMiddleware(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.GET("/host-ok", func(c *gin.Context) {
		c.Request = c.Request.WithContext(tenantctx.WithHost(c.Request.Context()))
	}, HostGuardMiddleware(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/host", nil))
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("host guard without host status = %d, want 403", recorder.Code)
	}

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/host-ok", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("host guard with host status = %d, want 200", recorder.Code)
	}
}

func TestErrorHandlerHidesInternalErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/err", func(c *gin.Context) {
		_ = c.Error(errors.New("tenant tenant-a failed"))
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/err", nil))
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("error handler status = %d, want 500", recorder.Code)
	}
	if body := recorder.Body.String(); body != "{\"error\":\"internal_error\"}" {
		t.Fatalf("error handler body = %s, want generic error", body)
	}
}
