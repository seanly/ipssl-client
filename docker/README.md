# Docker 部署配置

这个目录包含了所有Docker相关的配置文件。

## 文件说明

- `docker-compose.yml` - Docker Compose 配置文件
- `env.example` - 环境变量配置模板
- `.env` - 实际的环境变量配置文件（需要手动创建）
- `config/` - Caddy 配置文件目录
- `data/` - 数据存储目录

## 快速开始

1. **进入docker目录**：
   ```bash
   cd docker
   ```

2. **复制配置文件**：
   ```bash
   cp env.example .env
   ```

3. **编辑配置**：
   ```bash
   nano .env  # 修改IP地址和API密钥
   ```

4. **启动服务**：
   ```bash
   docker-compose up -d
   ```

## 目录结构

```
docker/
├── docker-compose.yml    # Docker Compose 配置
├── env.example          # 环境变量模板
├── .env                 # 实际配置文件（需手动创建）
├── config/              # 配置文件
│   └── caddy/
│       └── Caddyfile    # Caddy 配置
└── data/                # 数据目录
    └── caddy/
        ├── data/        # Caddy 数据
        ├── config/      # Caddy 配置
        ├── logs/        # 日志文件
        ├── webroot/     # Web 根目录
        └── ipssl/       # SSL 证书存储
```

## 环境变量说明

### 必需配置
- `IPSSL_API_KEY` - ZeroSSL API 密钥
- `IPSSL_CLIENT_IP` - 要申请证书的IP地址

### 可选配置
- `IP_WEB_DOMAIN` - Web服务器域名/IP（默认与CLIENT_IP相同）

### 高级配置
- `IPSSL_VALIDATION_DIR` - 验证文件目录（默认：/usr/share/caddy/）
- `IPSSL_SSL_DIR` - SSL证书存储目录（默认：/ipssl/）
- `IPSSL_CONTAINER_NAME` - 要重载的容器名称（默认：caddy-1，留空则禁用Docker重载功能）
- `RENEWAL_INTERVAL` - 续签检查间隔（默认：24h）
- `CERT_VALIDITY` - 证书有效期（默认：720h = 30天）

## 常用命令

```bash
# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down

# 重启服务
docker-compose restart

# 查看服务状态
docker-compose ps
```

## 注意事项

⚠️ **重要提醒**：

1. 确保 `.env` 文件中的 `IPSSL_CLIENT_IP` 和 `IP_WEB_DOMAIN` 设置正确
2. `IPSSL_API_KEY` 必须是有效的 ZeroSSL API 密钥
3. 确保服务器IP地址可以从外网访问（用于证书验证）
4. 首次运行可能需要几分钟时间来申请和验证证书

## 禁用 Docker 功能

如果你不需要自动重载 Docker 容器功能，可以将 `IPSSL_CONTAINER_NAME` 留空：

```bash
# 在 .env 文件中设置
IPSSL_CONTAINER_NAME=
```

这样 IPSSL 客户端将：
- 不会初始化 Docker 客户端
- 不会尝试重载容器
- 仍然会正常申请和保存 SSL 证书
- 适用于非 Docker 环境或手动管理证书的场景
