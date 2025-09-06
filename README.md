# IPSSL Client

一个使用ZeroSSL签名IP地址证书的Go工具，支持90天自动续签。

## 功能特性

- 🔐 使用ZeroSSL API为IP地址签名SSL证书
- 🔄 支持90天自动续签
- 🐳 集成Docker API，自动重载Caddy服务
- 📁 自动管理证书文件和验证文件目录
- ⚙️ 灵活的配置选项
- 📊 结构化日志记录

## 项目结构

```
ipssl-client/
├── main.go                 # 主程序入口
├── go.mod                  # Go模块文件
├── go.sum                  # 依赖校验文件
├── Dockerfile              # Docker构建文件
├── Makefile               # 构建脚本
├── env.example            # 环境变量示例
├── .gitignore             # Git忽略文件
├── docker/                # Docker部署配置
│   ├── docker-compose.yml # Docker Compose配置
│   ├── env.example        # 环境变量模板
│   ├── README.md          # Docker部署说明
│   ├── config/            # 配置文件
│   │   └── caddy/         # Caddy配置
│   └── data/              # 数据目录
│       └── caddy/         # Caddy数据
└── internal/              # 内部包
    ├── config/            # 配置管理
    ├── logger/            # 日志记录
    ├── ipssl/             # IPSSL客户端
    ├── zerossl/           # ZeroSSL API集成
    └── docker/            # Docker API集成
```

## 快速开始

### 1. 克隆项目

```bash
git clone <repository-url>
cd ipssl-client
```

### 2. 配置环境变量

```bash
cp env.example .env
# 编辑 .env 文件，设置你的配置
```

### 3. 本地开发

```bash
# 安装依赖
make deps

# 构建项目
make build

# 运行项目
make run
```

### 4. Docker部署

#### 方法一：使用 Makefile（推荐）

```bash
# 设置Docker环境
make docker-setup

# 编辑配置文件（修改IP地址和API密钥）
nano docker/.env

# 启动服务
make docker-run
```

#### 方法二：手动操作

```bash
# 进入docker目录
cd docker

# 复制配置文件
cp env.example .env

# 编辑配置文件（修改IP地址和API密钥）
nano .env

# 启动服务
docker-compose up -d
```

## 配置说明

### 环境变量

| 变量名 | 描述 | 默认值 | 必需 |
|--------|------|--------|------|
| `CLIENT_IP` | 要获取证书的IP地址 | `47.108.170.58` | 是 |
| `IPSSL_API_KEY` | ZeroSSL API密钥 | - | 是 |
| `IPSSL_VALIDATION_DIR` | 验证文件目录 | `/usr/share/caddy/` | 否 |
| `IPSSL_SSL_DIR` | SSL证书存储目录 | `/ipssl/` | 否 |
| `IPSSL_CONTAINER_NAME` | 要重载的容器名称（留空禁用Docker功能） | `caddy-1` | 否 |
| `RENEWAL_INTERVAL` | 续签检查间隔 | `24h` | 否 |
| `CERT_VALIDITY` | 证书有效期 | `2160h` (90天) | 否 |

### Docker Compose配置

项目包含完整的Docker Compose配置，包括：

- **Caddy服务**: 处理HTTP/HTTPS请求
- **IPSSL Client服务**: 管理SSL证书

## 部署架构

```
┌─────────────────┐    ┌─────────────────┐
│   Caddy Server  │    │  IPSSL Client   │
│                 │    │                 │
│  - HTTP/HTTPS   │    │  - ZeroSSL API  │
│  - Port 80/443  │    │  - Auto Renewal │
│  - SSL Certs    │◄───┤  - Docker API   │
└─────────────────┘    └─────────────────┘
         │                       │
         │                       │
    ┌────▼────┐              ┌───▼────┐
    │ Webroot │              │ /ipssl │
    │ Files   │              │ Certs  │
    └─────────┘              └────────┘
```

## 开发指南

### 构建命令

```bash
make build        # 构建应用
make run          # 运行应用
make test         # 运行测试
make clean        # 清理构建文件
make fmt          # 格式化代码
make lint         # 代码检查
```

### 添加新功能

1. 在相应的包中添加新功能
2. 更新配置结构（如需要）
3. 添加测试用例
4. 更新文档

## 注意事项

⚠️ **重要提醒**:

1. **ZeroSSL API集成**: 已集成官方 [caddyserver/zerossl](https://github.com/caddyserver/zerossl) 库
2. **IP地址证书**: 支持IP地址的SSL证书申请和验证
3. **安全性**: 确保API密钥和私钥文件的安全存储
4. **容器权限**: 确保Docker socket访问权限正确配置
5. **验证文件**: 自动创建HTTP验证文件到webroot目录

## 许可证

[添加许可证信息]

## 贡献

欢迎提交Issue和Pull Request！

## 更新日志

### v1.0.0
- 初始版本
- 基础IPSSL客户端功能
- Docker集成
- 自动续签支持