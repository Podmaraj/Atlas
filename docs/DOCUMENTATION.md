# EdgeCore Enterprise API Gateway Platform - Complete Technical Documentation

Welcome to the comprehensive technical documentation for **EdgeCore**, an enterprise-grade API Gateway and Control Plane platform built for high-throughput, low-latency API traffic management, zero-downtime dynamic reconfigurations, and centralized microservices governance.

---

## Table of Contents
1. [Architecture Overview](#architecture-overview)
2. [Component Breakdown](#component-breakdown)
   - [Data Plane Gateway Node](#1-data-plane-gateway-appsgateway)
   - [Control Plane Management Service](#2-control-plane-management-appscontrol-plane)
   - [Admin Dashboard](#3-admin-dashboard-appsdashboard)
3. [Control Plane REST API Reference](#control-plane-rest-api-reference)
   - [Authentication](#authentication)
   - [Service Management](#service-management)
   - [Route Ingress Management](#route-ingress-management)
   - [Plugin Hub Management](#plugin-hub-management)
   - [API Key Management](#api-key-management)
   - [Cluster Nodes & Heartbeats](#cluster-nodes--heartbeats)
4. [Gateway Plugin Pipeline Execution](#gateway-plugin-pipeline-execution)
   - [Supported Middleware Plugins](#supported-middleware-plugins)
5. [Ingress Routing Engine](#ingress-routing-engine)
6. [Load Balancing Strategies](#load-balancing-strategies)
7. [Zero-Downtime Dynamic Reloading](#zero-downtime-dynamic-reloading)
8. [Observability & Analytics](#observability--analytics)
9. [Deployment & Environment Configuration](#deployment--environment-configuration)
10. [Verification & Testing](#verification--testing)

---

## Architecture Overview

```
                      Client / Internet Traffic
                                  │
                                  ▼
                     ┌──────────────────────────┐
                     │   Edge Load Balancer     │
                     └────────────┬─────────────┘
                                  │
                  ┌───────────────┴───────────────┐
                  │                               │
       ┌──────────────────────┐        ┌──────────────────────┐
       │ Data Plane Node #1   │        │ Data Plane Node #2   │
       │ (Port 8080)          │        │ (Port 8082)          │
       └──────────┬───────────┘        └──────────┬───────────┘
                  │                               │
                  └───────────────┬───────────────┘
                                  │
                      Redis Pub/Sub Config Sync (`edgecore:config:events`)
                                  │
                   ┌──────────────┴──────────────┐
                   │    Go Control Plane API     │
                   │    (Port 8081 - Fiber & RBAC)│
                   └──────────────┬──────────────┘
                                  │
                         PostgreSQL Database
                                  │
                   ┌──────────────┴──────────────┐
                   │   Next.js Admin Dashboard   │
                   │   (Port 3000 - App Router)   │
                   └─────────────────────────────┘
```

---

## Component Breakdown

### 1. Data Plane Gateway (`apps/gateway`)
The **Data Plane** handles all incoming client HTTP/WebSocket traffic.
- **High Performance Reverse Proxy**: Powered by standard library `net/http/httputil` and custom connection pooling.
- **Sliding-Window Distributed Rate Limiting**: Redis ZSET-backed rate limiter with non-blocking memory fallback.
- **Circuit Breaker**: Microsecond failure threshold monitoring preventing cascading upstream backend failures.
- **WebSocket Pass-Through**: Full TCP hijacking support for real-time WebSockets (`ws://`, `wss://`) and Server-Sent Events (SSE).
- **Asynchronous Access Logging**: Non-blocking worker queue processing HTTP latency and access metrics without impacting request response time.

### 2. Control Plane Management (`apps/control-plane`)
The **Control Plane** provides administrative REST APIs to configure services, routing rules, plugins, and API keys.
- **Go Fiber REST API**: Lightweight, ultra-fast web framework.
- **Casbin RBAC Engine**: Fine-grained role-based authorization enforcing permissions (`superadmin`, `admin`, `viewer`).
- **GORM PostgreSQL Storage**: Production persistence layer with automatic schema migrations.
- **Redis Pub/Sub Event Dispatcher**: Pushes live configuration mutations instantly to Data Plane nodes for zero-downtime reconfigurations.

### 3. Admin Dashboard (`apps/dashboard`)
A modern Next.js 15 web interface for API platform operators:
- Built with **React 19**, **Tailwind CSS**, **Shadcn UI**, **TanStack Query**, and **Recharts**.
- Visual route rule builder, active node monitor, plugin toggles, and live traffic analytics.

---

## Control Plane REST API Reference

Base URL: `http://localhost:8081/api/v1`

### Authentication
All management endpoints (except `/auth/login` and `/health`) require standard JWT Bearer Authorization header:
```http
Authorization: Bearer <JWT_TOKEN>
```

#### `POST /api/v1/auth/login`
Authenticate admin user and receive JWT session token.
- **Request Body**:
  ```json
  {
    "username": "admin",
    "password": "password123"
  }
  ```
- **Response (200 OK)**:
  ```json
  {
    "token": "eyJhbGciOiJIUzI1NiIsIn...",
    "user": {
      "username": "admin",
      "role": "superadmin"
    }
  }
  ```

---

### Service Management

#### `GET /api/v1/services`
Fetch all registered microservices.

#### `GET /api/v1/services/:id`
Fetch details for a single microservice by UUID.

#### `POST /api/v1/services`
Register a new upstream backend service.
- **Request Body**:
  ```json
  {
    "name": "catalog-service",
    "protocol": "http",
    "host": "catalog-api.internal",
    "port": 8080,
    "lb_strategy": "round_robin",
    "health_check_path": "/health"
  }
  ```
- **Response (201 Created)**

#### `PUT /api/v1/services/:id`
Update an existing service configuration.

#### `DELETE /api/v1/services/:id`
Delete an upstream service registration.

---

### Route Ingress Management

#### `GET /api/v1/routes`
List all active gateway routing rules.

#### `POST /api/v1/routes`
Create a new API ingress routing rule.
- **Request Body**:
  ```json
  {
    "service_id": "3b2e7a10-8b9a-4c12-9e80-1a2b3c4d5e6f",
    "name": "catalog-products-route",
    "paths": ["/api/v1/products", "/api/v1/products/*"],
    "methods": ["GET", "POST"],
    "strip_path": true,
    "priority": 10
  }
  ```
- **Response (201 Created)**

#### `PUT /api/v1/routes/:id`
Update an existing route definition.

#### `DELETE /api/v1/routes/:id`
Delete a routing rule.

---

### Plugin Hub Management

#### `GET /api/v1/plugins`
List all installed plugins.

#### `POST /api/v1/plugins`
Attach a plugin to a global, service, or route scope.
- **Request Body**:
  ```json
  {
    "scope": "route",
    "target_id": "3b2e7a10-8b9a-4c12-9e80-1a2b3c4d5e6f",
    "name": "rate-limit",
    "enabled": true,
    "config": {
      "limit": 500,
      "window_seconds": 60,
      "key_strategy": "ip"
    },
    "priority": 100
  }
  ```

---

### API Key Management

#### `GET /api/v1/api-keys`
List active client API keys.

#### `POST /api/v1/api-keys`
Generate a new API key.
- **Request Body**:
  ```json
  {
    "tenant_id": "11111111-2222-3333-4444-555555555555",
    "name": "Mobile App Production Key",
    "prefix": "edge_",
    "rate_limit": 1000,
    "scopes": ["read", "write"]
  }
  ```

---

### Cluster Nodes & Heartbeats

#### `GET /api/v1/nodes`
Get current status of all cluster Gateway Data Plane nodes.

#### `POST /api/v1/nodes/heartbeat`
Report node health metrics (CPU, Memory, Active Connections).

---

## Gateway Plugin Pipeline Execution

The Data Plane executes middleware plugins sequentially based on assigned **priority** (highest priority runs first).

```
Client Request
      │
      ▼
┌──────────────────────────────────────┐
│  Rate Limiting Plugin (priority: 100)│ ---> Exceeded? Abort 429
└─────────────────┬────────────────────┘
                  │
┌──────────────────────────────────────┐
│  JWT Auth Plugin (priority: 90)      │ ---> Invalid? Abort 401
└─────────────────┬────────────────────┘
                  │
┌──────────────────────────────────────┐
│  CORS & Security Headers (priority: 80)
└─────────────────┬────────────────────┘
                  │
┌──────────────────────────────────────┐
│  Request Transformer (priority: 50)  │
└─────────────────┬────────────────────┘
                  │
                  ▼
         Forward to Upstream Service Target
```

### Supported Middleware Plugins
1. **`rate-limit`**: Enforces IP, Header, or API-Key sliding-window request limits.
2. **`jwt`**: Validates Authorization Bearer claims, signature, and expiration.
3. **`api-key`**: Validates `X-API-Key` headers or `apikey` query parameters.
4. **`oauth2-introspect`**: Performs RFC 7662 OAuth2 token introspection with Redis caching.
5. **`cors-security`**: Enforces strict CORS headers (`Access-Control-Allow-Origin`) and security policies (`X-Frame-Options`, `X-Content-Type-Options`).
6. **`transform`**: Modifies incoming request headers/query params or outgoing response headers.
7. **`cache`**: Caches HTTP GET responses in Redis for specified TTLs.

---

## Ingress Routing Engine

The routing engine supports composite matching rules ordered by priority:
- **Exact Path Match**: `/api/v1/checkout`
- **Prefix Path Match**: `/api/v1/catalog/*`
- **Wildcard Match**: `/products/*/reviews`
- **Regex Path Match**: `^/users/[0-9]+$`
- **Host Matching**: `api.company.com`
- **Header Rule Matching**: `X-Client-Version: 2.0`
- **Query Parameter Matching**: `version=beta`

---

## Load Balancing Strategies

Supported algorithms for upstream instance selection:
1. **Round Robin (`round_robin`)**: Uniform sequential request distribution.
2. **Least Connections (`least_connections`)**: Directs traffic to instance with lowest active HTTP connections.
3. **Random (`random`)**: Cryptographically random target selection.
4. **Weighted Round Robin (`weighted`)**: Distributes traffic proportionally according to instance target weights.
5. **IP Hash (`ip_hash`)**: Hashes client IP for sticky session persistence.
6. **Consistent Hashing (`consistent_hashing`)**: Ketama virtual-node hashing ring minimizing redistribution during scaling events.

---

## Zero-Downtime Dynamic Reloading

When an administrative user creates, updates, or deletes a route, service, or plugin in the Control Plane:
1. Control Plane persists change in PostgreSQL DB.
2. Control Plane publishes a payload to Redis channel `edgecore:config:events`.
3. All running Data Plane instances receive the PubSub notification.
4. Data Plane nodes dynamically reload their in-memory compiled route table (`gw.router`) without dropping active connections.

---

## Verification & Testing

### Running Go Backend Unit Tests
```bash
go test -v ./...
```

### Running Next.js Dashboard Build
```bash
npm --prefix apps/dashboard run build
```
