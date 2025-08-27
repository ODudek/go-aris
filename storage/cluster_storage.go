package storage

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ODudek/go-aris/cluster"
)

// ClusterStorage extends Storage with cluster-aware capabilities
type ClusterStorage struct {
	*Storage          // embed local storage
	mu                sync.RWMutex
	hashRing          *cluster.HashRing
	discovery         *cluster.Discovery
	replicationFactor int
	selfNodeID        string
	grpcClient        *cluster.GRPCClient
	syncInterval      time.Duration
}

// ReplicationConfig holds replication settings
type ReplicationConfig struct {
	Factor   int           // number of replicas
	Timeout  time.Duration // replication timeout
	Strategy string        // sync, async
}

// NewClusterStorage creates a new cluster-aware storage
func NewClusterStorage(nodeID string, discovery *cluster.Discovery) *ClusterStorage {
	return &ClusterStorage{
		Storage:           NewStorage(),
		hashRing:          cluster.NewHashRing(150),
		discovery:         discovery,
		replicationFactor: 3, // default replication factor
		selfNodeID:        nodeID,
		grpcClient:        cluster.NewGRPCClient(),
		syncInterval:      30 * time.Second,
	}
}

// Start initializes the cluster storage
func (cs *ClusterStorage) Start(ctx context.Context) error {
	// Add self to hash ring
	selfNode := cs.discovery.GetSelfNode()
	cs.hashRing.AddNode(selfNode)

	// Start synchronization processes
	go cs.syncWithCluster(ctx)
	go cs.monitorClusterChanges(ctx)

	return nil
}

// Set stores a key-value pair with cluster replication
func (cs *ClusterStorage) Set(key, value string) error {
	// Determine which nodes should store this key
	nodes := cs.hashRing.GetReplicationNodes(key, cs.replicationFactor)

	// Check if this node should store the key
	shouldStore := false
	for _, node := range nodes {
		if node.ID == cs.selfNodeID {
			shouldStore = true
			break
		}
	}

	if shouldStore {
		// Store locally
		cs.Storage.Set(key, value)
	}

	// Replicate to other nodes asynchronously
	go cs.replicateSet(key, value, nodes)

	return nil
}

// Get retrieves a value with cluster fallback
func (cs *ClusterStorage) Get(key string) (string, bool, error) {
	// First try local storage
	if value, exists, err := cs.Storage.Get(key); err == nil && exists {
		return value, true, nil
	}

	// If not found locally, check if we should have it
	responsibleNodes := cs.hashRing.GetReplicationNodes(key, cs.replicationFactor)
	shouldHave := false
	for _, node := range responsibleNodes {
		if node.ID == cs.selfNodeID {
			shouldHave = true
			break
		}
	}

	if shouldHave {
		// We should have it but don't - try to sync from peers
		if value, err := cs.fetchFromPeers(key, responsibleNodes); err == nil {
			// Store locally for future access
			cs.Storage.Set(key, value)
			return value, true, nil
		}
	}

	// Route to responsible node
	primaryNode := cs.hashRing.GetNode(key)
	if primaryNode != nil && primaryNode.ID != cs.selfNodeID {
		return cs.fetchFromNode(key, primaryNode)
	}

	return "", false, nil
}

// HSet stores a hash field with cluster replication
func (cs *ClusterStorage) HSet(key, field, value string) error {
	nodes := cs.hashRing.GetReplicationNodes(key, cs.replicationFactor)

	shouldStore := false
	for _, node := range nodes {
		if node.ID == cs.selfNodeID {
			shouldStore = true
			break
		}
	}

	if shouldStore {
		cs.Storage.HSet(key, field, value)
	}

	go cs.replicateHSet(key, field, value, nodes)
	return nil
}

// HGet retrieves a hash field with cluster fallback
func (cs *ClusterStorage) HGet(key, field string) (string, bool, error) {
	if value, exists, err := cs.Storage.HGet(key, field); err == nil && exists {
		return value, true, nil
	}

	// Try to fetch from responsible nodes
	responsibleNodes := cs.hashRing.GetReplicationNodes(key, cs.replicationFactor)
	for _, node := range responsibleNodes {
		if node.ID != cs.selfNodeID {
			if value, exists, err := cs.fetchHGetFromNode(key, field, node); err == nil && exists {
				// Cache locally
				cs.Storage.HSet(key, field, value)
				return value, true, nil
			}
		}
	}

	return "", false, nil
}

// HGetAll retrieves all hash fields with cluster fallback
func (cs *ClusterStorage) HGetAll(key string) (map[string]string, error) {
	localData, err := cs.Storage.HGetAll(key)
	if err != nil {
		return nil, err
	}

	// If we have data locally, return it
	if len(localData) > 0 {
		return localData, nil
	}

	// Try to fetch from responsible nodes
	responsibleNodes := cs.hashRing.GetReplicationNodes(key, cs.replicationFactor)
	for _, node := range responsibleNodes {
		if node.ID != cs.selfNodeID {
			if data, err := cs.fetchHGetAllFromNode(key, node); err == nil && len(data) > 0 {
				// Cache locally
				for field, value := range data {
					cs.Storage.HSet(key, field, value)
				}
				return data, nil
			}
		}
	}

	return make(map[string]string), nil
}

// replicateSet replicates a SET operation to other nodes via gRPC
func (cs *ClusterStorage) replicateSet(key, value string, nodes []*cluster.Node) {
	for _, node := range nodes {
		if node.ID != cs.selfNodeID && node.Status == "active" {
			go func(n *cluster.Node) {
				err := cs.grpcClient.SetKey(n, key, value, true, cs.selfNodeID)
				if err != nil {
					log.Printf("Failed to replicate SET to node %s: %v", n.ID, err)
				}
			}(node)
		}
	}
}

// replicateHSet replicates an HSET operation to other nodes via gRPC
func (cs *ClusterStorage) replicateHSet(key, field, value string, nodes []*cluster.Node) {
	for _, node := range nodes {
		if node.ID != cs.selfNodeID && node.Status == "active" {
			go func(n *cluster.Node) {
				err := cs.grpcClient.HSetKey(n, key, field, value, true, cs.selfNodeID)
				if err != nil {
					log.Printf("Failed to replicate HSET to node %s: %v", n.ID, err)
				}
			}(node)
		}
	}
}

// fetchFromNode retrieves a value from a specific node via gRPC
func (cs *ClusterStorage) fetchFromNode(key string, node *cluster.Node) (string, bool, error) {
	return cs.grpcClient.GetKey(node, key, cs.selfNodeID)
}

// fetchFromPeers tries to fetch a key from peer nodes
func (cs *ClusterStorage) fetchFromPeers(key string, nodes []*cluster.Node) (string, error) {
	for _, node := range nodes {
		if node.ID != cs.selfNodeID && node.Status == "active" {
			if value, exists, err := cs.fetchFromNode(key, node); err == nil && exists {
				return value, nil
			}
		}
	}
	return "", fmt.Errorf("key not found in any peer")
}

// fetchHGetFromNode retrieves a hash field from a specific node via gRPC
func (cs *ClusterStorage) fetchHGetFromNode(key, field string, node *cluster.Node) (string, bool, error) {
	return cs.grpcClient.HGetKey(node, key, field, cs.selfNodeID)
}

// fetchHGetAllFromNode retrieves all hash fields from a specific node via gRPC
func (cs *ClusterStorage) fetchHGetAllFromNode(key string, node *cluster.Node) (map[string]string, error) {
	return cs.grpcClient.HGetAllKey(node, key, cs.selfNodeID)
}

// syncWithCluster periodically synchronizes with other cluster nodes
func (cs *ClusterStorage) syncWithCluster(ctx context.Context) {
	ticker := time.NewTicker(cs.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cs.performSync()
		}
	}
}

// monitorClusterChanges watches for cluster topology changes
func (cs *ClusterStorage) monitorClusterChanges(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	lastNodeCount := 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Update hash ring with current active nodes
			activeNodes := cs.discovery.GetActiveNodes()
			currentNodeCount := len(activeNodes)

			if currentNodeCount != lastNodeCount {
				log.Printf("Cluster size changed: %d -> %d nodes", lastNodeCount, currentNodeCount)
				cs.updateHashRing(activeNodes)
				lastNodeCount = currentNodeCount

				// Trigger data rebalancing if needed
				go cs.rebalanceData()
			}
		}
	}
}

// updateHashRing updates the hash ring with current active nodes
func (cs *ClusterStorage) updateHashRing(nodes []*cluster.Node) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Create new hash ring
	newRing := cluster.NewHashRing(150)

	// Add self
	selfNode := cs.discovery.GetSelfNode()
	newRing.AddNode(selfNode)

	// Add active nodes
	for _, node := range nodes {
		if node.ID != cs.selfNodeID {
			newRing.AddNode(node)
		}
	}

	cs.hashRing = newRing
}

// performSync synchronizes data with peer nodes
func (cs *ClusterStorage) performSync() {
	// Implementation would sync data based on hash ring ownership
	log.Printf("Performing cluster sync...")
}

// rebalanceData redistributes data when cluster topology changes
func (cs *ClusterStorage) rebalanceData() {
	// Implementation would move data between nodes as needed
	log.Printf("Starting data rebalancing...")
}

// GetClusterStatus returns information about cluster health
func (cs *ClusterStorage) GetClusterStatus() map[string]interface{} {
	activeNodes := cs.discovery.GetActiveNodes()

	return map[string]interface{}{
		"node_id":            cs.selfNodeID,
		"cluster_size":       len(activeNodes) + 1, // +1 for self
		"replication_factor": cs.replicationFactor,
		"hash_ring_size":     cs.hashRing.Size(),
		"active_nodes":       activeNodes,
		"is_healthy":         cs.discovery.IsClusterHealthy(),
	}
}

// SetReplicationFactor updates the replication factor
func (cs *ClusterStorage) SetReplicationFactor(factor int) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.replicationFactor = factor
}
