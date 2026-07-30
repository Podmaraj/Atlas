package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"edgecore/internal/gateway/pipeline"
	"edgecore/internal/models"
)

type JWTPlugin struct {
	secret      []byte
	algorithm   string
	issuer      string
	audience    string
	headerName  string
	paramName   string
}

func NewJWTPlugin() *JWTPlugin {
	return &JWTPlugin{
		algorithm:  "HS256",
		headerName: "Authorization",
		paramName:  "token",
	}
}

func (p *JWTPlugin) Name() string {
	return "jwt"
}

func (p *JWTPlugin) Init(config models.JSONMap) error {
	if secretStr, ok := config["secret"].(string); ok {
		p.secret = []byte(secretStr)
	}
	if algoStr, ok := config["algorithm"].(string); ok {
		p.algorithm = algoStr
	}
	if issStr, ok := config["issuer"].(string); ok {
		p.issuer = issStr
	}
	if audStr, ok := config["audience"].(string); ok {
		p.audience = audStr
	}
	return nil
}

func (p *JWTPlugin) ExecuteRequest(ctx *pipeline.PipelineContext) error {
	tokenStr := p.extractToken(ctx.Request)
	if tokenStr == "" {
		ctx.Abort(http.StatusUnauthorized, `{"error":"Unauthorized","message":"Missing Authorization Bearer token"}`)
		return nil
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if p.algorithm != "" && token.Method.Alg() != p.algorithm {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return p.secret, nil
	})

	if err != nil || !token.Valid {
		ctx.Abort(http.StatusUnauthorized, fmt.Sprintf(`{"error":"Unauthorized","message":"Invalid or expired token: %s"}`, err.Error()))
		return nil
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		ctx.Abort(http.StatusUnauthorized, `{"error":"Unauthorized","message":"Invalid token claims structure"}`)
		return nil
	}

	// Validate Issuer if specified
	if p.issuer != "" {
		if iss, ok := claims["iss"].(string); !ok || iss != p.issuer {
			ctx.Abort(http.StatusUnauthorized, `{"error":"Unauthorized","message":"Invalid token issuer"}`)
			return nil
		}
	}

	// Validate Expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			ctx.Abort(http.StatusUnauthorized, `{"error":"Unauthorized","message":"Token expired"}`)
			return nil
		}
	}

	// Store claims in context metadata
	if sub, ok := claims["sub"].(string); ok {
		ctx.SetMetadata("user_id", sub)
	}
	if tenantID, ok := claims["tenant_id"].(string); ok {
		ctx.SetMetadata("tenant_id", tenantID)
	}
	if scopes, ok := claims["scopes"].([]interface{}); ok {
		ctx.SetMetadata("scopes", scopes)
	}

	return nil
}

func (p *JWTPlugin) ExecuteResponse(ctx *pipeline.PipelineContext) error {
	return nil
}

func (p *JWTPlugin) extractToken(r *http.Request) string {
	authHeader := r.Header.Get(p.headerName)
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return parts[1]
		}
	}
	return r.URL.Query().Get(p.paramName)
}
