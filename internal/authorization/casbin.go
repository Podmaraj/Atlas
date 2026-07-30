package authorization

import (
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
)

type Enforcer struct {
	e *casbin.Enforcer
}

// NewEnforcer initializes in-memory Casbin RBAC enforcer
func NewEnforcer() (*Enforcer, error) {
	text := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch(r.obj, p.obj) && (r.act == p.act || p.act == "*")
`

	m, err := model.NewModelFromString(text)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin model: %w", err)
	}

	e, err := casbin.NewEnforcer(m)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin enforcer: %w", err)
	}

	// Default RBAC policies
	_, _ = e.AddPolicy("superadmin", "*", "*")
	_, _ = e.AddPolicy("admin", "/api/v1/*", "*")
	_, _ = e.AddPolicy("viewer", "/api/v1/*", "GET")

	// Role mappings
	_, _ = e.AddGroupingPolicy("admin", "admin")
	_, _ = e.AddGroupingPolicy("viewer", "viewer")

	return &Enforcer{e: e}, nil
}

// Enforce checks if role/user has permission to perform action on resource
func (e *Enforcer) Enforce(sub, obj, act string) bool {
	ok, _ := e.e.Enforce(sub, obj, act)
	return ok
}
