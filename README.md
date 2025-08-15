# FX Gin Scaffold

一个基于 Gin 框架和 Uber FX 依赖注入的生产级 Go REST API 脚手架，遵循清洁架构原则和领域驱动设计。

## ✨ 特性

- 🏗️ **清洁架构**: 领域驱动设计，关注点清晰分离
- 💉 **依赖注入**: 使用 Uber FX 实现类型安全的依赖注入
- 🗄️ **多数据库支持**: SQLite、PostgreSQL、MongoDB 统一接口
- 🔐 **JWT 认证**: 安全的 JWT 中间件认证
- 📝 **Swagger 文档**: 自动生成的 API 文档
- 🧪 **测试**: 基于 testify 的完整测试套件
- 📊 **结构化日志**: 遵循最佳实践的 Zap 日志
- 🔄 **数据库迁移**: 内置手动迁移系统
- 🚀 **热重载**: 开发模式自动重启
- 🐳 **Docker 支持**: 容器化支持

## 🚀 快速开始

### 环境要求

- Go 1.21 或更高版本
- Make（可选，用于便捷命令）

### 安装

1. **克隆仓库**
   ```bash
   git clone https://github.com/luxixing/fx-gin-scaffold.git
   cd fx-gin-scaffold
   ```

2. **设置开发环境**
   ```bash
   make setup
   ```

3. **配置环境变量**
   ```bash
   # 复制并编辑环境配置文件
   cp .env.example .env
   # 编辑 .env 文件，设置必要的配置
   ```

4. **运行数据库迁移**
   ```bash
   make migrate
   ```

5. **启动开发服务器**
   ```bash
   make dev
   ```

API 将在 `http://localhost:8080` 可用

## 📖 API 文档

服务器启动后，可访问：
- **Swagger UI**: `http://localhost:8080/swagger/index.html`
- **健康检查**: `http://localhost:8080/health`

## 🏛️ 项目架构

```
fx-gin-scaffold/
├── cmd/
│   ├── server/              # 应用主入口
│   └── migrate/             # 迁移工具入口
├── internal/
│   ├── bootstrap/           # 应用生命周期和依赖注入配置
│   ├── config/              # 配置管理
│   ├── domain/              # 业务领域模型和接口
│   ├── service/             # 业务逻辑实现
│   ├── repo/                # 数据访问层
│   ├── http/                # HTTP 传输层
│   │   ├── handler/         # HTTP 处理器
│   │   └── middleware/      # HTTP 中间件
│   └── migration/           # 数据库迁移系统
│       ├── migrations/      # 迁移文件
│       └── seeders/         # 种子数据
├── pkg/
│   ├── logger/              # 日志工具
│   ├── database/            # 数据库连接
│   └── utils/               # 通用工具
└── docs/
    ├── swagger/             # Swagger 文档
    └── MIGRATION.md         # 迁移系统文档
```

### 核心原则

- **领域驱动设计**: 业务逻辑与基础设施分离
- **依赖倒置**: 高层模块不依赖于低层模块
- **接口隔离**: 小而专注的接口
- **单一职责**: 每个模块只有一个变化原因

## 🗄️ 数据库支持

### SQLite（默认）
```bash
DB_DRIVER=sqlite
SQLITE_PATH=./data/app.db
DB_TABLE_PREFIX=fx_
```

### PostgreSQL
```bash
DB_DRIVER=postgres
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password
POSTGRES_DATABASE=fx_gin_scaffold
DB_TABLE_PREFIX=fx_
```

### MongoDB
```bash
DB_DRIVER=mongo
MONGO_URI=mongodb://localhost:27017
MONGO_DATABASE=fx_gin_scaffold
DB_TABLE_PREFIX=fx_
```

## 🔄 数据库迁移

本项目使用手动迁移系统，提供完全的迁移时机控制：

```bash
# 检查待执行的迁移
make check-migrations

# 预览迁移内容（干运行）
make migrate-dry-run

# 执行迁移
make migrate
```

详细的迁移系统文档请参考 [MIGRATION.md](docs/MIGRATION.md)。

## 🔐 认证系统

脚手架包含基于 JWT 的认证系统：

1. **注册/登录**: 获取 JWT 令牌
2. **受保护路由**: 在请求头中包含 `Authorization: Bearer <token>`
3. **中间件**: 自动令牌验证

### 使用示例

```bash
# 注册新用户
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123","name":"张三"}'

# 登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'

# 访问受保护的路由
curl -X GET http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer <your-jwt-token>"
```

## 🧪 测试

```bash
# 运行所有测试
make test

# 运行测试并生成覆盖率报告
make test-coverage

# 仅运行仓储层测试
make test-repo
```

## 🛠️ 开发命令

```bash
# 开发服务器（热重载）
make dev

# 构建应用
make build

# 运行代码检查
make lint

# 格式化代码
make fmt

# 生成 Swagger 文档
make swagger

# 数据库迁移
make migrate               # 执行迁移
make check-migrations      # 检查待执行迁移
make migrate-dry-run      # 迁移预览

# 清理构建文件
make clean

# 安装开发工具
make install-tools
```

## 📝 添加新功能

### 1. 定义领域模型

```go
// internal/domain/product.go
type Product struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    Name        string    `json:"name" gorm:"not null;size:255"`
    Price       float64   `json:"price" gorm:"not null"`
    Description string    `json:"description" gorm:"type:text"`
    CreatedAt   time.Time `json:"created_at"`
}

func (Product) TableName() string {
    return GetTableName("products")
}

type ProductRepository interface {
    Create(ctx context.Context, product *Product) error
    GetByID(ctx context.Context, id uint) (*Product, error)
    // ... 其他方法
}
```

### 2. 实现仓储层

```go
// internal/repo/product_gorm.go
type productRepository struct {
    db *gorm.DB
}

func NewProductRepository(db *gorm.DB) domain.ProductRepository {
    return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, product *domain.Product) error {
    return r.db.WithContext(ctx).Create(product).Error
}
```

### 3. 创建服务层

```go
// internal/service/product.go
type ProductService struct {
    productRepo domain.ProductRepository
}

func NewProductService(productRepo domain.ProductRepository) *ProductService {
    return &ProductService{productRepo: productRepo}
}

func (s *ProductService) CreateProduct(ctx context.Context, product *domain.Product) error {
    return s.productRepo.Create(ctx, product)
}
```

### 4. 添加 HTTP 处理器

```go
// internal/http/handler/product.go
type ProductHandler struct {
    productService *service.ProductService
}

func NewProductHandler(productService *service.ProductService) *ProductHandler {
    return &ProductHandler{productService: productService}
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
    // 实现逻辑
}
```

### 5. 注册到 FX 容器

```go
// internal/bootstrap/bootstrap.go
// 在 GetModule() 函数中添加：
fx.Provide(repo.NewProductRepository),
fx.Provide(service.NewProductService),
fx.Provide(handler.NewProductHandler),
```

## 🚀 部署

### Docker

```bash
# 构建镜像
docker build -t fx-gin-scaffold .

# 运行容器
docker run -p 8080:8080 fx-gin-scaffold
```

### 二进制文件

```bash
# 生产构建
make build

# 运行二进制文件
./bin/fx-gin-scaffold
```

### 生产环境部署流程

1. **备份数据库**
2. **运行迁移**: `go run ./cmd/migrate/main.go`
3. **启动应用**: `./bin/fx-gin-scaffold`

## 📋 环境变量

| 变量 | 描述 | 默认值 |
|------|------|--------|
| `APP_ENV` | 应用环境 | `development` |
| `APP_HOST` | 服务器主机 | `localhost` |
| `APP_PORT` | 服务器端口 | `8080` |
| `DB_DRIVER` | 数据库驱动 (sqlite/postgres/mongo) | `sqlite` |
| `DB_TABLE_PREFIX` | 数据库表前缀 | `fx_` |
| `JWT_SECRET` | JWT 签名密钥 | **必需** |
| `LOG_LEVEL` | 日志级别 | `info` |
| `LOG_FORMAT` | 日志格式 | `json` |

完整的配置选项请参考 `.env.example` 文件。

## 🛡️ 安全

- JWT 令牌认证
- 密码 bcrypt 加密
- 输入验证和清理
- CORS 配置
- 生产环境安全头设置

## 📈 性能

- 连接池配置
- 优雅关闭
- 请求超时控制
- 数据库索引优化

## 🧩 扩展性

- 模块化架构设计
- 插件化中间件
- 多数据库支持
- 环境配置分离

## 🤝 贡献

1. Fork 本仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 开启 Pull Request

提交前请确保：
- 运行 `make lint` 通过代码检查
- 运行 `make test` 通过所有测试
- 添加必要的测试用例

## 📚 相关文档

- [数据库迁移系统](docs/MIGRATION.md)
- [API 文档](http://localhost:8080/swagger/index.html)

## 📄 许可证

本项目基于 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 🙏 致谢

- 基于 [Gin](https://gin-gonic.com/) 和 [Uber FX](https://uber-go.github.io/fx/) 构建
- 遵循 [Go 项目布局](https://github.com/golang-standards/project-layout) 标准
- 使用 [Gorm](https://gorm.io/) ORM 和 [Zap](https://github.com/uber-go/zap) 日志库

---

> 💡 **提示**: 这是一个生产就绪的脚手架，但在部署到生产环境前，请确保根据您的具体需求调整配置和安全设置。