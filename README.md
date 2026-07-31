# EdgeCore Enterprise API Gateway Platform

[![Go Version](https://img.shields.io/badge/Go-1.24%2B-00ADD8?style=flat&logo=go)](https://golang.org)
[![Next.js Version](https://img.shields.io/badge/Next.js-15.2-black?style=flat&logo=next.js)](https://nextjs.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**EdgeCore** is a production-grade, enterprise API Gateway platform built with **Go** (Data Plane & Control Plane) and **Next.js 15** (Admin Dashboard).

---

## 📖 Complete Documentation
Detailed architecture, REST API reference, plugin pipeline execution, and routing guide can be found in the [Full Technical Documentation](docs/DOCUMENTATION.md).

---

## Architecture

```
                Internet Traffic
                       │
                       ▼
           ┌──────────────────────┐
           │   Load Balancer       │
           └──────────┬───────────┘
                      │
         ┌────────────┴────────────┐
         │                         │
  ┌──────────────┐          ┌──────────────┐
  │ Gateway #1   │          │ Gateway #2   │
  │ Data Plane   │          │ Data Plane   │
  └──────┬───────┘          └──────┬───────┘
         │                         │
         └────────────┬────────────┘
                      │
              Redis Config Sync (Pub/Sub)
                      │
       ┌──────────────┴──────────────┐
       │     Go Control Plane        │
       └──────────────┬──────────────┘
                      │
               PostgreSQL DB
                      │
              Next.js Dashboard
```

---

## Key Features

- ⚡ **High-Throughput Data Plane (`apps/gateway`)**: Low-latency HTTP reverse proxy with tuned connection pooling based on `net/http/httputil`.
- 🔄 **Zero-Downtime Hot Reloading**: Dynamic configuration updates dispatched across Redis PubSub (`edgecore:config:events`) to all gateway nodes without process restarts.
- 🛡️ **Extensible Plugin Pipeline**: Distributed Redis sliding-window rate limiting, JWT validation, OAuth2 introspection, API key validation, CORS & Security headers, request/response transformers, and response caching.
- 🚦 **Dynamic Ingress Router**: Support for Exact, Prefix, Wildcard, Regex, Header, Host, Query Parameter, and Priority routing rules.
- ⚖️ **Advanced Load Balancing**: Round Robin, Least Connections, Random, Weighted Round Robin, IP Hash, and Consistent Hashing (Ketama ring).
- 🔐 **Control Plane Security (`apps/control-plane`)**: Management REST API built on Go Fiber with JWT authentication and Casbin RBAC role enforcement (`superadmin`, `admin`, `viewer`).
- 📊 **Real-time Observability**: Prometheus metrics exporter (`/metrics` on port 9090), OpenTelemetry W3C trace context propagation, and non-blocking asynchronous access log analytics collector.
- 💻 **Admin Dashboard (`apps/dashboard`)**: Modern web dashboard built on Next.js 15 App Router, React 19, Tailwind CSS, and TanStack Query.

---

## Getting Started

### Prerequisites
- Go 1.24+
- Node.js 20+
- Docker & Docker Compose

### Quick Start with Docker Compose
```bash
docker-compose -f deployments/docker-compose.yml up --build
```

Services will start at:
- **Data Plane Gateway Node 1**: `http://localhost:8080`
- **Data Plane Gateway Node 2**: `http://localhost:8082`
- **Prometheus Metrics**: `http://localhost:9090/metrics`
- **Control Plane API**: `http://localhost:8081`
- **Next.js Admin Dashboard**: `http://localhost:3000`

---

## Running Unit Tests

Run unit tests across all Go packages:
```bash
go test -v ./...
```

Verify Next.js Admin Dashboard production build:
```bash
npm --prefix apps/dashboard run build
```

---

## License
[MIT](LICENSE)
