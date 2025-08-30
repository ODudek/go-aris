package cluster

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Node represents a cluster node
type Node struct {
	ID       string    `json:"id"`
	Address  string    `json:"address"`
	Port     string    `json:"port"`
	Status   string    `json:"status"` // active, inactive, joining
	LastSeen time.Time `json:"last_seen"`
}

// Discovery handles cluster node discovery
type Discovery struct {
	mu       sync.RWMutex
	nodes    map[string]*Node
	selfNode *Node
	client   *http.Client
}

// NewDiscovery creates a new discovery instance
func NewDiscovery(selfID, address, port string) *Discovery {
	return &Discovery{
		nodes: make(map[string]*Node),
		selfNode: &Node{
			ID:       selfID,
			Address:  address,
			Port:     port,
			Status:   "active",
			LastSeen: time.Now(),
		},
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Start begins the discovery process
func (d *Discovery) Start(ctx context.Context) error {
	// Discover nodes via different methods
	go d.discoverViaKubernetes(ctx)
	go d.discoverViaEnvironment(ctx)
	go d.startHealthCheck(ctx)

	return nil
}

// discoverViaKubernetes uses Kubernetes service discovery
func (d *Discovery) discoverViaKubernetes(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.k8sDiscovery()
		}
	}
}

// k8sDiscovery performs Kubernetes-based service discovery
func (d *Discovery) k8sDiscovery() {
	// Use headless service for peer discovery
	serviceName := os.Getenv("K8S_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "aris-headless"
	}

	namespace := os.Getenv("POD_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	// DNS lookup for headless service
	fqdn := fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, namespace)
	ips, err := net.LookupIP(fqdn)
	if err != nil {
		log.Printf("Failed to lookup service %s: %v", fqdn, err)
		return
	}

	for _, ip := range ips {
		// Skip self
		if ip.String() == d.selfNode.Address {
			continue
		}

		nodeID := fmt.Sprintf("node-%s", strings.ReplaceAll(ip.String(), ".", "-"))
		d.addOrUpdateNode(&Node{
			ID:       nodeID,
			Address:  ip.String(),
			Port:     d.selfNode.Port,
			Status:   "active",
			LastSeen: time.Now(),
		})
	}
}

// discoverViaEnvironment uses environment variables for discovery
func (d *Discovery) discoverViaEnvironment(ctx context.Context) {
	peers := os.Getenv("CLUSTER_PEERS")
	if peers == "" {
		return
	}

	peerList := strings.Split(peers, ",")
	for _, peer := range peerList {
		parts := strings.Split(strings.TrimSpace(peer), ":")
		if len(parts) == 2 {
			nodeID := fmt.Sprintf("node-%s", strings.ReplaceAll(parts[0], ".", "-"))
			d.addOrUpdateNode(&Node{
				ID:       nodeID,
				Address:  parts[0],
				Port:     ":" + parts[1],
				Status:   "active",
				LastSeen: time.Now(),
			})
		}
	}
}

// startHealthCheck periodically checks node health
func (d *Discovery) startHealthCheck(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.healthCheckNodes()
		}
	}
}

// healthCheckNodes checks health of all known nodes
func (d *Discovery) healthCheckNodes() {
	d.mu.RLock()
	nodes := make([]*Node, 0, len(d.nodes))
	for _, node := range d.nodes {
		nodes = append(nodes, node)
	}
	d.mu.RUnlock()

	for _, node := range nodes {
		go d.checkNodeHealth(node)
	}
}

// checkNodeHealth checks if a specific node is healthy
func (d *Discovery) checkNodeHealth(node *Node) {
	url := fmt.Sprintf("http://%s%s/health", node.Address, node.Port)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		d.markNodeInactive(node.ID)
		return
	}

	resp, err := d.client.Do(req)
	if err != nil {
		d.markNodeInactive(node.ID)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		d.markNodeActive(node.ID)
	} else {
		d.markNodeInactive(node.ID)
	}
}

// addOrUpdateNode adds or updates a node in the cluster
func (d *Discovery) addOrUpdateNode(node *Node) {
	d.mu.Lock()
	defer d.mu.Unlock()

	existing, exists := d.nodes[node.ID]
	if exists {
		existing.LastSeen = time.Now()
		existing.Status = node.Status
	} else {
		d.nodes[node.ID] = node
		log.Printf("Discovered new node: %s at %s%s", node.ID, node.Address, node.Port)
	}
}

// markNodeActive marks a node as active
func (d *Discovery) markNodeActive(nodeID string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if node, exists := d.nodes[nodeID]; exists {
		if node.Status != "active" {
			log.Printf("Node %s is now active", nodeID)
		}
		node.Status = "active"
		node.LastSeen = time.Now()
	}
}

// markNodeInactive marks a node as inactive
func (d *Discovery) markNodeInactive(nodeID string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if node, exists := d.nodes[nodeID]; exists {
		if node.Status != "inactive" {
			log.Printf("Node %s is now inactive", nodeID)
		}
		node.Status = "inactive"
	}
}

// GetActiveNodes returns all active nodes
func (d *Discovery) GetActiveNodes() []*Node {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var active []*Node
	for _, node := range d.nodes {
		if node.Status == "active" {
			active = append(active, node)
		}
	}
	return active
}

// GetAllNodes returns all known nodes
func (d *Discovery) GetAllNodes() []*Node {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var all []*Node
	for _, node := range d.nodes {
		all = append(all, node)
	}
	return all
}

// GetSelfNode returns information about this node
func (d *Discovery) GetSelfNode() *Node {
	return d.selfNode
}

// RemoveStaleNodes removes nodes that haven't been seen for a while
func (d *Discovery) RemoveStaleNodes() {
	d.mu.Lock()
	defer d.mu.Unlock()

	staleThreshold := time.Now().Add(-5 * time.Minute)
	for id, node := range d.nodes {
		if node.LastSeen.Before(staleThreshold) {
			log.Printf("Removing stale node: %s", id)
			delete(d.nodes, id)
		}
	}
}

// GetNodeCount returns the number of active nodes
func (d *Discovery) GetNodeCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()

	count := 0
	for _, node := range d.nodes {
		if node.Status == "active" {
			count++
		}
	}
	return count + 1 // +1 for self
}

// IsClusterHealthy checks if the cluster has minimum required nodes
func (d *Discovery) IsClusterHealthy() bool {
	minNodes := 1
	if env := os.Getenv("MIN_CLUSTER_SIZE"); env != "" {
		fmt.Sscanf(env, "%d", &minNodes)
	}

	return d.GetNodeCount() >= minNodes
}
