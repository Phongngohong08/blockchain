#include "leveldb_wrapper.h"
#include <leveldb/write_batch.h>

// Additional batch operations for more complex use cases

extern "C" {

// Iterator over batch contents (for debugging/testing)
typedef struct {
    const char* key;
    size_t key_len;
    const char* value;
    size_t value_len;
    int is_delete;
} leveldb_batch_entry_t;

// Batch size estimation
size_t leveldb_writebatch_approximate_size(leveldb_writebatch_t* batch) {
    return batch->batch.ApproximateSize();
}

// Put with slice optimization
void leveldb_writebatch_put_slice(leveldb_writebatch_t* batch,
                                 const char* key, size_t keylen,
                                 const char* val, size_t vallen) {
    leveldb::Slice k(key, keylen);
    leveldb::Slice v(val, vallen);
    batch->batch.Put(k, v);
}

// Delete with slice optimization  
void leveldb_writebatch_delete_slice(leveldb_writebatch_t* batch,
                                    const char* key, size_t keylen) {
    leveldb::Slice k(key, keylen);
    batch->batch.Delete(k);
}

// Batch iteration handler for debugging
typedef void (*leveldb_batch_handler_t)(void* state, 
                                        const char* key, size_t key_len,
                                        const char* value, size_t value_len);

// Custom handler implementation
class BatchHandler : public leveldb::WriteBatch::Handler {
public:
    BatchHandler(void* state, leveldb_batch_handler_t put_handler, leveldb_batch_handler_t delete_handler)
        : state_(state), put_handler_(put_handler), delete_handler_(delete_handler) {}
    
    void Put(const leveldb::Slice& key, const leveldb::Slice& value) override {
        if (put_handler_) {
            put_handler_(state_, key.data(), key.size(), value.data(), value.size());
        }
    }
    
    void Delete(const leveldb::Slice& key) override {
        if (delete_handler_) {
            delete_handler_(state_, key.data(), key.size(), nullptr, 0);
        }
    }

private:
    void* state_;
    leveldb_batch_handler_t put_handler_;
    leveldb_batch_handler_t delete_handler_;
};

// Iterate over batch contents
void leveldb_writebatch_iterate(leveldb_writebatch_t* batch,
                               void* state,
                               leveldb_batch_handler_t put_handler,
                               leveldb_batch_handler_t delete_handler) {
    BatchHandler handler(state, put_handler, delete_handler);
    batch->batch.Iterate(&handler);
}

} // extern "C"
