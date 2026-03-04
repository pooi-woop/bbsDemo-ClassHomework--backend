# API 接口文档

## 1. 认证接口

### 1.1 发送验证码

**请求：**
```http
POST /api/auth/send-code
Content-Type: application/json

{
  "email": "user@example.com",
  "type": "register"
}
```

**响应：**
```json
{
  "message": "Verification code sent"
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 400 | `{"error": "email is required"}` | 邮箱不能为空 |
| 400 | `{"error": "type is required"}` | 类型不能为空 |
| 400 | `{"error": "invalid email format"}` | 邮箱格式错误 |
| 400 | `{"error": "invalid type"}` | 类型错误，只能是 register 或 reset |
| 409 | `{"error": "user already exists"}` | 注册类型时用户已存在 |
| 429 | `{"error": "verification code already sent"}` | 验证码已发送，请勿重复请求 |
| 500 | `{"error": "Failed to push email to queue"}` | 邮件推送失败 |

**测试用例：**
1. 正常发送：`{"email": "test@example.com", "type": "register"}`
2. 邮箱格式错误：`{"email": "invalid-email", "type": "register"}`
3. 重复发送：连续两次发送到同一邮箱

### 1.2 注册

**请求：**
```http
POST /api/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "code": "123456"
}
```

**响应：**
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "nickname": "",
    "avatar": "",
    "is_verified": true
  }
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 400 | `{"error": "email is required"}` | 邮箱不能为空 |
| 400 | `{"error": "password is required"}` | 密码不能为空 |
| 400 | `{"error": "code is required"}` | 验证码不能为空 |
| 400 | `{"error": "invalid email format"}` | 邮箱格式错误 |
| 400 | `{"error": "password must be at least 6 characters"}` | 密码过短 |
| 400 | `{"error": "invalid verification code"}` | 验证码错误 |
| 400 | `{"error": "verification code expired"}` | 验证码过期 |
| 500 | `{"error": "Failed to hash password"}` | 密码哈希失败 |
| 500 | `{"error": "Failed to create user"}` | 用户创建失败 |

**测试用例：**
1. 正常注册：`{"email": "newuser@example.com", "password": "password123", "code": "123456"}`
2. 验证码错误：`{"email": "user@example.com", "password": "password123", "code": "654321"}`
3. 密码过短：`{"email": "user@example.com", "password": "123", "code": "123456"}`

### 1.3 登录

**请求：**
```http
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

**响应：**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "nickname": "",
    "avatar": ""
  }
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 400 | `{"error": "email is required"}` | 邮箱不能为空 |
| 400 | `{"error": "password is required"}` | 密码不能为空 |
| 400 | `{"error": "invalid email format"}` | 邮箱格式错误 |
| 401 | `{"error": "user not found"}` | 用户不存在 |
| 401 | `{"error": "invalid password"}` | 密码错误 |
| 401 | `{"error": "email not verified"}` | 邮箱未验证 |
| 500 | `{"error": "Failed to generate tokens"}` | 令牌生成失败 |
| 500 | `{"error": "Failed to save refresh token"}` | 刷新令牌保存失败 |

**测试用例：**
1. 正常登录：`{"email": "user@example.com", "password": "password123"}`
2. 邮箱不存在：`{"email": "nonexistent@example.com", "password": "password123"}`
3. 密码错误：`{"email": "user@example.com", "password": "wrongpassword"}`

### 1.4 刷新令牌

**请求：**
```http
POST /api/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**响应：**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 400 | `{"error": "refresh_token is required"}` | 刷新令牌不能为空 |
| 401 | `{"error": "invalid token"}` | 令牌无效 |
| 401 | `{"error": "token expired"}` | 令牌过期 |
| 500 | `{"error": "Failed to generate tokens"}` | 令牌生成失败 |
| 500 | `{"error": "Failed to save refresh token"}` | 刷新令牌保存失败 |

**测试用例：**
1. 正常刷新：使用有效的 refresh_token
2. 无效令牌：使用伪造的 refresh_token
3. 过期令牌：使用过期的 refresh_token

### 1.5 退出登录

**请求：**
```http
POST /api/logout
Content-Type: application/json
Authorization: Bearer <access_token>

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**响应：**
```json
{
  "message": "Logged out"
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 401 | `{"error": "Invalid authorization header format"}` | 认证头格式错误 |
| 401 | `{"error": "Invalid token"}` | 令牌无效 |
| 500 | `{"error": "Failed to revoke token"}` | 令牌撤销失败 |

**测试用例：**
1. 正常退出：`{"refresh_token": "valid_refresh_token"}`
2. 无效令牌：`{"refresh_token": "invalid_refresh_token"}`
3. 缺少令牌：不提供 refresh_token

### 1.6 退出所有设备

**请求：**
```http
POST /api/logout-all
Authorization: Bearer <access_token>
```

**响应：**
```json
{
  "message": "Logged out from all devices"
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 401 | `{"error": "Invalid authorization header format"}` | 认证头格式错误 |
| 401 | `{"error": "Invalid token"}` | 令牌无效 |
| 500 | `{"error": "Failed to revoke all tokens"}` | 令牌撤销失败 |

**测试用例：**
1. 正常退出：发送请求

## 2. 个人资料接口

### 2.1 获取个人资料

**请求：**
```http
GET /api/profile
Authorization: Bearer <access_token>
```

**响应：**
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "nickname": "User",
    "avatar": "/uploads/avatar_1_1234567890.jpg",
    "is_verified": true,
    "last_login_at": "2023-01-01T00:00:00Z",
    "last_login_ip": "127.0.0.1"
  }
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 401 | `{"error": "Invalid authorization header format"}` | 认证头格式错误 |
| 401 | `{"error": "Invalid token"}` | 令牌无效 |

**测试用例：**
1. 正常获取：发送请求

### 2.2 更新昵称

**请求：**
```http
PUT /api/profile/nickname
Content-Type: application/json
Authorization: Bearer <access_token>

{
  "nickname": "New Nickname"
}
```

**响应：**
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "nickname": "New Nickname",
    "avatar": "/uploads/avatar_1_1234567890.jpg"
  }
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 400 | `{"error": "nickname is required"}` | 昵称不能为空 |
| 400 | `{"error": "nickname must not exceed 50 characters"}` | 昵称过长 |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 404 | `{"error": "User not found"}` | 用户不存在 |
| 500 | `{"error": "Failed to update nickname"}` | 更新失败 |

**测试用例：**
1. 正常更新：`{"nickname": "New Nickname"}`
2. 昵称过长：`{"nickname": "A" × 51}`
3. 空昵称：`{"nickname": ""}`

### 2.3 上传头像

**请求：**
```http
POST /api/profile/avatar
Content-Type: multipart/form-data
Authorization: Bearer <access_token>

[Form data with file field "avatar"]
```

**响应：**
```json
{
  "message": "Avatar uploaded successfully",
  "avatar": "/uploads/avatar_1_1234567890.jpg"
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 400 | `{"error": "Failed to get file"}` | 文件获取失败 |
| 400 | `{"error": "File too large"}` | 文件过大 |
| 400 | `{"error": "Invalid file type"}` | 文件类型错误 |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 500 | `{"error": "Failed to create upload directory"}` | 目录创建失败 |
| 500 | `{"error": "Failed to create file"}` | 文件创建失败 |
| 500 | `{"error": "Failed to save file"}` | 文件保存失败 |
| 500 | `{"error": "Failed to update avatar"}` | 头像更新失败 |

**测试用例：**
1. 正常上传：选择有效的图片文件
2. 文件过大：选择超过 5MB 的文件
3. 无效文件类型：选择非图片文件

## 3. 帖子接口

### 3.1 列表帖子

**请求：**
```http
GET /api/posts?page=1&page_size=10
```

**响应：**
```json
{
  "posts": [
    {
      "id": 1,
      "user_id": 1,
      "title": "Hello World",
      "content": "This is a test post",
      "views": 10,
      "created_at": "2023-01-01T00:00:00Z",
      "user": {
        "id": 1,
        "email": "user@example.com",
        "nickname": "User"
      }
    }
  ],
  "total": 1
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 500 | `{"error": "Failed to get posts"}` | 获取失败 |

**测试用例：**
1. 第一页：`GET /api/posts?page=1&page_size=10`
2. 第二页：`GET /api/posts?page=2&page_size=10`
3. 大页面：`GET /api/posts?page=1&page_size=100`

### 3.2 获取帖子

**请求：**
```http
GET /api/posts/1
```

**响应：**
```json
{
  "id": 1,
  "user_id": 1,
  "title": "Hello World",
  "content": "This is a test post",
  "views": 11,
  "created_at": "2023-01-01T00:00:00Z",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "nickname": "User"
  }
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 404 | `{"error": "post not found"}` | 帖子不存在 |
| 500 | `{"error": "Failed to get post"}` | 获取失败 |

**测试用例：**
1. 存在的帖子：`GET /api/posts/1`
2. 不存在的帖子：`GET /api/posts/999`

### 3.3 创建帖子

**请求：**
```http
POST /api/posts
Content-Type: application/json
Authorization: Bearer <access_token>

{
  "title": "New Post",
  "content": "This is a new post"
}
```

**响应：**
```json
{
  "id": 2,
  "user_id": 1,
  "title": "New Post",
  "content": "This is a new post",
  "views": 0,
  "created_at": "2023-01-02T00:00:00Z"
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 400 | `{"error": "title is required"}` | 标题不能为空 |
| 400 | `{"error": "content is required"}` | 内容不能为空 |
| 400 | `{"error": "title must not exceed 200 characters"}` | 标题过长 |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 500 | `{"error": "Failed to create post"}` | 创建失败 |

**测试用例：**
1. 正常创建：`{"title": "Test Post", "content": "Test content"}`
2. 标题过长：`{"title": "A" × 201, "content": "Test"}`
3. 内容为空：`{"title": "Test", "content": ""}`

### 3.4 更新帖子

**请求：**
```http
PUT /api/posts/1
Content-Type: application/json
Authorization: Bearer <access_token>

{
  "title": "Updated Title",
  "content": "Updated content"
}
```

**响应：**
```json
{
  "id": 1,
  "user_id": 1,
  "title": "Updated Title",
  "content": "Updated content",
  "views": 10,
  "created_at": "2023-01-01T00:00:00Z"
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 400 | `{"error": "title must not exceed 200 characters"}` | 标题过长 |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 403 | `{"error": "unauthorized"}` | 无权限更新 |
| 404 | `{"error": "post not found"}` | 帖子不存在 |
| 500 | `{"error": "Failed to update post"}` | 更新失败 |

**测试用例：**
1. 正常更新：`{"title": "Updated Title", "content": "Updated content"}`
2. 无权限更新：使用其他用户的 token
3. 不存在的帖子：`PUT /api/posts/999`

### 3.5 删除帖子

**请求：**
```http
DELETE /api/posts/1
Authorization: Bearer <access_token>
```

**响应：**
```json
{
  "message": "Post deleted"
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 403 | `{"error": "unauthorized"}` | 无权限删除 |
| 404 | `{"error": "post not found"}` | 帖子不存在 |
| 500 | `{"error": "Failed to delete post"}` | 删除失败 |

**测试用例：**
1. 正常删除：`DELETE /api/posts/1`
2. 无权限删除：使用其他用户的 token
3. 不存在的帖子：`DELETE /api/posts/999`

## 4. 评论接口

### 4.1 获取评论

**请求：**
```http
GET /api/posts/1/comments?page=1&page_size=10
```

**响应：**
```json
{
  "comments": [
    {
      "id": 1,
      "post_id": 1,
      "comment_id": null,
      "user_id": 1,
      "content": "Great post!",
      "created_at": "2023-01-01T00:00:00Z",
      "user": {
        "id": 1,
        "email": "user@example.com",
        "nickname": "User"
      }
    }
  ],
  "total": 1
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 500 | `{"error": "Failed to get comments"}` | 获取失败 |

**测试用例：**
1. 第一页：`GET /api/posts/1/comments?page=1&page_size=10`
2. 第二页：`GET /api/posts/1/comments?page=2&page_size=10`
3. 不存在的帖子：`GET /api/posts/999/comments`

### 4.2 获取回复

**请求：**
```http
GET /api/comments/1/replies?page=1&page_size=10
```

**响应：**
```json
{
  "comments": [
    {
      "id": 2,
      "post_id": 1,
      "comment_id": 1,
      "user_id": 2,
      "content": "Thanks!",
      "created_at": "2023-01-01T01:00:00Z",
      "user": {
        "id": 2,
        "email": "user2@example.com",
        "nickname": "User2"
      }
    }
  ],
  "total": 1
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 500 | `{"error": "Failed to get replies"}` | 获取失败 |

**测试用例：**
1. 正常获取：`GET /api/comments/1/replies?page=1&page_size=10`
2. 不存在的评论：`GET /api/comments/999/replies`

### 4.3 创建评论

**请求：**
```http
POST /api/comments
Content-Type: application/json
Authorization: Bearer <access_token>

{
  "post_id": 1,
  "content": "This is a comment"
}
```

**响应：**
```json
{
  "id": 3,
  "post_id": 1,
  "comment_id": null,
  "user_id": 1,
  "content": "This is a comment",
  "created_at": "2023-01-02T00:00:00Z"
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 400 | `{"error": "content is required"}` | 内容不能为空 |
| 400 | `{"error": "post_id or comment_id is required"}` | 必须指定帖子或评论 |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 500 | `{"error": "Failed to create comment"}` | 创建失败 |

**测试用例：**
1. 正常评论：`{"post_id": 1, "content": "Test comment"}`
2. 回复评论：`{"comment_id": 1, "content": "Test reply"}`
3. 内容为空：`{"post_id": 1, "content": ""}`

### 4.4 删除评论

**请求：**
```http
DELETE /api/comments/1
Authorization: Bearer <access_token>
```

**响应：**
```json
{
  "message": "Comment deleted"
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 403 | `{"error": "unauthorized"}` | 无权限删除 |
| 404 | `{"error": "comment not found"}` | 评论不存在 |
| 500 | `{"error": "Failed to delete comment"}` | 删除失败 |

**测试用例：**
1. 正常删除：`DELETE /api/comments/1`
2. 无权限删除：使用其他用户的 token
3. 不存在的评论：`DELETE /api/comments/999`

## 5. 点赞接口

### 5.1 点赞帖子

**请求：**
```http
POST /api/posts/1/like
Authorization: Bearer <access_token>
```

**响应：**
```json
{
  "message": "Post liked"
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 409 | `{"error": "already liked"}` | 已经点赞 |
| 500 | `{"error": "Failed to like post"}` | 点赞失败 |

**测试用例：**
1. 正常点赞：`POST /api/posts/1/like`
2. 重复点赞：`POST /api/posts/1/like`（已点赞后再点赞）
3. 不存在的帖子：`POST /api/posts/999/like`

### 5.2 取消点赞帖子

**请求：**
```http
DELETE /api/posts/1/like
Authorization: Bearer <access_token>
```

**响应：**
```json
{
  "message": "Post unliked"
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 404 | `{"error": "not liked"}` | 未点赞 |
| 500 | `{"error": "Failed to unlike post"}` | 取消点赞失败 |

**测试用例：**
1. 正常取消：`DELETE /api/posts/1/like`
2. 未点赞取消：`DELETE /api/posts/1/like`（未点赞时取消）

### 5.3 点赞评论

**请求：**
```http
POST /api/comments/1/like
Authorization: Bearer <access_token>
```

**响应：**
```json
{
  "message": "Comment liked"
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 409 | `{"error": "already liked"}` | 已经点赞 |
| 500 | `{"error": "Failed to like comment"}` | 点赞失败 |

**测试用例：**
1. 正常点赞：`POST /api/comments/1/like`
2. 重复点赞：`POST /api/comments/1/like`（已点赞后再点赞）
3. 不存在的评论：`POST /api/comments/999/like`

### 5.4 取消点赞评论

**请求：**
```http
DELETE /api/comments/1/like
Authorization: Bearer <access_token>
```

**响应：**
```json
{
  "message": "Comment unliked"
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 404 | `{"error": "not liked"}` | 未点赞 |
| 500 | `{"error": "Failed to unlike comment"}` | 取消点赞失败 |

**测试用例：**
1. 正常取消：`DELETE /api/comments/1/like`
2. 未点赞取消：`DELETE /api/comments/1/like`（未点赞时取消）

## 6. 收藏接口

### 6.1 获取收藏夹

**请求：**
```http
GET /api/folders
Authorization: Bearer <access_token>
```

**响应：**
```json
{
  "folders": [
    {
      "id": 1,
      "user_id": 1,
      "name": "默认收藏夹",
      "is_default": true
    }
  ]
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 500 | `{"error": "Failed to get folders"}` | 获取失败 |

**测试用例：**
1. 正常获取：`GET /api/folders`

### 6.2 创建收藏夹

**请求：**
```http
POST /api/folders
Content-Type: application/json
Authorization: Bearer <access_token>

{
  "name": "技术文章"
}
```

**响应：**
```json
{
  "id": 2,
  "user_id": 1,
  "name": "技术文章",
  "is_default": false
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 400 | `{"error": "name is required"}` | 名称不能为空 |
| 400 | `{"error": "name must not exceed 50 characters"}` | 名称过长 |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 409 | `{"error": "folder already exists"}` | 收藏夹已存在 |
| 500 | `{"error": "Failed to create folder"}` | 创建失败 |

**测试用例：**
1. 正常创建：`{"name": "技术文章"}`
2. 名称过长：`{"name": "A" × 51}`
3. 重复创建：`{"name": "技术文章"}`（已存在后再创建）

### 6.3 更新收藏夹

**请求：**
```http
PUT /api/folders/2
Content-Type: application/json
Authorization: Bearer <access_token>

{
  "name": "技术分享"
}
```

**响应：**
```json
{
  "id": 2,
  "user_id": 1,
  "name": "技术分享",
  "is_default": false
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 400 | `{"error": "name is required"}` | 名称不能为空 |
| 400 | `{"error": "name must not exceed 50 characters"}` | 名称过长 |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 404 | `{"error": "folder not found"}` | 收藏夹不存在 |
| 403 | `{"error": "folder not yours"}` | 无权限更新 |
| 409 | `{"error": "folder already exists"}` | 名称已存在 |
| 500 | `{"error": "Failed to update folder"}` | 更新失败 |

**测试用例：**
1. 正常更新：`{"name": "技术分享"}`
2. 无权限更新：使用其他用户的 token
3. 不存在的收藏夹：`PUT /api/folders/999`

### 6.4 删除收藏夹

**请求：**
```http
DELETE /api/folders/2
Authorization: Bearer <access_token>
```

**响应：**
```json
{
  "message": "Folder deleted"
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 404 | `{"error": "folder not found"}` | 收藏夹不存在 |
| 403 | `{"error": "folder not yours"}` | 无权限删除 |
| 400 | `{"error": "cannot delete default folder"}` | 不能删除默认收藏夹 |
| 500 | `{"error": "Failed to delete folder"}` | 删除失败 |

**测试用例：**
1. 正常删除：`DELETE /api/folders/2`
2. 无权限删除：使用其他用户的 token
3. 删除默认收藏夹：`DELETE /api/folders/1`

### 6.5 收藏帖子

**请求：**
```http
POST /api/favorites
Content-Type: application/json
Authorization: Bearer <access_token>

{
  "post_id": 1,
  "folder_id": 1
}
```

**响应：**
```json
{
  "message": "Post favorited"
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 400 | `{"error": "post_id is required"}` | 帖子 ID 不能为空 |
| 400 | `{"error": "folder_id is required"}` | 收藏夹 ID 不能为空 |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 404 | `{"error": "post not found"}` | 帖子不存在 |
| 404 | `{"error": "folder not found"}` | 收藏夹不存在 |
| 403 | `{"error": "folder not yours"}` | 收藏夹不属于用户 |
| 409 | `{"error": "already favorited"}` | 已经收藏 |
| 500 | `{"error": "Failed to favorite post"}` | 收藏失败 |

**测试用例：**
1. 正常收藏：`{"post_id": 1, "folder_id": 1}`
2. 重复收藏：`{"post_id": 1, "folder_id": 1}`（已收藏后再收藏）
3. 不存在的帖子：`{"post_id": 999, "folder_id": 1}`

### 6.6 取消收藏

**请求：**
```http
DELETE /api/posts/1/favorite
Authorization: Bearer <access_token>
```

**响应：**
```json
{
  "message": "Post unfavorited"
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 404 | `{"error": "not favorited"}` | 未收藏 |
| 500 | `{"error": "Failed to unfavorite post"}` | 取消收藏失败 |

**测试用例：**
1. 正常取消：`DELETE /api/posts/1/favorite`
2. 未收藏取消：`DELETE /api/posts/1/favorite`（未收藏时取消）

### 6.7 移动收藏

**请求：**
```http
PUT /api/posts/1/favorite
Content-Type: application/json
Authorization: Bearer <access_token>

{
  "folder_id": 2
}
```

**响应：**
```json
{
  "message": "Favorite moved"
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 400 | `{"error": "folder_id is required"}` | 收藏夹 ID 不能为空 |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 404 | `{"error": "not favorited"}` | 未收藏 |
| 404 | `{"error": "folder not found"}` | 收藏夹不存在 |
| 403 | `{"error": "folder not yours"}` | 收藏夹不属于用户 |
| 500 | `{"error": "Failed to move favorite"}` | 移动失败 |

**测试用例：**
1. 正常移动：`{"folder_id": 2}`
2. 不存在的收藏夹：`{"folder_id": 999}`
3. 未收藏移动：对未收藏的帖子操作

### 6.8 获取收藏

**请求：**
```http
GET /api/my/favorites?page=1&page_size=10
Authorization: Bearer <access_token>
```

**响应：**
```json
{
  "favorites": [
    {
      "post": {
        "id": 1,
        "title": "Hello World",
        "content": "This is a test post",
        "views": 10,
        "created_at": "2023-01-01T00:00:00Z",
        "user": {
          "id": 1,
          "email": "user@example.com",
          "nickname": "User"
        }
      },
      "folder_id": 1,
      "folder_name": "默认收藏夹",
      "created_at": "2023-01-01T00:00:00Z"
    }
  ],
  "total": 1
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 500 | `{"error": "Failed to get favorites"}` | 获取失败 |

**测试用例：**
1. 正常获取：`GET /api/my/favorites?page=1&page_size=10`

### 6.9 按收藏夹获取收藏

**请求：**
```http
GET /api/folders/1/posts?page=1&page_size=10
Authorization: Bearer <access_token>
```

**响应：**
```json
{
  "favorites": [
    {
      "post": {
        "id": 1,
        "title": "Hello World",
        "content": "This is a test post",
        "views": 10,
        "created_at": "2023-01-01T00:00:00Z",
        "user": {
          "id": 1,
          "email": "user@example.com",
          "nickname": "User"
        }
      },
      "created_at": "2023-01-01T00:00:00Z"
    }
  ],
  "total": 1
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 404 | `{"error": "folder not found"}` | 收藏夹不存在 |
| 403 | `{"error": "folder not yours"}` | 收藏夹不属于用户 |
| 500 | `{"error": "Failed to get favorites"}` | 获取失败 |

**测试用例：**
1. 正常获取：`GET /api/folders/1/posts?page=1&page_size=10`
2. 不存在的收藏夹：`GET /api/folders/999/posts`
3. 无权限访问：使用其他用户的 token

## 7. 拉黑接口

### 7.1 拉黑用户

**请求：**
```http
POST /api/users/2/block
Authorization: Bearer <access_token>
```

**响应：**
```json
{
  "message": "User blocked"
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 400 | `{"error": "cannot block yourself"}` | 不能拉黑自己 |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 409 | `{"error": "already blocked"}` | 已经拉黑 |
| 500 | `{"error": "Failed to block user"}` | 拉黑失败 |

**测试用例：**
1. 正常拉黑：`POST /api/users/2/block`
2. 重复拉黑：`POST /api/users/2/block`（已拉黑后再拉黑）
3. 拉黑自己：`POST /api/users/1/block`（尝试拉黑自己）

### 7.2 取消拉黑

**请求：**
```http
DELETE /api/users/2/block
Authorization: Bearer <access_token>
```

**响应：**
```json
{
  "message": "User unblocked"
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 404 | `{"error": "not blocked"}` | 未拉黑 |
| 500 | `{"error": "Failed to unblock user"}` | 取消拉黑失败 |

**测试用例：**
1. 正常取消：`DELETE /api/users/2/block`
2. 未拉黑取消：`DELETE /api/users/2/block`（未拉黑时取消）

### 7.3 获取拉黑列表

**请求：**
```http
GET /api/my/blocked?page=1&page_size=10
Authorization: Bearer <access_token>
```

**响应：**
```json
{
  "users": [
    {
      "id": 2,
      "email": "user2@example.com",
      "nickname": "User2",
      "avatar": ""
    }
  ],
  "total": 1
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 500 | `{"error": "Failed to get blocked users"}` | 获取失败 |

**测试用例：**
1. 正常获取：`GET /api/my/blocked?page=1&page_size=10`

## 8. 我的帖子

### 8.1 获取我的帖子

**请求：**
```http
GET /api/my/posts?page=1&page_size=10
Authorization: Bearer <access_token>
```

**响应：**
```json
{
  "posts": [
    {
      "id": 1,
      "user_id": 1,
      "title": "Hello World",
      "content": "This is a test post",
      "views": 10,
      "created_at": "2023-01-01T00:00:00Z"
    }
  ],
  "total": 1
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 500 | `{"error": "Failed to get posts"}` | 获取失败 |

**测试用例：**
1. 正常获取：`GET /api/my/posts?page=1&page_size=10`
