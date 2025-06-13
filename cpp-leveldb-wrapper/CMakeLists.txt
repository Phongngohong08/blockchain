cmake_minimum_required(VERSION 3.10)
project(cpp_leveldb_wrapper)

set(CMAKE_CXX_STANDARD 11)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Find LevelDB
find_path(LEVELDB_INCLUDE_DIR leveldb/db.h)
find_library(LEVELDB_LIBRARY leveldb)

if(NOT LEVELDB_INCLUDE_DIR OR NOT LEVELDB_LIBRARY)
    message(FATAL_ERROR "LevelDB not found. Please install leveldb development packages.")
endif()

# Include directories
include_directories(${LEVELDB_INCLUDE_DIR})
include_directories(include)

# Source files
set(SOURCES
    src/leveldb_wrapper.cpp
    src/iterator_wrapper.cpp
    src/batch_wrapper.cpp
)

# Create shared library
add_library(cpp_leveldb_wrapper SHARED ${SOURCES})

# Link LevelDB
target_link_libraries(cpp_leveldb_wrapper ${LEVELDB_LIBRARY})

# Set output directory
set_target_properties(cpp_leveldb_wrapper PROPERTIES
    LIBRARY_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/lib
)

# Install targets
install(TARGETS cpp_leveldb_wrapper
    LIBRARY DESTINATION lib
    ARCHIVE DESTINATION lib
)

install(DIRECTORY include/
    DESTINATION include
    FILES_MATCHING PATTERN "*.h"
)
