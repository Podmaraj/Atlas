package loadbalancer

import (
	"crypto/md5"
	"errors"
	"fmt"
	"hash/fnv"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"

	"edgecore/internal/models"
)

var ErrNoHealthyInstances = errors.New("no healthy upstream service instances available")

// LoadBalancer interface implemented by all load balancing algorithms
type LoadBalancer interface {
	SelectTarget(instances []*models.ServiceInstance, clientIP string) (*models.ServiceInstance, error)
	Name() string
}

// 1. Round Robin Strategy
type RoundRobinLB struct {
	counter uint64
}

func NewRoundRobinLB() *RoundRobinLB {
	return &RoundRobinLB{}
}

func (lb *RoundRobinLB) Name() string { return "round_robin" }

func (lb *RoundRobinLB) SelectTarget(instances []*models.ServiceInstance, clientIP string) (*models.ServiceInstance, error) {
	healthy := filterHealthy(instances)
	if len(healthy) == 0 {
		return nil, ErrNoHealthyInstances
	}
	idx := atomic.AddUint64(&lb.counter, 1) - 1
	return healthy[idx%uint64(len(healthy))], nil
}

// 2. Least Connections Strategy
type LeastConnLB struct {
	mu     sync.RWMutex
	conns  map[string]*int64 // Instance ID -> active connections count
}

func NewLeastConnLB() *LeastConnLB {
	return &LeastConnLB{
		conns: make(map[string]*int64),
	}
}

func (lb *LeastConnLB) Name() string { return "least_connections" }

func (lb *LeastConnLB) SelectTarget(instances []*models.ServiceInstance, clientIP string) (*models.ServiceInstance, error) {
	healthy := filterHealthy(instances)
	if len(healthy) == 0 {
		return nil, ErrNoHealthyInstances
	}

	lb.mu.Lock()
	var selected *models.ServiceInstance
	var minConns int64 = 1<<63 - 1

	for _, inst := range healthy {
		instID := inst.ID.String()
		connPtr, exists := lb.conns[instID]
		if !exists {
			var zero int64 = 0
			connPtr = &zero
			lb.conns[instID] = connPtr
		}

		current := atomic.LoadInt64(connPtr)
		if current < minConns {
			minConns = current
			selected = inst
		}
	}
	lb.mu.Unlock()

	if selected != nil {
		lb.IncConn(selected.ID.String())
	}

	return selected, nil
}

func (lb *LeastConnLB) IncConn(instanceID string) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	if ptr, ok := lb.conns[instanceID]; ok {
		atomic.AddInt64(ptr, 1)
	}
}

func (lb *LeastConnLB) DecConn(instanceID string) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	if ptr, ok := lb.conns[instanceID]; ok {
		atomic.AddInt64(ptr, -1)
	}
}

// 3. Random Strategy
type RandomLB struct{}

func NewRandomLB() *RandomLB { return &RandomLB{} }

func (lb *RandomLB) Name() string { return "random" }

func (lb *RandomLB) SelectTarget(instances []*models.ServiceInstance, clientIP string) (*models.ServiceInstance, error) {
	healthy := filterHealthy(instances)
	if len(healthy) == 0 {
		return nil, ErrNoHealthyInstances
	}
	return healthy[rand.Intn(len(healthy))], nil
}

// 4. Weighted Round Robin Strategy
type WeightedLB struct {
	counter uint64
}

func NewWeightedLB() *WeightedLB { return &WeightedLB{} }

func (lb *WeightedLB) Name() string { return "weighted" }

func (lb *WeightedLB) SelectTarget(instances []*models.ServiceInstance, clientIP string) (*models.ServiceInstance, error) {
	healthy := filterHealthy(instances)
	if len(healthy) == 0 {
		return nil, ErrNoHealthyInstances
	}

	var totalWeight int
	for _, inst := range healthy {
		w := inst.Weight
		if w <= 0 {
			w = 1
		}
		totalWeight += w
	}

	idx := atomic.AddUint64(&lb.counter, 1)
	val := int(idx % uint64(totalWeight))

	for _, inst := range healthy {
		w := inst.Weight
		if w <= 0 {
			w = 1
		}
		if val < w {
			return inst, nil
		}
		val -= w
	}

	return healthy[0], nil
}

// 5. IP Hash Strategy
type IPHashLB struct{}

func NewIPHashLB() *IPHashLB { return &IPHashLB{} }

func (lb *IPHashLB) Name() string { return "ip_hash" }

func (lb *IPHashLB) SelectTarget(instances []*models.ServiceInstance, clientIP string) (*models.ServiceInstance, error) {
	healthy := filterHealthy(instances)
	if len(healthy) == 0 {
		return nil, ErrNoHealthyInstances
	}

	h := fnv.New32a()
	_, _ = h.Write([]byte(clientIP))
	hashVal := h.Sum32()

	return healthy[hashVal%uint32(len(healthy))], nil
}

// 6. Consistent Hashing Strategy (Ketama Hash Ring)
type ConsistentHashLB struct {
	mu       sync.RWMutex
	vNodes   int
	ring     []uint32
	nodeMap  map[uint32]*models.ServiceInstance
}

func NewConsistentHashLB(vNodes int) *ConsistentHashLB {
	if vNodes <= 0 {
		vNodes = 100
	}
	return &ConsistentHashLB{
		vNodes:  vNodes,
		nodeMap: make(map[uint32]*models.ServiceInstance),
	}
}

func (lb *ConsistentHashLB) Name() string { return "consistent_hashing" }

func (lb *ConsistentHashLB) SelectTarget(instances []*models.ServiceInstance, clientIP string) (*models.ServiceInstance, error) {
	healthy := filterHealthy(instances)
	if len(healthy) == 0 {
		return nil, ErrNoHealthyInstances
	}

	lb.mu.Lock()
	lb.ring = nil
	lb.nodeMap = make(map[uint32]*models.ServiceInstance)

	for _, inst := range healthy {
		for i := 0; i < lb.vNodes; i++ {
			vNodeKey := fmt.Sprintf("%s:%d#%d", inst.Host, inst.Port, i)
			hash := hashKey(vNodeKey)
			lb.ring = append(lb.ring, hash)
			lb.nodeMap[hash] = inst
		}
	}
	sort.Slice(lb.ring, func(i, j int) bool { return lb.ring[i] < lb.ring[j] })
	lb.mu.Unlock()

	clientHash := hashKey(clientIP)
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	idx := sort.Search(len(lb.ring), func(i int) bool {
		return lb.ring[i] >= clientHash
	})

	if idx == len(lb.ring) {
		idx = 0
	}

	return lb.nodeMap[lb.ring[idx]], nil
}

func hashKey(key string) uint32 {
	hasher := md5.New()
	_, _ = hasher.Write([]byte(key))
	b := hasher.Sum(nil)
	return (uint32(b[3]) << 24) | (uint32(b[2]) << 16) | (uint32(b[1]) << 8) | uint32(b[0])
}

func filterHealthy(instances []*models.ServiceInstance) []*models.ServiceInstance {
	healthy := make([]*models.ServiceInstance, 0, len(instances))
	for _, inst := range instances {
		if inst.Healthy {
			healthy = append(healthy, inst)
		}
	}
	return healthy
}

// Factory function to get load balancer implementation by name
func GetLoadBalancer(strategy string) LoadBalancer {
	switch strategy {
	case "least_connections":
		return NewLeastConnLB()
	case "random":
		return NewRandomLB()
	case "weighted":
		return NewWeightedLB()
	case "ip_hash":
		return NewIPHashLB()
	case "consistent_hashing":
		return NewConsistentHashLB(100)
	default:
		return NewRoundRobinLB()
	}
}
