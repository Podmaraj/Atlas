package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"edgecore/internal/gateway/pipeline"
	"edgecore/internal/models"
)

type APIKeyValidatorFunc func(keyHash string) (*models.ApiKey, bool)

type APIKeyPlugin struct {
	headerName string
	paramName  string
	validator  APIKeyValidatorFunc
}

func NewAPIKeyPlugin(validator APIKeyValidatorFunc) *APIKeyPlugin {
	return &APIKeyPlugin{
		headerName: "X-API-Key",
		paramName:  "apikey",
		validator:  validator,
	}
}

func (p *APIKeyPlugin) Name() string {
	return "api-key"
}

func (p *APIKeyPlugin) Init(config models.JSONMap) error {
	if header, ok := config["header_name"].(string); ok && header != "" {
		p.headerName = header
	}
	if param, ok := config["param_name"].(string); ok && param != "" {
		p.paramName = param
	}
	return nil
}

func (p *APIKeyPlugin) ExecuteRequest(ctx *pipeline.PipelineContext) error {
	rawKey := p.extractKey(ctx.Request)
	if rawKey == "" {
		ctx.Abort(http.StatusUnauthorized, `{"error":"Unauthorized","message":"Missing X-API-Key header or query parameter"}`)
		return nil
	}

	keyHash := hashAPIKey(rawKey)

	if p.validator == nil {
		ctx.Abort(http.StatusInternalServerError, `{"error":"Internal Error","message":"API Key validator unconfigured"}`)
		return nil
	}

	apiKeyObj, valid := p.validator(keyHash)
	if !valid || apiKeyObj.Status != "active" {
		ctx.Abort(http.StatusForbidden, `{"error":"Forbidden","message":"Invalid or inactive API Key"}`)
		return nil
	}

	// Store ApiKey and TenantID in metadata
	ctx.ApiKey = apiKeyObj
	ctx.SetMetadata("tenant_id", apiKeyObj.TenantID.String())
	ctx.SetMetadata("api_key_id", apiKeyObj.ID.String())

	return nil
}

func (p *APIKeyPlugin) ExecuteResponse(ctx *pipeline.PipelineContext) error {
	return nil
}

func (p *APIKeyPlugin) extractKey(r *http.Request) string {
	key := r.Header.Get(p.headerName)
	if key != "" {
		return key
	}
	return r.URL.Query().Get(p.paramName)
}

func hashAPIKey(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
