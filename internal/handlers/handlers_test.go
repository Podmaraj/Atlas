package handlers

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	"edgecore/internal/config"
)

func TestHandler_LoginFallback(t *testing.T) {
	cfg := &config.Config{
		ControlPlane: config.ControlPlaneConfig{
			AdminUsername: "admin",
			AdminPassword: "password123",
			JWTSecret:     "secret-key-1234567890",
			JWTExpiration: 3600,
		},
	}

	h := NewHandler(nil, nil, cfg)

	app := fiber.New()
	app.Post("/login", h.Login)

	body := `{"username":"admin","password":"password123"}`
	req := httptest.NewRequest("POST", "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status code 200, got %d", resp.StatusCode)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var res map[string]interface{}
	if err := json.Unmarshal(respBody, &res); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	if res["token"] == nil {
		t.Errorf("expected token in response")
	}
}
