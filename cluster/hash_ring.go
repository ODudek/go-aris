package cluster

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"sync"
)

// HashRing implements consistent hashing for distributing keys across nodes
type HashRing struct {
	mu           sync.RWMutex
	nodes        map[string]*Node  // node_id -> Node
	ring         map[uint64]string // hash -> node_id
	sortedHashes []uint64          // sorted hash values
	replicas     int               // number of virtual nodes per physical node
}

// NewHashRing creates a new consistent hash ring
func NewHashRing(replicas int) *HashRing {
	if replicas <= 0 {
		replicas = 150 // default number of virtual nodes
	}

	return &HashRing{
		nodes:    make(map[string]*Node),
		ring:     make(map[uint64]string),
		replicas: replicas,
	}
}

// AddNode adds a node to the hash ring
func (hr *HashRing) AddNode(node *Node) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	hr.nodes[node.ID] = node

	// Add virtual nodes to the ring
	for i := 0; i < hr.replicas; i++ {
		virtualKey := fmt.Sprintf("%s:%d", node.ID, i)
		hash := hr.hashKey(virtualKey)
		hr.ring[hash] = node.ID
	}

	hr.updateSortedHashes()
}

// RemoveNode removes a node from the hash ring
func (hr *HashRing) RemoveNode(nodeID string) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	delete(hr.nodes, nodeID)

	// Remove virtual nodes from the ring
	for i := 0; i < hr.replicas; i++ {
		virtualKey := fmt.Sprintf("%s:%d", nodeID, i)
		hash := hr.hashKey(virtualKey)
		delete(hr.ring, hash)
	}

	hr.updateSortedHashes()
}

// GetNode returns the node responsible for a given key
func (hr *HashRing) GetNode(key string) *Node {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	if len(hr.ring) == 0 {
		return nil
	}

	hash := hr.hashKey(key)

	// Find the first node hash that is >= key hash
	idx := sort.Search(len(hr.sortedHashes), func(i int) bool {
		return hr.sortedHashes[i] >= hash
	})

	// Wrap around if necessary
	if idx == len(hr.sortedHashes) {
		idx = 0
	}

	nodeID := hr.ring[hr.sortedHashes[idx]]
	return hr.nodes[nodeID]
}

// GetNodes returns N nodes responsible for a given key (for replication)
func (hr *HashRing) GetNodes(key string, count int) []*Node {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	if len(hr.nodes) == 0 || count <= 0 {
		return nil
	}

	if count > len(hr.nodes) {
		count = len(hr.nodes)
	}

	hash := hr.hashKey(key)

	// Find the starting position
	idx := sort.Search(len(hr.sortedHashes), func(i int) bool {
		return hr.sortedHashes[i] >= hash
	})

	if idx == len(hr.sortedHashes) {
		idx = 0
	}

	seen := make(map[string]bool)
	var result []*Node

	// Collect unique nodes
	for len(result) < count && len(seen) < len(hr.nodes) {
		nodeID := hr.ring[hr.sortedHashes[idx]]

		if !seen[nodeID] {
			if node, exists := hr.nodes[nodeID]; exists {
				result = append(result, node)
				seen[nodeID] = true
			}
		}

		idx = (idx + 1) % len(hr.sortedHashes)
	}

	return result
}

// GetPredecessorNode returns the node that comes before the given key
func (hr *HashRing) GetPredecessorNode(key string) *Node {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	if len(hr.ring) == 0 {
		return nil
	}

	hash := hr.hashKey(key)

	// Find the last node hash that is < key hash
	idx := sort.Search(len(hr.sortedHashes), func(i int) bool {
		return hr.sortedHashes[i] >= hash
	})

	// Get the predecessor
	if idx == 0 {
		idx = len(hr.sortedHashes) - 1
	} else {
		idx--
	}

	nodeID := hr.ring[hr.sortedHashes[idx]]
	return hr.nodes[nodeID]
}

// GetSuccessorNode returns the node that comes after the given key
func (hr *HashRing) GetSuccessorNode(key string) *Node {
	// This is the same as GetNode
	return hr.GetNode(key)
}

// GetKeyRange returns the range of keys that a node is responsible for
func (hr *HashRing) GetKeyRange(nodeID string) (start, end uint64, exists bool) {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	if _, exists = hr.nodes[nodeID]; !exists {
		return 0, 0, false
	}

	// Find the first virtual node for this physical node
	var nodeHashes []uint64
	for hash, id := range hr.ring {
		if id == nodeID {
			nodeHashes = append(nodeHashes, hash)
		}
	}

	if len(nodeHashes) == 0 {
		return 0, 0, false
	}

	sort.Slice(nodeHashes, func(i, j int) bool {
		return nodeHashes[i] < nodeHashes[j]
	})

	// The range starts from the predecessor of the smallest hash
	// and ends at the largest hash
	smallestHash := nodeHashes[0]
	largestHash := nodeHashes[len(nodeHashes)-1]

	// Find predecessor of smallest hash
	idx := sort.Search(len(hr.sortedHashes), func(i int) bool {
		return hr.sortedHashes[i] >= smallestHash
	})

	if idx == 0 {
		start = hr.sortedHashes[len(hr.sortedHashes)-1]
	} else {
		start = hr.sortedHashes[idx-1]
	}

	end = largestHash
	exists = true
	return
}

// IsResponsible checks if a node is responsible for a given key
func (hr *HashRing) IsResponsible(nodeID, key string) bool {
	node := hr.GetNode(key)
	return node != nil && node.ID == nodeID
}

// GetNodeLoad returns the approximate load distribution
func (hr *HashRing) GetNodeLoad() map[string]float64 {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	if len(hr.nodes) == 0 {
		return nil
	}

	load := make(map[string]float64)
	totalVirtualNodes := len(hr.ring)

	for nodeID := range hr.nodes {
		virtualNodeCount := 0
		for _, id := range hr.ring {
			if id == nodeID {
				virtualNodeCount++
			}
		}
		load[nodeID] = float64(virtualNodeCount) / float64(totalVirtualNodes)
	}

	return load
}

// GetAllNodes returns all nodes in the ring
func (hr *HashRing) GetAllNodes() []*Node {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	nodes := make([]*Node, 0, len(hr.nodes))
	for _, node := range hr.nodes {
		nodes = append(nodes, node)
	}

	return nodes
}

// IsEmpty returns true if the ring has no nodes
func (hr *HashRing) IsEmpty() bool {
	hr.mu.RLock()
	defer hr.mu.RUnlock()
	return len(hr.nodes) == 0
}

// Size returns the number of nodes in the ring
func (hr *HashRing) Size() int {
	hr.mu.RLock()
	defer hr.mu.RUnlock()
	return len(hr.nodes)
}

// hashKey generates a hash for a given key
func (hr *HashRing) hashKey(key string) uint64 {
	hasher := sha256.New()
	hasher.Write([]byte(key))
	hash := hasher.Sum(nil)

	// Convert first 8 bytes to uint64
	var result uint64
	for i := 0; i < 8; i++ {
		result = result<<8 + uint64(hash[i])
	}

	return result
}

// updateSortedHashes updates the sorted hash slice
func (hr *HashRing) updateSortedHashes() {
	hr.sortedHashes = make([]uint64, 0, len(hr.ring))
	for hash := range hr.ring {
		hr.sortedHashes = append(hr.sortedHashes, hash)
	}
	sort.Slice(hr.sortedHashes, func(i, j int) bool {
		return hr.sortedHashes[i] < hr.sortedHashes[j]
	})
}

// Rebalance can be called when nodes are added or removed to redistribute data
func (hr *HashRing) Rebalance() map[string][]string {
	// Returns a map of nodeID -> list of keys that should be moved to this node
	// This would be used by the storage layer to actually move data
	return make(map[string][]string)
}

// GetReplicationNodes returns nodes for storing replicas of a key
func (hr *HashRing) GetReplicationNodes(key string, replicationFactor int) []*Node {
	return hr.GetNodes(key, replicationFactor)
}

// String returns a string representation of the hash ring
func (hr *HashRing) String() string {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	return fmt.Sprintf("HashRing{nodes: %d, virtual_nodes: %d, replicas: %d}",
		len(hr.nodes), len(hr.ring), hr.replicas)
}
