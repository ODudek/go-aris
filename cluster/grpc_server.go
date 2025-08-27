package cluster

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

// GRPCServer handles incoming gRPC requests from other cluster nodes
type GRPCServer struct {
	mu        sync.RWMutex
	server    *grpc.Server
	listener  net.Listener
	storage   ClusterStorageInterface
	discovery *Discovery
	nodeID    string
	address   string
	port      string
	startTime time.Time
	metrics   *NodeMetrics
}

// ClusterStorageInterface defines the interface for storage operations
type ClusterStorageInterface interface {
	Set(key, value string) error
	Get(key string) (string, bool, error)
	HSet(key, field, value string) error
	HGet(key, field string) (string, bool, error)
	HGetAll(key string) (map[string]string, error)
	GetClusterStatus() map[string]interface{}
}

// NewGRPCServer creates a new gRPC server for cluster communication
func NewGRPCServer(nodeID, address, port string, storage ClusterStorageInterface, discovery *Discovery) *GRPCServer {
	return &GRPCServer{
		storage:   storage,
		discovery: discovery,
		nodeID:    nodeID,
		address:   address,
		port:      port,
		startTime: time.Now(),
		metrics:   &NodeMetrics{},
	}
}

// Start initializes and starts the gRPC server
func (s *GRPCServer) Start(ctx context.Context) error {
	// Parse port number
	portNum := strings.TrimPrefix(s.port, ":")
	grpcPort, err := strconv.Atoi(portNum)
	if err != nil {
		return fmt.Errorf("invalid port format: %s", s.port)
	}

	// Use a different port for gRPC (Redis port + 1000)
	grpcPort += 1000
	grpcAddr := fmt.Sprintf("%s:%d", s.address, grpcPort)

	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", grpcAddr, err)
	}
	s.listener = listener

	// Create gRPC server with keepalive settings
	s.server = grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Second,
			MaxConnectionAge:      30 * time.Second,
			MaxConnectionAgeGrace: 5 * time.Second,
			Time:                  5 * time.Second,
			Timeout:               1 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.MaxRecvMsgSize(4*1024*1024), // 4MB
		grpc.MaxSendMsgSize(4*1024*1024), // 4MB
	)

	// Register the service
	RegisterClusterServiceServer(s.server, s)

	// Enable reflection for debugging
	reflection.Register(s.server)

	log.Printf("Starting gRPC server for cluster communication on %s", grpcAddr)

	// Start server in goroutine
	go func() {
		if err := s.server.Serve(listener); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	// Wait for context cancellation
	go func() {
		<-ctx.Done()
		s.Stop()
	}()

	return nil
}

// Stop gracefully stops the gRPC server
func (s *GRPCServer) Stop() error {
	if s.server != nil {
		s.server.GracefulStop()
	}
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// Implement ClusterServiceServer interface

// Set handles SET operations from other nodes
func (s *GRPCServer) Set(ctx context.Context, req *SetRequest) (*SetResponse, error) {
	s.updateMetrics("set")

	if req.Key == "" {
		return NewSetResponse(false, fmt.Errorf("key cannot be empty")), nil
	}

	// If this is a replication, store directly without further replication loops
	if req.IsReplication {
		// For replication, we need to access the underlying storage directly
		// This is a simplified implementation - in practice you'd need proper interface
		log.Printf("Storing replicated SET: %s = %s from node %s", req.Key, req.Value, req.SourceNode)
		err := s.storage.Set(req.Key, req.Value)
		return NewSetResponse(err == nil, err), nil
	}

	// Normal operation
	err := s.storage.Set(req.Key, req.Value)
	return NewSetResponse(err == nil, err), nil
}

// Get handles GET operations from other nodes
func (s *GRPCServer) Get(ctx context.Context, req *GetRequest) (*GetResponse, error) {
	s.updateMetrics("get")

	if req.Key == "" {
		return NewGetResponse("", false, fmt.Errorf("key cannot be empty")), nil
	}

	value, found, err := s.storage.Get(req.Key)
	return NewGetResponse(value, found, err), nil
}

// HSet handles HSET operations from other nodes
func (s *GRPCServer) HSet(ctx context.Context, req *HSetRequest) (*HSetResponse, error) {
	s.updateMetrics("hset")

	if req.Key == "" || req.Field == "" {
		return NewHSetResponse(false, fmt.Errorf("key and field cannot be empty")), nil
	}

	// If this is a replication, store directly without further replication
	if req.IsReplication {
		// For replication, we need to access the underlying storage directly
		log.Printf("Storing replicated HSET: %s.%s = %s from node %s", req.Key, req.Field, req.Value, req.SourceNode)
		err := s.storage.HSet(req.Key, req.Field, req.Value)
		return NewHSetResponse(err == nil, err), nil
	}

	err := s.storage.HSet(req.Key, req.Field, req.Value)
	return NewHSetResponse(err == nil, err), nil
}

// HGet handles HGET operations from other nodes
func (s *GRPCServer) HGet(ctx context.Context, req *HGetRequest) (*HGetResponse, error) {
	s.updateMetrics("hget")

	if req.Key == "" || req.Field == "" {
		return NewHGetResponse("", false, fmt.Errorf("key and field cannot be empty")), nil
	}

	value, found, err := s.storage.HGet(req.Key, req.Field)
	return NewHGetResponse(value, found, err), nil
}

// HGetAll handles HGETALL operations from other nodes
func (s *GRPCServer) HGetAll(ctx context.Context, req *HGetAllRequest) (*HGetAllResponse, error) {
	s.updateMetrics("hgetall")

	if req.Key == "" {
		return NewHGetAllResponse(nil, fmt.Errorf("key cannot be empty")), nil
	}

	fields, err := s.storage.HGetAll(req.Key)
	return NewHGetAllResponse(fields, err), nil
}

// Delete handles DELETE operations from other nodes
func (s *GRPCServer) Delete(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error) {
	s.updateMetrics("delete")

	// For now, return not implemented
	return &DeleteResponse{
		Success: false,
		Error:   "DELETE operation not yet implemented",
	}, nil
}

// Ping handles ping requests from other nodes
func (s *GRPCServer) Ping(ctx context.Context, req *PingRequest) (*PingResponse, error) {
	s.updateMetrics("ping")

	return NewPingResponse(s.nodeID, "active", "1.0.0"), nil
}

// NodeInfo returns information about this node
func (s *GRPCServer) NodeInfo(ctx context.Context, req *NodeInfoRequest) (*NodeInfoResponse, error) {
	s.updateMetrics("nodeinfo")

	uptime := time.Since(s.startTime).Milliseconds()

	return &NodeInfoResponse{
		NodeId:  s.nodeID,
		Address: s.address,
		Port:    s.port,
		Status:  "active",
		Uptime:  uptime,
		Metrics: s.getNodeMetrics(),
		Roles:   []string{"data", "cluster"},
	}, nil
}

// ClusterStatus returns cluster status information
func (s *GRPCServer) ClusterStatus(ctx context.Context, req *ClusterStatusRequest) (*ClusterStatusResponse, error) {
	s.updateMetrics("clusterstatus")

	status := s.storage.GetClusterStatus()

	// Convert to gRPC response format
	activeNodes := int32(0)
	totalNodes := int32(0)
	isHealthy := false
	replicationFactor := int32(3)

	if clusterSize, ok := status["cluster_size"].(int); ok {
		totalNodes = int32(clusterSize)
		activeNodes = int32(clusterSize) // Assume all are active for now
	}

	if healthy, ok := status["is_healthy"].(bool); ok {
		isHealthy = healthy
	}

	if factor, ok := status["replication_factor"].(int); ok {
		replicationFactor = int32(factor)
	}

	return &ClusterStatusResponse{
		TotalNodes:        totalNodes,
		ActiveNodes:       activeNodes,
		Nodes:             []*NodeInfo{}, // TODO: populate with actual node info
		IsHealthy:         isHealthy,
		ReplicationFactor: replicationFactor,
		ClusterId:         "aris-cluster-1",
	}, nil
}

// Join handles node join requests
func (s *GRPCServer) Join(ctx context.Context, req *JoinRequest) (*JoinResponse, error) {
	s.updateMetrics("join")

	log.Printf("Node %s requesting to join cluster", req.NodeId)

	// Create new node info
	newNode := &Node{
		ID:       req.NodeId,
		Address:  req.Address,
		Port:     req.Port,
		Status:   "joining",
		LastSeen: time.Now(),
	}

	// Add to discovery
	if s.discovery != nil {
		s.discovery.addOrUpdateNode(newNode)
	}

	return &JoinResponse{
		Accepted:      true,
		ClusterId:     "aris-cluster-1",
		ExistingNodes: []*NodeInfo{}, // TODO: populate with existing nodes
		Config: &ClusterConfig{
			ReplicationFactor:          3,
			HashRingReplicas:           150,
			SyncIntervalSeconds:        30,
			HealthCheckIntervalSeconds: 15,
		},
	}, nil
}

// Leave handles node leave requests
func (s *GRPCServer) Leave(ctx context.Context, req *LeaveRequest) (*LeaveResponse, error) {
	s.updateMetrics("leave")

	log.Printf("Node %s leaving cluster (graceful: %t)", req.NodeId, req.Graceful)

	return &LeaveResponse{
		Success: true,
	}, nil
}

// SyncData handles data synchronization requests
func (s *GRPCServer) SyncData(ctx context.Context, req *SyncDataRequest) (*SyncDataResponse, error) {
	s.updateMetrics("syncdata")

	// TODO: Implement actual data synchronization
	return &SyncDataResponse{
		Data:     []*KeyValue{},
		HashData: []*HashKeyValue{},
		HasMore:  false,
	}, nil
}

// ReplicateOperation handles replication requests
func (s *GRPCServer) ReplicateOperation(ctx context.Context, req *ReplicateRequest) (*ReplicateResponse, error) {
	s.updateMetrics("replicate")

	var err error

	switch req.Operation {
	case "SET":
		err = s.storage.Set(req.Key, req.Value)
	case "HSET":
		err = s.storage.HSet(req.Key, req.Field, req.Value)
	default:
		err = fmt.Errorf("unsupported operation: %s", req.Operation)
	}

	return &ReplicateResponse{
		Success: err == nil,
		Error: func() string {
			if err != nil {
				return err.Error()
			}
			return ""
		}(),
		NodeId: s.nodeID,
	}, nil
}

// TransferKeys handles key transfer requests
func (s *GRPCServer) TransferKeys(ctx context.Context, req *TransferKeysRequest) (*TransferKeysResponse, error) {
	s.updateMetrics("transferkeys")

	// TODO: Implement key transfer logic
	return &TransferKeysResponse{
		TransferredCount: 0,
		FailedKeys:       []string{},
	}, nil
}

// HealthCheck handles health check requests
func (s *GRPCServer) HealthCheck(ctx context.Context, req *HealthCheckRequest) (*HealthCheckResponse, error) {
	s.updateMetrics("healthcheck")

	return NewHealthCheckResponse("healthy", s.getNodeMetrics(), []string{}), nil
}

// GetMetrics handles metrics requests
func (s *GRPCServer) GetMetrics(ctx context.Context, req *MetricsRequest) (*MetricsResponse, error) {
	s.updateMetrics("getmetrics")

	metrics := make(map[string]float64)
	metrics["uptime_seconds"] = time.Since(s.startTime).Seconds()
	metrics["total_requests"] = float64(s.metrics.OperationsPerSecond)

	return &MetricsResponse{
		Metrics:     metrics,
		NodeMetrics: s.getNodeMetrics(),
		Timestamp:   time.Now().UnixNano(),
	}, nil
}

// Helper methods

func (s *GRPCServer) updateMetrics(operation string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.metrics.OperationsPerSecond++
	// TODO: Implement more detailed metrics tracking
}

func (s *GRPCServer) getNodeMetrics() *NodeMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return &NodeMetrics{
		TotalKeys:           s.metrics.TotalKeys,
		TotalMemoryUsage:    s.metrics.TotalMemoryUsage,
		Connections:         s.metrics.Connections,
		OperationsPerSecond: s.metrics.OperationsPerSecond,
		CpuUsage:            s.metrics.CpuUsage,
		MemoryUsagePercent:  s.metrics.MemoryUsagePercent,
	}
}

// getLocalStorage is removed as we now use the storage interface directly

// Dummy registration function - would be replaced by protobuf generated code
func RegisterClusterServiceServer(s *grpc.Server, srv ClusterServiceServer) {
	// This would be implemented by protobuf generated code
	log.Println("ClusterService registered (dummy implementation)")
}
