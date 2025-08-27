package cluster

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// GRPCClient handles gRPC communication with other cluster nodes
type GRPCClient struct {
	mu          sync.RWMutex
	connections map[string]*grpc.ClientConn
	timeout     time.Duration
}

// NewGRPCClient creates a new gRPC client for cluster communication
func NewGRPCClient() *GRPCClient {
	return &GRPCClient{
		connections: make(map[string]*grpc.ClientConn),
		timeout:     5 * time.Second,
	}
}

// GetConnection returns a gRPC connection to the specified node
func (c *GRPCClient) GetConnection(address, port string) (*grpc.ClientConn, error) {
	target := fmt.Sprintf("%s%s", address, port)

	c.mu.RLock()
	if conn, exists := c.connections[target]; exists {
		c.mu.RUnlock()
		// Check if connection is still valid
		if conn.GetState().String() != "SHUTDOWN" {
			return conn, nil
		}
	} else {
		c.mu.RUnlock()
	}

	// Create new connection
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if conn, exists := c.connections[target]; exists {
		if conn.GetState().String() != "SHUTDOWN" {
			return conn, nil
		}
		// Close existing broken connection
		conn.Close()
	}

	// Create new gRPC connection
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(4*1024*1024), // 4MB
			grpc.MaxCallSendMsgSize(4*1024*1024), // 4MB
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %v", target, err)
	}

	c.connections[target] = conn
	log.Printf("Established gRPC connection to %s", target)

	return conn, nil
}

// SetKey sends a SET operation to a remote node
func (c *GRPCClient) SetKey(node *Node, key, value string, isReplication bool, sourceNode string) error {
	conn, err := c.GetConnection(node.Address, node.Port)
	if err != nil {
		return err
	}

	client := NewClusterServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req := &SetRequest{
		Key:           key,
		Value:         value,
		IsReplication: isReplication,
		SourceNode:    sourceNode,
		Timestamp:     time.Now().UnixNano(),
	}

	resp, err := client.Set(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC Set failed: %v", err)
	}

	if !resp.Success {
		return fmt.Errorf("remote Set failed: %s", resp.Error)
	}

	return nil
}

// GetKey retrieves a key from a remote node
func (c *GRPCClient) GetKey(node *Node, key string, requestingNode string) (string, bool, error) {
	conn, err := c.GetConnection(node.Address, node.Port)
	if err != nil {
		return "", false, err
	}

	client := NewClusterServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req := &GetRequest{
		Key:            key,
		RequestingNode: requestingNode,
	}

	resp, err := client.Get(ctx, req)
	if err != nil {
		return "", false, fmt.Errorf("gRPC Get failed: %v", err)
	}

	if resp.Error != "" {
		return "", false, fmt.Errorf("remote Get failed: %s", resp.Error)
	}

	return resp.Value, resp.Found, nil
}

// HSetKey sends an HSET operation to a remote node
func (c *GRPCClient) HSetKey(node *Node, key, field, value string, isReplication bool, sourceNode string) error {
	conn, err := c.GetConnection(node.Address, node.Port)
	if err != nil {
		return err
	}

	client := NewClusterServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req := &HSetRequest{
		Key:           key,
		Field:         field,
		Value:         value,
		IsReplication: isReplication,
		SourceNode:    sourceNode,
		Timestamp:     time.Now().UnixNano(),
	}

	resp, err := client.HSet(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC HSet failed: %v", err)
	}

	if !resp.Success {
		return fmt.Errorf("remote HSet failed: %s", resp.Error)
	}

	return nil
}

// HGetKey retrieves a hash field from a remote node
func (c *GRPCClient) HGetKey(node *Node, key, field string, requestingNode string) (string, bool, error) {
	conn, err := c.GetConnection(node.Address, node.Port)
	if err != nil {
		return "", false, err
	}

	client := NewClusterServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req := &HGetRequest{
		Key:            key,
		Field:          field,
		RequestingNode: requestingNode,
	}

	resp, err := client.HGet(ctx, req)
	if err != nil {
		return "", false, fmt.Errorf("gRPC HGet failed: %v", err)
	}

	if resp.Error != "" {
		return "", false, fmt.Errorf("remote HGet failed: %s", resp.Error)
	}

	return resp.Value, resp.Found, nil
}

// HGetAllKey retrieves all hash fields from a remote node
func (c *GRPCClient) HGetAllKey(node *Node, key string, requestingNode string) (map[string]string, error) {
	conn, err := c.GetConnection(node.Address, node.Port)
	if err != nil {
		return nil, err
	}

	client := NewClusterServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req := &HGetAllRequest{
		Key:            key,
		RequestingNode: requestingNode,
	}

	resp, err := client.HGetAll(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("gRPC HGetAll failed: %v", err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("remote HGetAll failed: %s", resp.Error)
	}

	return resp.Fields, nil
}

// Ping sends a ping to a remote node
func (c *GRPCClient) Ping(node *Node, nodeID string) (*PingResponse, error) {
	conn, err := c.GetConnection(node.Address, node.Port)
	if err != nil {
		return nil, err
	}

	client := NewClusterServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req := &PingRequest{
		NodeId:    nodeID,
		Timestamp: time.Now().UnixNano(),
	}

	resp, err := client.Ping(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("gRPC Ping failed: %v", err)
	}

	return resp, nil
}

// GetNodeInfo retrieves information about a remote node
func (c *GRPCClient) GetNodeInfo(node *Node, requestingNode string) (*NodeInfoResponse, error) {
	conn, err := c.GetConnection(node.Address, node.Port)
	if err != nil {
		return nil, err
	}

	client := NewClusterServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req := &NodeInfoRequest{
		RequestingNode: requestingNode,
	}

	resp, err := client.NodeInfo(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("gRPC NodeInfo failed: %v", err)
	}

	return resp, nil
}

// GetClusterStatus retrieves cluster status from a remote node
func (c *GRPCClient) GetClusterStatus(node *Node, requestingNode string) (*ClusterStatusResponse, error) {
	conn, err := c.GetConnection(node.Address, node.Port)
	if err != nil {
		return nil, err
	}

	client := NewClusterServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req := &ClusterStatusRequest{
		RequestingNode: requestingNode,
	}

	resp, err := client.ClusterStatus(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("gRPC ClusterStatus failed: %v", err)
	}

	return resp, nil
}

// HealthCheck performs a health check on a remote node
func (c *GRPCClient) HealthCheck(node *Node, requestingNode string) (*HealthCheckResponse, error) {
	conn, err := c.GetConnection(node.Address, node.Port)
	if err != nil {
		return nil, err
	}

	client := NewClusterServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req := &HealthCheckRequest{
		RequestingNode: requestingNode,
		Timestamp:      time.Now().UnixNano(),
	}

	resp, err := client.HealthCheck(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("gRPC HealthCheck failed: %v", err)
	}

	return resp, nil
}

// ReplicateOperation sends a replication operation to a remote node
func (c *GRPCClient) ReplicateOperation(node *Node, operation, key, field, value, sourceNode string) error {
	conn, err := c.GetConnection(node.Address, node.Port)
	if err != nil {
		return err
	}

	client := NewClusterServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req := &ReplicateRequest{
		Operation:  operation,
		Key:        key,
		Field:      field,
		Value:      value,
		SourceNode: sourceNode,
		Timestamp:  time.Now().UnixNano(),
	}

	resp, err := client.ReplicateOperation(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC ReplicateOperation failed: %v", err)
	}

	if !resp.Success {
		return fmt.Errorf("remote replication failed: %s", resp.Error)
	}

	return nil
}

// Close closes all gRPC connections
func (c *GRPCClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errors []error
	for target, conn := range c.connections {
		if err := conn.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close connection to %s: %v", target, err))
		}
		delete(c.connections, target)
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing connections: %v", errors)
	}

	return nil
}

// SetTimeout updates the gRPC call timeout
func (c *GRPCClient) SetTimeout(timeout time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.timeout = timeout
}

// GetConnectionCount returns the number of active connections
func (c *GRPCClient) GetConnectionCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.connections)
}

// CleanupStaleConnections removes connections to nodes that are no longer active
func (c *GRPCClient) CleanupStaleConnections(activeNodes []*Node) {
	c.mu.Lock()
	defer c.mu.Unlock()

	activeTargets := make(map[string]bool)
	for _, node := range activeNodes {
		target := fmt.Sprintf("%s%s", node.Address, node.Port)
		activeTargets[target] = true
	}

	for target, conn := range c.connections {
		if !activeTargets[target] {
			log.Printf("Closing stale gRPC connection to %s", target)
			conn.Close()
			delete(c.connections, target)
		}
	}
}
