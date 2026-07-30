package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JSONMap helper type for JSONB columns in GORM
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = JSONMap{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("type assertion to []byte/string failed")
	}
	return json.Unmarshal(bytes, j)
}

// StringArray helper type for storing string arrays in Postgres
type StringArray []string

func (s StringArray) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	return json.Marshal(s)
}

func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = StringArray{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("type assertion to []byte/string failed")
	}
	return json.Unmarshal(bytes, s)
}

// Base struct with UUID primary keys
type BaseModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// Organization entity
type Organization struct {
	BaseModel
	Name    string   `gorm:"size:255;not null" json:"name"`
	Slug    string   `gorm:"size:255;uniqueIndex;not null" json:"slug"`
	Plan    string   `gorm:"size:50;default:'free'" json:"plan"`
	Status  string   `gorm:"size:50;default:'active'" json:"status"`
	Tenants []Tenant `gorm:"foreignKey:OrganizationID" json:"tenants,omitempty"`
	Users   []User   `gorm:"foreignKey:OrganizationID" json:"users,omitempty"`
}

// Tenant entity (Multi-tenancy isolation boundary)
type Tenant struct {
	BaseModel
	OrganizationID uuid.UUID `gorm:"type:uuid;not null;index" json:"organization_id"`
	Name           string    `gorm:"size:255;not null" json:"name"`
	Slug           string    `gorm:"size:255;uniqueIndex;not null" json:"slug"`
	Environment    string    `gorm:"size:50;default:'production'" json:"environment"`
	RateLimitQuota int       `gorm:"default:10000" json:"rate_limit_quota"`
	Status         string    `gorm:"size:50;default:'active'" json:"status"`
	Services       []Service `gorm:"foreignKey:TenantID" json:"services,omitempty"`
	APIKeys        []ApiKey  `gorm:"foreignKey:TenantID" json:"api_keys,omitempty"`
}

// User entity
type User struct {
	BaseModel
	OrganizationID uuid.UUID `gorm:"type:uuid;not null;index" json:"organization_id"`
	Username       string    `gorm:"size:100;uniqueIndex;not null" json:"username"`
	Email          string    `gorm:"size:255;uniqueIndex;not null" json:"email"`
	PasswordHash   string    `gorm:"size:255;not null" json:"-"`
	Role           string    `gorm:"size:50;default:'admin'" json:"role"`
	Status         string    `gorm:"size:50;default:'active'" json:"status"`
}

// Role entity for RBAC
type Role struct {
	BaseModel
	Name        string      `gorm:"size:100;uniqueIndex;not null" json:"name"`
	Description string      `gorm:"size:255" json:"description"`
	Permissions StringArray `gorm:"type:jsonb" json:"permissions"`
}

// ApiKey entity
type ApiKey struct {
	BaseModel
	TenantID    uuid.UUID   `gorm:"type:uuid;not null;index" json:"tenant_id"`
	UserID      *uuid.UUID  `gorm:"type:uuid;index" json:"user_id,omitempty"`
	Name        string      `gorm:"size:100;not null" json:"name"`
	KeyHash     string      `gorm:"size:255;uniqueIndex;not null" json:"-"`
	Prefix      string      `gorm:"size:20;not null" json:"prefix"`
	Scopes      StringArray `gorm:"type:jsonb" json:"scopes"`
	RateLimit   int         `gorm:"default:100" json:"rate_limit"` // requests per minute
	ExpiresAt   *time.Time  `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time  `json:"last_used_at,omitempty"`
	Status      string      `gorm:"size:50;default:'active'" json:"status"`
}

// Service entity represents an upstream backend microservice
type Service struct {
	BaseModel
	TenantID            uuid.UUID         `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Name                string            `gorm:"size:100;not null" json:"name"`
	Protocol            string            `gorm:"size:20;default:'http'" json:"protocol"` // http, https, ws, wss, grpc
	Host                string            `gorm:"size:255;not null" json:"host"`
	Port                int               `gorm:"default:80" json:"port"`
	BasePath            string            `gorm:"size:255;default:''" json:"base_path"`
	Retries             int               `gorm:"default:3" json:"retries"`
	ConnectTimeout      int               `gorm:"default:5000" json:"connect_timeout"` // ms
	ReadTimeout         int               `gorm:"default:15000" json:"read_timeout"`   // ms
	WriteTimeout        int               `gorm:"default:15000" json:"write_timeout"`  // ms
	HealthCheckPath     string            `gorm:"size:255;default:'/health'" json:"health_check_path"`
	HealthCheckInterval int               `gorm:"default:10" json:"health_check_interval"` // seconds
	LBStrategy          string            `gorm:"size:50;default:'round_robin'" json:"lb_strategy"`
	Status              string            `gorm:"size:50;default:'active'" json:"status"`
	Tags                StringArray       `gorm:"type:jsonb" json:"tags"`
	Routes              []Route           `gorm:"foreignKey:ServiceID" json:"routes,omitempty"`
	Instances           []ServiceInstance `gorm:"foreignKey:ServiceID" json:"instances,omitempty"`
}

// ServiceInstance entity represents a specific physical/virtual replica of a service
type ServiceInstance struct {
	BaseModel
	ServiceID uuid.UUID  `gorm:"type:uuid;not null;index" json:"service_id"`
	Host      string     `gorm:"size:255;not null" json:"host"`
	Port      int        `gorm:"not null" json:"port"`
	Weight    int        `gorm:"default:1" json:"weight"`
	Region    string     `gorm:"size:50;default:'us-east-1'" json:"region"`
	Healthy   bool       `gorm:"default:true;index" json:"healthy"`
	LastPing  *time.Time `json:"last_ping,omitempty"`
}

// Route entity represents an API ingress routing rule
type Route struct {
	BaseModel
	ServiceID     uuid.UUID   `gorm:"type:uuid;not null;index" json:"service_id"`
	TenantID      uuid.UUID   `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Name          string      `gorm:"size:100;not null" json:"name"`
	Paths         StringArray `gorm:"type:jsonb;not null" json:"paths"`   // e.g. ["/api/v1/users", "/api/v1/users/*"]
	Methods       StringArray `gorm:"type:jsonb" json:"methods"`          // e.g. ["GET", "POST"]
	Hosts         StringArray `gorm:"type:jsonb" json:"hosts"`            // e.g. ["api.company.com"]
	Headers       JSONMap     `gorm:"type:jsonb" json:"headers"`          // required matching headers
	QueryParams   JSONMap     `gorm:"type:jsonb" json:"query_params"`     // required matching query parameters
	StripPath     bool        `gorm:"default:true" json:"strip_path"`
	PreserveHost  bool        `gorm:"default:false" json:"preserve_host"`
	Priority      int         `gorm:"default:0;index" json:"priority"`
	Status        string      `gorm:"size:50;default:'active'" json:"status"`
	Plugins       []Plugin    `gorm:"foreignKey:TargetID" json:"plugins,omitempty"`
}

// Plugin entity defines dynamic gateway middleware configurations
type Plugin struct {
	BaseModel
	Scope    string    `gorm:"size:20;not null;index" json:"scope"` // "global", "service", "route"
	TargetID *uuid.UUID `gorm:"type:uuid;index" json:"target_id,omitempty"`
	Name     string    `gorm:"size:100;not null;index" json:"name"` // "rate-limit", "jwt", "cors", "transform", etc.
	Enabled  bool      `gorm:"default:true;index" json:"enabled"`
	Config   JSONMap   `gorm:"type:jsonb" json:"config"`
	Priority int       `gorm:"default:0" json:"priority"`
}

// GatewayNode represents an active cluster node running the Data Plane
type GatewayNode struct {
	BaseModel
	NodeID            string    `gorm:"size:100;uniqueIndex;not null" json:"node_id"`
	Hostname          string    `gorm:"size:255;not null" json:"hostname"`
	IPAddress         string    `gorm:"size:50;not null" json:"ip_address"`
	Version           string    `gorm:"size:50;not null" json:"version"`
	Status            string    `gorm:"size:50;default:'online'" json:"status"`
	ActiveConnections int64     `gorm:"default:0" json:"active_connections"`
	CPUUsage          float64   `gorm:"default:0" json:"cpu_usage"`
	MemoryUsage       float64   `gorm:"default:0" json:"memory_usage"`
	LastHeartbeat     time.Time `gorm:"index" json:"last_heartbeat"`
}

// Certificate entity for TLS termination
type Certificate struct {
	BaseModel
	Domain    string    `gorm:"size:255;uniqueIndex;not null" json:"domain"`
	CertPEM   string    `gorm:"type:text;not null" json:"cert_pem"`
	KeyPEM    string    `gorm:"type:text;not null" json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	AutoRenew bool      `gorm:"default:true" json:"auto_renew"`
}

// AuditLog records Control Plane management operations
type AuditLog struct {
	BaseModel
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	TenantID  uuid.UUID `gorm:"type:uuid;index" json:"tenant_id"`
	Action    string    `gorm:"size:100;not null;index" json:"action"`
	Resource  string    `gorm:"size:100;not null" json:"resource"`
	Payload   JSONMap   `gorm:"type:jsonb" json:"payload"`
	IPAddress string    `gorm:"size:50" json:"ip_address"`
	UserAgent string    `gorm:"size:255" json:"user_agent"`
}

// RateLimitRule defines global/granular rate limit rules
type RateLimitRule struct {
	BaseModel
	Scope         string     `gorm:"size:20;not null;index" json:"scope"` // global, route, tenant, ip
	TargetID      *uuid.UUID `gorm:"type:uuid;index" json:"target_id,omitempty"`
	Limit         int        `gorm:"not null" json:"limit"`
	PeriodSeconds int        `gorm:"default:60" json:"period_seconds"`
	BlockSeconds  int        `gorm:"default:0" json:"block_seconds"`
}
