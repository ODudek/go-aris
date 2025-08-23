# Aris

![Go](https://github.com/ODudek/go-aris/workflows/Go/badge.svg)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/ODudek/go-aris)](https://goreportcard.com/report/github.com/ODudek/go-aris)

Aris is a simple in-memory database fully compatible with the Redis protocol, written in Go. It implements the Redis Serialization Protocol (RESP) for client-server communication.

## Features

- ✅ RESP (Redis Serialization Protocol) parsing
- ✅ High-performance gnet-based server
- ✅ Standard Go net server fallback
- ✅ Redis commands: SET, GET, PING, HSET, HGET, HGETALL
- ✅ Thread-safe in-memory storage
- ✅ Concurrent client connections
- 🔲 Data persistence
- 🔲 Additional Redis commands
- 🔲 Pub/Sub support

## Quick Start

### Installation

```bash
git clone https://github.com/ODudek/go-aris.git
cd go-aris
go build .
```

### Running the Server

#### With gnet (high performance - default)
```bash
go run .
# or explicitly
go run . -gnet=true -port=:8089
```

#### With standard Go net (fallback)
```bash
go run . -gnet=false -port=:8089
```

The server will start on port `8089` and accept Redis-compatible commands.

### Testing with Redis CLI

```bash
redis-cli -p 8089
```

## Development

### Prerequisites

- Go 1.20 or higher

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build .
```

## Project Structure

```
.
├── main.go           # Application entry point
├── server/           # Server implementations
│   ├── handler.go    # Server interface and config
│   ├── gnet_server.go        # High-performance gnet server
│   └── standard_server.go    # Standard Go net server
├── commands/         # Redis command handlers
│   ├── command.go    # Command interface
│   ├── parser.go     # Command parsing logic
│   ├── ping.go       # PING command
│   ├── set.go        # SET command
│   ├── get.go        # GET command
│   ├── hset.go       # HSET command
│   ├── hget.go       # HGET command
│   └── hgetall.go    # HGETALL command
├── storage/          # In-memory storage
│   └── storage.go    # Thread-safe storage implementation
└── resp/            # Redis Serialization Protocol
    ├── reader.go     # RESP message parsing
    ├── writer.go     # RESP message formatting
    └── *.go         # RESP data types (str, int, bulk, array, null, err)
```

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Performance

Aris uses [gnet](https://github.com/panjf2000/gnet) by default for exceptional performance:

- **Event-driven architecture** - No goroutines per connection overhead
- **Multi-core support** - Scales across CPU cores
- **Memory efficient** - Lower memory footprint than standard net
- **High throughput** - Handles thousands of concurrent connections

You can switch to standard Go net server using `-gnet=false` for compatibility.

## Roadmap

- [x] Implement handlers for SET, GET, PING, HSET, HGET, HGETALL commands
- [x] Implement thread-safe in-memory storage
- [x] Support concurrent client connections  
- [x] High-performance server with gnet
- [ ] Add data persistence to file
- [ ] Add more Redis commands (DEL, EXISTS, EXPIRE, etc.)
- [ ] Add comprehensive test coverage
- [ ] Add Docker support
- [ ] Performance benchmarking vs Redis
- [ ] Pub/Sub support
- [ ] Clustering support

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.