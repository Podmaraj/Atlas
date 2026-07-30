package auth

import (
	"net/http"
	"strings"

	"edgecore/internal/gateway/pipeline"
	"edgecore/internal/models"
)

type SecurityHeadersPlugin struct {
	allowOrigin      string
	allowMethods     string
	allowHeaders     string
	allowCredentials bool
	maxAge           string
	enableHSTS       bool
}

func NewSecurityHeadersPlugin() *SecurityHeadersPlugin {
	return &SecurityHeadersPlugin{
		allowOrigin:      "*",
		allowMethods:     "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		allowHeaders:     "Content-Type, Authorization, X-API-Key, X-Request-ID",
		allowCredentials: true,
		maxAge:           "86400",
		enableHSTS:       true,
	}
}

func (p *SecurityHeadersPlugin) Name() string {
	return "cors-security"
}

func (p *SecurityHeadersPlugin) Init(config models.JSONMap) error {
	if origin, ok := config["allow_origin"].(string); ok {
		p.allowOrigin = origin
	}
	if methods, ok := config["allow_methods"].(string); ok {
		p.allowMethods = methods
	}
	if headers, ok := config["allow_headers"].(string); ok {
		p.allowHeaders = headers
	}
	if hsts, ok := config["enable_hsts"].(bool); ok {
		p.enableHSTS = hsts
	}
	return nil
}

func (p *SecurityHeadersPlugin) ExecuteRequest(ctx *pipeline.PipelineContext) error {
	w := ctx.Writer
	r := ctx.Request

	// Standard Security Headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	if p.enableHSTS {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	}

	// CORS Headers
	origin := r.Header.Get("Origin")
	if origin != "" {
		if p.allowOrigin == "*" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else if strings.Contains(p.allowOrigin, origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
	} else {
		w.Header().Set("Access-Control-Allow-Origin", p.allowOrigin)
	}

	w.Header().Set("Access-Control-Allow-Methods", p.allowMethods)
	w.Header().Set("Access-Control-Allow-Headers", p.allowHeaders)
	if p.allowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}
	w.Header().Set("Access-Control-Max-Age", p.maxAge)

	// Intercept OPTIONS Preflight Request
	if r.Method == http.MethodOptions {
		ctx.Abort(http.StatusNoContent, "")
		return nil
	}

	return nil
}

func (p *SecurityHeadersPlugin) ExecuteResponse(ctx *pipeline.PipelineContext) error {
	return nil
}
