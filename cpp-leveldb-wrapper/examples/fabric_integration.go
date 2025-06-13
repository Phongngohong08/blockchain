package main

import (
	"fmt"
	"log"
	"os"

	leveldb "github.com/fabric/cpp-leveldb-wrapper/go"
)

func main() {
	// Create temporary directory for example
	tmpDir := "/tmp/fabric_leveldb_example"
	defer os.RemoveAll(tmpDir)

	fmt.Println("=== Fabric LevelDB C++ Wrapper Example ===")

	// 1. Create provider (equivalent to Fabric's VersionedDBProvider)
	fmt.Println("\n1. Creating Fabric DB Provider...")
	provider, err := leveldb.NewFabricDBProvider(tmpDir)
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Close()

	// 2. Get database handle for a channel
	fmt.Println("2. Getting database handle for channel 'mychannel'...")
	db, err := provider.GetDBHandle("mychannel", nil)
	if err != nil {
		log.Fatalf("Failed to get database handle: %v", err)
	}

	// 3. Simulate chaincode operations
	fmt.Println("3. Simulating chaincode operations...")

	namespace := "mycc" // chaincode name
	key1 := "asset1"
	key2 := "asset2"

	// Create versioned values (like what Fabric would create)
	value1 := &leveldb.VersionedValue{
		Value:    []byte(`{"name":"Asset1","owner":"Alice","value":100}`),
		Metadata: nil,
		Version: &leveldb.Height{
			BlockNum: 1,
			TxNum:    1,
		},
	}

	value2 := &leveldb.VersionedValue{
		Value:    []byte(`{"name":"Asset2","owner":"Bob","value":200}`),
		Metadata: nil,
		Version: &leveldb.Height{
			BlockNum: 1,
			TxNum:    2,
		},
	}

	// 4. Create update batch (like what Fabric does during block commit)
	fmt.Println("4. Creating update batch...")
	batch := &leveldb.UpdateBatch{
		Updates: map[string]map[string]*leveldb.VersionedValue{
			namespace: {
				key1: value1,
				key2: value2,
			},
		},
	}

	// 5. Apply updates (commit the block)
	fmt.Println("5. Applying updates (committing block)...")
	height := &leveldb.Height{BlockNum: 1, TxNum: 2}
	err = db.ApplyUpdates(batch, height)
	if err != nil {
		log.Fatalf("Failed to apply updates: %v", err)
	}

	// 6. Query individual states
	fmt.Println("6. Querying individual states...")

	retrievedValue1, err := db.GetState(namespace, key1)
	if err != nil {
		log.Fatalf("Failed to get state for %s: %v", key1, err)
	}
	fmt.Printf("   %s: %s (Block: %d, Tx: %d)\n",
		key1, string(retrievedValue1.Value),
		retrievedValue1.Version.BlockNum,
		retrievedValue1.Version.TxNum)

	retrievedValue2, err := db.GetState(namespace, key2)
	if err != nil {
		log.Fatalf("Failed to get state for %s: %v", key2, err)
	}
	fmt.Printf("   %s: %s (Block: %d, Tx: %d)\n",
		key2, string(retrievedValue2.Value),
		retrievedValue2.Version.BlockNum,
		retrievedValue2.Version.TxNum)

	// 7. Range scan (like what happens during range queries)
	fmt.Println("7. Performing range scan...")
	iter, err := db.GetStateRangeScanIterator(namespace, "", "")
	if err != nil {
		log.Fatalf("Failed to create iterator: %v", err)
	}
	defer iter.Close()

	count := 0
	for {
		kv, err := iter.Next()
		if err != nil {
			log.Fatalf("Iterator error: %v", err)
		}
		if kv == nil {
			break // End of iteration
		}

		fmt.Printf("   Found: %s = %s (Block: %d, Tx: %d)\n",
			kv.CompositeKey.Key,
			string(kv.VersionedValue.Value),
			kv.VersionedValue.Version.BlockNum,
			kv.VersionedValue.Version.TxNum)
		count++
	}
	fmt.Printf("   Total entries found: %d\n", count)

	// 8. Check latest save point
	fmt.Println("8. Checking latest save point...")
	savePoint, err := db.GetLatestSavePoint()
	if err != nil {
		log.Fatalf("Failed to get save point: %v", err)
	}
	fmt.Printf("   Latest save point: Block %d, Tx %d\n",
		savePoint.BlockNum, savePoint.TxNum)

	// 9. Simulate delete operation
	fmt.Println("9. Simulating delete operation...")
	deleteValue := &leveldb.VersionedValue{
		Value:   nil, // nil indicates delete
		Version: &leveldb.Height{BlockNum: 2, TxNum: 1},
	}

	deleteBatch := &leveldb.UpdateBatch{
		Updates: map[string]map[string]*leveldb.VersionedValue{
			namespace: {
				key1: deleteValue,
			},
		},
	}

	newHeight := &leveldb.Height{BlockNum: 2, TxNum: 1}
	err = db.ApplyUpdates(deleteBatch, newHeight)
	if err != nil {
		log.Fatalf("Failed to apply delete batch: %v", err)
	}

	// Verify deletion
	deletedValue, err := db.GetState(namespace, key1)
	if err != nil {
		log.Fatalf("Failed to check deleted state: %v", err)
	}
	if deletedValue != nil {
		fmt.Printf("   Warning: %s still exists after delete\n", key1)
	} else {
		fmt.Printf("   Successfully deleted %s\n", key1)
	}

	// 10. Multiple keys query
	fmt.Println("10. Querying multiple keys...")
	keys := []string{key1, key2, "nonexistent"}
	values, err := db.GetStateMultipleKeys(namespace, keys)
	if err != nil {
		log.Fatalf("Failed to get multiple keys: %v", err)
	}

	for i, key := range keys {
		if values[i] != nil {
			fmt.Printf("    %s: %s\n", key, string(values[i].Value))
		} else {
			fmt.Printf("    %s: <nil>\n", key)
		}
	}

	fmt.Println("\n=== Example completed successfully! ===")
}
