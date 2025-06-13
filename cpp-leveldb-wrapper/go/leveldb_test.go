package leveldb

import (
	"os"
	"testing"
)

func TestBasicOperations(t *testing.T) {
	// Create temporary directory for test
	tmpDir := "/tmp/test_leveldb"
	defer os.RemoveAll(tmpDir)

	// Create options
	options := &Options{
		CreateIfMissing: true,
		WriteBufferSize: 1024 * 1024, // 1MB
		MaxOpenFiles:    100,
		BlockSize:       4096,
		Compression:     1, // Snappy
	}

	// Open database
	db, err := Open(tmpDir, options)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test Put and Get
	key := []byte("test_key")
	value := []byte("test_value")

	err = db.Put(nil, key, value)
	if err != nil {
		t.Fatalf("Failed to put value: %v", err)
	}

	retrievedValue, err := db.Get(nil, key)
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}

	if string(retrievedValue) != string(value) {
		t.Errorf("Expected %s, got %s", value, retrievedValue)
	}

	// Test Delete
	err = db.Delete(nil, key)
	if err != nil {
		t.Fatalf("Failed to delete key: %v", err)
	}

	retrievedValue, err = db.Get(nil, key)
	if err != nil {
		t.Fatalf("Failed to get value after delete: %v", err)
	}

	if retrievedValue != nil {
		t.Errorf("Expected nil after delete, got %s", retrievedValue)
	}
}

func TestIterator(t *testing.T) {
	// Create temporary directory for test
	tmpDir := "/tmp/test_leveldb_iter"
	defer os.RemoveAll(tmpDir)

	// Create options
	options := &Options{
		CreateIfMissing: true,
	}

	// Open database
	db, err := Open(tmpDir, options)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Insert test data
	testData := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	for k, v := range testData {
		err = db.Put(nil, []byte(k), []byte(v))
		if err != nil {
			t.Fatalf("Failed to put %s: %v", k, err)
		}
	}

	// Test iterator
	iter := db.NewIterator(nil)
	defer iter.Close()

	count := 0
	for iter.SeekToFirst(); iter.Valid(); iter.Next() {
		key := string(iter.Key())
		value := string(iter.Value())

		expectedValue, exists := testData[key]
		if !exists {
			t.Errorf("Unexpected key: %s", key)
		}

		if value != expectedValue {
			t.Errorf("For key %s, expected %s, got %s", key, expectedValue, value)
		}

		count++
	}

	if count != len(testData) {
		t.Errorf("Expected %d items, got %d", len(testData), count)
	}

	// Check for iterator errors
	if err := iter.Error(); err != nil {
		t.Errorf("Iterator error: %v", err)
	}
}

func TestWriteBatch(t *testing.T) {
	// Create temporary directory for test
	tmpDir := "/tmp/test_leveldb_batch"
	defer os.RemoveAll(tmpDir)

	// Create options
	options := &Options{
		CreateIfMissing: true,
	}

	// Open database
	db, err := Open(tmpDir, options)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create write batch
	batch := NewWriteBatch()
	defer batch.Close()

	// Add operations to batch
	batch.Put([]byte("batch_key1"), []byte("batch_value1"))
	batch.Put([]byte("batch_key2"), []byte("batch_value2"))
	batch.Delete([]byte("batch_key3")) // This key doesn't exist, but delete should not fail

	// Write batch
	err = db.Write(nil, batch)
	if err != nil {
		t.Fatalf("Failed to write batch: %v", err)
	}

	// Verify batch operations
	value1, err := db.Get(nil, []byte("batch_key1"))
	if err != nil {
		t.Fatalf("Failed to get batch_key1: %v", err)
	}
	if string(value1) != "batch_value1" {
		t.Errorf("Expected batch_value1, got %s", value1)
	}

	value2, err := db.Get(nil, []byte("batch_key2"))
	if err != nil {
		t.Fatalf("Failed to get batch_key2: %v", err)
	}
	if string(value2) != "batch_value2" {
		t.Errorf("Expected batch_value2, got %s", value2)
	}
}

func TestFabricAdapter(t *testing.T) {
	// Create temporary directory for test
	tmpDir := "/tmp/test_fabric_adapter"
	defer os.RemoveAll(tmpDir)

	// Create provider
	provider, err := NewFabricDBProvider(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Close()

	// Get database handle
	db, err := provider.GetDBHandle("test_channel", nil)
	if err != nil {
		t.Fatalf("Failed to get database handle: %v", err)
	}

	// Test basic operations
	namespace := "test_chaincode"
	key := "test_key"

	// Create versioned value
	versionedValue := &VersionedValue{
		Value:    []byte("test_value"),
		Metadata: []byte("test_metadata"),
		Version: &Height{
			BlockNum: 1,
			TxNum:    2,
		},
	}

	// Create update batch
	batch := &UpdateBatch{
		Updates: map[string]map[string]*VersionedValue{
			namespace: {
				key: versionedValue,
			},
		},
	}

	// Apply updates
	height := &Height{BlockNum: 1, TxNum: 2}
	err = db.ApplyUpdates(batch, height)
	if err != nil {
		t.Fatalf("Failed to apply updates: %v", err)
	}

	// Get state
	retrievedValue, err := db.GetState(namespace, key)
	if err != nil {
		t.Fatalf("Failed to get state: %v", err)
	}

	if retrievedValue == nil {
		t.Fatal("Retrieved value is nil")
	}

	if string(retrievedValue.Value) != string(versionedValue.Value) {
		t.Errorf("Expected %s, got %s", versionedValue.Value, retrievedValue.Value)
	}

	if string(retrievedValue.Metadata) != string(versionedValue.Metadata) {
		t.Errorf("Expected %s, got %s", versionedValue.Metadata, retrievedValue.Metadata)
	}

	if retrievedValue.Version.BlockNum != versionedValue.Version.BlockNum {
		t.Errorf("Expected block num %d, got %d", versionedValue.Version.BlockNum, retrievedValue.Version.BlockNum)
	}

	if retrievedValue.Version.TxNum != versionedValue.Version.TxNum {
		t.Errorf("Expected tx num %d, got %d", versionedValue.Version.TxNum, retrievedValue.Version.TxNum)
	}

	// Test GetLatestSavePoint
	savePoint, err := db.GetLatestSavePoint()
	if err != nil {
		t.Fatalf("Failed to get save point: %v", err)
	}

	if savePoint.BlockNum != height.BlockNum || savePoint.TxNum != height.TxNum {
		t.Errorf("Expected save point %+v, got %+v", height, savePoint)
	}
}
