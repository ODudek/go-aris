package cluster

import (
	"context"
	"testing"
	"time"
)

func TestNewDiscovery(t *testing.T) {
	discovery := NewDiscovery("test-node", "127.0.0.1", ":8089")

	if discovery == nil {
		t.Fatal("Expected discovery instance, got nil")
	}

	if discovery.selfNode.ID != "test-node" {
		t.Errorf("Expected node ID 'test-node', got '%s'", discovery.selfNode.ID)
	}

	if discovery.selfNode.Address != "127.0.0.1" {
		t.Errorf("Expected address '127.0.0.1', got '%s'", discovery.selfNode.Address)
	}

	if discovery.selfNode.Port != ":8089" {
		t.Errorf("Expected port ':8089', got '%s'", discovery.selfNode.Port)
	}

	if discovery.selfNode.Status != "active" {
		t.Errorf("Expected status 'active', got '%s'", discovery.selfNode.Status)
	}
}

func TestAddOrUpdateNode(t *testing.T) {
	discovery := NewDiscovery("test-node", "127.0.0.1", ":8089")

	node := &Node{
		ID:       "peer-node",
		Address:  "127.0.0.2",
		Port:     ":8089",
		Status:   "active",
		LastSeen: time.Now(),
	}

	discovery.addOrUpdateNode(node)

	nodes := discovery.GetAllNodes()
	if len(nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(nodes))
	}

	if nodes[0].ID != "peer-node" {
		t.Errorf("Expected node ID 'peer-node', got '%s'", nodes[0].ID)
	}
}

func TestGetActiveNodes(t *testing.T) {
	discovery := NewDiscovery("test-node", "127.0.0.1", ":8089")

	// Add active node
	activeNode := &Node{
		ID:       "active-node",
		Address:  "127.0.0.2",
		Port:     ":8089",
		Status:   "active",
		LastSeen: time.Now(),
	}
	discovery.addOrUpdateNode(activeNode)

	// Add inactive node
	inactiveNode := &Node{
		ID:       "inactive-node",
		Address:  "127.0.0.3",
		Port:     ":8089",
		Status:   "inactive",
		LastSeen: time.Now(),
	}
	discovery.addOrUpdateNode(inactiveNode)

	activeNodes := discovery.GetActiveNodes()
	if len(activeNodes) != 1 {
		t.Errorf("Expected 1 active node, got %d", len(activeNodes))
	}

	if activeNodes[0].ID != "active-node" {
		t.Errorf("Expected active node ID 'active-node', got '%s'", activeNodes[0].ID)
	}
}

func TestMarkNodeActive(t *testing.T) {
	discovery := NewDiscovery("test-node", "127.0.0.1", ":8089")

	node := &Node{
		ID:       "peer-node",
		Address:  "127.0.0.2",
		Port:     ":8089",
		Status:   "inactive",
		LastSeen: time.Now(),
	}
	discovery.addOrUpdateNode(node)

	discovery.markNodeActive("peer-node")

	activeNodes := discovery.GetActiveNodes()
	if len(activeNodes) != 1 {
		t.Errorf("Expected 1 active node, got %d", len(activeNodes))
	}

	if activeNodes[0].Status != "active" {
		t.Errorf("Expected node status 'active', got '%s'", activeNodes[0].Status)
	}
}

func TestMarkNodeInactive(t *testing.T) {
	discovery := NewDiscovery("test-node", "127.0.0.1", ":8089")

	node := &Node{
		ID:       "peer-node",
		Address:  "127.0.0.2",
		Port:     ":8089",
		Status:   "active",
		LastSeen: time.Now(),
	}
	discovery.addOrUpdateNode(node)

	discovery.markNodeInactive("peer-node")

	activeNodes := discovery.GetActiveNodes()
	if len(activeNodes) != 0 {
		t.Errorf("Expected 0 active nodes, got %d", len(activeNodes))
	}

	allNodes := discovery.GetAllNodes()
	if len(allNodes) != 1 {
		t.Errorf("Expected 1 total node, got %d", len(allNodes))
	}

	if allNodes[0].Status != "inactive" {
		t.Errorf("Expected node status 'inactive', got '%s'", allNodes[0].Status)
	}
}

func TestGetNodeCount(t *testing.T) {
	discovery := NewDiscovery("test-node", "127.0.0.1", ":8089")

	// Initially should have 1 node (self)
	count := discovery.GetNodeCount()
	if count != 1 {
		t.Errorf("Expected node count 1, got %d", count)
	}

	// Add another active node
	node := &Node{
		ID:       "peer-node",
		Address:  "127.0.0.2",
		Port:     ":8089",
		Status:   "active",
		LastSeen: time.Now(),
	}
	discovery.addOrUpdateNode(node)

	count = discovery.GetNodeCount()
	if count != 2 {
		t.Errorf("Expected node count 2, got %d", count)
	}

	// Add inactive node - shouldn't count
	inactiveNode := &Node{
		ID:       "inactive-node",
		Address:  "127.0.0.3",
		Port:     ":8089",
		Status:   "inactive",
		LastSeen: time.Now(),
	}
	discovery.addOrUpdateNode(inactiveNode)

	count = discovery.GetNodeCount()
	if count != 2 {
		t.Errorf("Expected node count 2 (inactive shouldn't count), got %d", count)
	}
}

func TestRemoveStaleNodes(t *testing.T) {
	discovery := NewDiscovery("test-node", "127.0.0.1", ":8089")

	// Add a stale node (old timestamp)
	staleNode := &Node{
		ID:       "stale-node",
		Address:  "127.0.0.2",
		Port:     ":8089",
		Status:   "active",
		LastSeen: time.Now().Add(-10 * time.Minute), // 10 minutes ago
	}
	discovery.addOrUpdateNode(staleNode)

	// Add a fresh node
	freshNode := &Node{
		ID:       "fresh-node",
		Address:  "127.0.0.3",
		Port:     ":8089",
		Status:   "active",
		LastSeen: time.Now(),
	}
	discovery.addOrUpdateNode(freshNode)

	allNodes := discovery.GetAllNodes()
	if len(allNodes) != 2 {
		t.Errorf("Expected 2 nodes before cleanup, got %d", len(allNodes))
	}

	discovery.RemoveStaleNodes()

	allNodes = discovery.GetAllNodes()
	if len(allNodes) != 1 {
		t.Errorf("Expected 1 node after cleanup, got %d", len(allNodes))
	}

	if allNodes[0].ID != "fresh-node" {
		t.Errorf("Expected remaining node to be 'fresh-node', got '%s'", allNodes[0].ID)
	}
}

func TestIsClusterHealthy(t *testing.T) {
	discovery := NewDiscovery("test-node", "127.0.0.1", ":8089")

	// Single node should be healthy (default min is 1)
	if !discovery.IsClusterHealthy() {
		t.Error("Expected single node cluster to be healthy")
	}

	// Add another node
	node := &Node{
		ID:       "peer-node",
		Address:  "127.0.0.2",
		Port:     ":8089",
		Status:   "active",
		LastSeen: time.Now(),
	}
	discovery.addOrUpdateNode(node)

	if !discovery.IsClusterHealthy() {
		t.Error("Expected two node cluster to be healthy")
	}
}

func TestGetSelfNode(t *testing.T) {
	discovery := NewDiscovery("test-node", "127.0.0.1", ":8089")

	selfNode := discovery.GetSelfNode()
	if selfNode == nil {
		t.Fatal("Expected self node, got nil")
	}

	if selfNode.ID != "test-node" {
		t.Errorf("Expected self node ID 'test-node', got '%s'", selfNode.ID)
	}

	if selfNode.Address != "127.0.0.1" {
		t.Errorf("Expected self node address '127.0.0.1', got '%s'", selfNode.Address)
	}
}

func TestStartDiscovery(t *testing.T) {
	discovery := NewDiscovery("test-node", "127.0.0.1", ":8089")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := discovery.Start(ctx)
	if err != nil {
		t.Errorf("Expected no error starting discovery, got %v", err)
	}

	// Wait a bit for goroutines to start
	time.Sleep(50 * time.Millisecond)

	// Cancel context and wait for cleanup
	cancel()
	time.Sleep(50 * time.Millisecond)
}
