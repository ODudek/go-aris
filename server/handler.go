package server

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/ODudek/go-aris/cluster"
	"github.com/ODudek/go-aris/commands"
	"github.com/ODudek/go-aris/storage"
)

type Handler interface {
	Start() error
	Stop() error
}

type Config struct {
	Port           string
	Storage        *storage.Storage
	ClusterStorage *storage.ClusterStorage
	Parser         *commands.Parser
	Discovery      *cluster.Discovery
	GRPCServer     *cluster.GRPCServer
	IsClusterMode  bool
}

func NewConfig(port string) *Config {
	localStorage := storage.NewStorage()
	parser := commands.NewParser(localStorage)

	// Check if cluster mode is enabled
	clusterMode := os.Getenv("CLUSTER_ENABLED") == "true"

	config := &Config{
		Port:          port,
		Storage:       localStorage,
		Parser:        parser,
		IsClusterMode: clusterMode,
	}

	if clusterMode {
		// Initialize cluster components
		nodeID := os.Getenv("POD_NAME")
		if nodeID == "" {
			nodeID = "node-standalone"
		}

		address := os.Getenv("POD_IP")
		if address == "" {
			address = "127.0.0.1"
		}

		discovery := cluster.NewDiscovery(nodeID, address, port)
		clusterStorage := storage.NewClusterStorage(nodeID, discovery)
		grpcServer := cluster.NewGRPCServer(nodeID, address, port, clusterStorage, discovery)

		config.Discovery = discovery
		config.ClusterStorage = clusterStorage
		config.GRPCServer = grpcServer

		// Update parser to use cluster storage
		config.Parser = commands.NewParser(clusterStorage)
	}

	return config
}

// HealthCheckHandler handles health check requests
func (c *Config) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"cluster":   c.IsClusterMode,
	}

	if c.IsClusterMode && c.ClusterStorage != nil {
		clusterStatus := c.ClusterStorage.GetClusterStatus()
		status["cluster_info"] = clusterStatus
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}

// InternalSetHandler handles internal SET replication requests
func (c *Config) InternalSetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.Header.Get("X-Key")
	value := r.Header.Get("X-Value")

	if key == "" || value == "" {
		http.Error(w, "Missing key or value", http.StatusBadRequest)
		return
	}

	// Store directly in local storage (this is internal replication)
	c.Storage.Set(key, value)
	w.WriteHeader(http.StatusOK)
}

// InternalGetHandler handles internal GET requests
func (c *Config) InternalGetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing key", http.StatusBadRequest)
		return
	}

	value, exists, err := c.Storage.Get(key)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}

	w.Header().Set("X-Value", value)
	w.WriteHeader(http.StatusOK)
}

// InternalHSetHandler handles internal HSET replication requests
func (c *Config) InternalHSetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.Header.Get("X-Key")
	field := r.Header.Get("X-Field")
	value := r.Header.Get("X-Value")

	if key == "" || field == "" || value == "" {
		http.Error(w, "Missing key, field, or value", http.StatusBadRequest)
		return
	}

	c.Storage.HSet(key, field, value)
	w.WriteHeader(http.StatusOK)
}

// InternalHGetHandler handles internal HGET requests
func (c *Config) InternalHGetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	field := r.URL.Query().Get("field")

	if key == "" || field == "" {
		http.Error(w, "Missing key or field", http.StatusBadRequest)
		return
	}

	value, exists, err := c.Storage.HGet(key, field)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Key or field not found", http.StatusNotFound)
		return
	}

	w.Header().Set("X-Value", value)
	w.WriteHeader(http.StatusOK)
}

// StartCluster initializes cluster components if in cluster mode
func (c *Config) StartCluster(ctx context.Context) error {
	if !c.IsClusterMode {
		return nil
	}

	if c.Discovery != nil {
		if err := c.Discovery.Start(ctx); err != nil {
			return err
		}
	}

	if c.ClusterStorage != nil {
		if err := c.ClusterStorage.Start(ctx); err != nil {
			return err
		}
	}

	if c.GRPCServer != nil {
		if err := c.GRPCServer.Start(ctx); err != nil {
			return err
		}
	}

	return nil
}
