#!/bin/bash

# Build script for C++ LevelDB Wrapper
set -e

echo "Building C++ LevelDB Wrapper for Hyperledger Fabric..."

# Check dependencies
echo "Checking dependencies..."

# Check if LevelDB is installed
if ! pkg-config --exists leveldb 2>/dev/null; then
    echo "Error: LevelDB not found. Please install LevelDB development packages."
    echo "Ubuntu/Debian: sudo apt-get install libleveldb-dev"
    echo "CentOS/RHEL: sudo yum install leveldb-devel"
    echo "macOS: brew install leveldb"
    exit 1
fi

# Check if CMake is available
if ! command -v cmake &> /dev/null; then
    echo "Error: CMake not found. Please install CMake 3.10+."
    exit 1
fi

# Check if Go is available
if ! command -v go &> /dev/null; then
    echo "Error: Go not found. Please install Go 1.19+."
    exit 1
fi

# Create build directory
echo "Creating build directory..."
mkdir -p build
cd build

# Configure with CMake
echo "Configuring with CMake..."
cmake ..

# Build the shared library
echo "Building shared library..."
make -j$(nproc)

# Go back to project root
cd ..

# Test the Go wrapper (optional)
if [ "$1" != "--no-test" ]; then
    echo "Testing Go wrapper..."
    cd go
    
    # Set library path for testing
    export CGO_LDFLAGS="-L../build/lib -lcpp_leveldb_wrapper -lleveldb -lstdc++"
    export LD_LIBRARY_PATH="../build/lib:$LD_LIBRARY_PATH"
    
    # Run tests
    go test -v
    
    cd ..
fi

echo "Build completed successfully!"
echo "Shared library location: build/lib/libcpp_leveldb_wrapper.so"
echo ""
echo "To use in Fabric:"
echo "1. Copy the shared library to /usr/local/lib/ or add to LD_LIBRARY_PATH"
echo "2. Modify Fabric's statedb implementation to use this wrapper"
echo "3. Rebuild Fabric with the new statedb implementation"
