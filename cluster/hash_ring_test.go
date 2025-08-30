package cluster

import (
	"testing"
)

func TestNewHashRing(t *testing.T) {
	ring := NewHashRing(100)

	if ring == nil {
		t.Fatal("Expected hash ring instance, got nil")
	}

	if ring.replicas != 100 {
		t.Errorf("Expected 100 replicas, got %d", ring.replicas)
	}

	if len(ring.nodes) != 0 {
		t.Errorf("Expected empty nodes map, got %d nodes", len(ring.nodes))
	}

	if len(ring.ring) != 0 {
		t.Errorf("Expected empty ring, got %d entries", len(ring.ring))
	}
}

func TestNewHashRingDefaultReplicas(t *testing.T) {
	ring := NewHashRing(0)

	if ring.replicas != 150 {
		t.Errorf("Expected default 150 replicas, got %d", ring.replicas)
	}
}

func TestAddNode(t *testing.T) {
	ring := NewHashRing(10)

	node := &Node{
		ID:      "node1",
		Address: "127.0.0.1",
		Port:    ":8089",
		Status:  "active",
	}

	ring.AddNode(node)

	if len(ring.nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(ring.nodes))
	}

	if ring.nodes["node1"] != node {
		t.Error("Node not found in nodes map")
	}

	// Should have 10 virtual nodes in the ring
	if len(ring.ring) != 10 {
		t.Errorf("Expected 10 virtual nodes in ring, got %d", len(ring.ring))
	}

	if len(ring.sortedHashes) != 10 {
		t.Errorf("Expected 10 sorted hashes, got %d", len(ring.sortedHashes))
	}
}

func TestRemoveNode(t *testing.T) {
	ring := NewHashRing(10)

	node := &Node{
		ID:      "node1",
		Address: "127.0.0.1",
		Port:    ":8089",
		Status:  "active",
	}

	ring.AddNode(node)
	ring.RemoveNode("node1")

	if len(ring.nodes) != 0 {
		t.Errorf("Expected 0 nodes after removal, got %d", len(ring.nodes))
	}

	if len(ring.ring) != 0 {
		t.Errorf("Expected 0 virtual nodes after removal, got %d", len(ring.ring))
	}

	if len(ring.sortedHashes) != 0 {
		t.Errorf("Expected 0 sorted hashes after removal, got %d", len(ring.sortedHashes))
	}
}

func TestGetNode(t *testing.T) {
	ring := NewHashRing(10)

	node1 := &Node{
		ID:      "node1",
		Address: "127.0.0.1",
		Port:    ":8089",
		Status:  "active",
	}

	node2 := &Node{
		ID:      "node2",
		Address: "127.0.0.2",
		Port:    ":8089",
		Status:  "active",
	}

	ring.AddNode(node1)
	ring.AddNode(node2)

	// Test getting node for a key
	responsibleNode := ring.GetNode("test-key")
	if responsibleNode == nil {
		t.Error("Expected a responsible node, got nil")
	}

	if responsibleNode.ID != "node1" && responsibleNode.ID != "node2" {
		t.Errorf("Expected node1 or node2, got %s", responsibleNode.ID)
	}

	// Same key should always return the same node
	responsibleNode2 := ring.GetNode("test-key")
	if responsibleNode.ID != responsibleNode2.ID {
		t.Error("Same key should return the same node")
	}
}

func TestGetNodeEmptyRing(t *testing.T) {
	ring := NewHashRing(10)

	node := ring.GetNode("test-key")
	if node != nil {
		t.Error("Expected nil for empty ring, got a node")
	}
}

func TestGetNodes(t *testing.T) {
	ring := NewHashRing(10)

	node1 := &Node{
		ID:      "node1",
		Address: "127.0.0.1",
		Port:    ":8089",
		Status:  "active",
	}

	node2 := &Node{
		ID:      "node2",
		Address: "127.0.0.2",
		Port:    ":8089",
		Status:  "active",
	}

	node3 := &Node{
		ID:      "node3",
		Address: "127.0.0.3",
		Port:    ":8089",
		Status:  "active",
	}

	ring.AddNode(node1)
	ring.AddNode(node2)
	ring.AddNode(node3)

	// Get 2 nodes for replication
	nodes := ring.GetNodes("test-key", 2)
	if len(nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(nodes))
	}

	// Should be unique nodes
	if nodes[0].ID == nodes[1].ID {
		t.Error("Expected unique nodes, got duplicates")
	}

	// Test requesting more nodes than available
	nodes = ring.GetNodes("test-key", 5)
	if len(nodes) != 3 {
		t.Errorf("Expected 3 nodes (max available), got %d", len(nodes))
	}
}

func TestGetNodesEmptyRing(t *testing.T) {
	ring := NewHashRing(10)

	nodes := ring.GetNodes("test-key", 2)
	if nodes != nil {
		t.Error("Expected nil for empty ring, got nodes")
	}
}

func TestIsResponsible(t *testing.T) {
	ring := NewHashRing(10)

	node1 := &Node{
		ID:      "node1",
		Address: "127.0.0.1",
		Port:    ":8089",
		Status:  "active",
	}

	node2 := &Node{
		ID:      "node2",
		Address: "127.0.0.2",
		Port:    ":8089",
		Status:  "active",
	}

	ring.AddNode(node1)
	ring.AddNode(node2)

	key := "test-key"
	responsibleNode := ring.GetNode(key)

	if !ring.IsResponsible(responsibleNode.ID, key) {
		t.Error("Node should be responsible for the key")
	}

	// Find the other node
	otherNodeID := "node1"
	if responsibleNode.ID == "node1" {
		otherNodeID = "node2"
	}

	if ring.IsResponsible(otherNodeID, key) {
		t.Error("Other node should not be responsible for the key")
	}
}

func TestGetNodeLoad(t *testing.T) {
	ring := NewHashRing(20)

	node1 := &Node{
		ID:      "node1",
		Address: "127.0.0.1",
		Port:    ":8089",
		Status:  "active",
	}

	node2 := &Node{
		ID:      "node2",
		Address: "127.0.0.2",
		Port:    ":8089",
		Status:  "active",
	}

	ring.AddNode(node1)
	ring.AddNode(node2)

	load := ring.GetNodeLoad()
	if len(load) != 2 {
		t.Errorf("Expected load info for 2 nodes, got %d", len(load))
	}

	if load["node1"] <= 0 || load["node1"] >= 1 {
		t.Errorf("Expected load between 0 and 1 for node1, got %f", load["node1"])
	}

	if load["node2"] <= 0 || load["node2"] >= 1 {
		t.Errorf("Expected load between 0 and 1 for node2, got %f", load["node2"])
	}

	// Total load should be approximately 1.0
	totalLoad := load["node1"] + load["node2"]
	if totalLoad < 0.99 || totalLoad > 1.01 {
		t.Errorf("Expected total load around 1.0, got %f", totalLoad)
	}
}

func TestGetAllNodes(t *testing.T) {
	ring := NewHashRing(10)

	node1 := &Node{
		ID:      "node1",
		Address: "127.0.0.1",
		Port:    ":8089",
		Status:  "active",
	}

	node2 := &Node{
		ID:      "node2",
		Address: "127.0.0.2",
		Port:    ":8089",
		Status:  "active",
	}

	ring.AddNode(node1)
	ring.AddNode(node2)

	nodes := ring.GetAllNodes()
	if len(nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(nodes))
	}

	// Check that we got both nodes
	nodeIDs := make(map[string]bool)
	for _, node := range nodes {
		nodeIDs[node.ID] = true
	}

	if !nodeIDs["node1"] || !nodeIDs["node2"] {
		t.Error("Expected both node1 and node2 in result")
	}
}

func TestIsEmpty(t *testing.T) {
	ring := NewHashRing(10)

	if !ring.IsEmpty() {
		t.Error("Expected empty ring to return true for IsEmpty()")
	}

	node := &Node{
		ID:      "node1",
		Address: "127.0.0.1",
		Port:    ":8089",
		Status:  "active",
	}

	ring.AddNode(node)

	if ring.IsEmpty() {
		t.Error("Expected non-empty ring to return false for IsEmpty()")
	}
}

func TestSize(t *testing.T) {
	ring := NewHashRing(10)

	if ring.Size() != 0 {
		t.Errorf("Expected size 0, got %d", ring.Size())
	}

	node1 := &Node{
		ID:      "node1",
		Address: "127.0.0.1",
		Port:    ":8089",
		Status:  "active",
	}

	ring.AddNode(node1)

	if ring.Size() != 1 {
		t.Errorf("Expected size 1, got %d", ring.Size())
	}

	node2 := &Node{
		ID:      "node2",
		Address: "127.0.0.2",
		Port:    ":8089",
		Status:  "active",
	}

	ring.AddNode(node2)

	if ring.Size() != 2 {
		t.Errorf("Expected size 2, got %d", ring.Size())
	}

	ring.RemoveNode("node1")

	if ring.Size() != 1 {
		t.Errorf("Expected size 1 after removal, got %d", ring.Size())
	}
}

func TestConsistentHashing(t *testing.T) {
	ring := NewHashRing(100)

	node1 := &Node{
		ID:      "node1",
		Address: "127.0.0.1",
		Port:    ":8089",
		Status:  "active",
	}

	node2 := &Node{
		ID:      "node2",
		Address: "127.0.0.2",
		Port:    ":8089",
		Status:  "active",
	}

	ring.AddNode(node1)
	ring.AddNode(node2)

	// Test that keys are consistently mapped
	keys := []string{"key1", "key2", "key3", "key4", "key5"}
	mapping1 := make(map[string]string)

	for _, key := range keys {
		node := ring.GetNode(key)
		mapping1[key] = node.ID
	}

	// Add another node
	node3 := &Node{
		ID:      "node3",
		Address: "127.0.0.3",
		Port:    ":8089",
		Status:  "active",
	}

	ring.AddNode(node3)

	// Check how many keys moved
	moved := 0
	for _, key := range keys {
		node := ring.GetNode(key)
		if mapping1[key] != node.ID {
			moved++
		}
	}

	// With consistent hashing, not all keys should move
	if moved == len(keys) {
		t.Error("All keys moved when adding a node - consistent hashing not working properly")
	}

	// But some keys should move to achieve load balancing
	if moved == 0 {
		t.Error("No keys moved when adding a node - this is unexpected")
	}
}

func TestGetReplicationNodes(t *testing.T) {
	ring := NewHashRing(10)

	node1 := &Node{
		ID:      "node1",
		Address: "127.0.0.1",
		Port:    ":8089",
		Status:  "active",
	}

	node2 := &Node{
		ID:      "node2",
		Address: "127.0.0.2",
		Port:    ":8089",
		Status:  "active",
	}

	ring.AddNode(node1)
	ring.AddNode(node2)

	replicationNodes := ring.GetReplicationNodes("test-key", 2)
	if len(replicationNodes) != 2 {
		t.Errorf("Expected 2 replication nodes, got %d", len(replicationNodes))
	}

	// Should be the same as GetNodes
	nodes := ring.GetNodes("test-key", 2)
	if len(nodes) != len(replicationNodes) {
		t.Error("GetReplicationNodes should return same as GetNodes")
	}

	for i, node := range nodes {
		if node.ID != replicationNodes[i].ID {
			t.Error("GetReplicationNodes should return same nodes as GetNodes")
		}
	}
}

func TestString(t *testing.T) {
	ring := NewHashRing(10)

	str := ring.String()
	if str == "" {
		t.Error("Expected non-empty string representation")
	}

	// Should contain basic info
	expected := "HashRing{nodes: 0, virtual_nodes: 0, replicas: 10}"
	if str != expected {
		t.Errorf("Expected '%s', got '%s'", expected, str)
	}

	// Add a node and check again
	node := &Node{
		ID:      "node1",
		Address: "127.0.0.1",
		Port:    ":8089",
		Status:  "active",
	}

	ring.AddNode(node)
	str = ring.String()
	expected = "HashRing{nodes: 1, virtual_nodes: 10, replicas: 10}"
	if str != expected {
		t.Errorf("Expected '%s', got '%s'", expected, str)
	}
}
