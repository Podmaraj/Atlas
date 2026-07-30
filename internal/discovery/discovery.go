package discovery

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"

	"edgecore/internal/logger"
	"edgecore/internal/models"
)

// ServiceRegistry maintains the in-memory health-checked active upstream instances
type ServiceRegistry struct {
	mu        sync.RWMutex
	services  map[string]*models.Service
	instances map[string][]*models.ServiceInstance // ServiceID -> []ServiceInstance
	client    *http.Client
	stopChan  chan struct{}
}

// NewServiceRegistry initializes a new service discovery registry
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services:  make(map[string]*models.Service),
		instances: make(map[string][]*models.ServiceInstance),
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		stopChan: make(chan struct{}),
	}
}

// RegisterService registers a service and its initial instances
func (sr *ServiceRegistry) RegisterService(svc *models.Service, insts []*models.ServiceInstance) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	svcID := svc.ID.String()
	sr.services[svcID] = svc
	sr.instances[svcID] = insts

	logger.Info("Service registered in Data Plane registry",
		zap.String("service", svc.Name),
		zap.Int("instances", len(insts)),
	)
}

// GetHealthyInstances returns only instances currently marked healthy
func (sr *ServiceRegistry) GetHealthyInstances(serviceID string) []*models.ServiceInstance {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	insts, exists := sr.instances[serviceID]
	if !exists {
		return nil
	}

	healthy := make([]*models.ServiceInstance, 0, len(insts))
	for _, inst := range insts {
		if inst.Healthy {
			healthy = append(healthy, inst)
		}
	}
	return healthy
}

// StartHealthChecker launches background ticker pinging service targets every 10 seconds
func (sr *ServiceRegistry) StartHealthChecker(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 10 * time.Second
	}

	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-sr.stopChan:
				ticker.Stop()
				return
			case <-ticker.C:
				sr.checkAllServices(ctx)
			}
		}
	}()

	logger.Info("Active background Health Checker started", zap.Duration("interval", interval))
}

func (sr *ServiceRegistry) checkAllServices(ctx context.Context) {
	sr.mu.RLock()
	svcs := make([]*models.Service, 0, len(sr.services))
	for _, s := range sr.services {
		svcs = append(svcs, s)
	}
	sr.mu.RUnlock()

	var wg sync.WaitGroup
	for _, svc := range svcs {
		svcID := svc.ID.String()
		sr.mu.RLock()
		insts := sr.instances[svcID]
		sr.mu.RUnlock()

		for _, inst := range insts {
			wg.Add(1)
			go func(s *models.Service, instance *models.ServiceInstance) {
				defer wg.Done()
				sr.pingInstance(ctx, s, instance)
			}(svc, inst)
		}
	}
	wg.Wait()
}

func (sr *ServiceRegistry) pingInstance(ctx context.Context, svc *models.Service, inst *models.ServiceInstance) {
	checkURL := fmt.Sprintf("%s://%s:%d%s", svc.Protocol, inst.Host, inst.Port, svc.HealthCheckPath)
	req, err := http.NewRequestWithContext(ctx, "GET", checkURL, nil)
	if err != nil {
		sr.updateHealth(inst, false)
		return
	}

	resp, err := sr.client.Do(req)
	if err != nil {
		sr.updateHealth(inst, false)
		return
	}
	defer resp.Body.Close()

	healthy := resp.StatusCode >= 200 && resp.StatusCode < 400
	sr.updateHealth(inst, healthy)
}

func (sr *ServiceRegistry) updateHealth(inst *models.ServiceInstance, healthy bool) {
	now := time.Now()
	inst.Healthy = healthy
	inst.LastPing = &now

	if !healthy {
		logger.Warn("Service instance health check failed",
			zap.String("host", inst.Host),
			zap.Int("port", inst.Port),
		)
	}
}

// Stop terminates the health checker loop
func (sr *ServiceRegistry) Stop() {
	close(sr.stopChan)
}
