package cluster

import (
	"context"
	"time"
)

// Simplified gRPC message structures for cluster communication
// These replace the protobuf-generated structures for now

// Data operation messages
type SetRequest struct {
	Key           string
	Value         string
	IsReplication bool
	SourceNode    string
	Timestamp     int64
}

type SetResponse struct {
	Success bool
	Error   string
}

type GetRequest struct {
	Key            string
	RequestingNode string
}

type GetResponse struct {
	Value string
	Found bool
	Error string
}

type HSetRequest struct {
	Key           string
	Field         string
	Value         string
	IsReplication bool
	SourceNode    string
	Timestamp     int64
}

type HSetResponse struct {
	Success bool
	Error   string
}

type HGetRequest struct {
	Key            string
	Field          string
	RequestingNode string
}

type HGetResponse struct {
	Value string
	Found bool
	Error string
}

type HGetAllRequest struct {
	Key            string
	RequestingNode string
}

type HGetAllResponse struct {
	Fields map[string]string
	Error  string
}

type DeleteRequest struct {
	Key           string
	IsReplication bool
	SourceNode    string
	Timestamp     int64
}

type DeleteResponse struct {
	Success bool
	Error   string
}

// Cluster management messages
type PingRequest struct {
	NodeId    string
	Timestamp int64
}

type PingResponse struct {
	NodeId    string
	Status    string
	Timestamp int64
	Version   string
}

type NodeInfoRequest struct {
	RequestingNode string
}

type NodeInfoResponse struct {
	NodeId  string
	Address string
	Port    string
	Status  string
	Uptime  int64
	Metrics *NodeMetrics
	Roles   []string
}

type NodeMetrics struct {
	TotalKeys           int64
	TotalMemoryUsage    int64
	Connections         int64
	OperationsPerSecond int64
	CpuUsage            float64
	MemoryUsagePercent  float64
}

type ClusterStatusRequest struct {
	RequestingNode string
}

type ClusterStatusResponse struct {
	TotalNodes        int32
	ActiveNodes       int32
	Nodes             []*NodeInfo
	IsHealthy         bool
	ReplicationFactor int32
	ClusterId         string
}

type NodeInfo struct {
	NodeId   string
	Address  string
	Port     string
	Status   string
	LastSeen int64
	Metrics  *NodeMetrics
}

type JoinRequest struct {
	NodeId   string
	Address  string
	Port     string
	Version  string
	Metadata map[string]string
}

type JoinResponse struct {
	Accepted      bool
	ClusterId     string
	ExistingNodes []*NodeInfo
	Error         string
	Config        *ClusterConfig
}

type ClusterConfig struct {
	ReplicationFactor          int32
	HashRingReplicas           int32
	SyncIntervalSeconds        int32
	HealthCheckIntervalSeconds int32
}

type LeaveRequest struct {
	NodeId   string
	Graceful bool
}

type LeaveResponse struct {
	Success bool
	Error   string
}

// Data synchronization messages
type SyncDataRequest struct {
	RequestingNode string
	KeyPrefixes    []string
	SinceTimestamp int64
	BatchSize      int32
}

type SyncDataResponse struct {
	Data              []*KeyValue
	HashData          []*HashKeyValue
	HasMore           bool
	ContinuationToken string
	Error             string
}

type KeyValue struct {
	Key       string
	Value     string
	Timestamp int64
}

type HashKeyValue struct {
	Key       string
	Fields    map[string]string
	Timestamp int64
}

type ReplicateRequest struct {
	Operation  string // SET, HSET, DELETE, etc.
	Key        string
	Field      string // for hash operations
	Value      string
	SourceNode string
	Timestamp  int64
	Metadata   map[string]string
}

type ReplicateResponse struct {
	Success bool
	Error   string
	NodeId  string
}

type TransferKeysRequest struct {
	TargetNode          string
	Keys                []string
	DeleteAfterTransfer bool
}

type TransferKeysResponse struct {
	TransferredCount int32
	FailedKeys       []string
	Error            string
}

// Health and monitoring messages
type HealthCheckRequest struct {
	RequestingNode string
	Timestamp      int64
}

type HealthCheckResponse struct {
	Status    string // healthy, unhealthy, degraded
	Metrics   *NodeMetrics
	Issues    []string
	Timestamp int64
}

type MetricsRequest struct {
	RequestingNode string
	MetricNames    []string
}

type MetricsResponse struct {
	Metrics     map[string]float64
	NodeMetrics *NodeMetrics
	Timestamp   int64
}

// Simplified gRPC service interface
type ClusterServiceClient interface {
	// Data operations
	Set(ctx context.Context, req *SetRequest) (*SetResponse, error)
	Get(ctx context.Context, req *GetRequest) (*GetResponse, error)
	HSet(ctx context.Context, req *HSetRequest) (*HSetResponse, error)
	HGet(ctx context.Context, req *HGetRequest) (*HGetResponse, error)
	HGetAll(ctx context.Context, req *HGetAllRequest) (*HGetAllResponse, error)
	Delete(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error)

	// Cluster management
	Ping(ctx context.Context, req *PingRequest) (*PingResponse, error)
	NodeInfo(ctx context.Context, req *NodeInfoRequest) (*NodeInfoResponse, error)
	ClusterStatus(ctx context.Context, req *ClusterStatusRequest) (*ClusterStatusResponse, error)
	Join(ctx context.Context, req *JoinRequest) (*JoinResponse, error)
	Leave(ctx context.Context, req *LeaveRequest) (*LeaveResponse, error)

	// Data synchronization
	SyncData(ctx context.Context, req *SyncDataRequest) (*SyncDataResponse, error)
	ReplicateOperation(ctx context.Context, req *ReplicateRequest) (*ReplicateResponse, error)
	TransferKeys(ctx context.Context, req *TransferKeysRequest) (*TransferKeysResponse, error)

	// Health and monitoring
	HealthCheck(ctx context.Context, req *HealthCheckRequest) (*HealthCheckResponse, error)
	GetMetrics(ctx context.Context, req *MetricsRequest) (*MetricsResponse, error)
}

type ClusterServiceServer interface {
	// Data operations
	Set(context.Context, *SetRequest) (*SetResponse, error)
	Get(context.Context, *GetRequest) (*GetResponse, error)
	HSet(context.Context, *HSetRequest) (*HSetResponse, error)
	HGet(context.Context, *HGetRequest) (*HGetResponse, error)
	HGetAll(context.Context, *HGetAllRequest) (*HGetAllResponse, error)
	Delete(context.Context, *DeleteRequest) (*DeleteResponse, error)

	// Cluster management
	Ping(context.Context, *PingRequest) (*PingResponse, error)
	NodeInfo(context.Context, *NodeInfoRequest) (*NodeInfoResponse, error)
	ClusterStatus(context.Context, *ClusterStatusRequest) (*ClusterStatusResponse, error)
	Join(context.Context, *JoinRequest) (*JoinResponse, error)
	Leave(context.Context, *LeaveRequest) (*LeaveResponse, error)

	// Data synchronization
	SyncData(context.Context, *SyncDataRequest) (*SyncDataResponse, error)
	ReplicateOperation(context.Context, *ReplicateRequest) (*ReplicateResponse, error)
	TransferKeys(context.Context, *TransferKeysRequest) (*TransferKeysResponse, error)

	// Health and monitoring
	HealthCheck(context.Context, *HealthCheckRequest) (*HealthCheckResponse, error)
	GetMetrics(context.Context, *MetricsRequest) (*MetricsResponse, error)
}

// Helper functions for creating responses
func NewSetResponse(success bool, err error) *SetResponse {
	resp := &SetResponse{Success: success}
	if err != nil {
		resp.Error = err.Error()
	}
	return resp
}

func NewGetResponse(value string, found bool, err error) *GetResponse {
	resp := &GetResponse{Value: value, Found: found}
	if err != nil {
		resp.Error = err.Error()
	}
	return resp
}

func NewHSetResponse(success bool, err error) *HSetResponse {
	resp := &HSetResponse{Success: success}
	if err != nil {
		resp.Error = err.Error()
	}
	return resp
}

func NewHGetResponse(value string, found bool, err error) *HGetResponse {
	resp := &HGetResponse{Value: value, Found: found}
	if err != nil {
		resp.Error = err.Error()
	}
	return resp
}

func NewHGetAllResponse(fields map[string]string, err error) *HGetAllResponse {
	resp := &HGetAllResponse{Fields: fields}
	if err != nil {
		resp.Error = err.Error()
	}
	return resp
}

func NewPingResponse(nodeId, status, version string) *PingResponse {
	return &PingResponse{
		NodeId:    nodeId,
		Status:    status,
		Timestamp: time.Now().UnixNano(),
		Version:   version,
	}
}

func NewHealthCheckResponse(status string, metrics *NodeMetrics, issues []string) *HealthCheckResponse {
	return &HealthCheckResponse{
		Status:    status,
		Metrics:   metrics,
		Issues:    issues,
		Timestamp: time.Now().UnixNano(),
	}
}

// Dummy client implementation for compilation
type dummyClusterServiceClient struct{}

func NewClusterServiceClient(conn interface{}) ClusterServiceClient {
	return &dummyClusterServiceClient{}
}

func (c *dummyClusterServiceClient) Set(ctx context.Context, req *SetRequest) (*SetResponse, error) {
	return NewSetResponse(false, nil), nil
}

func (c *dummyClusterServiceClient) Get(ctx context.Context, req *GetRequest) (*GetResponse, error) {
	return NewGetResponse("", false, nil), nil
}

func (c *dummyClusterServiceClient) HSet(ctx context.Context, req *HSetRequest) (*HSetResponse, error) {
	return NewHSetResponse(false, nil), nil
}

func (c *dummyClusterServiceClient) HGet(ctx context.Context, req *HGetRequest) (*HGetResponse, error) {
	return NewHGetResponse("", false, nil), nil
}

func (c *dummyClusterServiceClient) HGetAll(ctx context.Context, req *HGetAllRequest) (*HGetAllResponse, error) {
	return NewHGetAllResponse(make(map[string]string), nil), nil
}

func (c *dummyClusterServiceClient) Delete(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error) {
	return &DeleteResponse{Success: false}, nil
}

func (c *dummyClusterServiceClient) Ping(ctx context.Context, req *PingRequest) (*PingResponse, error) {
	return NewPingResponse("dummy", "active", "1.0.0"), nil
}

func (c *dummyClusterServiceClient) NodeInfo(ctx context.Context, req *NodeInfoRequest) (*NodeInfoResponse, error) {
	return &NodeInfoResponse{}, nil
}

func (c *dummyClusterServiceClient) ClusterStatus(ctx context.Context, req *ClusterStatusRequest) (*ClusterStatusResponse, error) {
	return &ClusterStatusResponse{}, nil
}

func (c *dummyClusterServiceClient) Join(ctx context.Context, req *JoinRequest) (*JoinResponse, error) {
	return &JoinResponse{}, nil
}

func (c *dummyClusterServiceClient) Leave(ctx context.Context, req *LeaveRequest) (*LeaveResponse, error) {
	return &LeaveResponse{}, nil
}

func (c *dummyClusterServiceClient) SyncData(ctx context.Context, req *SyncDataRequest) (*SyncDataResponse, error) {
	return &SyncDataResponse{}, nil
}

func (c *dummyClusterServiceClient) ReplicateOperation(ctx context.Context, req *ReplicateRequest) (*ReplicateResponse, error) {
	return &ReplicateResponse{}, nil
}

func (c *dummyClusterServiceClient) TransferKeys(ctx context.Context, req *TransferKeysRequest) (*TransferKeysResponse, error) {
	return &TransferKeysResponse{}, nil
}

func (c *dummyClusterServiceClient) HealthCheck(ctx context.Context, req *HealthCheckRequest) (*HealthCheckResponse, error) {
	return NewHealthCheckResponse("healthy", &NodeMetrics{}, []string{}), nil
}

func (c *dummyClusterServiceClient) GetMetrics(ctx context.Context, req *MetricsRequest) (*MetricsResponse, error) {
	return &MetricsResponse{}, nil
}
