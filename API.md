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
| 400 | `{"error": "invalid type"}` | 类型错误，只能是 register、reset 或 delete |
| 409 | `{"error": "user already exists"}` | 注册类型时用户已存在 |
| 429 | `{"error": "verification code already sent"}` | 验证码已发送，请勿重复请求 |
| 500 | `{"error": "Failed to push email to queue"}` | 邮件推送失败 |

**测试用例：**
1. 正常发送：`{"email": "test@example.com", "type": "register"}`
2. 邮箱格式错误：`{"email": "invalid-email", "type": "register"}`
3. 重复发送：连续两次发送到同一邮箱
4. 重置密码发送：`{"email": "user@example.com", "type": "reset"}`
5. 删除账户发送：`{"email": "user@example.com", "type": "delete"}`

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
    "id": "1234567890123456789",
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
  "message": "Login successful",
  "tokens": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 400 | `{"error": "email is required"}` | 邮箱不能为空 |
| 400 | `{"error": "password is required"}` | 密码不能为空 |
| 401 | `{"error": "Invalid email or password"}` | 邮箱或密码错误 |
| 403 | `{"error": "Email not verified"}` | 邮箱未验证 |
| 403 | `{"error": "account is banned"}` | 账号被禁言 |
| 500 | `{"error": "Login failed"}` | 登录失败 |

**测试用例：**
1. 正常登录：`{"email": "user@example.com", "password": "password123"}`
2. 密码错误：`{"email": "user@example.com", "password": "wrongpassword"}`
3. 账号被禁言：使用被禁言的账号登录

### 1.4 刷新令牌

**请求：**
```http
POST /api/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**响应：**
```json
{
  "message": "Token refreshed",
  "tokens": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 400 | `{"error": "refresh_token is required"}` | 刷新令牌不能为空 |
| 401 | `{"error": "Invalid token"}` | 令牌无效 |
| 401 | `{"error": "Token expired"}` | 令牌过期 |
| 500 | `{"error": "Token refresh failed"}` | 刷新失败 |

**测试用例：**
1. 正常刷新：`{"refresh_token": "valid_refresh_token"}`
2. 无效令牌：`{"refresh_token": "invalid_token"}`
3. 过期令牌：`{"refresh_token": "expired_token"}`

### 1.5 登出

**请求：**
```http
POST /api/logout
Content-Type: application/json
Authorization: Bearer <access_token>

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**响应：**
```json
{
  "message": "Logout successful"
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 500 | `{"error": "Logout failed"}` | 登出失败 |

**测试用例：**
1. 正常登出：提供有效的 access_token 和 refresh_token
2. 无效令牌：使用无效的 refresh_token

### 1.6 登出所有设备

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
| 500 | `{"error": "Logout failed"}` | 登出失败 |

**测试用例：**
1. 正常登出：提供有效的 access_token
2. 多设备登录后登出：在其他设备上登录后，使用此接口登出所有设备

### 1.7 重置密码

**请求：**
```http
POST /api/auth/reset-password
Content-Type: application/json

{
  "email": "user@example.com",
  "code": "123456",
  "password": "newpassword123"
}
```

**响应：**
```json
{
  "message": "Password reset successfully"
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 400 | `{"error": "invalid verification code"}` | 验证码错误 |
| 400 | `{"error": "verification code already used"}` | 验证码已使用 |
| 400 | `{"error": "verification code expired"}` | 验证码过期 |
| 404 | `{"error": "User not found"}` | 用户不存在 |
| 500 | `{"error": "Failed to reset password"}` | 重置密码失败 |

**测试用例：**
1. 正常重置：`{"email": "user@example.com", "code": "123456", "password": "newpassword123"}`
2. 验证码错误：`{"email": "user@example.com", "code": "654321", "password": "newpassword123"}`
3. 验证码过期：使用过期的验证码

### 1.8 删除账户

**请求：**
```http
POST /api/auth/delete-account
Content-Type: application/json

{
  "email": "user@example.com",
  "code": "123456"
}
```

**响应：**
```json
{
  "message": "Account deleted successfully"
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 400 | `{"error": "invalid verification code"}` | 验证码错误 |
| 400 | `{"error": "verification code already used"}` | 验证码已使用 |
| 400 | `{"error": "verification code expired"}` | 验证码过期 |
| 404 | `{"error": "User not found"}` | 用户不存在 |
| 500 | `{"error": "Failed to delete account"}` | 删除账户失败 |

**测试用例：**
1. 正常删除：`{"email": "user@example.com", "code": "123456"}`
2. 验证码错误：`{"email": "user@example.com", "code": "654321"}`
3. 验证码过期：使用过期的验证码

## 2. 用户接口

### 2.1 获取用户信息

**请求：**
```http
GET /api/profile
Authorization: Bearer <access_token>
```

**响应：**
```json
{
  "user": {
    "id": "1234567890123456789",
    "email": "user@example.com",
    "nickname": "User",
    "bio": "",
    "avatar": "/uploads/avatar_1234567890123456789_1234567890.jpg",
    "status": 1,
    "is_admin": false,
    "is_verified": true,
    "created_at": "2023-01-01T00:00:00Z",
    "last_login_at": "2023-01-01T00:00:00Z"
  }
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 404 | `{"error": "User not found"}` | 用户不存在 |
| 500 | `{"error": "Failed to get profile"}` | 获取失败 |

**测试用例：**
1. 正常获取：提供有效的 access_token
2. 用户不存在：删除用户后尝试获取

### 2.2 获取其他用户信息

**请求：**
```http
GET /api/users/1234567890123456789
Authorization: Bearer <access_token>
```

**响应：**
```json
{
  "user": {
    "id": "1234567890123456789",
    "email": "user@example.com",
    "nickname": "User",
    "bio": "",
    "avatar": "/uploads/avatar_1234567890123456789_1234567890.jpg",
    "status": 1,
    "is_admin": false,
    "is_verified": true,
    "created_at": "2023-01-01T00:00:00Z",
    "last_login_at": "2023-01-01T00:00:00Z"
  }
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 400 | `{"error": "Invalid user ID"}` | 用户ID格式错误 |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 404 | `{"error": "User not found"}` | 用户不存在 |
| 500 | `{"error": "Failed to get user info"}` | 获取失败 |

**测试用例：**
1. 正常获取：提供有效的用户ID
2. 用户不存在：使用不存在的用户ID
3. ID格式错误：使用非数字ID

### 2.3 更新昵称

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
    "id": "1234567890123456789",
    "email": "user@example.com",
    "nickname": "New Nickname",
    "avatar": "",
    "status": 1,
    "is_admin": false,
    "is_verified": true
  }
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 400 | `{"error": "nickname is required"}` | 昵称不能为空 |
| 400 | `{"error": "nickname must be at most 50 characters"}` | 昵称过长 |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 404 | `{"error": "User not found"}` | 用户不存在 |
| 500 | `{"error": "Failed to update nickname"}` | 更新失败 |

**测试用例：**
1. 正常更新：`{"nickname": "New Nickname"}`
2. 昵称过长：`{"nickname": "a very long nickname that exceeds the maximum length of 50 characters"}`
3. 空昵称：`{"nickname": ""}`

### 2.3 上传头像

**请求：**
```http
POST /api/profile/avatar
Content-Type: multipart/form-data
Authorization: Bearer <access_token>

avatar: <file>
```

**响应：**
```json
{
  "message": "Avatar uploaded successfully",
  "avatar": "/uploads/avatar_1234567890123456789_1234567890.jpg"
}
```

**注意：** 所有 ID 字段（如 `id`、`user_id`、`post_id` 等）在 JSON 响应中均为**字符串类型**，以避免 JavaScript 数字精度丢失问题。
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 400 | `{"error": "Failed to get file"}` | 文件获取失败 |
| 400 | `{"error": "File too large"}` | 文件过大（超过 5MB） |
| 400 | `{"error": "Invalid file type"}` | 文件类型不支持（仅支持 jpg, jpeg, png, gif, webp） |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 500 | `{"error": "Failed to upload avatar"}` | 上传失败 |

**测试用例：**
1. 正常上传：上传 jpg/png 图片
2. 文件过大：上传 10MB 的图片
3. 无效类型：上传 pdf 文件

### 2.4 更新简介

**请求：**
```http
PUT /api/profile/bio
Content-Type: application/json
Authorization: Bearer <access_token>

{
  "bio": "This is my bio"
}
```

**响应：**
```json
{
  "user": {
    "id": "1234567890123456789",
    "email": "user@example.com",
    "nickname": "User",
    "bio": "This is my bio",
    "avatar": "",
    "status": 1,
    "is_admin": false,
    "is_verified": true
  }
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 400 | `{"error": "bio must be at most 500 characters"}` | 简介过长 |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 404 | `{"error": "User not found"}` | 用户不存在 |
| 500 | `{"error": "Failed to update bio"}` | 更新失败 |

**测试用例：**
1. 正常更新：`{"bio": "This is my bio"}`
2. 简介过长：`{"bio": "a very long bio that exceeds the maximum length of 500 characters and should be rejected by the server. This is just a test to see if the validation works correctly. The bio should not be longer than 500 characters. Let's see if this is enough to trigger the error."}`

## 3. 帖子接口

### 3.1 获取帖子列表

**请求：**
```http
GET /api/posts?page=1&page_size=10
```

**响应：**
```json
{
  "posts": [
    {
      "id": "1234567890123456789",
      "user_id": "1234567890123456789",
      "title": "Hello World",
      "content": "This is a test post",
      "views": 10,
      "like_count": 5,
      "comment_count": 3,
      "created_at": "2023-01-01T00:00:00Z",
      "updated_at": "2023-01-01T00:00:00Z",
      "is_liked": false,
      "is_favorited": false,
      "user": {
        "id": "1234567890123456789",
        "email": "user@example.com",
        "nickname": "User",
        "avatar": ""
      }
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 10
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 500 | `{"error": "Failed to get posts"}` | 获取失败 |

**测试用例：**
1. 正常获取：`GET /api/posts?page=1&page_size=10`
2. 分页获取：`GET /api/posts?page=2&page_size=5`

### 3.2 搜索帖子

**请求：**
```http
GET /api/posts/search?keyword=Hello&page=1&page_size=10
```

**响应：**
```json
{
  "posts": [
    {
      "id": "1234567890123456789",
      "user_id": "1234567890123456789",
      "title": "Hello World",
      "content": "This is a test post",
      "views": 10,
      "like_count": 5,
      "comment_count": 3,
      "created_at": "2023-01-01T00:00:00Z",
      "updated_at": "2023-01-01T00:00:00Z",
      "is_liked": false,
      "is_favorited": false,
      "user": {
        "id": "1234567890123456789",
        "email": "user@example.com",
        "nickname": "User",
        "avatar": ""
      }
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 10,
  "keyword": "Hello"
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 400 | `{"error": "keyword is required"}` | 关键词不能为空 |
| 500 | `{"error": "Failed to search posts"}` | 搜索失败 |

**测试用例：**
1. 正常搜索：`GET /api/posts/search?keyword=Hello&page=1&page_size=10`
2. 空关键词：`GET /api/posts/search?keyword=`
3. 无结果搜索：`GET /api/posts/search?keyword=不存在的关键词`

### 3.3 获取帖子详情

**请求：**
```http
GET /api/posts/1234567890123456789
```

**响应：**
```json
{
  "post": {
    "id": "1234567890123456789",
    "user_id": "1234567890123456789",
    "title": "Hello World",
    "content": "This is a test post",
    "views": 10,
    "like_count": 5,
    "comment_count": 3,
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z",
    "is_liked": false,
    "is_favorited": false,
    "user": {
      "id": "1234567890123456789",
      "email": "user@example.com",
      "nickname": "User",
      "avatar": ""
    },
    "comments": [
      {
        "id": 1,
        "post_id": "1234567890123456789",
        "user_id": "1234567890123456789",
        "content": "Great post!",
        "like_count": 2,
        "created_at": "2023-01-01T00:00:00Z",
        "user": {
          "id": "1234567890123456789",
          "email": "user@example.com",
          "nickname": "User",
          "avatar": ""
        }
      }
    ]
  }
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 404 | `{"error": "Post not found"}` | 帖子不存在 |
| 500 | `{"error": "Failed to get post"}` | 获取失败 |

**测试用例：**
1. 正常获取：`GET /api/posts/1`
2. 不存在的帖子：`GET /api/posts/999`

### 3.4 创建帖子

**请求：**
```http
POST /api/posts
Content-Type: application/json
Authorization: Bearer <access_token>

{
  "title": "Hello World",
  "content": "This is a test post"
}
```

**响应：**
```json
{
  "post": {
    "id": "1234567890123456789",
    "user_id": "1234567890123456789",
    "title": "Hello World",
    "content": "This is a test post",
    "views": 0,
    "like_count": 0,
    "comment_count": 0,
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z"
  }
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 400 | `{"error": "title is required"}` | 标题不能为空 |
| 400 | `{"error": "content is required"}` | 内容不能为空 |
| 400 | `{"error": "title must be at most 200 characters"}` | 标题过长 |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 500 | `{"error": "Failed to create post"}` | 创建失败 |

**测试用例：**
1. 正常创建：`{"title": "Hello World", "content": "This is a test post"}`
2. 标题过长：`{"title": "a very long title that exceeds the maximum length of 200 characters and should be rejected by the server", "content": "content"}`
3. 空内容：`{"title": "Title", "content": ""}`

### 3.5 更新帖子

**请求：**
```http
PUT /api/posts/1234567890123456789
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
  "post": {
    "id": "1234567890123456789",
    "user_id": "1234567890123456789",
    "title": "Updated Title",
    "content": "Updated content",
    "views": 10,
    "like_count": 5,
    "comment_count": 3,
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T01:00:00Z"
  }
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 400 | `{"error": "title must be at most 200 characters"}` | 标题过长 |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 403 | `{"error": "unauthorized"}` | 无权修改（非帖子作者） |
| 404 | `{"error": "Post not found"}` | 帖子不存在 |
| 500 | `{"error": "Failed to update post"}` | 更新失败 |

**测试用例：**
1. 正常更新：`{"title": "Updated Title", "content": "Updated content"}`
2. 无权更新：使用其他用户的 token 更新帖子
3. 不存在的帖子：`PUT /api/posts/999`

### 3.6 删除帖子

**请求：**
```http
DELETE /api/posts/1234567890123456789
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
| 403 | `{"error": "unauthorized"}` | 无权删除（非帖子作者） |
| 404 | `{"error": "Post not found"}` | 帖子不存在 |
| 500 | `{"error": "Failed to delete post"}` | 删除失败 |

**测试用例：**
1. 正常删除：`DELETE /api/posts/1`
2. 无权删除：使用其他用户的 token 删除帖子
3. 不存在的帖子：`DELETE /api/posts/999`

## 4. 评论接口

### 4.1 获取评论

**请求：**
```http
GET /api/posts/1234567890123456789/comments?page=1&page_size=10
```

**响应：**
```json
{
  "comments": [
    {
      "id": "1",
      "post_id": "1234567890123456789",
      "user_id": "1234567890123456789",
      "content": "Great post!",
      "like_count": 2,
      "created_at": "2023-01-01T00:00:00Z",
      "user": {
        "id": "1234567890123456789",
        "email": "user@example.com",
        "nickname": "User",
        "avatar": ""
      }
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 10
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 404 | `{"error": "Post not found"}` | 帖子不存在 |
| 500 | `{"error": "Failed to get comments"}` | 获取失败 |

**测试用例：**
1. 正常获取：`GET /api/posts/1/comments?page=1&page_size=10`
2. 不存在的帖子：`GET /api/posts/999/comments`

### 4.2 创建评论

**请求：**
```http
POST /api/comments
Content-Type: application/json
Authorization: Bearer <access_token>

{
  "post_id": 1234567890123456789,
  "content": "Great post!"
}
```

**响应：**
```json
{
  "comment": {
    "id": 1,
    "post_id": 1234567890123456789,
    "user_id": 1234567890123456789,
    "content": "Great post!",
    "like_count": 0,
    "created_at": "2023-01-01T00:00:00Z"
  }
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 400 | `{"error": "post_id is required"}` | 帖子 ID 不能为空 |
| 400 | `{"error": "content is required"}` | 内容不能为空 |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 404 | `{"error": "Post not found"}` | 帖子不存在 |
| 500 | `{"error": "Failed to create comment"}` | 创建失败 |

**测试用例：**
1. 正常创建：`{"post_id": 1, "content": "Great post!"}`
2. 不存在的帖子：`{"post_id": 999, "content": "Great post!"}`
3. 空内容：`{"post_id": 1, "content": ""}`

### 4.3 删除评论

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
| 403 | `{"error": "unauthorized"}` | 无权删除（非评论作者） |
| 404 | `{"error": "Comment not found"}` | 评论不存在 |
| 500 | `{"error": "Failed to delete comment"}` | 删除失败 |

**测试用例：**
1. 正常删除：`DELETE /api/comments/1`
2. 无权删除：使用其他用户的 token 删除评论
3. 不存在的评论：`DELETE /api/comments/999`

### 4.4 获取回复（楼中楼）

**请求：**
```http
GET /api/comments/1/replies?page=1&page_size=10
```

**响应：**
```json
{
  "replies": [
    {
      "id": 2,
      "post_id": 1234567890123456789,
      "user_id": 1234567890123456789,
      "parent_id": 1,
      "content": "Thanks!",
      "like_count": 1,
      "created_at": "2023-01-01T00:00:00Z",
      "user": {
        "id": 1234567890123456789,
        "email": "user@example.com",
        "nickname": "User",
        "avatar": ""
      }
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 10
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 404 | `{"error": "Comment not found"}` | 评论不存在 |
| 500 | `{"error": "Failed to get replies"}` | 获取失败 |

**测试用例：**
1. 正常获取：`GET /api/comments/1/replies?page=1&page_size=10`
2. 不存在的评论：`GET /api/comments/999/replies`

## 5. 点赞接口

### 5.1 点赞帖子

**请求：**
```http
POST /api/posts/1234567890123456789/like
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
| 404 | `{"error": "Post not found"}` | 帖子不存在 |
| 409 | `{"error": "already liked"}` | 已经点赞 |
| 500 | `{"error": "Failed to like post"}` | 点赞失败 |

**测试用例：**
1. 正常点赞：`POST /api/posts/1/like`
2. 重复点赞：`POST /api/posts/1/like`（已点赞后再点赞）
3. 不存在的帖子：`POST /api/posts/999/like`

### 5.2 取消点赞帖子

**请求：**
```http
DELETE /api/posts/1234567890123456789/like
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
      "id": 1234567890123456789,
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
  "folder": {
    "id": 1234567890123456789,
    "user_id": 1234567890123456789,
    "name": "技术文章",
    "is_default": false
  }
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 400 | `{"error": "name is required"}` | 名称不能为空 |
| 400 | `{"error": "name must be at most 50 characters"}` | 名称过长 |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 409 | `{"error": "folder already exists"}` | 收藏夹已存在 |
| 500 | `{"error": "Failed to create folder"}` | 创建失败 |

**测试用例：**
1. 正常创建：`{"name": "技术文章"}`
2. 重复创建：`{"name": "技术文章"}`（已存在时再创建）
3. 空名称：`{"name": ""}`

### 6.3 更新收藏夹

**请求：**
```http
PUT /api/folders/1234567890123456789
Content-Type: application/json
Authorization: Bearer <access_token>

{
  "name": "Updated Name"
}
```

**响应：**
```json
{
  "folder": {
    "id": 1234567890123456789,
    "user_id": 1234567890123456789,
    "name": "Updated Name",
    "is_default": false
  }
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 400 | `{"error": "name is required"}` | 名称不能为空 |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 403 | `{"error": "folder not yours"}` | 收藏夹不属于用户 |
| 404 | `{"error": "folder not found"}` | 收藏夹不存在 |
| 409 | `{"error": "folder already exists"}` | 收藏夹名称已存在 |
| 500 | `{"error": "Failed to update folder"}` | 更新失败 |

**测试用例：**
1. 正常更新：`{"name": "Updated Name"}`
2. 无权限更新：使用其他用户的 token
3. 不存在的收藏夹：`PUT /api/folders/999`

### 6.4 删除收藏夹

**请求：**
```http
DELETE /api/folders/1234567890123456789
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
| 403 | `{"error": "folder not yours"}` | 收藏夹不属于用户 |
| 403 | `{"error": "cannot delete default folder"}` | 不能删除默认收藏夹 |
| 404 | `{"error": "folder not found"}` | 收藏夹不存在 |
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
  "post_id": 1234567890123456789,
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
DELETE /api/posts/1234567890123456789/favorite
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
PUT /api/posts/1234567890123456789/favorite
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
        "id": 1234567890123456789,
        "title": "Hello World",
        "content": "This is a test post",
        "views": 10,
        "created_at": "2023-01-01T00:00:00Z",
        "user": {
          "id": 1234567890123456789,
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
        "id": 1234567890123456789,
        "title": "Hello World",
        "content": "This is a test post",
        "views": 10,
        "created_at": "2023-01-01T00:00:00Z",
        "user": {
          "id": 1234567890123456789,
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
POST /api/users/1234567890123456789/block
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
4. 重新拉黑：`POST /api/users/2/block`（已取消拉黑后再拉黑）

### 7.2 取消拉黑

**请求：**
```http
DELETE /api/users/1234567890123456789/block
Authorization: Bearer <access_token>
```

**响应：**
```json
{
  "message": "User unblocked"
}
```

**说明：**
- 取消拉黑后不会删除记录，而是保留记录并更新 `unblocked_at` 字段
- 可以通过 `unblocked_at` 字段判断用户是否已取消拉黑
- 取消拉黑后可以重新拉黑该用户

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 404 | `{"error": "not blocked"}` | 未拉黑 |
| 500 | `{"error": "Failed to unblock user"}` | 取消拉黑失败 |

**测试用例：**
1. 正常取消：`DELETE /api/users/2/block`
2. 未拉黑取消：`DELETE /api/users/2/block`（未拉黑时取消）
3. 重复取消：`DELETE /api/users/2/block`（已取消拉黑后再次取消）

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
      "id": 1234567890123456789,
      "email": "user2@example.com",
      "nickname": "User2",
      "bio": "这是用户简介",
      "avatar": "",
      "is_admin": false,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z",
      "deleted_at": null,
      "blocked_at": "2024-01-01T00:00:00Z",
      "unblocked_at": null
    },
    {
      "id": 1234567890123456790,
      "email": "user3@example.com",
      "nickname": "User3",
      "bio": "",
      "avatar": "",
      "is_admin": false,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z",
      "deleted_at": null,
      "blocked_at": "2024-01-01T00:00:00Z",
      "unblocked_at": "2024-01-02T00:00:00Z"
    }
  ],
  "total": 2
}
```

**字段说明：**
- `blocked_at`: 拉黑时间
- `unblocked_at`: 取消拉黑时间（null 表示未取消拉黑）
- 可以通过 `unblocked_at` 字段判断用户是否已取消拉黑

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
      "id": 1234567890123456789,
      "user_id": 1234567890123456789,
      "title": "Hello World",
      "content": "This is a test post",
      "views": 10,
      "like_count": 5,
      "comment_count": 3,
      "created_at": "2023-01-01T00:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 10
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 500 | `{"error": "Failed to get posts"}` | 获取失败 |

**测试用例：**
1. 正常获取：`GET /api/my/posts?page=1&page_size=10`

## 9. 管理员接口

### 9.1 管理员删除帖子

**请求：**
```http
DELETE /api/admin/posts/1234567890123456789
Authorization: Bearer <admin_access_token>
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
| 403 | `{"error": "Admin access required"}` | 需要管理员权限 |
| 404 | `{"error": "Post not found"}` | 帖子不存在 |
| 500 | `{"error": "Failed to delete post"}` | 删除失败 |

**测试用例：**
1. 管理员删除：`DELETE /api/admin/posts/1`（使用管理员 token）
2. 普通用户尝试：使用普通用户 token 访问
3. 不存在的帖子：`DELETE /api/admin/posts/999`

### 9.2 管理员删除评论

**请求：**
```http
DELETE /api/admin/comments/1
Authorization: Bearer <admin_access_token>
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
| 403 | `{"error": "Admin access required"}` | 需要管理员权限 |
| 404 | `{"error": "Comment not found"}` | 评论不存在 |
| 500 | `{"error": "Failed to delete comment"}` | 删除失败 |

**测试用例：**
1. 管理员删除：`DELETE /api/admin/comments/1`（使用管理员 token）
2. 普通用户尝试：使用普通用户 token 访问
3. 不存在的评论：`DELETE /api/admin/comments/999`

### 9.3 禁言用户

**请求：**
```http
PUT /api/admin/users/1234567890123456789/ban
Authorization: Bearer <admin_access_token>
```

**响应：**
```json
{
  "message": "User banned"
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 403 | `{"error": "Admin access required"}` | 需要管理员权限 |
| 404 | `{"error": "User not found"}` | 用户不存在 |
| 500 | `{"error": "Failed to ban user"}` | 禁言失败 |

**测试用例：**
1. 管理员禁言：`PUT /api/admin/users/2/ban`（使用管理员 token）
2. 普通用户尝试：使用普通用户 token 访问
3. 不存在的用户：`PUT /api/admin/users/999/ban`

### 9.4 解除禁言

**请求：**
```http
PUT /api/admin/users/1234567890123456789/unban
Authorization: Bearer <admin_access_token>
```

**响应：**
```json
{
  "message": "User unbanned"
}
```

**错误返回：**
| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 403 | `{"error": "Admin access required"}` | 需要管理员权限 |
| 404 | `{"error": "User not found"}` | 用户不存在 |
| 500 | `{"error": "Failed to unban user"}` | 解除禁言失败 |

**测试用例：**
1. 管理员解禁：`PUT /api/admin/users/2/unban`（使用管理员 token）
2. 普通用户尝试：使用普通用户 token 访问
3. 不存在的用户：`PUT /api/admin/users/999/unban`

## 10. 中间件说明

### 10.1 AuthRequired 中间件

**功能：** 验证用户登录状态

**使用方式：**
```go
router.Use(middleware.AuthRequired())
```

**行为：**
- 检查请求头中的 `Authorization` 字段
- 验证 JWT 令牌的有效性
- 将用户信息（userID, email, isAdmin）存入上下文
- 令牌无效或过期时返回 401 错误

### 10.2 AdminRequired 中间件

**功能：** 验证管理员权限

**使用方式：**
```go
router.Use(middleware.AuthRequired())
router.Use(middleware.AdminRequired())
```

**行为：**
- 检查上下文中是否存在用户信息（需要先使用 AuthRequired）
- 验证用户是否为管理员（isAdmin = true）
- 非管理员用户返回 403 错误

**注意：** AdminRequired 中间件必须与 AuthRequired 中间件一起使用，且 AuthRequired 必须在前面。

## 11. 错误代码汇总

| 状态码 | 含义 | 常见场景 |
|--------|------|---------|
| 200 | 成功 | 请求成功处理 |
| 201 | 创建成功 | 资源创建成功 |
| 400 | 请求参数错误 | 缺少必填字段、格式错误 |
| 401 | 未授权 | 缺少认证头、令牌无效或过期 |
| 403 | 禁止访问 | 无权限操作、需要管理员权限 |
| 404 | 资源不存在 | 帖子/评论/用户不存在 |
| 409 | 冲突 | 资源已存在（重复操作） |
| 429 | 请求过于频繁 | 验证码重复发送 |
| 500 | 服务器错误 | 内部错误、数据库操作失败 |
