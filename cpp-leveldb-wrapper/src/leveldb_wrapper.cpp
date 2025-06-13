#include "leveldb_wrapper.h"
#include <leveldb/db.h>
#include <leveldb/options.h>
#include <leveldb/write_batch.h>
#include <string>
#include <memory>

// Internal structures
struct leveldb_t {
    std::unique_ptr<leveldb::DB> db;
};

struct leveldb_iterator_t {
    std::unique_ptr<leveldb::Iterator> iter;
};

struct leveldb_writebatch_t {
    leveldb::WriteBatch batch;
};

// Helper function to create error
static leveldb_error_t* create_error(const std::string& message) {
    leveldb_error_t* err = new leveldb_error_t;
    err->message = new char[message.length() + 1];
    strcpy(err->message, message.c_str());
    return err;
}

// Convert options
static leveldb::Options convert_options(const leveldb_options_t* options) {
    leveldb::Options opts;
    if (options) {
        opts.create_if_missing = options->create_if_missing;
        opts.error_if_exists = options->error_if_exists;
        opts.paranoid_checks = options->paranoid_checks;
        opts.write_buffer_size = options->write_buffer_size;
        opts.max_open_files = options->max_open_files;
        opts.block_size = options->block_size;
        opts.block_restart_interval = options->block_restart_interval;
        opts.max_file_size = options->max_file_size;
        
        switch (options->compression) {
            case 0:
                opts.compression = leveldb::kNoCompression;
                break;
            case 1:
                opts.compression = leveldb::kSnappyCompression;
                break;
            default:
                opts.compression = leveldb::kSnappyCompression;
        }
    }
    return opts;
}

static leveldb::ReadOptions convert_read_options(const leveldb_readoptions_t* options) {
    leveldb::ReadOptions opts;
    if (options) {
        opts.verify_checksums = options->verify_checksums;
        opts.fill_cache = options->fill_cache;
    }
    return opts;
}

static leveldb::WriteOptions convert_write_options(const leveldb_writeoptions_t* options) {
    leveldb::WriteOptions opts;
    if (options) {
        opts.sync = options->sync;
    }
    return opts;
}

// Database operations
extern "C" {

leveldb_t* leveldb_open(const char* name, const leveldb_options_t* options, leveldb_error_t** errptr) {
    leveldb::DB* db;
    leveldb::Options opts = convert_options(options);
    leveldb::Status status = leveldb::DB::Open(opts, name, &db);
    
    if (!status.ok()) {
        if (errptr) {
            *errptr = create_error(status.ToString());
        }
        return nullptr;
    }
    
    leveldb_t* result = new leveldb_t;
    result->db.reset(db);
    return result;
}

void leveldb_close(leveldb_t* db) {
    delete db;
}

void leveldb_put(leveldb_t* db, const leveldb_writeoptions_t* options,
                const char* key, size_t keylen,
                const char* val, size_t vallen,
                leveldb_error_t** errptr) {
    leveldb::WriteOptions opts = convert_write_options(options);
    leveldb::Slice k(key, keylen);
    leveldb::Slice v(val, vallen);
    
    leveldb::Status status = db->db->Put(opts, k, v);
    if (!status.ok() && errptr) {
        *errptr = create_error(status.ToString());
    }
}

char* leveldb_get(leveldb_t* db, const leveldb_readoptions_t* options,
                 const char* key, size_t keylen,
                 size_t* vallen, leveldb_error_t** errptr) {
    leveldb::ReadOptions opts = convert_read_options(options);
    leveldb::Slice k(key, keylen);
    std::string value;
    
    leveldb::Status status = db->db->Get(opts, k, &value);
    if (!status.ok()) {
        if (errptr) {
            *errptr = create_error(status.ToString());
        }
        return nullptr;
    }
    
    char* result = new char[value.size()];
    memcpy(result, value.c_str(), value.size());
    if (vallen) {
        *vallen = value.size();
    }
    return result;
}

void leveldb_delete(leveldb_t* db, const leveldb_writeoptions_t* options,
                   const char* key, size_t keylen,
                   leveldb_error_t** errptr) {
    leveldb::WriteOptions opts = convert_write_options(options);
    leveldb::Slice k(key, keylen);
    
    leveldb::Status status = db->db->Delete(opts, k);
    if (!status.ok() && errptr) {
        *errptr = create_error(status.ToString());
    }
}

// Batch operations
leveldb_writebatch_t* leveldb_writebatch_create() {
    return new leveldb_writebatch_t;
}

void leveldb_writebatch_destroy(leveldb_writebatch_t* batch) {
    delete batch;
}

void leveldb_writebatch_clear(leveldb_writebatch_t* batch) {
    batch->batch.Clear();
}

void leveldb_writebatch_put(leveldb_writebatch_t* batch,
                           const char* key, size_t keylen,
                           const char* val, size_t vallen) {
    leveldb::Slice k(key, keylen);
    leveldb::Slice v(val, vallen);
    batch->batch.Put(k, v);
}

void leveldb_writebatch_delete(leveldb_writebatch_t* batch,
                              const char* key, size_t keylen) {
    leveldb::Slice k(key, keylen);
    batch->batch.Delete(k);
}

void leveldb_write(leveldb_t* db, const leveldb_writeoptions_t* options,
                  leveldb_writebatch_t* batch, leveldb_error_t** errptr) {
    leveldb::WriteOptions opts = convert_write_options(options);
    leveldb::Status status = db->db->Write(opts, &batch->batch);
    
    if (!status.ok() && errptr) {
        *errptr = create_error(status.ToString());
    }
}

// Options management
leveldb_options_t* leveldb_options_create() {
    leveldb_options_t* options = new leveldb_options_t;
    options->create_if_missing = 0;
    options->error_if_exists = 0;
    options->paranoid_checks = 0;
    options->write_buffer_size = 4 << 20; // 4MB
    options->max_open_files = 1000;
    options->block_size = 4096;
    options->block_restart_interval = 16;
    options->max_file_size = 2 << 20; // 2MB
    options->compression = 1; // Snappy
    return options;
}

void leveldb_options_destroy(leveldb_options_t* options) {
    delete options;
}

leveldb_readoptions_t* leveldb_readoptions_create() {
    leveldb_readoptions_t* options = new leveldb_readoptions_t;
    options->verify_checksums = 0;
    options->fill_cache = 1;
    return options;
}

void leveldb_readoptions_destroy(leveldb_readoptions_t* options) {
    delete options;
}

leveldb_writeoptions_t* leveldb_writeoptions_create() {
    leveldb_writeoptions_t* options = new leveldb_writeoptions_t;
    options->sync = 0;
    return options;
}

void leveldb_writeoptions_destroy(leveldb_writeoptions_t* options) {
    delete options;
}

// Error management
void leveldb_error_destroy(leveldb_error_t* err) {
    if (err) {
        delete[] err->message;
        delete err;
    }
}

// Memory management
void leveldb_free(void* ptr) {
    delete[] static_cast<char*>(ptr);
}

// Utility functions
void leveldb_compact_range(leveldb_t* db, const char* start_key, size_t start_key_len,
                          const char* limit_key, size_t limit_key_len) {
    leveldb::Slice start(start_key, start_key_len);
    leveldb::Slice limit(limit_key, limit_key_len);
    db->db->CompactRange(&start, &limit);
}

char* leveldb_property_value(leveldb_t* db, const char* propname) {
    std::string value;
    if (!db->db->GetProperty(propname, &value)) {
        return nullptr;
    }
    
    char* result = new char[value.size() + 1];
    strcpy(result, value.c_str());
    return result;
}

void leveldb_approximate_sizes(leveldb_t* db, int num_ranges,
                              const char* const* range_start_key, const size_t* range_start_key_len,
                              const char* const* range_limit_key, const size_t* range_limit_key_len,
                              uint64_t* sizes) {
    std::vector<leveldb::Range> ranges(num_ranges);
    for (int i = 0; i < num_ranges; i++) {
        ranges[i].start = leveldb::Slice(range_start_key[i], range_start_key_len[i]);
        ranges[i].limit = leveldb::Slice(range_limit_key[i], range_limit_key_len[i]);
    }
    db->db->GetApproximateSizes(ranges.data(), num_ranges, sizes);
}

} // extern "C"
