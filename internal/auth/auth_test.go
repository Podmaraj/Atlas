package auth

import (
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"edgecore/internal/gateway/pipeline"
	"edgecore/internal/models"
)

func TestSecurityHeadersPlugin(t *testing.T) {
	plugin := NewSecurityHeadersPlugin()

	if plugin.Name() != "cors-security" {
		t.Errorf("expected plugin name 'cors-security', got %s", plugin.Name())
	}

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	ctx := pipeline.NewPipelineContext(rec, req)

	err := plugin.ExecuteRequest(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	h := rec.Header()
	if h.Get("X-Frame-Options") != "DENY" {
		t.Errorf("expected X-Frame-Options DENY, got %s", h.Get("X-Frame-Options"))
	}
	if h.Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("expected X-Content-Type-Options nosniff, got %s", h.Get("X-Content-Type-Options"))
	}
}

func TestAPIKeyPlugin(t *testing.T) {
	testKeyID := uuid.New()
	validator := func(keyHash string) (*models.ApiKey, bool) {
		return &models.ApiKey{
			BaseModel: models.BaseModel{ID: testKeyID},
			Status:    "active",
		}, true
	}

	plugin := NewAPIKeyPlugin(validator)
	if plugin.Name() != "api-key" {
		t.Errorf("expected plugin name 'api-key', got %s", plugin.Name())
	}

	req := httptest.NewRequest("GET", "/api/data", nil)
	req.Header.Set("X-API-Key", "test-secret-key")
	rec := httptest.NewRecorder()
	ctx := pipeline.NewPipelineContext(rec, req)

	err := plugin.ExecuteRequest(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if ctx.Aborted {
		t.Errorf("expected request not aborted for valid API key")
	}

	if ctx.ApiKey == nil || ctx.ApiKey.ID != testKeyID {
		t.Errorf("expected context ApiKey ID to match %s", testKeyID)
	}
}

func TestOAuth2Plugin_MissingToken(t *testing.T) {
	plugin := NewOAuth2Plugin(nil)
	if plugin.Name() != "oauth2-introspect" {
		t.Errorf("expected plugin name 'oauth2-introspect', got %s", plugin.Name())
	}

	req := httptest.NewRequest("GET", "/api/protected", nil)
	rec := httptest.NewRecorder()
	ctx := pipeline.NewPipelineContext(rec, req)

	err := plugin.ExecuteRequest(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !ctx.Aborted {
		t.Errorf("expected request to be aborted due to missing OAuth2 bearer token")
	}
}
