package loadbalancer_test

import (
	"testing"

	"github.com/google/uuid"

	"edgecore/internal/loadbalancer"
	"edgecore/internal/models"
)

func TestLoadBalancers(t *testing.T) {
	inst1 := &models.ServiceInstance{
		BaseModel: models.BaseModel{ID: uuid.New()},
		Host:      "10.0.0.1",
		Port:      8080,
		Weight:    1,
		Healthy:   true,
	}
	inst2 := &models.ServiceInstance{
		BaseModel: models.BaseModel{ID: uuid.New()},
		Host:      "10.0.0.2",
		Port:      8080,
		Weight:    3,
		Healthy:   true,
	}
	instances := []*models.ServiceInstance{inst1, inst2}

	t.Run("RoundRobin", func(t *testing.T) {
		lb := loadbalancer.GetLoadBalancer("round_robin")
		t1, err1 := lb.SelectTarget(instances, "127.0.0.1")
		t2, err2 := lb.SelectTarget(instances, "127.0.0.1")

		if err1 != nil || err2 != nil {
			t.Fatalf("expected no errors selecting targets: %v, %v", err1, err2)
		}
		if t1.ID == t2.ID {
			t.Errorf("expected round robin to alternate targets, got same target %s", t1.ID)
		}
	})

	t.Run("IPHash", func(t *testing.T) {
		lb := loadbalancer.GetLoadBalancer("ip_hash")
		t1, _ := lb.SelectTarget(instances, "192.168.1.50")
		t2, _ := lb.SelectTarget(instances, "192.168.1.50")

		if t1.ID != t2.ID {
			t.Errorf("expected IP Hash to yield consistent target for same IP, got %s and %s", t1.ID, t2.ID)
		}
	})

	t.Run("ConsistentHashing", func(t *testing.T) {
		lb := loadbalancer.GetLoadBalancer("consistent_hashing")
		t1, _ := lb.SelectTarget(instances, "user-client-10")
		t2, _ := lb.SelectTarget(instances, "user-client-10")

		if t1.ID != t2.ID {
			t.Errorf("expected Consistent Hash to pick consistent target for same key, got %s and %s", t1.ID, t2.ID)
		}
	})
}
