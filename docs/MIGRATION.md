# 数据库迁移系统文档

## 📋 概述

本项目使用基于Go代码的通用迁移系统，支持SQL数据库（SQLite、PostgreSQL）和NoSQL数据库（MongoDB）。迁移系统提供了版本控制、手动执行、环境感知的种子数据等功能。

## 🚀 迁移系统特性

### ✅ 核心功能
- **多数据库支持**: SQL (SQLite/PostgreSQL) + NoSQL (MongoDB)
- **手动执行**: 通过独立的迁移命令手动控制迁移时机
- **版本控制**: 基于时间戳的版本管理系统
- **迁移跟踪**: 自动记录已执行的迁移，防止重复执行
- **环境感知**: 不同环境可以运行不同的种子数据
- **类型安全**: 使用Go代码编写，编译时检查

### 📊 数据库支持对比

| 特性 | SQL数据库 | MongoDB |
|------|-----------|---------|
| 表/集合创建 | GORM AutoMigrate | 手动创建集合 |
| 索引管理 | GORM标签自动创建 | 手动创建索引 |
| 迁移跟踪 | `migrations` 表 | `migrations` 集合 |
| 事务支持 | ✅ | 部分支持 |

## 🏗️ 系统架构

### 文件结构
```
internal/migration/
├── migration.go              # 迁移引擎核心
├── registry.go              # 迁移和种子注册器
├── migrations/              # 迁移文件目录
│   └── 20240815120000_create_users_table.go
└── seeders/                 # 种子数据目录
    ├── admin_user_seeder.go
    └── test_users_seeder.go
```

### 核心接口

#### Migration 接口
```go
type Migration interface {
    Version() string                                           // 迁移版本号
    Description() string                                       // 迁移描述
    Up(ctx context.Context, db *database.Connection) error   // 执行迁移
    Down(ctx context.Context, db *database.Connection) error // 回滚迁移
}
```

#### Seeder 接口
```go
type Seeder interface {
    Name() string                                    // 种子名称
    Run(ctx context.Context, db *database.Connection) error // 执行种子
    ShouldRun(env string) bool                      // 是否应该在指定环境运行
}
```

## ⚡ 执行机制

### 1. 手动执行流程
迁移通过独立的命令手动执行：

```bash
# 运行迁移
make migrate

# 检查待执行的迁移
make check-migrations

# 查看迁移预览（干运行）
make migrate-dry-run

# 启动应用（不执行迁移）
make dev
```

### 2. 执行时序
```
手动执行迁移命令 → 加载配置 → 连接数据库 → 设置表前缀 → 
检查迁移表 → 执行待执行迁移 → 运行环境种子
```

### 3. 日志输出示例
```
🔄 Loading configuration...
🔗 Connecting to database...
🚀 Running migrations...
[INFO] running migration version=20240815120000 description="Create users table/collection"
[INFO] migration completed version=20240815120000
[INFO] running seeder name=AdminUserSeeder
[INFO] seeder completed name=AdminUserSeeder
✅ Migrations completed successfully
```

## 📝 编写迁移

### 1. 创建迁移文件

**文件命名规则**: `YYYYMMDDHHMMSS_description.go`

```go
// internal/migration/migrations/20240815130000_add_user_avatar.go
package migrations

import (
    "context"
    "github.com/luxixing/fx-gin-scaffold/internal/domain"
    "github.com/luxixing/fx-gin-scaffold/pkg/database"
)

type AddUserAvatar struct{}

func (m *AddUserAvatar) Version() string {
    return "20240815130000"
}

func (m *AddUserAvatar) Description() string {
    return "Add avatar field to users table"
}

func (m *AddUserAvatar) Up(ctx context.Context, db *database.Connection) error {
    if db.GORM != nil {
        // SQL数据库迁移
        return db.GORM.Exec("ALTER TABLE users ADD COLUMN avatar VARCHAR(255)").Error
    }

    if db.Mongo != nil {
        // MongoDB迁移（通常不需要schema变更）
        // 可以在这里创建新的索引或集合
        return nil
    }

    return nil
}

func (m *AddUserAvatar) Down(ctx context.Context, db *database.Connection) error {
    if db.GORM != nil {
        return db.GORM.Exec("ALTER TABLE users DROP COLUMN avatar").Error
    }
    return nil
}
```

### 2. 注册迁移

在 `internal/migration/registry.go` 中注册：

```go
func RegisterMigrations(migrator *Migrator) {
    migrator.AddMigration(&migrations.CreateUsersTable{})
    migrator.AddMigration(&migrations.AddUserAvatar{})  // 添加新迁移
}
```

## 🌱 编写种子数据

### 1. 创建种子文件

```go
// internal/migration/seeders/demo_data_seeder.go
package seeders

import (
    "context"
    "github.com/luxixing/fx-gin-scaffold/pkg/database"
)

type DemoDataSeeder struct{}

func (s *DemoDataSeeder) Name() string {
    return "DemoDataSeeder"
}

func (s *DemoDataSeeder) ShouldRun(env string) bool {
    // 只在开发和测试环境运行
    return env == "development" || env == "testing"
}

func (s *DemoDataSeeder) Run(ctx context.Context, db *database.Connection) error {
    // 实现种子数据逻辑
    return nil
}
```

### 2. 注册种子

```go
func RegisterSeeders(migrator *Migrator) {
    migrator.AddSeeder(&seeders.AdminUserSeeder{})
    migrator.AddSeeder(&seeders.TestUsersSeeder{})
    migrator.AddSeeder(&seeders.DemoDataSeeder{})  // 添加新种子
}
```

## 🏭 生产环境部署

### 1. 环境配置

```bash
# .env (生产环境)
APP_ENV=production
DB_TABLE_PREFIX=prod_         # 生产环境表前缀
DB_DRIVER=postgres           # 生产数据库驱动
```

### 2. 部署策略

#### 推荐方案: 手动迁移（适用于所有项目）
```bash
# 1. 停止应用
docker-compose stop app

# 2. 备份数据库
./scripts/backup-db.sh

# 3. 运行迁移
docker-compose run --rm migrate-tool go run ./cmd/migrate/main.go

# 4. 启动应用
docker-compose up -d app
```

### 3. 生产环境安全检查清单

- [ ] **迁移测试**: 在staging环境测试所有迁移
- [ ] **数据备份**: 确保有可恢复的数据备份
- [ ] **回滚计划**: 准备回滚脚本和流程
- [ ] **监控告警**: 设置迁移失败告警
- [ ] **停机窗口**: 评估是否需要维护窗口
- [ ] **版本兼容**: 确保新旧版本的兼容性

## 🧪 测试环境

### 1. CI/CD集成

```yaml
# .github/workflows/test.yml
- name: Run migrations and tests
  run: |
    export APP_ENV=testing
    export DB_DRIVER=sqlite
    export SQLITE_PATH=./test.db
    export DB_TABLE_PREFIX=test_
    # 先运行迁移
    go run ./cmd/migrate/main.go
    # 再启动应用和测试
    go run ./cmd/server/main.go &
    sleep 5
    go test ./...
```

### 2. 测试数据管理

```go
// 测试专用种子
func (s *TestDataSeeder) ShouldRun(env string) bool {
    return env == "testing"
}
```

## 🔧 高级功能

### 1. 条件迁移

```go
func (m *ConditionalMigration) Up(ctx context.Context, db *database.Connection) error {
    if db.GORM != nil {
        // 检查列是否存在
        if !db.GORM.Migrator().HasColumn(&domain.User{}, "avatar") {
            return db.GORM.Migrator().AddColumn(&domain.User{}, "avatar")
        }
    }
    return nil
}
```

### 2. 数据迁移

```go
func (m *DataMigration) Up(ctx context.Context, db *database.Connection) error {
    if db.GORM != nil {
        // 数据转换示例
        return db.GORM.Exec(`
            UPDATE users
            SET role = 'user'
            WHERE role IS NULL OR role = ''
        `).Error
    }
    return nil
}
```

### 3. 性能优化

```go
func (m *AddIndexMigration) Up(ctx context.Context, db *database.Connection) error {
    if db.GORM != nil {
        // 添加索引
        return db.GORM.Exec(`
            CREATE INDEX CONCURRENTLY idx_users_email_active
            ON users(email, active)
            WHERE active = true
        `).Error
    }
    return nil
}
```

## 🚨 故障排除

### 常见问题

1. **迁移失败**
   ```
   [ERROR] migration 20240815130000 failed: column already exists
   ```
   **解决**: 检查迁移逻辑，添加条件判断

2. **重复执行**
   ```
   [ERROR] migration already executed
   ```
   **解决**: 检查迁移跟踪表，确认版本号唯一

3. **种子数据冲突**
   ```
   [ERROR] duplicate key error
   ```
   **解决**: 在种子中添加存在性检查

### 调试技巧

1. **启用详细日志**
   ```go
   // 临时启用FX日志查看依赖注入过程
   app := fx.New(
       // fx.NopLogger, // 注释掉这行
       bootstrap.GetModule(),
       fx.Invoke(bootstrap.RegisterHooks),
   )
   ```

2. **单独运行迁移**
   ```go
   // 创建测试程序单独运行迁移
   func main() {
       db := database.NewConnection(config)
       migration.RunMigrations(context.Background(), db, "development")
   }
   ```

## 📚 最佳实践

### 1. 迁移设计原则
- ✅ **幂等性**: 迁移可以安全地重复执行
- ✅ **向后兼容**: 新迁移不破坏现有功能
- ✅ **原子性**: 每个迁移应该是一个原子操作
- ✅ **可回滚**: 提供Down方法用于回滚

### 2. 版本号管理
- 使用时间戳格式: `YYYYMMDDHHMMSS`
- 确保版本号的唯一性和递增性
- 团队协作时避免版本号冲突

### 3. 测试策略
- 在开发环境充分测试迁移
- 使用staging环境验证生产数据
- 编写迁移的单元测试

### 4. 性能考虑
- 大表迁移考虑分批处理
- 添加索引时使用CONCURRENTLY
- 避免在迁移中执行重负载操作

## 🔗 相关命令

```bash
# 开发环境
make migrate               # 运行数据库迁移
make check-migrations      # 检查待执行迁移
make migrate-dry-run      # 预览待执行迁移
make dev                  # 启动开发服务器
make swagger              # 生成API文档
make test                 # 运行测试

# 生产环境
go run ./cmd/migrate/main.go           # 运行迁移
go run ./cmd/migrate/main.go -check    # 检查待执行迁移
go run ./cmd/migrate/main.go -dry-run  # 预览待执行迁移
```

---

> 💡 **提示**: 这个迁移系统设计为手动控制和安全，通过独立的迁移命令确保迁移时机的可控性。在生产环境中始终建议进行充分的测试和备份。如有疑问，请查看 `internal/migration/` 目录下的示例代码。
