#include "leveldb_wrapper.h"
#include <leveldb/db.h>
#include <leveldb/iterator.h>

extern "C" {

leveldb_iterator_t* leveldb_create_iterator(leveldb_t* db, const leveldb_readoptions_t* options) {
    leveldb::ReadOptions opts;
    if (options) {
        opts.verify_checksums = options->verify_checksums;
        opts.fill_cache = options->fill_cache;
    }
    
    leveldb_iterator_t* iter = new leveldb_iterator_t;
    iter->iter.reset(db->db->NewIterator(opts));
    return iter;
}

void leveldb_iter_destroy(leveldb_iterator_t* iter) {
    delete iter;
}

int leveldb_iter_valid(const leveldb_iterator_t* iter) {
    return iter->iter->Valid() ? 1 : 0;
}

void leveldb_iter_seek_to_first(leveldb_iterator_t* iter) {
    iter->iter->SeekToFirst();
}

void leveldb_iter_seek_to_last(leveldb_iterator_t* iter) {
    iter->iter->SeekToLast();
}

void leveldb_iter_seek(leveldb_iterator_t* iter, const char* key, size_t keylen) {
    leveldb::Slice k(key, keylen);
    iter->iter->Seek(k);
}

void leveldb_iter_next(leveldb_iterator_t* iter) {
    iter->iter->Next();
}

void leveldb_iter_prev(leveldb_iterator_t* iter) {
    iter->iter->Prev();
}

const char* leveldb_iter_key(const leveldb_iterator_t* iter, size_t* keylen) {
    leveldb::Slice key = iter->iter->key();
    if (keylen) {
        *keylen = key.size();
    }
    return key.data();
}

const char* leveldb_iter_value(const leveldb_iterator_t* iter, size_t* vallen) {
    leveldb::Slice value = iter->iter->value();
    if (vallen) {
        *vallen = value.size();
    }
    return value.data();
}

void leveldb_iter_get_error(leveldb_iterator_t* iter, leveldb_error_t** errptr) {
    leveldb::Status status = iter->iter->status();
    if (!status.ok() && errptr) {
        leveldb_error_t* err = new leveldb_error_t;
        std::string msg = status.ToString();
        err->message = new char[msg.length() + 1];
        strcpy(err->message, msg.c_str());
        *errptr = err;
    }
}

} // extern "C"
