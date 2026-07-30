package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"edgecore/internal/gateway/pipeline"
	"edgecore/internal/models"
	"edgecore/pkg/redis"
)

type OAuth2IntrospectionResponse struct {
	Active   bool   `json:"active"`
	Subject  string `json:"sub,omitempty"`
	ClientID string `json:"client_id,omitempty"`
	Scope    string `json:"scope,omitempty"`
}

type OAuth2Plugin struct {
	introspectionURL string
	clientID         string
	clientSecret     string
	redisClient      *redis.Client
	httpClient       *http.Client
	ttl              time.Duration
}

func NewOAuth2Plugin(rc *redis.Client) *OAuth2Plugin {
	return &OAuth2Plugin{
		redisClient: rc,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		ttl: 300 * time.Second,
	}
}

func (p *OAuth2Plugin) Name() string {
	return "oauth2-introspect"
}

func (p *OAuth2Plugin) Init(config models.JSONMap) error {
	if u, ok := config["introspection_url"].(string); ok {
		p.introspectionURL = u
	}
	if id, ok := config["client_id"].(string); ok {
		p.clientID = id
	}
	if secret, ok := config["client_secret"].(string); ok {
		p.clientSecret = secret
	}
	return nil
}

func (p *OAuth2Plugin) ExecuteRequest(ctx *pipeline.PipelineContext) error {
	token := p.extractBearerToken(ctx.Request)
	if token == "" {
		ctx.Abort(http.StatusUnauthorized, `{"error":"Unauthorized","message":"Missing OAuth2 Bearer token"}`)
		return nil
	}

	// 1. Check Redis Cache for introspection result
	cacheKey := fmt.Sprintf("edgecore:oauth2:%s", token)
	if p.redisClient != nil {
		cachedJson, err := p.redisClient.Get(context.Background(), cacheKey)
		if err == nil && cachedJson != "" {
			var cachedResp OAuth2IntrospectionResponse
			if err := json.Unmarshal([]byte(cachedJson), &cachedResp); err == nil && cachedResp.Active {
				ctx.SetMetadata("user_id", cachedResp.Subject)
				ctx.SetMetadata("client_id", cachedResp.ClientID)
				ctx.SetMetadata("scopes", cachedResp.Scope)
				return nil
			}
		}
	}

	// 2. Perform Introspection Call if URL configured
	if p.introspectionURL == "" {
		// Mock pass-through if URL unconfigured
		ctx.SetMetadata("user_id", "oauth2-user")
		return nil
	}

	data := url.Values{}
	data.Set("token", token)

	req, err := http.NewRequest("POST", p.introspectionURL, strings.NewReader(data.Encode()))
	if err != nil {
		ctx.Abort(http.StatusInternalServerError, `{"error":"Internal Error","message":"Failed to construct introspection request"}`)
		return nil
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if p.clientID != "" && p.clientSecret != "" {
		req.SetBasicAuth(p.clientID, p.clientSecret)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		ctx.Abort(http.StatusUnauthorized, `{"error":"Unauthorized","message":"Token introspection failed"}`)
		return nil
	}
	defer resp.Body.Close()

	var introspectResp OAuth2IntrospectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&introspectResp); err != nil || !introspectResp.Active {
		ctx.Abort(http.StatusUnauthorized, `{"error":"Unauthorized","message":"Inactive or invalid OAuth2 token"}`)
		return nil
	}

	// Cache result in Redis
	if p.redisClient != nil {
		bytes, _ := json.Marshal(introspectResp)
		_ = p.redisClient.Set(context.Background(), cacheKey, string(bytes), p.ttl)
	}

	ctx.SetMetadata("user_id", introspectResp.Subject)
	ctx.SetMetadata("client_id", introspectResp.ClientID)
	ctx.SetMetadata("scopes", introspectResp.Scope)

	return nil
}

func (p *OAuth2Plugin) ExecuteResponse(ctx *pipeline.PipelineContext) error {
	return nil
}

func (p *OAuth2Plugin) extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		return strings.TrimSpace(auth[7:])
	}
	return ""
}
