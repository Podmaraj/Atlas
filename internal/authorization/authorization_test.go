package authorization

import (
	"testing"
)

func TestCasbinEnforcer(t *testing.T) {
	enforcer, err := NewEnforcer()
	if err != nil {
		t.Fatalf("failed to create casbin enforcer: %v", err)
	}

	// Test superadmin permissions
	if !enforcer.Enforce("superadmin", "/api/v1/services", "POST") {
		t.Errorf("expected superadmin to have access to POST /api/v1/services")
	}

	// Test viewer permissions (GET allowed, POST denied)
	if !enforcer.Enforce("viewer", "/api/v1/services", "GET") {
		t.Errorf("expected viewer to have access to GET /api/v1/services")
	}
	if enforcer.Enforce("viewer", "/api/v1/services", "POST") {
		t.Errorf("expected viewer to be denied access to POST /api/v1/services")
	}
}
