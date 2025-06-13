# Integration Guide: Replacing goleveldb with C++ LevelDB in Hyperledger Fabric

This guide explains how to integrate the C++ LevelDB wrapper into Hyperledger Fabric to replace goleveldb.

## Prerequisites

- Hyperledger Fabric 2.5.x source code
- Google LevelDB C++ library installed
- CMake 3.10+
- Go 1.19+
- GCC/Clang with C++11 support

## Step 1: Build the C++ LevelDB Wrapper

```bash
cd cpp-leveldb-wrapper
chmod +x build.sh
./build.sh
```

This will create `build/lib/libcpp_leveldb_wrapper.so`

## Step 2: Install the Shared Library

```bash
# Option 1: Install system-wide
sudo cp build/lib/libcpp_leveldb_wrapper.so /usr/local/lib/
sudo ldconfig

# Option 2: Add to library path
export LD_LIBRARY_PATH=/path/to/cpp-leveldb-wrapper/build/lib:$LD_LIBRARY_PATH
```

## Step 3: Modify Fabric's State Database

Navigate to your Fabric source directory and follow these steps:

### 3.1 Add the Go wrapper to Fabric

Copy the Go wrapper files to Fabric:

```bash
mkdir -p $FABRIC_SRC/core/ledger/kvledger/txmgmt/statedb/cppleveldb
cp go/*.go $FABRIC_SRC/core/ledger/kvledger/txmgmt/statedb/cppleveldb/
```

### 3.2 Update Fabric's go.mod

Add the following to `$FABRIC_SRC/go.mod`:

```go
replace github.com/fabric/cpp-leveldb-wrapper => ./core/ledger/kvledger/txmgmt/statedb/cppleveldb
```

### 3.3 Create New State DB Provider

Create `$FABRIC_SRC/core/ledger/kvledger/txmgmt/statedb/cppleveldb/provider.go`:

```go
package cppleveldb

import (
    "github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/statedb"
    leveldb "github.com/fabric/cpp-leveldb-wrapper/go"
)

// NewVersionedDBProvider creates a new C++ LevelDB provider
func NewVersionedDBProvider(dbPath string) (statedb.VersionedDBProvider, error) {
    return leveldb.NewFabricDBProvider(dbPath)
}
```

### 3.4 Modify State Database Factory

Edit `$FABRIC_SRC/core/ledger/kvledger/txmgmt/statedb/statedb.go` or create a new factory file:

```go
// Add this import
import cppleveldb "github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/statedb/cppleveldb"

// Modify the provider creation logic
func createStateDBProvider(stateDBConfig *StateDBConfig) (statedb.VersionedDBProvider, error) {
    if stateDBConfig.StateDatabase == "cppleveldb" {
        return cppleveldb.NewVersionedDBProvider(stateDBConfig.LevelDBPath)
    }
    // ... existing code for other providers
}
```

### 3.5 Update Configuration

Modify `$FABRIC_SRC/sampleconfig/core.yaml`:

```yaml
ledger:
  state:
    stateDatabase: cppleveldb  # Change from "goleveldb" to "cppleveldb"
    couchDBConfig:
      # ... existing couchdb config
    levelDBPath: /var/hyperledger/production/ledgersData/stateLeveldb
```

## Step 4: Rebuild Fabric

```bash
cd $FABRIC_SRC
make clean
make peer
make orderer
```

## Step 5: Test the Integration

### 5.1 Unit Tests

```bash
cd $FABRIC_SRC/core/ledger/kvledger/txmgmt/statedb/cppleveldb
go test -v
```

### 5.2 Integration Test with Test Network

```bash
cd $FABRIC_SRC/../fabric-samples/test-network
./network.sh down
./network.sh up createChannel -ca
./network.sh deployCC -ccn basic -ccp ../asset-transfer-basic/chaincode-go -ccl go
```

## Step 6: Performance Comparison

You can compare performance between goleveldb and the C++ LevelDB wrapper:

```bash
# Test with original goleveldb
./network.sh down
# Edit core.yaml to use "goleveldb"
./network.sh up createChannel -ca
# Run your performance tests

# Test with C++ LevelDB wrapper  
./network.sh down
# Edit core.yaml to use "cppleveldb"
./network.sh up createChannel -ca
# Run the same performance tests
```

## Troubleshooting

### CGO Compilation Issues

If you encounter CGO compilation issues:

```bash
export CGO_CFLAGS="-I/path/to/cpp-leveldb-wrapper/include"
export CGO_LDFLAGS="-L/path/to/cpp-leveldb-wrapper/build/lib -lcpp_leveldb_wrapper -lleveldb -lstdc++"
```

### Runtime Library Loading Issues

```bash
export LD_LIBRARY_PATH="/path/to/cpp-leveldb-wrapper/build/lib:$LD_LIBRARY_PATH"
# Or install the library system-wide:
sudo cp build/lib/libcpp_leveldb_wrapper.so /usr/local/lib/
sudo ldconfig
```

### Debug Mode

To enable debug logging in the wrapper:

```go
// Add to your main function or init
import "log"
log.SetFlags(log.LstdFlags | log.Lshortfile)
```

## Advanced Configuration

### Custom LevelDB Options

You can modify the LevelDB options in `fabric_adapter.go`:

```go
options := &Options{
    CreateIfMissing: true,
    WriteBufferSize: 8 * 1024 * 1024, // 8MB instead of 4MB
    MaxOpenFiles:    2000,             // More open files
    BlockSize:       8192,             // Larger block size
    Compression:     1,                // Snappy compression
}
```

### Memory Usage Optimization

For high-throughput scenarios:

```go
options := &Options{
    CreateIfMissing: true,
    WriteBufferSize: 16 * 1024 * 1024, // 16MB buffer
    MaxOpenFiles:    5000,              // More file handles
    BlockSize:       16384,             // Larger blocks
    Compression:     1,                 // Snappy
}
```

## Performance Tuning

### Operating System Level

```bash
# Increase file descriptor limits
echo "* soft nofile 65536" >> /etc/security/limits.conf
echo "* hard nofile 65536" >> /etc/security/limits.conf

# Optimize disk I/O
echo mq-deadline > /sys/block/sda/queue/scheduler  # Replace sda with your disk
```

### LevelDB Specific

```bash
# Ensure proper filesystem (ext4 or xfs recommended)
# Use SSDs for better performance
# Consider RAID 0 for write-heavy workloads
```

## Monitoring and Metrics

The wrapper provides basic metrics through LevelDB's property interface:

```go
// Get statistics
stats := db.PropertyValue("leveldb.stats")
fmt.Println("LevelDB Stats:", stats)

// Get approximate sizes
// Implementation depends on your monitoring needs
```

## Rollback Plan

If you need to rollback to goleveldb:

1. Stop all Fabric components
2. Change `stateDatabase` back to `"goleveldb"` in core.yaml
3. Rebuild Fabric without the C++ wrapper
4. Restart Fabric components

Note: The data format is compatible, so no data migration is needed.

## Example Implementation Files

### Provider Implementation

Create `$FABRIC_SRC/core/ledger/kvledger/txmgmt/statedb/cppleveldb/provider.go`:

```go
package cppleveldb

import (
    "github.com/hyperledger/fabric/core/ledger/internal/version"
    "github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/statedb"
    "github.com/pkg/errors"
)

// VersionedDBProvider implements statedb.VersionedDBProvider
type VersionedDBProvider struct {
    dbProvider *FabricDBProvider
}

// NewVersionedDBProvider creates a new C++ LevelDB provider
func NewVersionedDBProvider(dbPath string) (statedb.VersionedDBProvider, error) {
    provider, err := NewFabricDBProvider(dbPath)
    if err != nil {
        return nil, errors.Wrap(err, "failed to create C++ LevelDB provider")
    }
    
    return &VersionedDBProvider{
        dbProvider: provider,
    }, nil
}

// GetDBHandle returns a handle to a VersionedDB
func (provider *VersionedDBProvider) GetDBHandle(id string, namespaceProvider statedb.NamespaceProvider) (statedb.VersionedDB, error) {
    return provider.dbProvider.GetDBHandle(id, namespaceProvider)
}

// ImportFromSnapshot loads data from snapshot
func (provider *VersionedDBProvider) ImportFromSnapshot(id string, savepoint *version.Height, itr statedb.FullScanIterator) error {
    return provider.dbProvider.ImportFromSnapshot(id, convertHeight(savepoint), convertIterator(itr))
}

// BytesKeySupported returns true as LevelDB supports bytes keys
func (provider *VersionedDBProvider) BytesKeySupported() bool {
    return provider.dbProvider.BytesKeySupported()
}

// Close closes the provider
func (provider *VersionedDBProvider) Close() {
    provider.dbProvider.Close()
}

// Drop drops the database
func (provider *VersionedDBProvider) Drop(id string) error {
    return provider.dbProvider.Drop(id)
}

// Helper functions to convert between Fabric types and wrapper types
func convertHeight(h *version.Height) *Height {
    if h == nil {
        return nil
    }
    return &Height{
        BlockNum: h.BlockNum,
        TxNum:    h.TxNum,
    }
}

func convertIterator(itr statedb.FullScanIterator) FullScanIterator {
    // Implementation depends on your iterator conversion needs
    return nil
}
```

### Factory Integration

Modify `$FABRIC_SRC/core/ledger/kvledger/kvledger.go`:

```go
import (
    // ... existing imports
    cppleveldb "github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/statedb/cppleveldb"
)

func newStateDBProvider(conf *ledger.Config) (statedb.VersionedDBProvider, error) {
    switch conf.StateDBConfig.StateDatabase {
    case "cppleveldb":
        return cppleveldb.NewVersionedDBProvider(conf.StateDBConfig.LevelDBPath)
    case "goleveldb":
        return stateleveldb.NewVersionedDBProvider(conf.StateDBConfig.LevelDBPath)
    case "CouchDB":
        return statecouchdb.NewVersionedDBProvider(conf.StateDBConfig.CouchDBConfig, conf.StateDBConfig.LevelDBPath)
    default:
        return nil, errors.Errorf("unsupported state database: %s", conf.StateDBConfig.StateDatabase)
    }
}
```

This guide provides a complete integration path for replacing goleveldb with the C++ LevelDB wrapper in Hyperledger Fabric.
