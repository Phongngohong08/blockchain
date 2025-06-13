package leveldb

/*
#cgo CFLAGS: -I../include
#cgo LDFLAGS: -L../lib -lcpp_leveldb_wrapper -lleveldb -lstdc++
#include "leveldb_wrapper.h"
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"unsafe"
)

// DB represents a LevelDB database
type DB struct {
	db *C.leveldb_t
}

// Iterator represents a LevelDB iterator
type Iterator struct {
	iter *C.leveldb_iterator_t
}

// WriteBatch represents a batch of write operations
type WriteBatch struct {
	batch *C.leveldb_writebatch_t
}

// Options represents database options
type Options struct {
	CreateIfMissing      bool
	ErrorIfExists        bool
	ParanoidChecks       bool
	WriteBufferSize      int
	MaxOpenFiles         int
	BlockSize            int
	BlockRestartInterval int
	MaxFileSize          int
	Compression          int
}

// ReadOptions represents read options
type ReadOptions struct {
	VerifyChecksums bool
	FillCache       bool
}

// WriteOptions represents write options
type WriteOptions struct {
	Sync bool
}

// Open opens a database
func Open(name string, options *Options) (*DB, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	var coptions *C.leveldb_options_t
	if options != nil {
		coptions = C.leveldb_options_create()
		defer C.leveldb_options_destroy(coptions)

		if options.CreateIfMissing {
			coptions.create_if_missing = 1
		}
		if options.ErrorIfExists {
			coptions.error_if_exists = 1
		}
		if options.ParanoidChecks {
			coptions.paranoid_checks = 1
		}
		coptions.write_buffer_size = C.size_t(options.WriteBufferSize)
		coptions.max_open_files = C.int(options.MaxOpenFiles)
		coptions.block_size = C.size_t(options.BlockSize)
		coptions.block_restart_interval = C.int(options.BlockRestartInterval)
		coptions.max_file_size = C.size_t(options.MaxFileSize)
		coptions.compression = C.int(options.Compression)
	}

	var cerr *C.leveldb_error_t
	cdb := C.leveldb_open(cname, coptions, &cerr)

	if cerr != nil {
		defer C.leveldb_error_destroy(cerr)
		return nil, errors.New(C.GoString(cerr.message))
	}

	if cdb == nil {
		return nil, errors.New("failed to open database")
	}

	return &DB{db: cdb}, nil
}

// Close closes the database
func (db *DB) Close() {
	if db.db != nil {
		C.leveldb_close(db.db)
		db.db = nil
	}
}

// Put writes a key-value pair
func (db *DB) Put(options *WriteOptions, key, value []byte) error {
	var coptions *C.leveldb_writeoptions_t
	if options != nil {
		coptions = C.leveldb_writeoptions_create()
		defer C.leveldb_writeoptions_destroy(coptions)
		if options.Sync {
			coptions.sync = 1
		}
	}

	var cerr *C.leveldb_error_t
	C.leveldb_put(db.db, coptions,
		(*C.char)(unsafe.Pointer(&key[0])), C.size_t(len(key)),
		(*C.char)(unsafe.Pointer(&value[0])), C.size_t(len(value)),
		&cerr)

	if cerr != nil {
		defer C.leveldb_error_destroy(cerr)
		return errors.New(C.GoString(cerr.message))
	}

	return nil
}

// Get reads a value for a key
func (db *DB) Get(options *ReadOptions, key []byte) ([]byte, error) {
	var coptions *C.leveldb_readoptions_t
	if options != nil {
		coptions = C.leveldb_readoptions_create()
		defer C.leveldb_readoptions_destroy(coptions)
		if options.VerifyChecksums {
			coptions.verify_checksums = 1
		}
		if options.FillCache {
			coptions.fill_cache = 1
		}
	}

	var vallen C.size_t
	var cerr *C.leveldb_error_t
	cvalue := C.leveldb_get(db.db, coptions,
		(*C.char)(unsafe.Pointer(&key[0])), C.size_t(len(key)),
		&vallen, &cerr)

	if cerr != nil {
		defer C.leveldb_error_destroy(cerr)
		return nil, errors.New(C.GoString(cerr.message))
	}

	if cvalue == nil {
		return nil, nil // Key not found
	}
	defer C.leveldb_free(unsafe.Pointer(cvalue))

	value := C.GoBytes(unsafe.Pointer(cvalue), C.int(vallen))
	return value, nil
}

// Delete removes a key
func (db *DB) Delete(options *WriteOptions, key []byte) error {
	var coptions *C.leveldb_writeoptions_t
	if options != nil {
		coptions = C.leveldb_writeoptions_create()
		defer C.leveldb_writeoptions_destroy(coptions)
		if options.Sync {
			coptions.sync = 1
		}
	}

	var cerr *C.leveldb_error_t
	C.leveldb_delete(db.db, coptions,
		(*C.char)(unsafe.Pointer(&key[0])), C.size_t(len(key)),
		&cerr)

	if cerr != nil {
		defer C.leveldb_error_destroy(cerr)
		return errors.New(C.GoString(cerr.message))
	}

	return nil
}

// Write executes a batch of operations
func (db *DB) Write(options *WriteOptions, batch *WriteBatch) error {
	var coptions *C.leveldb_writeoptions_t
	if options != nil {
		coptions = C.leveldb_writeoptions_create()
		defer C.leveldb_writeoptions_destroy(coptions)
		if options.Sync {
			coptions.sync = 1
		}
	}

	var cerr *C.leveldb_error_t
	C.leveldb_write(db.db, coptions, batch.batch, &cerr)

	if cerr != nil {
		defer C.leveldb_error_destroy(cerr)
		return errors.New(C.GoString(cerr.message))
	}

	return nil
}

// NewIterator creates a new iterator
func (db *DB) NewIterator(options *ReadOptions) *Iterator {
	var coptions *C.leveldb_readoptions_t
	if options != nil {
		coptions = C.leveldb_readoptions_create()
		defer C.leveldb_readoptions_destroy(coptions)
		if options.VerifyChecksums {
			coptions.verify_checksums = 1
		}
		if options.FillCache {
			coptions.fill_cache = 1
		}
	}

	citer := C.leveldb_create_iterator(db.db, coptions)
	return &Iterator{iter: citer}
}

// Iterator methods

// Valid returns whether the iterator is positioned at a valid key-value pair
func (it *Iterator) Valid() bool {
	return C.leveldb_iter_valid(it.iter) != 0
}

// SeekToFirst positions at the first key in the database
func (it *Iterator) SeekToFirst() {
	C.leveldb_iter_seek_to_first(it.iter)
}

// SeekToLast positions at the last key in the database
func (it *Iterator) SeekToLast() {
	C.leveldb_iter_seek_to_last(it.iter)
}

// Seek positions at the first key >= target
func (it *Iterator) Seek(key []byte) {
	C.leveldb_iter_seek(it.iter,
		(*C.char)(unsafe.Pointer(&key[0])), C.size_t(len(key)))
}

// Next moves to the next entry
func (it *Iterator) Next() {
	C.leveldb_iter_next(it.iter)
}

// Prev moves to the previous entry
func (it *Iterator) Prev() {
	C.leveldb_iter_prev(it.iter)
}

// Key returns the key of the current entry
func (it *Iterator) Key() []byte {
	var keylen C.size_t
	ckey := C.leveldb_iter_key(it.iter, &keylen)
	if ckey == nil {
		return nil
	}
	return C.GoBytes(unsafe.Pointer(ckey), C.int(keylen))
}

// Value returns the value of the current entry
func (it *Iterator) Value() []byte {
	var vallen C.size_t
	cvalue := C.leveldb_iter_value(it.iter, &vallen)
	if cvalue == nil {
		return nil
	}
	return C.GoBytes(unsafe.Pointer(cvalue), C.int(vallen))
}

// Error returns any error encountered during iteration
func (it *Iterator) Error() error {
	var cerr *C.leveldb_error_t
	C.leveldb_iter_get_error(it.iter, &cerr)
	if cerr != nil {
		defer C.leveldb_error_destroy(cerr)
		return errors.New(C.GoString(cerr.message))
	}
	return nil
}

// Close releases the iterator
func (it *Iterator) Close() {
	if it.iter != nil {
		C.leveldb_iter_destroy(it.iter)
		it.iter = nil
	}
}

// WriteBatch methods

// NewWriteBatch creates a new write batch
func NewWriteBatch() *WriteBatch {
	return &WriteBatch{batch: C.leveldb_writebatch_create()}
}

// Put adds a put operation to the batch
func (wb *WriteBatch) Put(key, value []byte) {
	C.leveldb_writebatch_put(wb.batch,
		(*C.char)(unsafe.Pointer(&key[0])), C.size_t(len(key)),
		(*C.char)(unsafe.Pointer(&value[0])), C.size_t(len(value)))
}

// Delete adds a delete operation to the batch
func (wb *WriteBatch) Delete(key []byte) {
	C.leveldb_writebatch_delete(wb.batch,
		(*C.char)(unsafe.Pointer(&key[0])), C.size_t(len(key)))
}

// Clear clears all operations from the batch
func (wb *WriteBatch) Clear() {
	C.leveldb_writebatch_clear(wb.batch)
}

// Close releases the batch
func (wb *WriteBatch) Close() {
	if wb.batch != nil {
		C.leveldb_writebatch_destroy(wb.batch)
		wb.batch = nil
	}
}

// Utility functions

// CompactRange compacts the database in the given range
func (db *DB) CompactRange(start, limit []byte) {
	var startPtr, limitPtr *C.char
	var startLen, limitLen C.size_t

	if start != nil {
		startPtr = (*C.char)(unsafe.Pointer(&start[0]))
		startLen = C.size_t(len(start))
	}

	if limit != nil {
		limitPtr = (*C.char)(unsafe.Pointer(&limit[0]))
		limitLen = C.size_t(len(limit))
	}

	C.leveldb_compact_range(db.db, startPtr, startLen, limitPtr, limitLen)
}

// PropertyValue returns the value of a database property
func (db *DB) PropertyValue(property string) string {
	cprop := C.CString(property)
	defer C.free(unsafe.Pointer(cprop))

	cvalue := C.leveldb_property_value(db.db, cprop)
	if cvalue == nil {
		return ""
	}
	defer C.leveldb_free(unsafe.Pointer(cvalue))

	return C.GoString(cvalue)
}
