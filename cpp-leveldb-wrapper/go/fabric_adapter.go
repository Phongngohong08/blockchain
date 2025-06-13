package leveldb

import (
	"errors"
	"fmt"
	"sync"
)

// FabricDBProvider implements VersionedDBProvider using C++ LevelDB
type FabricDBProvider struct {
	dbPath string
	dbs    map[string]*FabricDB
	mutex  sync.RWMutex
}

// FabricDB implements VersionedDB using C++ LevelDB
type FabricDB struct {
	db     *DB
	dbName string
}

// FabricIterator implements ResultsIterator
type FabricIterator struct {
	iter      *Iterator
	namespace string
	startKey  string
	endKey    string
}

// NewFabricDBProvider creates a new provider
func NewFabricDBProvider(dbPath string) (*FabricDBProvider, error) {
	return &FabricDBProvider{
		dbPath: dbPath,
		dbs:    make(map[string]*FabricDB),
	}, nil
}

// GetDBHandle implements VersionedDBProvider
func (provider *FabricDBProvider) GetDBHandle(id string, namespaceProvider interface{}) (VersionedDB, error) {
	provider.mutex.Lock()
	defer provider.mutex.Unlock()

	if db, exists := provider.dbs[id]; exists {
		return db, nil
	}

	dbPath := fmt.Sprintf("%s/%s", provider.dbPath, id)
	options := &Options{
		CreateIfMissing: true,
		WriteBufferSize: 4 * 1024 * 1024, // 4MB
		MaxOpenFiles:    1000,
		BlockSize:       4096,
		Compression:     1, // Snappy
	}

	db, err := Open(dbPath, options)
	if err != nil {
		return nil, fmt.Errorf("failed to open leveldb: %v", err)
	}

	fabricDB := &FabricDB{
		db:     db,
		dbName: id,
	}

	provider.dbs[id] = fabricDB
	return fabricDB, nil
}

// ImportFromSnapshot implements VersionedDBProvider
func (provider *FabricDBProvider) ImportFromSnapshot(id string, savepoint *Height, itr FullScanIterator) error {
	// Implementation for importing from snapshot
	return errors.New("ImportFromSnapshot not implemented yet")
}

// BytesKeySupported implements VersionedDBProvider
func (provider *FabricDBProvider) BytesKeySupported() bool {
	return true
}

// Close implements VersionedDBProvider
func (provider *FabricDBProvider) Close() {
	provider.mutex.Lock()
	defer provider.mutex.Unlock()

	for _, db := range provider.dbs {
		db.Close()
	}
	provider.dbs = make(map[string]*FabricDB)
}

// Drop implements VersionedDBProvider
func (provider *FabricDBProvider) Drop(id string) error {
	provider.mutex.Lock()
	defer provider.mutex.Unlock()

	if db, exists := provider.dbs[id]; exists {
		db.Close()
		delete(provider.dbs, id)
	}

	// Note: This doesn't actually remove the directory
	// You might want to add filesystem operations here
	return nil
}

// FabricDB methods

// GetState implements VersionedDB
func (vdb *FabricDB) GetState(namespace string, key string) (*VersionedValue, error) {
	compositeKey := constructCompositeKey(namespace, key)

	value, err := vdb.db.Get(nil, []byte(compositeKey))
	if err != nil {
		return nil, err
	}

	if value == nil {
		return nil, nil
	}

	return decodeValue(value)
}

// GetVersion implements VersionedDB
func (vdb *FabricDB) GetVersion(namespace string, key string) (*Height, error) {
	versionedValue, err := vdb.GetState(namespace, key)
	if err != nil {
		return nil, err
	}

	if versionedValue == nil {
		return nil, nil
	}

	return versionedValue.Version, nil
}

// GetStateMultipleKeys implements VersionedDB
func (vdb *FabricDB) GetStateMultipleKeys(namespace string, keys []string) ([]*VersionedValue, error) {
	results := make([]*VersionedValue, len(keys))

	for i, key := range keys {
		value, err := vdb.GetState(namespace, key)
		if err != nil {
			return nil, err
		}
		results[i] = value
	}

	return results, nil
}

// GetStateRangeScanIterator implements VersionedDB
func (vdb *FabricDB) GetStateRangeScanIterator(namespace string, startKey string, endKey string) (ResultsIterator, error) {
	iter := vdb.db.NewIterator(nil)

	fabricIter := &FabricIterator{
		iter:      iter,
		namespace: namespace,
		startKey:  startKey,
		endKey:    endKey,
	}

	// Position the iterator
	compositeStartKey := constructCompositeKey(namespace, startKey)
	iter.Seek([]byte(compositeStartKey))

	return fabricIter, nil
}

// GetStateRangeScanIteratorWithPagination implements VersionedDB
func (vdb *FabricDB) GetStateRangeScanIteratorWithPagination(namespace string, startKey string, endKey string, pageSize int32) (QueryResultsIterator, error) {
	iter, err := vdb.GetStateRangeScanIterator(namespace, startKey, endKey)
	if err != nil {
		return nil, err
	}
	// Return the FabricIterator which implements QueryResultsIterator
	return iter.(*FabricIterator), nil
}

// ExecuteQuery implements VersionedDB
func (vdb *FabricDB) ExecuteQuery(namespace, query string) (ResultsIterator, error) {
	return nil, errors.New("ExecuteQuery not supported by LevelDB")
}

// ExecuteQueryWithPagination implements VersionedDB
func (vdb *FabricDB) ExecuteQueryWithPagination(namespace, query, bookmark string, pageSize int32) (QueryResultsIterator, error) {
	return nil, errors.New("ExecuteQueryWithPagination not supported by LevelDB")
}

// ApplyUpdates implements VersionedDB
func (vdb *FabricDB) ApplyUpdates(batch *UpdateBatch, height *Height) error {
	writeBatch := NewWriteBatch()
	defer writeBatch.Close()

	// Process all updates in the batch
	for ns, nsUpdates := range batch.Updates {
		for key, versionedValue := range nsUpdates {
			compositeKey := constructCompositeKey(ns, key)

			if versionedValue.Value == nil {
				// Delete operation
				writeBatch.Delete([]byte(compositeKey))
			} else {
				// Put operation
				encodedValue, err := encodeValue(versionedValue)
				if err != nil {
					return err
				}
				writeBatch.Put([]byte(compositeKey), encodedValue)
			}
		}
	}

	// Save the height as savepoint
	if height != nil {
		savePointValue, err := encodeHeight(height)
		if err != nil {
			return err
		}
		writeBatch.Put([]byte("savepoint"), savePointValue)
	}

	// Write the batch
	writeOptions := &WriteOptions{Sync: true}
	return vdb.db.Write(writeOptions, writeBatch)
}

// GetLatestSavePoint implements VersionedDB
func (vdb *FabricDB) GetLatestSavePoint() (*Height, error) {
	value, err := vdb.db.Get(nil, []byte("savepoint"))
	if err != nil {
		return nil, err
	}

	if value == nil {
		return nil, nil
	}

	return decodeHeight(value)
}

// ValidateKeyValue implements VersionedDB
func (vdb *FabricDB) ValidateKeyValue(key string, value []byte) error {
	// LevelDB supports any bytes as key/value
	return nil
}

// BytesKeySupported implements VersionedDB
func (vdb *FabricDB) BytesKeySupported() bool {
	return true
}

// GetFullScanIterator implements VersionedDB
func (vdb *FabricDB) GetFullScanIterator(skipNamespace func(string) bool) (FullScanIterator, error) {
	return nil, errors.New("GetFullScanIterator not implemented yet")
}

// Open implements VersionedDB
func (vdb *FabricDB) Open() error {
	// Database is already opened in GetDBHandle
	return nil
}

// Close implements VersionedDB
func (vdb *FabricDB) Close() {
	if vdb.db != nil {
		vdb.db.Close()
		vdb.db = nil
	}
}

// FabricIterator methods

// Next implements ResultsIterator
func (iter *FabricIterator) Next() (*VersionedKV, error) {
	for iter.iter.Valid() {
		key := iter.iter.Key()
		value := iter.iter.Value()

		// Parse the composite key
		namespace, actualKey, err := parseCompositeKey(string(key))
		if err != nil {
			iter.iter.Next()
			continue
		}

		// Check if we're still in the correct namespace
		if namespace != iter.namespace {
			return nil, nil // End of namespace
		}

		// Check if we've reached the end key
		if iter.endKey != "" && actualKey >= iter.endKey {
			return nil, nil
		}

		// Decode the value
		versionedValue, err := decodeValue(value)
		if err != nil {
			return nil, err
		}

		compositeKey := &CompositeKey{
			Namespace: namespace,
			Key:       actualKey,
		}

		result := &VersionedKV{
			CompositeKey:   compositeKey,
			VersionedValue: versionedValue,
		}

		iter.iter.Next()
		return result, nil
	}

	return nil, nil
}

// Close implements ResultsIterator
func (iter *FabricIterator) Close() {
	if iter.iter != nil {
		iter.iter.Close()
		iter.iter = nil
	}
}

// GetBookmarkAndClose implements QueryResultsIterator
func (iter *FabricIterator) GetBookmarkAndClose() string {
	iter.Close()
	return ""
}

// Helper functions

func constructCompositeKey(namespace, key string) string {
	return namespace + "\x00" + key
}

func parseCompositeKey(compositeKey string) (string, string, error) {
	parts := []string{}
	current := ""

	for _, char := range compositeKey {
		if char == '\x00' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	parts = append(parts, current)

	if len(parts) != 2 {
		return "", "", errors.New("invalid composite key format")
	}

	return parts[0], parts[1], nil
}
