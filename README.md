# FlexiHash - Consistent Hashing for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/mysamimi/flexiHash.svg)](https://pkg.go.dev/github.com/mysamimi/flexiHash)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

FlexiHash is a Go implementation of [consistent hashing](https://en.wikipedia.org/wiki/Consistent_hashing), designed to be **fully compatible** with the PHP [flexihash library](https://github.com/pda/flexihash). This allows Go and PHP applications to work together seamlessly using the same consistent hashing algorithm.

## Features

- ✅ **PHP Compatible**: Produces identical hash results as the PHP flexihash library
- 🔄 **Consistent Hashing**: Minimal key redistribution when nodes are added/removed
- ⚖️ **Weighted Targets**: Support for nodes with different capacities
- 🔌 **Pluggable Hash Functions**: Use CRC32 (default) or implement custom hashers
- 🎯 **Configurable Replicas**: Adjust virtual nodes for better distribution
- 🚀 **High Performance**: Efficient binary search for lookups
- 📦 **Zero Dependencies**: Uses only Go standard library
- 🧪 **Well Tested**: Comprehensive test suite

## What is Consistent Hashing?

Consistent hashing is a technique used in distributed systems to distribute data across multiple nodes. When nodes are added or removed, only a minimal amount of data needs to be redistributed, making it ideal for:

- **Distributed caches** (Redis, Memcached clusters)
- **Load balancers**
- **Distributed databases**
- **Content delivery networks**

## Installation

```bash
go get github.com/mysamimi/flexiHash
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    "github.com/mysamimi/flexiHash"
)

func main() {
    // Create a new FlexiHash instance
    hash := flexihash.NewFlexiHash()

    // Add cache servers
    targets := []string{"cache-1", "cache-2", "cache-3"}
    err := hash.AddTargets(targets, 1)
    if err != nil {
        log.Fatal(err)
    }

    // Look up which server should handle an object
    server, err := hash.Lookup("object-a")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("object-a -> %s\n", server) // e.g., "cache-2"

    // Get multiple servers for redundancy
    servers, err := hash.LookupList("object-b", 2)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("object-b -> %v\n", servers) // e.g., ["cache-1", "cache-3"]
}
```

## Usage

### Basic Operations

```go
// Create instance
hash := flexihash.NewFlexiHash()

// Add single target
err := hash.AddTarget("server-1", 1)

// Add multiple targets
targets := []string{"server-1", "server-2", "server-3"}
err := hash.AddTargets(targets, 1)

// Remove target
err := hash.RemoveTarget("server-1")

// Get all targets
allTargets := hash.GetAllTargets()

// Simple lookup
target, err := hash.Lookup("my-key")

// Lookup with fallbacks
targets, err := hash.LookupList("my-key", 3)
```

### Weighted Targets

Assign different weights to targets based on their capacity:

```go
hash := flexihash.NewFlexiHash()

// Small server with weight 1
hash.AddTarget("small-server", 1)

// Large server with weight 2 (2x capacity)
hash.AddTarget("large-server", 2)

// The large server will receive approximately twice as many keys
```

### Custom Configuration

```go
// Custom number of replicas (virtual nodes)
// Higher = better distribution, but more memory
hash := flexihash.NewFlexiHashWithHasher(nil, 128) // 128 instead of default 64

// Use built-in MD5 hasher (compatible with PHP Flexihash MD5)
md5Hasher := &flexihash.Md5Hasher{}
hash := flexihash.NewFlexiHashWithHasher(md5Hasher, 64)

// Custom hash function
type MyHasher struct{}

func (h *MyHasher) Hash(str string) int {
    // Your custom hash implementation
    return customHash(str)
}

customHasher := &MyHasher{}
hash := flexihash.NewFlexiHashWithHasher(customHasher, 64)
```

## PHP Interoperability

This library is designed to produce **identical results** to the PHP flexihash library when using the same configuration.

### Go Example

```go
hash := flexihash.NewFlexiHash() // Default: 64 replicas, CRC32
hash.AddTargets([]string{"cache-1", "cache-2", "cache-3"}, 1)
target, _ := hash.Lookup("object-a")
fmt.Println(target) // e.g., "cache-2"
```

### Equivalent PHP Code

```php
<?php
require 'vendor/autoload.php';

use Flexihash\Flexihash;

$hash = new Flexihash(); // Default: 64 replicas, CRC32
$hash->addTargets(['cache-1', 'cache-2', 'cache-3']);
echo $hash->lookup('object-a'); // "cache-2" (same as Go!)
?>
```

### Verification Script

See [`examples/php_interop/php_interop.go`](examples/php_interop/php_interop.go) for a complete example that generates test vectors for PHP verification.

## API Reference

### Types

#### `FlexiHash`

The main consistent hashing structure.

#### `Hasher` Interface

```go
type Hasher interface {
    Hash(string) int
}
```

Implement this interface to create custom hash functions.

#### Built-in Hashers

**`Crc32Hasher`** (default)
- Uses CRC32 algorithm
- Compatible with PHP `crc32()` function
- Fast and evenly distributed

**`Md5Hasher`**
- Uses MD5 algorithm (first 32 bits)
- Compatible with PHP Flexihash MD5 hasher
- Good for compatibility with existing MD5-based systems

```go
// Use MD5 hasher
md5Hasher := &flexihash.Md5Hasher{}
hash := flexihash.NewFlexiHashWithHasher(md5Hasher, 64)
```

### Functions

#### `NewFlexiHash() *FlexiHash`

Creates a new FlexiHash with default settings:
- 64 replicas (virtual nodes)
- CRC32 hash function

#### `NewFlexiHashWithHasher(hasher Hasher, replicas int) *FlexiHash`

Creates a FlexiHash with custom configuration:
- `hasher`: Custom hash function (nil = use CRC32)
- `replicas`: Number of virtual nodes per target (0 = use 64)

#### `AddTarget(target string, weight float64) error`

Adds a target with the specified weight.
- Returns error if target already exists
- Weight defaults to 1 if set to 0

#### `AddTargets(targets []string, weight float64) error`

Adds multiple targets with the same weight.

#### `RemoveTarget(target string) error`

Removes a target.
- Returns error if target doesn't exist

#### `GetAllTargets() []string`

Returns all currently registered targets.

#### `Lookup(resource string) (string, error)`

Finds the target for a given resource.
- Returns error if no targets exist

#### `LookupList(resource string, count int) ([]string, error)`

Returns multiple targets for a resource, in order of precedence.
- Useful for redundancy (e.g., storing data on multiple servers)
- Returns error if count < 1
- Returns fewer targets if count exceeds available targets

## How It Works

### Consistent Hashing Algorithm

1. **Virtual Nodes**: Each target is hashed multiple times (default: 64) to create "virtual nodes" on a hash ring
2. **Resource Hashing**: When looking up a resource, it's hashed to a position on the ring
3. **Clockwise Search**: The algorithm finds the first virtual node clockwise from the resource's position
4. **Target Mapping**: The virtual node maps back to its physical target

### Why Virtual Nodes?

Virtual nodes (replicas) ensure better distribution:
- **Without**: Adding/removing a node affects adjacent nodes unevenly
- **With**: Load is distributed more evenly across all nodes

### Performance Characteristics

- **Add Target**: O(R) where R = replicas
- **Remove Target**: O(R)
- **Lookup**: O(log N) where N = total virtual nodes (binary search)
- **Memory**: O(T × R) where T = number of targets

## Testing

Run the test suite:

```bash
go test -v
```

Run benchmarks:

```bash
go test -bench=. -benchmem
```

Example benchmark results:
```
BenchmarkAddTarget-8       100000    15420 ns/op    4328 B/op    67 allocs/op
BenchmarkLookup-8         5000000      251 ns/op       0 B/op     0 allocs/op
BenchmarkLookupList-8     3000000      489 ns/op     128 B/op     4 allocs/op
```

## Examples

See the [`examples/`](examples/) directory for complete examples:

- **[basic_usage.go](examples/basic_usage/basic_usage.go)**: Basic operations and features
- **[php_interop.go](examples/php_interop/php_interop.go)**: PHP compatibility demonstration
- **[custom_hasher.go](examples/custom_hasher/custom_hasher.go)**: Custom hash functions and configurations

Run examples:
```bash
go run examples/basic_usage/basic_usage.go
go run examples/php_interop/php_interop.go
go run examples/custom_hasher/custom_hasher.go
```

## Use Cases

### Distributed Cache

```go
hash := flexihash.NewFlexiHash()
hash.AddTargets([]string{
    "cache-server-1:11211",
    "cache-server-2:11211",
    "cache-server-3:11211",
}, 1)

// Determine which cache server stores a user's session
server, _ := hash.Lookup("user:123:session")
// Connect to that server and get/set the value
```

### Load Balancing

```go
hash := flexihash.NewFlexiHash()
hash.AddTarget("backend-1", 1) // Standard server
hash.AddTarget("backend-2", 2) // Powerful server (2x weight)
hash.AddTarget("backend-3", 1)

// Route request based on user ID (same user always goes to same server)
backend, _ := hash.Lookup("user:" + userID)
```

### Sharded Database

```go
hash := flexihash.NewFlexiHash()
hash.AddTargets([]string{"shard-1", "shard-2", "shard-3", "shard-4"}, 1)

// Determine which shard contains a record
shard, _ := hash.Lookup("record:" + recordID)

// For redundancy, store on multiple shards
shards, _ := hash.LookupList("record:" + recordID, 2)
```

## Differences from PHP Version

While this library maintains compatibility with PHP flexihash's hashing algorithm, there are some API differences due to language conventions:

| Feature | PHP | Go |
|---------|-----|-----|
| Error Handling | Exceptions | Return errors |
| Method Naming | camelCase | camelCase (same) |
| Fluent Interface | Yes (method chaining) | No (Go idiom) |
| Type System | Dynamic | Static |

## Contributing

Contributions are welcome! Please ensure:

1. Tests pass: `go test -v`
2. Code is formatted: `go fmt`
3. Benchmarks don't regress significantly
4. PHP compatibility is maintained

## References

- [Original PHP Flexihash](https://github.com/pda/flexihash)
- [Consistent Hashing on Wikipedia](https://en.wikipedia.org/wiki/Consistent_hashing)
- [Consistent Hashing and Random Trees](https://www.akamai.com/us/en/multimedia/documents/technical-publication/consistent-hashing-and-random-trees-distributed-caching-protocols-for-relieving-hot-spots-on-the-world-wide-web-technical-publication.pdf)

## License

MIT License - see [LICENSE](LICENSE) file for details.

This implementation is inspired by and compatible with [pda/flexihash](https://github.com/pda/flexihash) (PHP), also under MIT License.

## Changelog

### v1.0.0 (2024-12-30)

- Initial release
- Full PHP flexihash compatibility
- CRC32 hasher with signed integer support
- Weighted targets
- Custom hasher interface
- Comprehensive test suite
- Examples and documentation

## Support

- 📫 Issues: [GitHub Issues](https://github.com/mysamimi/flexiHash/issues)
- 📚 Documentation: [pkg.go.dev](https://pkg.go.dev/github.com/mysamimi/flexiHash)

---

Made with ❤️ for distributed systems
