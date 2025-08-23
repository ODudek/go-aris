# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Aris is a simple in-memory database fully compatible with Redis protocol. It's written in Go and implements the Redis Serialization Protocol (RESP) for client-server communication.

## Architecture

The codebase is organized into several key components:

- **Main entry point** (`main.go`): Command-line interface with server selection
- **Server package** (`server/`): High-performance and standard server implementations
  - `handler.go`: Server interface and configuration
  - `gnet_server.go`: Event-driven server using gnet for high performance
  - `standard_server.go`: Standard Go net server for compatibility
- **Storage package** (`storage/`): Thread-safe in-memory data storage
  - `storage.go`: Implements key-value and hash storage with mutex protection
- **Commands package** (`commands/`): Complete Redis command implementations
  - `command.go`: Command interface definition
  - `parser.go`: RESP to command parsing with argument validation
  - Individual command files: `ping.go`, `set.go`, `get.go`, `hset.go`, `hget.go`, `hgetall.go`
- **RESP package** (`resp/`): Complete Redis Serialization Protocol implementation
  - `reader.go`: Parses incoming RESP messages (arrays, bulk strings)
  - `writer.go`: Formats outgoing RESP responses
  - `value.go`: Defines the Value interface for RESP data types
  - Type implementations: `arr.go`, `bulk.go`, `str.go`, `int.go`, `null.go`, `err.go`

## Development Commands

### Running the Server
```bash
# With gnet (high performance - default)
go run .

# With standard Go net (fallback)
go run . -gnet=false

# Custom port
go run . -port=:9000
```

### Testing
```bash
go test ./...
```

### Building
```bash
go build .
```

### Running Individual Tests
```bash
go test ./commands -v
go test ./storage -v
go test ./resp -v
```

## Current State

Production-ready Redis-compatible server with the following features:
- **Complete RESP protocol implementation** with all data types
- **Full Redis command support**: PING, SET, GET, HSET, HGET, HGETALL
- **High-performance gnet server** for handling thousands of concurrent connections
- **Thread-safe in-memory storage** with proper synchronization
- **Comprehensive error handling** with proper Redis error responses
- **Extensive test coverage** for all components

## Key Implementation Notes

- **Dual server architecture**: Choose between gnet (high performance) and standard Go net
- **Event-driven processing**: No goroutine-per-connection overhead with gnet
- **Thread-safe storage**: Uses RWMutex for optimal read/write performance
- **Command pattern**: Each Redis command is a separate, testable component
- **Proper RESP protocol**: Full compatibility with Redis clients like redis-cli
- **Memory efficiency**: Optimized byte handling and minimal allocations

## Performance Characteristics

- **gnet server**: Event-driven, handles 10k+ concurrent connections efficiently
- **Standard server**: Goroutine-per-connection, suitable for moderate loads
- **Storage**: Thread-safe with separate key-value and hash storage
- **RESP parsing**: Zero-copy where possible, efficient buffer management

## Next Development Steps

- Add data persistence (RDB/AOF file formats)
- Implement additional Redis commands (DEL, EXISTS, EXPIRE, etc.)
- Add pub/sub functionality
- Performance benchmarking against Redis
- Clustering support