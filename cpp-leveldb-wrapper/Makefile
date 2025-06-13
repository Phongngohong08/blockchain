CC := gcc
CXX := g++
CXXFLAGS := -std=c++11 -fPIC -O2 -Wall -Wextra

# Directories
SRC_DIR := src
INCLUDE_DIR := include
BUILD_DIR := build
LIB_DIR := lib

# Source files
SOURCES := $(wildcard $(SRC_DIR)/*.cpp)
OBJECTS := $(SOURCES:$(SRC_DIR)/%.cpp=$(BUILD_DIR)/%.o)

# Library name
LIBRARY := $(LIB_DIR)/libcpp_leveldb_wrapper.so

# LevelDB flags
LEVELDB_CFLAGS := $(shell pkg-config --cflags leveldb 2>/dev/null || echo "-I/usr/include")
LEVELDB_LIBS := $(shell pkg-config --libs leveldb 2>/dev/null || echo "-lleveldb")

# Default target
all: $(LIBRARY)

# Create directories
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

$(LIB_DIR):
	mkdir -p $(LIB_DIR)

# Compile object files
$(BUILD_DIR)/%.o: $(SRC_DIR)/%.cpp | $(BUILD_DIR)
	$(CXX) $(CXXFLAGS) $(LEVELDB_CFLAGS) -I$(INCLUDE_DIR) -c $< -o $@

# Link shared library
$(LIBRARY): $(OBJECTS) | $(LIB_DIR)
	$(CXX) -shared -o $@ $^ $(LEVELDB_LIBS)

# Clean
clean:
	rm -rf $(BUILD_DIR) $(LIB_DIR)

# Install
install: $(LIBRARY)
	cp $(LIBRARY) /usr/local/lib/
	cp -r $(INCLUDE_DIR)/* /usr/local/include/
	ldconfig

# Uninstall
uninstall:
	rm -f /usr/local/lib/libcpp_leveldb_wrapper.so
	rm -f /usr/local/include/leveldb_wrapper.h

.PHONY: all clean install uninstall
