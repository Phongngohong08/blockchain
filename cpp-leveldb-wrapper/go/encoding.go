package leveldb

import (
	"encoding/binary"
	"encoding/json"
	"errors"
)

// Mock version package types for standalone compilation
type Height struct {
	BlockNum uint64 `json:"block_num"`
	TxNum    uint64 `json:"tx_num"`
}

// Mock statedb package types for standalone compilation
type VersionedValue struct {
	Value    []byte  `json:"value"`
	Metadata []byte  `json:"metadata"`
	Version  *Height `json:"version"`
}

type CompositeKey struct {
	Namespace string `json:"namespace"`
	Key       string `json:"key"`
}

type VersionedKV struct {
	*CompositeKey   `json:"composite_key"`
	*VersionedValue `json:"versioned_value"`
}

type ResultsIterator interface {
	Next() (*VersionedKV, error)
	Close()
}

type QueryResultsIterator interface {
	ResultsIterator
	GetBookmarkAndClose() string
}

type UpdateBatch struct {
	ContainsPostOrderWrites bool                                  `json:"contains_post_order_writes"`
	Updates                 map[string]map[string]*VersionedValue `json:"updates"`
}

type FullScanIterator interface {
	Next() (*VersionedKV, error)
	Close()
}

type VersionedDB interface {
	GetState(namespace string, key string) (*VersionedValue, error)
	GetVersion(namespace string, key string) (*Height, error)
	GetStateMultipleKeys(namespace string, keys []string) ([]*VersionedValue, error)
	GetStateRangeScanIterator(namespace string, startKey string, endKey string) (ResultsIterator, error)
	GetStateRangeScanIteratorWithPagination(namespace string, startKey string, endKey string, pageSize int32) (QueryResultsIterator, error)
	ExecuteQuery(namespace, query string) (ResultsIterator, error)
	ExecuteQueryWithPagination(namespace, query, bookmark string, pageSize int32) (QueryResultsIterator, error)
	ApplyUpdates(batch *UpdateBatch, height *Height) error
	GetLatestSavePoint() (*Height, error)
	ValidateKeyValue(key string, value []byte) error
	BytesKeySupported() bool
	GetFullScanIterator(skipNamespace func(string) bool) (FullScanIterator, error)
	Open() error
	Close()
}

type VersionedDBProvider interface {
	GetDBHandle(id string, namespaceProvider interface{}) (VersionedDB, error)
	ImportFromSnapshot(id string, savepoint *Height, itr FullScanIterator) error
	BytesKeySupported() bool
	Close()
	Drop(id string) error
}

// Encoding/Decoding functions

// ValueWrapper for encoding VersionedValue
type ValueWrapper struct {
	Value    []byte `json:"value"`
	Metadata []byte `json:"metadata"`
	BlockNum uint64 `json:"block_num"`
	TxNum    uint64 `json:"tx_num"`
}

// encodeValue serializes a VersionedValue to bytes
func encodeValue(vv *VersionedValue) ([]byte, error) {
	if vv == nil {
		return nil, nil
	}

	wrapper := ValueWrapper{
		Value:    vv.Value,
		Metadata: vv.Metadata,
	}

	if vv.Version != nil {
		wrapper.BlockNum = vv.Version.BlockNum
		wrapper.TxNum = vv.Version.TxNum
	}

	return json.Marshal(wrapper)
}

// decodeValue deserializes bytes to a VersionedValue
func decodeValue(data []byte) (*VersionedValue, error) {
	if data == nil {
		return nil, nil
	}

	var wrapper ValueWrapper
	err := json.Unmarshal(data, &wrapper)
	if err != nil {
		return nil, err
	}

	vv := &VersionedValue{
		Value:    wrapper.Value,
		Metadata: wrapper.Metadata,
		Version: &Height{
			BlockNum: wrapper.BlockNum,
			TxNum:    wrapper.TxNum,
		},
	}

	return vv, nil
}

// encodeHeight serializes a Height to bytes
func encodeHeight(height *Height) ([]byte, error) {
	if height == nil {
		return nil, nil
	}

	data := make([]byte, 16) // 8 bytes for BlockNum + 8 bytes for TxNum
	binary.BigEndian.PutUint64(data[0:8], height.BlockNum)
	binary.BigEndian.PutUint64(data[8:16], height.TxNum)

	return data, nil
}

// decodeHeight deserializes bytes to a Height
func decodeHeight(data []byte) (*Height, error) {
	if data == nil {
		return nil, nil
	}

	if len(data) != 16 {
		return nil, errors.New("invalid height data length")
	}

	height := &Height{
		BlockNum: binary.BigEndian.Uint64(data[0:8]),
		TxNum:    binary.BigEndian.Uint64(data[8:16]),
	}

	return height, nil
}
