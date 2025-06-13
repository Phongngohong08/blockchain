#ifndef LEVELDB_WRAPPER_H
#define LEVELDB_WRAPPER_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stddef.h>
#include <stdint.h>

// Error handling
typedef struct {
    char* message;
} leveldb_error_t;

// Database handle
typedef struct leveldb_t leveldb_t;

// Iterator handle  
typedef struct leveldb_iterator_t leveldb_iterator_t;

// WriteBatch handle
typedef struct leveldb_writebatch_t leveldb_writebatch_t;

// Options structures
typedef struct {
    int create_if_missing;
    int error_if_exists;
    int paranoid_checks;
    size_t write_buffer_size;
    int max_open_files;
    size_t block_size;
    int block_restart_interval;
    size_t max_file_size;
    int compression;
} leveldb_options_t;

typedef struct {
    int verify_checksums;
    int fill_cache;
} leveldb_readoptions_t;

typedef struct {
    int sync;
} leveldb_writeoptions_t;

// Database operations
leveldb_t* leveldb_open(const char* name, const leveldb_options_t* options, leveldb_error_t** errptr);
void leveldb_close(leveldb_t* db);
void leveldb_put(leveldb_t* db, const leveldb_writeoptions_t* options,
                const char* key, size_t keylen,
                const char* val, size_t vallen,
                leveldb_error_t** errptr);
char* leveldb_get(leveldb_t* db, const leveldb_readoptions_t* options,
                 const char* key, size_t keylen,
                 size_t* vallen, leveldb_error_t** errptr);
void leveldb_delete(leveldb_t* db, const leveldb_writeoptions_t* options,
                   const char* key, size_t keylen,
                   leveldb_error_t** errptr);

// Batch operations
leveldb_writebatch_t* leveldb_writebatch_create();
void leveldb_writebatch_destroy(leveldb_writebatch_t* batch);
void leveldb_writebatch_clear(leveldb_writebatch_t* batch);
void leveldb_writebatch_put(leveldb_writebatch_t* batch,
                           const char* key, size_t keylen,
                           const char* val, size_t vallen);
void leveldb_writebatch_delete(leveldb_writebatch_t* batch,
                              const char* key, size_t keylen);
void leveldb_write(leveldb_t* db, const leveldb_writeoptions_t* options,
                  leveldb_writebatch_t* batch, leveldb_error_t** errptr);

// Iterator operations
leveldb_iterator_t* leveldb_create_iterator(leveldb_t* db, const leveldb_readoptions_t* options);
void leveldb_iter_destroy(leveldb_iterator_t* iter);
int leveldb_iter_valid(const leveldb_iterator_t* iter);
void leveldb_iter_seek_to_first(leveldb_iterator_t* iter);
void leveldb_iter_seek_to_last(leveldb_iterator_t* iter);
void leveldb_iter_seek(leveldb_iterator_t* iter, const char* key, size_t keylen);
void leveldb_iter_next(leveldb_iterator_t* iter);
void leveldb_iter_prev(leveldb_iterator_t* iter);
const char* leveldb_iter_key(const leveldb_iterator_t* iter, size_t* keylen);
const char* leveldb_iter_value(const leveldb_iterator_t* iter, size_t* vallen);
void leveldb_iter_get_error(leveldb_iterator_t* iter, leveldb_error_t** errptr);

// Options management
leveldb_options_t* leveldb_options_create();
void leveldb_options_destroy(leveldb_options_t* options);
leveldb_readoptions_t* leveldb_readoptions_create();
void leveldb_readoptions_destroy(leveldb_readoptions_t* options);
leveldb_writeoptions_t* leveldb_writeoptions_create();
void leveldb_writeoptions_destroy(leveldb_writeoptions_t* options);

// Error management
void leveldb_error_destroy(leveldb_error_t* err);

// Memory management
void leveldb_free(void* ptr);

// Utility functions
void leveldb_compact_range(leveldb_t* db, const char* start_key, size_t start_key_len,
                          const char* limit_key, size_t limit_key_len);
char* leveldb_property_value(leveldb_t* db, const char* propname);
void leveldb_approximate_sizes(leveldb_t* db, int num_ranges,
                              const char* const* range_start_key, const size_t* range_start_key_len,
                              const char* const* range_limit_key, const size_t* range_limit_key_len,
                              uint64_t* sizes);

#ifdef __cplusplus
}
#endif

#endif // LEVELDB_WRAPPER_H
