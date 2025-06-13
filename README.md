# Hyperledger Fabric Development Environment

Môi trường phát triển tích hợp cho Hyperledger Fabric với tối ưu hóa hiệu năng thông qua C++ LevelDB wrapper.

## 📋 Tổng quan dự án

Dự án này bao gồm:

- **Hyperledger Fabric 2.5.13** - Nền tảng blockchain doanh nghiệp
- **Fabric Samples 2.2** - Bộ ví dụ và tutorial học tập 
- **C++ LevelDB Wrapper** - Thư viện tối ưu hóa hiệu năng cơ sở dữ liệu

## 🏗️ Cấu trúc dự án

```
fabric/
├── README.md                     # Tài liệu chính (file này)
├── cpp-leveldb-wrapper/          # Wrapper C++ cho LevelDB
│   ├── include/                  # Header files C++
│   ├── src/                      # Source code C++
│   ├── go/                       # Go bindings và CGO
│   ├── examples/                 # Ví dụ tích hợp
│   ├── CMakeLists.txt           # Cấu hình build CMake
│   ├── Makefile                 # Build Unix/Linux
│   ├── build.sh                 # Script build
│   ├── README.md                # Hướng dẫn wrapper
│   └── INTEGRATION.md           # Hướng dẫn tích hợp Fabric
├── fabric-2.5.13/               # Mã nguồn Hyperledger Fabric
│   ├── cmd/                     # Các công cụ CLI
│   ├── core/                    # Core functionality
│   ├── common/                  # Thư viện chung
│   ├── orderer/                 # Orderer service
│   ├── bccsp/                   # Cryptographic service provider
│   ├── msp/                     # Membership service provider
│   ├── gossip/                  # Gossip protocol
│   └── ...                     # Các module khác
└── fabric-samples-2.2/          # Ví dụ và tutorials
    ├── asset-transfer-basic/     # Ví dụ cơ bản
    ├── asset-transfer-events/    # Events và notifications
    ├── commercial-paper/         # Commercial paper use case
    ├── test-network/            # Mạng test
    └── ...                      # Các ví dụ khác
```

## 🚀 Bắt đầu nhanh

### Yêu cầu hệ thống

- **Go 1.19+**
- **Docker và Docker Compose**
- **CMake 3.10+**
- **GCC/Clang** với hỗ trợ C++11
- **Google LevelDB C++** library
- **Git**

### Cài đặt dependencies

#### Windows (PowerShell)
```powershell
# Cài đặt Go
winget install GoLang.Go

# Cài đặt Docker Desktop
winget install Docker.DockerDesktop

# Cài đặt CMake
winget install Kitware.CMake

# Cài đặt Git
winget install Git.Git
```

#### Ubuntu/Debian
```bash
# Cập nhật package manager
sudo apt update

# Cài đặt dependencies
sudo apt install -y golang-go docker.io docker-compose cmake build-essential git
sudo apt install -y libleveldb-dev libleveldb1d

# Thêm user vào docker group
sudo usermod -aG docker $USER
```

#### macOS
```bash
# Sử dụng Homebrew
brew install go docker cmake leveldb git
```

### Build và cài đặt

1. **Build C++ LevelDB Wrapper**
```bash
cd cpp-leveldb-wrapper
chmod +x build.sh
./build.sh
```

2. **Cài đặt shared library**
```bash
# Linux/macOS
sudo cp build/lib/libcpp_leveldb_wrapper.so /usr/local/lib/
sudo ldconfig  # Linux only

# Windows
# Copy libcpp_leveldb_wrapper.dll to system PATH
```

3. **Build Hyperledger Fabric**
```bash
cd fabric-2.5.13
make all
```

4. **Khởi chạy test network**
```bash
cd fabric-samples-2.2/test-network
./network.sh up
```

## 📚 Hướng dẫn sử dụng

### 1. Khởi động môi trường development

```bash
# Khởi động test network
cd fabric-samples-2.2/test-network
./network.sh up createChannel

# Deploy chaincode mẫu
./network.sh deployCC -ccn basic -ccp ../asset-transfer-basic/chaincode-go -ccl go
```

### 2. Chạy ví dụ Asset Transfer

```bash
cd fabric-samples-2.2/asset-transfer-basic/application-go
go mod tidy
go run .
```

### 3. Sử dụng C++ LevelDB Wrapper

Tham khảo file `cpp-leveldb-wrapper/INTEGRATION.md` để tích hợp wrapper vào Fabric.

### 4. Phát triển chaincode

```bash
# Tạo chaincode mới
mkdir my-chaincode
cd my-chaincode
go mod init my-chaincode

# Implement chaincode logic
# Deploy chaincode
cd ../fabric-samples-2.2/test-network
./network.sh deployCC -ccn my-chaincode -ccp ../my-chaincode -ccl go
```

## 🔧 Tối ưu hóa hiệu năng

### C++ LevelDB Wrapper

Dự án bao gồm wrapper C++ thay thế goleveldb để cải thiện hiệu năng:

- **Hiệu năng cao hơn** - Sử dụng implementation C++ gốc của Google LevelDB
- **Tiết kiệm bộ nhớ** - Tối ưu hóa memory usage
- **Tương thích hoàn toàn** - Drop-in replacement cho goleveldb
- **Hỗ trợ đa nền tảng** - Linux, macOS, Windows

## 🧪 Testing

### Chạy unit tests

```bash
# Test C++ wrapper
cd cpp-leveldb-wrapper
make test

# Test Fabric
cd fabric-2.5.13
make unit-test

# Test integration
make integration-test
```

### Performance testing

```bash
# Benchmark C++ wrapper
cd cpp-leveldb-wrapper/go
go test -bench=. -benchmem

# Fabric performance tests
cd fabric-2.5.13
make performance-test
```

## 📖 Tài liệu

### Tài liệu chính thức

- [Hyperledger Fabric Documentation](https://hyperledger-fabric.readthedocs.io/)
- [Fabric Samples Repository](https://github.com/hyperledger/fabric-samples)
- [Fabric API Documentation](https://godoc.org/github.com/hyperledger/fabric)

### Tài liệu dự án

- [`cpp-leveldb-wrapper/README.md`](cpp-leveldb-wrapper/README.md) - Hướng dẫn C++ wrapper
- [`cpp-leveldb-wrapper/INTEGRATION.md`](cpp-leveldb-wrapper/INTEGRATION.md) - Tích hợp với Fabric
- [`fabric-samples-2.2/README.md`](fabric-samples-2.2/README.md) - Hướng dẫn samples

### Tutorials quan trọng

1. **[Writing Your First Application](https://hyperledger-fabric.readthedocs.io/en/release-2.5/write_first_app.html)**
2. **[Commercial Paper Tutorial](https://hyperledger-fabric.readthedocs.io/en/release-2.5/tutorial/commercial_paper.html)**
3. **[Private Data Tutorial](https://hyperledger-fabric.readthedocs.io/en/release-2.5/private_data_tutorial.html)**

## 🔍 Các ví dụ có sẵn

### Asset Transfer Series

| Sample | Mô tả | Ngôn ngữ |
|--------|-------|----------|
| [asset-transfer-basic](fabric-samples-2.2/asset-transfer-basic) | Tạo và chuyển giao tài sản cơ bản | Go, JS, TS, Java |
| [asset-transfer-ledger-queries](fabric-samples-2.2/asset-transfer-ledger-queries) | Truy vấn ledger và CouchDB | Go, JS |
| [asset-transfer-private-data](fabric-samples-2.2/asset-transfer-private-data) | Private data collections | Go, Java |
| [asset-transfer-events](fabric-samples-2.2/asset-transfer-events) | Events và notifications | JS, Java |

## 🐛 Troubleshooting

### Lỗi thường gặp

1. **CGO compilation error**
```bash
# Đảm bảo CGO_ENABLED=1
export CGO_ENABLED=1
go build
```

2. **LevelDB library not found**
```bash
# Ubuntu/Debian
sudo apt install libleveldb-dev

# CentOS/RHEL  
sudo yum install leveldb-devel

# macOS
brew install leveldb
```

3. **Docker permission denied**
```bash
# Linux
sudo usermod -aG docker $USER
# Logout và login lại
```

4. **Port conflicts**
```bash
# Kiểm tra ports đang sử dụng
netstat -tulpn | grep :7051
netstat -tulpn | grep :7054

# Dọn dẹp containers
docker container prune -f
```

### Debug modes

```bash
# Fabric với debug logs
export FABRIC_LOGGING_SPEC=DEBUG

# Fabric với trace logs cho specific modules
export FABRIC_LOGGING_SPEC=peer.gossip=DEBUG:peer.chaincode=DEBUG
```

## 🤝 Đóng góp

### Quy trình development

1. Fork repository
2. Tạo feature branch: `git checkout -b feature/amazing-feature`
3. Commit changes: `git commit -m 'Add amazing feature'`
4. Push to branch: `git push origin feature/amazing-feature`
5. Tạo Pull Request

## 🔗 Links hữu ích

- [Hyperledger Fabric GitHub](https://github.com/hyperledger/fabric)
- [Fabric Discord Community](https://discord.gg/hyperledger)
- [Hyperledger Foundation](https://www.hyperledger.org/)
- [LevelDB Documentation](https://github.com/google/leveldb)
