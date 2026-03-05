# BBS 论坛系统

## 项目概述

这是一个基于 Go 语言开发的 BBS 论坛系统，具有完整的用户认证、帖子管理、评论系统、点赞功能、收藏功能和消息队列等特性。系统采用分层架构设计，使用 MySQL 作为主数据库，Redis 作为消息队列，支持高并发场景。

## 技术栈

| 技术 | 版本 | 用途 |
|------|------|------|
| Go | 1.20+ | 主要开发语言 |
| Gin | v1.9.0 | Web 框架 |
| GORM | v1.25.0 | ORM 框架 |
| MySQL | 8.0+ | 主数据库 |
| Redis | 7.0+ | 消息队列 |
| JWT | - | 认证令牌 |
| Zap | v1.24.0 | 日志库 |
| Viper | v1.15.0 | 配置管理 |
| 雪花算法 | - | 生成唯一用户 ID |

## 项目结构

```
bbsDemo/
├── config/            # 配置管理
│   └── config.go      # 配置结构和加载
├── database/          # 数据库相关
│   ├── mysql.go       # MySQL 连接和迁移
│   └── redis.go       # Redis 连接和消息队列
├── handler/           # HTTP 处理器
│   ├── auth.go        # 认证相关接口
│   └── post.go        # 帖子相关接口
├── logger/            # 日志管理
│   └── logger.go      # Zap 日志初始化
├── middleware/        # 中间件
│   ├── auth.go        # 认证中间件
│   └── admin.go       # 管理员权限中间件
├── models/            # 数据模型
│   ├── user.go        # 用户相关模型
│   └── post.go        # 帖子相关模型
├── queue/             # 消息队列
│   └── worker.go      # 消息消费者
├── router/            # 路由管理
│   └── router.go      # 路由注册
├── service/           # 业务逻辑
│   ├── user.go        # 用户业务逻辑
│   └── post.go        # 帖子业务逻辑
├── utils/             # 工具函数
│   ├── jwt.go         # JWT 工具
│   ├── hash.go        # 密码哈希工具
│   └── snowflake.go   # 雪花算法 ID 生成
├── config.yaml        # 配置文件
├── go.mod             # Go 模块文件
├── go.sum             # 依赖校验
├── main.go            # 主入口
├── README.md          # 项目说明
└── API.md             # API 接口文档
```

## 核心功能实现

### 1. 用户认证系统

**实现方式：**
- 使用 JWT 生成访问令牌（access token）和刷新令牌（refresh token）
- 密码采用 bcrypt 加盐哈希存储
- 邮箱验证码通过 Redis 消息队列异步发送
- 支持令牌刷新机制

**关键文件：**
- `utils/jwt.go` - JWT 令牌生成和解析
- `utils/hash.go` - 密码哈希和验证
- `service/user.go` - 认证业务逻辑
- `handler/auth.go` - 认证接口

### 2. 邮箱验证码发送系统

**实现方式：**
- **验证码生成**：使用随机数字生成 6 位验证码
- **存储机制**：将验证码存储在内存中，使用 `map` 结构存储，键格式为 `email:type`
- **过期时间**：验证码默认 10 分钟过期
- **防重复发送**：同一邮箱在短时间内只能发送一次验证码
- **异步发送**：通过 Redis 消息队列异步发送邮件，避免阻塞主流程
- **使用验证**：验证码只能使用一次，使用后标记为已使用
- **自动清理**：启动定时协程，每分钟清理过期或已使用的验证码

**流程说明：**
1. 用户请求发送验证码，指定邮箱和验证码类型（注册/重置密码）
2. 系统生成 6 位随机验证码
3. 验证码存储到内存，设置 10 分钟过期时间
4. 将邮件发送任务推送到 Redis 消息队列
5. 消息队列消费者（Worker）异步发送邮件
6. 用户收到邮件后，使用验证码进行注册或重置密码
7. 系统从内存中验证验证码，标记为已使用
8. 过期或已使用的验证码会被定时清理

**关键文件：**
- `service/user.go` - 验证码生成、存储和验证逻辑
- `database/redis.go` - 消息队列推送
- `queue/worker.go` - 邮件发送消费者

**安全性措施：**
- 验证码存储在内存中，系统重启后会自动清除
- 每次发送验证码时检查是否有未使用且未过期的验证码
- 验证码使用后立即标记为已使用
- 邮箱发送失败时记录日志，不影响用户体验
- 使用 `sync.RWMutex` 确保并发操作安全

### 3. 帖子管理

**实现方式：**
- 支持帖子的创建、更新、删除
- 浏览量通过 Redis 消息队列异步更新
- 支持分页查询
- 支持按标题和内容关键词搜索帖子

**关键文件：**
- `models/post.go` - 帖子数据模型
- `service/post.go` - 帖子业务逻辑
- `handler/post.go` - 帖子接口

### 4. 评论系统

**实现方式：**
- 支持评论和嵌套回复（楼中楼）
- 评论按时间倒序排列
- 支持分页查询

**关键文件：**
- `models/post.go` - 评论数据模型
- `service/post.go` - 评论业务逻辑
- `handler/post.go` - 评论接口

### 5. 点赞系统

**实现方式：**
- 支持对帖子和评论的点赞/取消点赞
- 点赞数通过 Redis 消息队列异步更新
- 防止重复点赞

**关键文件：**
- `models/post.go` - 点赞数据模型
- `service/post.go` - 点赞业务逻辑
- `handler/post.go` - 点赞接口

### 6. 收藏系统

**实现方式：**
- 支持帖子收藏到自定义收藏夹
- 默认收藏夹自动创建
- 支持收藏夹的创建、更新、删除
- 支持按收藏夹查看收藏的帖子

**关键文件：**
- `models/post.go` - 收藏和收藏夹数据模型
- `service/post.go` - 收藏业务逻辑
- `handler/post.go` - 收藏接口

### 7. 拉黑系统

**实现方式：**
- 支持拉黑其他用户
- 防止重复拉黑
- 支持查看和取消拉黑

**关键文件：**
- `models/post.go` - 拉黑数据模型
- `service/post.go` - 拉黑业务逻辑
- `handler/post.go` - 拉黑接口

### 8. 头像上传

**实现方式：**
- 支持图片上传
- 文件类型和大小验证
- 自动生成唯一文件名
- 旧头像自动删除

**关键文件：**
- `config/config.go` - 上传配置
- `service/user.go` - 头像上传逻辑
- `handler/auth.go` - 头像上传接口

### 9. 消息队列

**实现方式：**
- 使用 Redis List 作为消息队列
- 支持邮件发送、浏览量更新、点赞数更新
- 多协程并发消费
- 优雅关闭

**关键文件：**
- `database/redis.go` - 消息队列操作
- `queue/worker.go` - 消息消费者

### 10. 管理员系统

**实现方式：**
- 用户模型添加 `is_admin` 字段标识管理员身份
- 使用 `AuthRequired` 中间件验证用户登录状态
- 使用 `AdminRequired` 中间件验证管理员权限
- 支持管理员删除帖子、删除评论、禁言/解禁用户
- 被禁言用户无法登录系统

**关键文件：**
- `middleware/auth.go` - 认证中间件
- `middleware/admin.go` - 管理员权限中间件
- `service/post.go` - 管理员业务逻辑
- `handler/post.go` - 管理员接口

**管理员权限说明：**
- 管理员可以删除任意帖子（不受用户限制）
- 管理员可以删除任意评论（不受用户限制）
- 管理员可以禁言/解禁任意用户
- 被禁言用户（status=0）无法登录系统

## 配置说明

### 配置文件（config.yaml）

```yaml
mysql:
  host: localhost
  port: 3306
  user: root
  password: password
  database: bbs

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

server:
  port: 8080

logger:
  level: info
  max_size: 100
  max_backups: 10
  max_age: 30
  compress: false
  output_path: ./logs

jwt:
  secret: your_jwt_secret_key

email:
  host: smtp.example.com
  port: 587
  username: your_email@example.com
  password: your_email_password
  from: your_email@example.com

upload:
  path: ./uploads
  max_size: 5242880  # 5MB
  allowed_ext: .jpg,.jpeg,.png,.gif,.webp
```

## 部署指南

### 1. 环境准备

- Go 1.20+
- MySQL 8.0+
- Redis 7.0+

### 2. 数据库配置

1. 创建 MySQL 数据库：
   ```sql
   CREATE DATABASE bbs CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
   ```

2. 配置 config.yaml 文件中的数据库连接信息

### 3. 启动服务

```bash
# 安装依赖
go mod tidy

# 构建
go build

# 启动服务
./bbsDemo
```

### 4. 自动迁移

服务启动时会自动执行数据库迁移，创建所需的表结构。

## 项目使用

### 1. API 接口使用

#### 1.1 认证流程

1. **发送验证码**：
   ```bash
   curl -X POST http://localhost:8080/api/auth/send-code \
     -H "Content-Type: application/json" \
     -d '{"email": "user@example.com", "type": "register"}'
   ```

2. **注册**：
   ```bash
   curl -X POST http://localhost:8080/api/auth/register \
     -H "Content-Type: application/json" \
     -d '{"email": "user@example.com", "password": "password123", "code": "123456"}'
   ```

3. **登录**：
   ```bash
   curl -X POST http://localhost:8080/api/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email": "user@example.com", "password": "password123"}'
   ```

4. **刷新令牌**：
   ```bash
   curl -X POST http://localhost:8080/api/auth/refresh \
     -H "Content-Type: application/json" \
     -d '{"refresh_token": "your_refresh_token"}'
   ```

5. **重置密码**：
   - 5.1 发送重置密码验证码：
     ```bash
     curl -X POST http://localhost:8080/api/auth/send-code \
       -H "Content-Type: application/json" \
       -d '{"email": "user@example.com", "type": "reset"}'
     ```
   - 5.2 使用验证码重置密码：
     ```bash
     curl -X POST http://localhost:8080/api/auth/reset-password \
       -H "Content-Type: application/json" \
       -d '{"email": "user@example.com", "code": "123456", "password": "newpassword123"}'
     ```

6. **注销账户**：
   - 6.1 发送注销验证码：
     ```bash
     curl -X POST http://localhost:8080/api/auth/send-code \
       -H "Content-Type: application/json" \
       -d '{"email": "user@example.com", "type": "delete"}'
     ```
   - 6.2 使用验证码注销账户：
     ```bash
     curl -X POST http://localhost:8080/api/auth/delete-account \
       -H "Content-Type: application/json" \
       -d '{"email": "user@example.com", "code": "123456"}'
     ```

#### 1.2 帖子操作

1. **创建帖子**：
   ```bash
   curl -X POST http://localhost:8080/api/posts \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer your_access_token" \
     -d '{"title": "Hello World", "content": "This is a test post"}'
   ```

2. **获取帖子列表**：
   ```bash
   curl http://localhost:8080/api/posts?page=1&page_size=10
   ```

3. **搜索帖子**：
   ```bash
   curl "http://localhost:8080/api/posts/search?keyword=Hello&page=1&page_size=10"
   ```

4. **获取帖子详情**：
   ```bash
   curl http://localhost:8080/api/posts/1
   ```

5. **更新帖子**：
   ```bash
   curl -X PUT http://localhost:8080/api/posts/1 \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer your_access_token" \
     -d '{"title": "Updated Title", "content": "Updated content"}'
   ```

6. **删除帖子**：
   ```bash
   curl -X DELETE http://localhost:8080/api/posts/1 \
     -H "Authorization: Bearer your_access_token"
   ```

#### 1.3 评论操作

1. **创建评论**：
   ```bash
   curl -X POST http://localhost:8080/api/comments \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer your_access_token" \
     -d '{"post_id": 1, "content": "Great post!"}'
   ```

2. **获取评论**：
   ```bash
   curl http://localhost:8080/api/posts/1/comments?page=1&page_size=10
   ```

#### 1.4 点赞操作

1. **点赞帖子**：
   ```bash
   curl -X POST http://localhost:8080/api/posts/1/like \
     -H "Authorization: Bearer your_access_token"
   ```

2. **取消点赞**：
   ```bash
   curl -X DELETE http://localhost:8080/api/posts/1/like \
     -H "Authorization: Bearer your_access_token"
   ```

#### 1.5 收藏操作

1. **创建收藏夹**：
   ```bash
   curl -X POST http://localhost:8080/api/folders \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer your_access_token" \
     -d '{"name": "技术文章"}'
   ```

2. **收藏帖子**：
   ```bash
   curl -X POST http://localhost:8080/api/favorites \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer your_access_token" \
     -d '{"post_id": 1, "folder_id": 1}'
   ```

3. **获取收藏**：
   ```bash
   curl http://localhost:8080/api/my/favorites?page=1&page_size=10 \
     -H "Authorization: Bearer your_access_token"
   ```

#### 1.6 个人资料操作

1. **获取当前用户信息**：
   ```bash
   curl http://localhost:8080/api/profile \
     -H "Authorization: Bearer your_access_token"
   ```

2. **获取其他用户信息**：
   ```bash
   curl http://localhost:8080/api/users/1234567890123456789 \
     -H "Authorization: Bearer your_access_token"
   ```

3. **更新昵称**：
   ```bash
   curl -X PUT http://localhost:8080/api/profile/nickname \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer your_access_token" \
     -d '{"nickname": "New Nickname"}'
   ```

4. **上传头像**：
   ```bash
   curl -X POST http://localhost:8080/api/profile/avatar \
     -H "Authorization: Bearer your_access_token" \
     -F "avatar=@/path/to/avatar.jpg"
   ```

5. **更新简介**：
   ```bash
   curl -X PUT http://localhost:8080/api/profile/bio \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer your_access_token" \
     -d '{"bio": "This is my bio"}'
   ```

### 2. 前端集成

1. **认证状态管理**：
   - 存储 access_token 和 refresh_token
   - 实现 token 过期自动刷新
   - 处理 401 错误（未授权）

2. **API 调用封装**：
   - 统一处理请求头（Authorization）
   - 统一处理错误响应
   - 实现请求拦截和响应拦截

3. **功能模块**：
   - 登录/注册页面
   - 帖子列表和详情页面
   - 评论和回复功能
   - 个人中心（资料、收藏、拉黑）

### 3. 常见问题

#### 3.1 验证码相关
- **问题**：验证码未收到
  **解决**：检查邮箱配置是否正确，查看服务端日志

- **问题**：验证码无效
  **解决**：确认验证码是否过期，是否输入正确

#### 3.2 认证相关
- **问题**：Token 过期
  **解决**：使用 refresh_token 刷新获取新 token

- **问题**：401 错误
  **解决**：检查 token 是否有效，是否在请求头中正确设置

#### 3.3 上传相关
- **问题**：头像上传失败
  **解决**：检查文件大小和类型是否符合要求

### 4. 开发建议

1. **本地开发**：
   - 使用 `go run main.go` 启动开发服务器
   - 开启 Gin 的 debug 模式查看详细日志

2. **测试**：
   - 使用 Postman 或 curl 测试 API 接口
   - 编写单元测试和集成测试

3. **部署**：
   - 配置生产环境的 config.yaml
   - 使用 PM2 或 systemd 管理服务
   - 配置反向代理（如 Nginx）

### 5. 服务器管理

1. **优雅关闭**：
   - 服务器支持优雅关闭，会等待所有工作处理完毕后再关闭
   - 默认等待超时为 30 秒

2. **通过指令关闭**：
   ```bash
   ./bbsDemo shutdown
   ```
   - 此命令会尝试关闭服务器
   - 当前实现为基础版本，实际生产环境建议使用进程管理工具
   - 推荐使用 PM2 或 systemd 等进程管理工具来管理服务的启动和关闭

### 6. 用户权限管理

1. **通过 MySQL 控制台设置管理员**：

   首先连接到 MySQL 数据库：
   ```bash
   mysql -u root -p bbs
   ```

   查看用户列表：
   ```sql
   SELECT id, email, nickname, is_admin FROM users;
   ```

   将指定用户设置为管理员：
   ```sql
   UPDATE users SET is_admin = TRUE WHERE id = <用户ID>;
   ```

   取消用户的管理员权限：
   ```sql
   UPDATE users SET is_admin = FALSE WHERE id = <用户ID>;
   ```

   示例：将邮箱为 admin@example.com 的用户设置为管理员：
   ```sql
   UPDATE users SET is_admin = TRUE WHERE email = 'admin@example.com';
   ```

2. **管理员权限说明**：
   - 管理员拥有更高的权限，可以执行一些普通用户无法进行的操作
   - 管理员权限在数据库中通过 `is_admin` 字段控制

## 开发说明

### 代码规范

- 使用 Go 标准包和常用第三方库
- 遵循 Go 代码风格
- 分层架构：handler → service → database
- 错误处理统一使用自定义错误类型
- 日志使用 Zap 结构化日志

### 扩展建议

1. **缓存优化**：使用 Redis 缓存热点数据
2. **搜索功能**：集成 Elasticsearch 实现全文搜索
3. **通知系统**：实现站内信和邮件通知
4. **权限管理**：添加角色和权限控制
5. **国际化**：支持多语言

## 许可证

MIT License
