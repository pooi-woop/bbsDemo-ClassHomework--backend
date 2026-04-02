# EyuForum（恶雨论坛）

## 项目概述

基于 Go 语言开发的 EyuForum（恶雨论坛）系统，具有完整的用户认证、帖子管理、评论系统、点赞功能、收藏功能、消息队列等特性。系统采用分层架构设计，使用 MySQL 作为主数据库，Redis 作为消息队列，Kafka 作为消息中间件，支持高并发场景。

## 技术栈

| 技术 | 版本 | 用途 |
|------|------|------|
| Go | 1.20+ | 主要开发语言 |
| Gin | v1.9.0 | Web 框架 |
| GORM | v1.25.0 | ORM 框架 |
| MySQL | 8.0+ | 主数据库 |
| Redis | 7.0+ | 缓存和消息队列 |
| Kafka | 4.0+ | 消息中间件 |
| Elasticsearch | 9.0+ | 搜索和AI知识库 |
| JWT | - | 认证令牌 |
| Zap | v1.24.0 | 日志库 |
| Viper | v1.15.0 | 配置管理 |
| Snowflake | - | 生成唯一ID |
| Eino | - | RAG流程实现 |
| Vue3 | - | 前端框架 |
| Element Plus | - | 前端组件库 |
| GitHub Actions | - | CI/CD |

## 项目结构

```
bbsDemo/
├── config/            # 配置管理
│   └── config.go      # 配置结构和加载
├── database/          # 数据库相关
│   ├── mysql.go       # MySQL 连接和迁移
│   ├── redis.go       # Redis 连接和消息队列
│   ├── kafka.go       # Kafka 连接和消息处理
│   └── elasticsearch.go # Elasticsearch 连接和索引
├── handler/           # HTTP 处理器
│   ├── auth.go        # 认证相关接口
│   ├── post.go        # 帖子相关接口
│   └── weather.go     # 天气相关接口
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
│   ├── post.go        # 帖子业务逻辑
│   ├── weather.go     # 天气业务逻辑
│   └── ai.go          # AI业务逻辑
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

**禁言功能说明：**
- 被禁言用户无法登录系统
- 被禁言用户无法发帖和评论
- 管理员可以通过接口禁言/解禁用户

**实现方式：**
- 使用 JWT 生成访问令牌（access token）和刷新令牌（refresh token）
- 密码采用 bcrypt 加盐哈希存储
- 邮箱验证码通过 Redis 消息队列异步发送
- 支持令牌刷新机制
- Refresh Token 存储在 Redis 中，提高安全性

**关键文件：**
- `utils/jwt.go` - JWT 令牌生成和解析
- `utils/hash.go` - 密码哈希和验证
- `service/user.go` - 认证业务逻辑
- `handler/auth.go` - 认证接口

### 2. 邮箱验证码发送系统

**实现方式：**
- **验证码生成**：使用随机数字生成 6 位验证码
- **存储机制**：将验证码存储在 Redis 中，使用 `map` 结构存储，键格式为 `email:type`
- **过期时间**：验证码默认 10 分钟过期
- **防重复发送**：同一邮箱在短时间内只能发送一次验证码
- **异步发送**：通过 Redis 消息队列异步发送邮件，避免阻塞主流程
- **使用验证**：验证码只能使用一次，使用后标记为已使用
- **自动清理**：启动定时协程，每分钟清理过期或已使用的验证码

**流程说明：**
1. 用户请求发送验证码，指定邮箱和验证码类型（注册/重置密码）
2. 系统生成 6 位随机验证码
3. 验证码存储到 Redis，设置 10 分钟过期时间
4. 将邮件发送任务推送到 Redis 消息队列
5. 消息队列消费者（Worker）异步发送邮件
6. 用户收到邮件后，使用验证码进行注册或重置密码
7. 系统从 Redis 中验证验证码，标记为已使用
8. 过期或已使用的验证码会被定时清理

**关键文件：**
- `service/user.go` - 验证码生成、存储和验证逻辑
- `database/redis.go` - 消息队列推送
- `queue/worker.go` - 邮件发送消费者

**安全性措施：**
- 验证码存储在 Redis 中，系统重启后会自动清除
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
- 帖子内容同步到 Elasticsearch，支持全文搜索

**关键文件：**
- `models/post.go` - 帖子数据模型
- `service/post.go` - 帖子业务逻辑
- `handler/post.go` - 帖子接口
- `database/elasticsearch.go` - 帖子索引同步

### 4. 评论系统

**实现方式：**
- 支持评论和嵌套回复（楼中楼）
- 评论按时间倒序排列
- 支持分页查询
- 评论内容同步到 Elasticsearch，支持搜索

**关键文件：**
- `models/post.go` - 评论数据模型
- `service/post.go` - 评论业务逻辑
- `handler/post.go` - 评论接口
- `database/elasticsearch.go` - 评论索引同步

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

### 11. Kafka 消息中间件

**实现方式：**
- 使用 Kafka 作为消息中间件，提高消息处理速度
- 支持高并发场景
- 与 Redis 配合使用，实现消息的可靠传递

**关键文件：**
- `database/kafka.go` - Kafka 连接和消息处理
- `queue/worker.go` - Kafka 消息消费者

### 12. Elasticsearch 搜索和 AI 知识库

**实现方式：**
- 帖子和评论内容同步到 Elasticsearch
- 支持全文搜索功能
- 通过 Eino 库实现 RAG 流程，支持 AI 问答功能

**关键文件：**
- `database/elasticsearch.go` - Elasticsearch 连接和索引
- `service/ai.go` - AI 业务逻辑
- `handler/ai.go` - AI 接口

### 13. 天气预报功能

**实现方式：**
- 使用高德地图 API 实现天气预报功能
- 支持根据 IP 地址自动获取地理位置
- 支持指定 IP 查询天气

**关键文件：**
- `service/weather.go` - 天气业务逻辑
- `handler/weather.go` - 天气接口

### 14. CI/CD 持续集成

**实现方式：**
- 通过 GitHub Actions 实现 CI/CD
- 自动构建、测试和部署

**关键文件：**
- `.github/workflows/deploy.yml` - CI/CD 配置

### 15. 前端实现

**实现方式：**
- 使用 Vue3 框架和 Element Plus 组件库实现前端页面
- 支持响应式设计
- 与后端 API 无缝集成

**关键功能：**
- 登录/注册页面
- 帖子列表和详情页面
- 评论和回复功能
- 个人中心（资料、收藏、拉黑）
- 天气信息展示

### 16. 安全性措施

**实现方式：**
- 密码采用加盐哈希方式存储
- 前后端 ID 通过 string 类型传输，防止 JSON 精度丢失
- 使用 Snowflake 算法生成唯一 ID，提高安全性
- JWT 令牌过期机制
- 验证码防重复发送
- 邮箱验证码存储在 Redis 中

**关键文件：**
- `utils/hash.go` - 密码哈希工具
- `utils/snowflake.go` - 雪花算法 ID 生成
- `utils/jwt.go` - JWT 工具

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

kafka:
  brokers: ["localhost:9092"]
  topic: bbs_demo
  group_id: bbs_demo_group

elasticsearch:
  host: localhost
  port: 9200
  index: eyuforum

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
  host: smtp.qq.com
  port: 587
  username: your_qq_email@qq.com
  password: your_qq_email_password
  from: your_qq_email@qq.com

upload:
  path: ./uploads
  max_size: 5242880  # 5MB
  allowed_ext: .jpg,.jpeg,.png,.gif,.webp

weather:
  gaode_api_key: your_gaode_api_key

ai:
  model: deepseek-chat
  api_base: https://api.deepseek.com
  api_key: your_ai_api_key
  timeout: 300
  max_tokens: 1000
  temperature: 0.7
```

## 部署指南

### 1. 环境准备

- Go 1.20+
- MySQL 8.0+
- Redis 7.0+
- Kafka 4.0+
- Elasticsearch 9.0+

### 2. 数据库配置

1. 创建 MySQL 数据库：
   ```sql
   CREATE DATABASE bbs CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
   ```

2. 配置 config.yaml 文件中的数据库连接信息

### 3. 服务启动

```bash
# 安装依赖
go mod tidy

# 构建
go build

# 启动服务
./bbsDemo
```

### 4. 关闭服务

在服务运行的控制台中输入以下指令关闭服务器：

```bash
shutdown
```

**说明：**
- 服务器支持优雅关闭，会等待所有正在处理的请求完成后再关闭
- 关闭时会自动清理资源，包括关闭数据库连接、停止消息队列工作线程等
- 此指令只能在服务运行的控制台中使用，不支持 HTTP 请求调用

### 5. 自动迁移

服务启动时会自动执行数据库迁移，创建所需的表结构。

### 6. 外部服务配置

#### 6.1 Kafka 配置
- 启动 Kafka 服务
- 创建主题 `bbs_demo`

#### 6.2 Elasticsearch 配置
- 启动 Elasticsearch 服务
- 服务启动时会自动创建索引 `eyuforum`

#### 6.3 高德地图 API 配置
- 注册高德地图开发者账号
- 获取 API Key
- 在 config.yaml 中配置 `weather.gaode_api_key`

#### 6.4 AI 服务配置
- 获取 DeepSeek API Key
- 在 config.yaml 中配置 `ai.api_key`

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

#### 1.7 天气接口

1. **获取当前天气**（自动识别IP）：
   ```bash
   curl http://localhost:8080/api/weather
   ```

2. **根据IP获取天气**：
   ```bash
   curl "http://localhost:8080/api/weather/by-ip?ip=202.106.0.20"
   ```

#### 1.8 AI 接口

1. **AI 问答**：
   ```bash
   curl -X POST http://localhost:8080/api/ai/ask \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer your_access_token" \
     -d '{"question": "如何使用这个论坛系统？"}'
   ```

2. **AI 问答（流式）**：
   ```bash
   curl -X POST http://localhost:8080/api/ai/ask/stream \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer your_access_token" \
     -d '{"question": "如何使用这个论坛系统？"}'
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
   - 天气信息展示
   - AI 问答功能

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

#### 3.4 天气相关
- **问题**：天气信息获取失败
  **解决**：检查高德地图 API Key 是否配置正确

#### 3.5 AI 相关
- **问题**：AI 问答失败
  **解决**：检查 AI API Key 是否配置正确

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

## 7. 许可证

本项目使用 MIT 许可证，详情请查看 [LICENSE](LICENSE) 文件。
