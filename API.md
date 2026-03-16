# API 接口文档

## 重要说明

### ID 字段类型

**所有 ID 字段（如** **`id`、`user_id`、`post_id`、`comment_id`、`folder_id`** **等）在 JSON 请求和响应中均为字符串类型**，以避免 JavaScript 数字精度丢失问题。

例如：

```json
{
  "id": "1234567890123456789",
  "user_id": "1234567890123456789",
  "post_id": "1234567890123456789"
}
```

### 认证方式

需要认证的接口必须在请求头中携带 `Authorization` 字段：

```http
Authorization: Bearer <access_token>
```

### 时间格式

所有时间字段均采用 ISO 8601 格式（UTC）：

```
2023-01-01T00:00:00Z
```

***

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

| 状态码 | 错误信息                                          | 说明                               |
| --- | --------------------------------------------- | -------------------------------- |
| 400 | `{"error": "email is required"}`              | 邮箱不能为空                           |
| 400 | `{"error": "type is required"}`               | 类型不能为空                           |
| 400 | `{"error": "invalid email format"}`           | 邮箱格式错误                           |
| 400 | `{"error": "invalid type"}`                   | 类型错误，只能是 register、reset 或 delete |
| 409 | `{"error": "user already exists"}`            | 注册类型时用户已存在                       |
| 429 | `{"error": "verification code already sent"}` | 验证码已发送，请勿重复请求                    |
| 500 | `{"error": "Failed to push email to queue"}`  | 邮件推送失败                           |

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

| 状态码 | 错误信息                                                  | 说明      |
| --- | ----------------------------------------------------- | ------- |
| 400 | `{"error": "email is required"}`                      | 邮箱不能为空  |
| 400 | `{"error": "password is required"}`                   | 密码不能为空  |
| 400 | `{"error": "code is required"}`                       | 验证码不能为空 |
| 400 | `{"error": "invalid email format"}`                   | 邮箱格式错误  |
| 400 | `{"error": "password must be at least 6 characters"}` | 密码过短    |
| 400 | `{"error": "invalid verification code"}`              | 验证码错误   |
| 400 | `{"error": "verification code expired"}`              | 验证码过期   |
| 500 | `{"error": "Failed to hash password"}`                | 密码哈希失败  |
| 500 | `{"error": "Failed to create user"}`                  | 用户创建失败  |

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

| 状态码 | 错误信息                                     | 说明      |
| --- | ---------------------------------------- | ------- |
| 400 | `{"error": "email is required"}`         | 邮箱不能为空  |
| 400 | `{"error": "password is required"}`      | 密码不能为空  |
| 401 | `{"error": "Invalid email or password"}` | 邮箱或密码错误 |
| 403 | `{"error": "Email not verified"}`        | 邮箱未验证   |
| 403 | `{"error": "account is banned"}`         | 账号被禁言   |
| 500 | `{"error": "Login failed"}`              | 登录失败    |

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

| 状态码 | 错误信息                                     | 说明       |
| --- | ---------------------------------------- | -------- |
| 400 | `{"error": "refresh_token is required"}` | 刷新令牌不能为空 |
| 401 | `{"error": "Invalid token"}`             | 令牌无效     |
| 401 | `{"error": "Token expired"}`             | 令牌过期     |
| 500 | `{"error": "Token refresh failed"}`      | 刷新失败     |

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

| 状态码 | 错误信息                                         | 说明    |
| --- | -------------------------------------------- | ----- |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 500 | `{"error": "Logout failed"}`                 | 登出失败  |

**测试用例：**

1. 正常登出：提供有效的 access\_token 和 refresh\_token
2. 无效令牌：使用无效的 refresh\_token

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

| 状态码 | 错误信息                                         | 说明    |
| --- | -------------------------------------------- | ----- |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 500 | `{"error": "Logout failed"}`                 | 登出失败  |

**测试用例：**

1. 正常登出：提供有效的 access\_token
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

| 状态码 | 错误信息                                          | 说明     |
| --- | --------------------------------------------- | ------ |
| 400 | `{"error": "invalid verification code"}`      | 验证码错误  |
| 400 | `{"error": "verification code already used"}` | 验证码已使用 |
| 400 | `{"error": "verification code expired"}`      | 验证码过期  |
| 404 | `{"error": "User not found"}`                 | 用户不存在  |
| 500 | `{"error": "Failed to reset password"}`       | 重置密码失败 |

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

| 状态码 | 错误信息                                          | 说明     |
| --- | --------------------------------------------- | ------ |
| 400 | `{"error": "invalid verification code"}`      | 验证码错误  |
| 400 | `{"error": "verification code already used"}` | 验证码已使用 |
| 400 | `{"error": "verification code expired"}`      | 验证码过期  |
| 404 | `{"error": "User not found"}`                 | 用户不存在  |
| 500 | `{"error": "Failed to delete account"}`       | 删除账户失败 |

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

| 状态码 | 错误信息                                         | 说明    |
| --- | -------------------------------------------- | ----- |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 404 | `{"error": "User not found"}`                | 用户不存在 |
| 500 | `{"error": "Failed to get profile"}`         | 获取失败  |

**测试用例：**

1. 正常获取：提供有效的 access\_token
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

| 状态码 | 错误信息                                         | 说明       |
| --- | -------------------------------------------- | -------- |
| 400 | `{"error": "Invalid user ID"}`               | 用户ID格式错误 |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头    |
| 404 | `{"error": "User not found"}`                | 用户不存在    |
| 500 | `{"error": "Failed to get user info"}`       | 获取失败     |

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

| 状态码 | 错误信息                                                  | 说明     |
| --- | ----------------------------------------------------- | ------ |
| 400 | `{"error": "nickname is required"}`                   | 昵称不能为空 |
| 400 | `{"error": "nickname must be at most 50 characters"}` | 昵称过长   |
| 401 | `{"error": "Authorization header required"}`          | 缺少认证头  |
| 404 | `{"error": "User not found"}`                         | 用户不存在  |
| 500 | `{"error": "Failed to update nickname"}`              | 更新失败   |

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

````

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
````

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

| 状态码 | 错误信息                                              | 说明    |
| --- | ------------------------------------------------- | ----- |
| 400 | `{"error": "bio must be at most 500 characters"}` | 简介过长  |
| 401 | `{"error": "Authorization header required"}`      | 缺少认证头 |
| 404 | `{"error": "User not found"}`                     | 用户不存在 |
| 500 | `{"error": "Failed to update bio"}`               | 更新失败  |

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

| 状态码 | 错误信息                               | 说明   |
| --- | ---------------------------------- | ---- |
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

| 状态码 | 错误信息                                  | 说明      |
| --- | ------------------------------------- | ------- |
| 400 | `{"error": "keyword is required"}`    | 关键词不能为空 |
| 500 | `{"error": "Failed to search posts"}` | 搜索失败    |

**测试用例：**

1. 正常搜索：`GET /api/posts/search?keyword=Hello&page=1&page_size=10`
2. 空关键词：`GET /api/posts/search?keyword=`
3. 无结果搜索：`GET /api/posts/search?keyword=不存在的关键词`

### 3.3 综合搜索（用户和帖子）

**请求：**

```http
GET /api/search?keyword=Hello&page=1&page_size=10
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
  "users": [
    {
      "id": "1234567890123456789",
      "email": "user@example.com",
      "nickname": "HelloUser",
      "bio": "Hello everyone!",
      "avatar": "",
      "status": 1,
      "is_admin": false,
      "is_verified": true,
      "created_at": "2023-01-01T00:00:00Z",
      "last_login_at": "2023-01-01T00:00:00Z"
    }
  ],
  "total": {
    "posts": 1,
    "users": 1
  },
  "page": 1,
  "page_size": 10,
  "keyword": "Hello"
}
```

**参数说明：**

- `keyword`：搜索关键词（必填）
- `page`：页码，默认 1
- `page_size`：每页数量，默认 10

**搜索范围：**

- **帖子**：标题或内容包含关键词
- **用户**：昵称或邮箱包含关键词

**错误返回：**

| 状态码 | 错误信息                               | 说明      |
| --- | ---------------------------------- | ------- |
| 400 | `{"error": "keyword is required"}` | 关键词不能为空 |
| 500 | `{"error": "Failed to search"}`    | 搜索失败    |

**测试用例：**

1. 正常搜索：`GET /api/search?keyword=Hello&page=1&page_size=10`
2. 空关键词：`GET /api/search?keyword=`
3. 只搜索用户：`GET /api/search?keyword=user@example.com`
4. 分页查询：`GET /api/search?keyword=Hello&page=2&page_size=5`

### 3.4 获取帖子详情

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

| 状态码 | 错误信息                              | 说明    |
| --- | --------------------------------- | ----- |
| 404 | `{"error": "Post not found"}`     | 帖子不存在 |
| 500 | `{"error": "Failed to get post"}` | 获取失败  |

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

| 状态码 | 错误信息                                                | 说明     |
| --- | --------------------------------------------------- | ------ |
| 400 | `{"error": "title is required"}`                    | 标题不能为空 |
| 400 | `{"error": "content is required"}`                  | 内容不能为空 |
| 400 | `{"error": "title must be at most 200 characters"}` | 标题过长   |
| 401 | `{"error": "Authorization header required"}`        | 缺少认证头  |
| 500 | `{"error": "Failed to create post"}`                | 创建失败   |

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

| 状态码 | 错误信息                                                | 说明          |
| --- | --------------------------------------------------- | ----------- |
| 400 | `{"error": "title must be at most 200 characters"}` | 标题过长        |
| 401 | `{"error": "Authorization header required"}`        | 缺少认证头       |
| 403 | `{"error": "unauthorized"}`                         | 无权修改（非帖子作者） |
| 404 | `{"error": "Post not found"}`                       | 帖子不存在       |
| 500 | `{"error": "Failed to update post"}`                | 更新失败        |

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

| 状态码 | 错误信息                                         | 说明          |
| --- | -------------------------------------------- | ----------- |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头       |
| 403 | `{"error": "unauthorized"}`                  | 无权删除（非帖子作者） |
| 404 | `{"error": "Post not found"}`                | 帖子不存在       |
| 500 | `{"error": "Failed to delete post"}`         | 删除失败        |

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

| 状态码 | 错误信息                                  | 说明    |
| --- | ------------------------------------- | ----- |
| 404 | `{"error": "Post not found"}`         | 帖子不存在 |
| 500 | `{"error": "Failed to get comments"}` | 获取失败  |

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
  "post_id": "1234567890123456789",
  "content": "Great post!"
}
```

**响应：**

```json
{
  "comment": {
    "id": "1",
    "post_id": "1234567890123456789",
    "user_id": "1234567890123456789",
    "content": "Great post!",
    "like_count": 0,
    "created_at": "2023-01-01T00:00:00Z"
  }
}
```

**错误返回：**

| 状态码 | 错误信息                                         | 说明         |
| --- | -------------------------------------------- | ---------- |
| 400 | `{"error": "post_id is required"}`           | 帖子 ID 不能为空 |
| 400 | `{"error": "content is required"}`           | 内容不能为空     |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头      |
| 404 | `{"error": "Post not found"}`                | 帖子不存在      |
| 500 | `{"error": "Failed to create comment"}`      | 创建失败       |

**测试用例：**

1. 正常创建：`{"post_id": "1234567890123456789", "content": "Great post!"}`
2. 不存在的帖子：`{"post_id": "999", "content": "Great post!"}`
3. 空内容：`{"post_id": "1", "content": ""}`

### 4.3 创建评论回复（楼中楼）

**请求：**

```http
POST /api/comments
Content-Type: application/json
Authorization: Bearer <access_token>

{
  "comment_id": "1",
  "content": "Thanks for your reply!"
}
```

**参数说明：**

- `comment_id`：父评论 ID（必填），表示回复哪条评论
- `content`：回复内容（必填）
- 注意：创建回复时不需要提供 `post_id`，系统会自动从父评论获取

**响应：**

```json
{
  "comment": {
    "id": "2",
    "post_id": "1234567890123456789",
    "comment_id": "1",
    "user_id": "1234567890123456789",
    "content": "Thanks for your reply!",
    "like_count": 0,
    "created_at": "2023-01-01T00:00:00Z"
  }
}
```

**错误返回：**

| 状态码 | 错误信息                                         | 说明         |
| --- | -------------------------------------------- | ---------- |
| 400 | `{"error": "comment_id is required"}`        | 评论 ID 不能为空 |
| 400 | `{"error": "content is required"}`           | 内容不能为空     |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头      |
| 404 | `{"error": "Parent comment not found"}`      | 父评论不存在     |
| 500 | `{"error": "Failed to create comment"}`      | 创建失败       |

**测试用例：**

1. 正常回复：`{"comment_id": "1", "content": "Thanks!"}`
2. 不存在的评论：`{"comment_id": "999", "content": "Thanks!"}`
3. 空内容：`{"comment_id": "1", "content": ""}`

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

| 状态码 | 错误信息                                         | 说明          |
| --- | -------------------------------------------- | ----------- |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头       |
| 403 | `{"error": "unauthorized"}`                  | 无权删除（非评论作者） |
| 404 | `{"error": "Comment not found"}`             | 评论不存在       |
| 500 | `{"error": "Failed to delete comment"}`      | 删除失败        |

**测试用例：**

1. 正常删除：`DELETE /api/comments/1`
2. 无权删除：使用其他用户的 token 删除评论
3. 不存在的评论：`DELETE /api/comments/999`

### 4.5 搜索评论

**请求：**

```http
GET /api/comments?keyword=Hello&page=1&page_size=10
```

**响应：**

```json
{
  "comments": [
    {
      "id": "1",
      "post_id": "1234567890123456789",
      "user_id": "1234567890123456789",
      "content": "Hello, great post!",
      "is_deleted": false,
      "created_at": "2023-01-01T00:00:00Z",
      "updated_at": "2023-01-01T00:00:00Z",
      "user": {
        "id": "1234567890123456789",
        "email": "user@example.com",
        "nickname": "User",
        "avatar": ""
      },
      "post": {
        "id": "1234567890123456789",
        "title": "Hello World",
        "content": "This is a test post"
      }
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 10,
  "keyword": "Hello"
}
```

**参数说明：**

- `keyword`：搜索关键词（必填），匹配评论内容
- `page`：页码，默认 1
- `page_size`：每页数量，默认 10

**错误返回：**

| 状态码 | 错误信息                                     | 说明      |
| --- | ---------------------------------------- | ------- |
| 400 | `{"error": "keyword is required"}`       | 关键词不能为空 |
| 500 | `{"error": "Failed to search comments"}` | 搜索失败    |

**测试用例：**

1. 正常搜索：`GET /api/comments?keyword=Hello&page=1&page_size=10`
2. 空关键词：`GET /api/comments?keyword=`
3. 无结果搜索：`GET /api/comments?keyword=不存在的关键词`

### 4.6 获取回复（楼中楼）

**请求：**

```http
GET /api/comments/1/replies?page=1&page_size=10
```

**响应：**

```json
{
  "replies": [
    {
      "id": "2",
      "post_id": "1234567890123456789",
      "user_id": "1234567890123456789",
      "comment_id": "1",
      "content": "Thanks!",
      "like_count": 1,
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

| 状态码 | 错误信息                                 | 说明    |
| --- | ------------------------------------ | ----- |
| 404 | `{"error": "Comment not found"}`     | 评论不存在 |
| 500 | `{"error": "Failed to get replies"}` | 获取失败  |

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

| 状态码 | 错误信息                                         | 说明    |
| --- | -------------------------------------------- | ----- |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 404 | `{"error": "Post not found"}`                | 帖子不存在 |
| 409 | `{"error": "already liked"}`                 | 已经点赞  |
| 500 | `{"error": "Failed to like post"}`           | 点赞失败  |

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

| 状态码 | 错误信息                                         | 说明     |
| --- | -------------------------------------------- | ------ |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头  |
| 404 | `{"error": "not liked"}`                     | 未点赞    |
| 500 | `{"error": "Failed to unlike post"}`         | 取消点赞失败 |

**测试用例：**

1. 正常取消：`DELETE /api/posts/1/like`
2. 未点赞取消：`DELETE /api/posts/1/like`（未点赞时取消）

### 5.3 查询帖子点赞状态和数量

**请求：**

```http
GET /api/posts/1234567890123456789/like
Authorization: Bearer <access_token>
```

**响应：**

```json
{
  "post_id": "1234567890123456789",
  "is_liked": true,
  "like_count": 10
}
```

**响应字段说明：**

| 字段           | 类型      | 说明           |
| ------------ | ------- | ------------ |
| `post_id`    | string  | 帖子 ID        |
| `is_liked`   | boolean | 当前用户是否点赞了该帖子 |
| `like_count` | number  | 该帖子的总点赞数量    |

**错误返回：**

| 状态码 | 错误信息                                         | 说明       |
| --- | -------------------------------------------- | -------- |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头    |
| 400 | `{"error": "Invalid post ID"}`               | 无效的帖子ID  |
| 500 | `{"error": "Failed to get like status"}`     | 查询点赞状态失败 |
| 500 | `{"error": "Failed to get like count"}`      | 查询点赞数量失败 |

**测试用例：**

1. 已点赞查询：`GET /api/posts/1234567890123456789/like`（已点赞的帖子）
2. 未点赞查询：`GET /api/posts/1234567890123456789/like`（未点赞的帖子）
3. 不存在的帖子：`GET /api/posts/999/like`

### 5.4 点赞评论

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

| 状态码 | 错误信息                                         | 说明    |
| --- | -------------------------------------------- | ----- |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 409 | `{"error": "already liked"}`                 | 已经点赞  |
| 500 | `{"error": "Failed to like comment"}`        | 点赞失败  |

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

| 状态码 | 错误信息                                         | 说明     |
| --- | -------------------------------------------- | ------ |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头  |
| 404 | `{"error": "not liked"}`                     | 未点赞    |
| 500 | `{"error": "Failed to unlike comment"}`      | 取消点赞失败 |

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

| 状态码 | 错误信息                                         | 说明    |
| --- | -------------------------------------------- | ----- |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 500 | `{"error": "Failed to get folders"}`         | 获取失败  |

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

| 状态码 | 错误信息                                              | 说明     |
| --- | ------------------------------------------------- | ------ |
| 400 | `{"error": "name is required"}`                   | 名称不能为空 |
| 400 | `{"error": "name must be at most 50 characters"}` | 名称过长   |
| 401 | `{"error": "Authorization header required"}`      | 缺少认证头  |
| 409 | `{"error": "folder already exists"}`              | 收藏夹已存在 |
| 500 | `{"error": "Failed to create folder"}`            | 创建失败   |

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

| 状态码 | 错误信息                                         | 说明       |
| --- | -------------------------------------------- | -------- |
| 400 | `{"error": "name is required"}`              | 名称不能为空   |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头    |
| 403 | `{"error": "folder not yours"}`              | 收藏夹不属于用户 |
| 404 | `{"error": "folder not found"}`              | 收藏夹不存在   |
| 409 | `{"error": "folder already exists"}`         | 收藏夹名称已存在 |
| 500 | `{"error": "Failed to update folder"}`       | 更新失败     |

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

| 状态码 | 错误信息                                         | 说明        |
| --- | -------------------------------------------- | --------- |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头     |
| 403 | `{"error": "folder not yours"}`              | 收藏夹不属于用户  |
| 403 | `{"error": "cannot delete default folder"}`  | 不能删除默认收藏夹 |
| 404 | `{"error": "folder not found"}`              | 收藏夹不存在    |
| 500 | `{"error": "Failed to delete folder"}`       | 删除失败      |

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
  "post_id": "1234567890123456789",
  "folder_id": "1"
}
```

**响应：**

```json
{
  "message": "Post favorited"
}
```

**错误返回：**

| 状态码 | 错误信息                                         | 说明          |
| --- | -------------------------------------------- | ----------- |
| 400 | `{"error": "post_id is required"}`           | 帖子 ID 不能为空  |
| 400 | `{"error": "folder_id is required"}`         | 收藏夹 ID 不能为空 |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头       |
| 404 | `{"error": "post not found"}`                | 帖子不存在       |
| 404 | `{"error": "folder not found"}`              | 收藏夹不存在      |
| 403 | `{"error": "folder not yours"}`              | 收藏夹不属于用户    |
| 409 | `{"error": "already favorited"}`             | 已经收藏        |
| 500 | `{"error": "Failed to favorite post"}`       | 收藏失败        |

**测试用例：**

1. 正常收藏：`{"post_id": "1", "folder_id": "1"}`
2. 重复收藏：`{"post_id": "1", "folder_id": "1"}`（已收藏后再收藏）
3. 不存在的帖子：`{"post_id": "999", "folder_id": "1"}`

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

| 状态码 | 错误信息                                         | 说明     |
| --- | -------------------------------------------- | ------ |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头  |
| 404 | `{"error": "not favorited"}`                 | 未收藏    |
| 500 | `{"error": "Failed to unfavorite post"}`     | 取消收藏失败 |

**测试用例：**

1. 正常取消：`DELETE /api/posts/1/favorite`
2. 未收藏取消：`DELETE /api/posts/1/favorite`（未收藏时取消）

### 6.7 查询帖子收藏状态和收藏夹信息

**请求：**

```http
GET /api/posts/1234567890123456789/favorite
Authorization: Bearer <access_token>
```

**响应（已收藏）：**

```json
{
  "post_id": "1234567890123456789",
  "is_favorited": true,
  "folders": [
    {
      "id": "1",
      "user_id": "1234567890123456789",
      "name": "我的收藏",
      "is_default": false,
      "created_at": "2023-01-01T00:00:00Z"
    },
    {
      "id": "2",
      "user_id": "1234567890123456789",
      "name": "技术文章",
      "is_default": false,
      "created_at": "2023-01-02T00:00:00Z"
    }
  ]
}
```

**响应（未收藏）：**

```json
{
  "post_id": "1234567890123456789",
  "is_favorited": false,
  "folders": []
}
```

**响应字段说明：**

| 字段                     | 类型      | 说明                      |
| ---------------------- | ------- | ----------------------- |
| `post_id`              | string  | 帖子 ID                   |
| `is_favorited`         | boolean | 当前用户是否收藏了该帖子            |
| `folders`              | array   | 该帖子被收藏到的收藏夹列表（未收藏时为空数组） |
| `folders[].id`         | string  | 收藏夹 ID                  |
| `folders[].name`       | string  | 收藏夹名称                   |
| `folders[].is_default` | boolean | 是否为默认收藏夹                |

**错误返回：**

| 状态码 | 错误信息                                         | 说明      |
| --- | -------------------------------------------- | ------- |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头   |
| 400 | `{"error": "Invalid post ID"}`               | 无效的帖子ID |
| 500 | `{"error": "Failed to get favorite status"}` | 查询失败    |

**测试用例：**

1. 已收藏查询（单收藏夹）：`GET /api/posts/1234567890123456789/favorite`
2. 已收藏查询（多收藏夹）：帖子被收藏到多个收藏夹
3. 未收藏查询：`GET /api/posts/1234567890123456789/favorite`（未收藏的帖子）
4. 不存在的帖子：`GET /api/posts/999/favorite`

### 6.8 移动收藏

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

| 状态码 | 错误信息                                         | 说明          |
| --- | -------------------------------------------- | ----------- |
| 400 | `{"error": "folder_id is required"}`         | 收藏夹 ID 不能为空 |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头       |
| 404 | `{"error": "not favorited"}`                 | 未收藏         |
| 404 | `{"error": "folder not found"}`              | 收藏夹不存在      |
| 403 | `{"error": "folder not yours"}`              | 收藏夹不属于用户    |
| 500 | `{"error": "Failed to move favorite"}`       | 移动失败        |

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

| 状态码 | 错误信息                                         | 说明    |
| --- | -------------------------------------------- | ----- |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 500 | `{"error": "Failed to get favorites"}`       | 获取失败  |

**测试用例：**

1. 正常获取：`GET /api/my/favorites?page=1&page_size=10`

## 7. AI 接口

### 7.1 AI 问答

**请求：**

```http
POST /api/ai/ask
Content-Type: application/json

{
  "question": "如何使用Elasticsearch？"
}
```

**响应：**

```json
{
  "answer": "Elasticsearch是一个开源的搜索引擎，主要用于全文搜索、结构化搜索和分析。在EyuForum中，我们使用Elasticsearch来存储和搜索帖子和评论..."
}
```

**错误返回：**

| 状态码 | 错误信息                                            | 说明       |
| --- | ----------------------------------------------- | -------- |
| 400 | `{"error": "Invalid request"}`                  | 请求参数错误   |
| 500 | `{"error": "Failed to get relevant documents"}` | 获取相关文档失败 |

### 7.2 AI 问答（流式传输）

**请求：**

```http
POST /api/ai/ask/stream
Content-Type: application/json

{
  "question": "如何使用Elasticsearch？"
}
```

**响应：**

- 响应类型：`text/event-stream`
- 响应格式：SSE (Server-Sent Events)

```
event: end
data: 

event: data
data: 基于论坛内容的回答：

event: data
data: 

event: data
data: 相关内容：

event: data
data: 

event: data
data: 1. 标题: Elasticsearch使用指南
内容: Elasticsearch是一个强大的搜索引擎，可用于论坛的内容搜索和分析。

event: data
data: 

event: data
data: 问题：如何使用Elasticsearch？

event: data
data: 

event: data
data: 这是一个基于论坛内容的智能回答。

event: end
data: 
```

**实现说明：**

- 后端接收DeepSeek API的完整JSON响应
- 将AI回答内容分割成小块（每块约100字符）流式发送
- 每个chunk之间有50ms延迟，模拟真实的流式体验
- 支持实时显示AI回答内容

**错误返回：**

| 状态码 | 错误信息                                            | 说明       |
| --- | ----------------------------------------------- | -------- |
| 400 | `{"error": "Invalid request"}`                  | 请求参数错误   |
| 500 | `{"error": "Failed to get relevant documents"}` | 获取相关文档失败 |

### 7.3 AI 接口实现说明

- **技术栈**：使用DeepSeek API实现RAG (Retrieval-Augmented Generation) 模式
- **数据流**：
  1. 接收用户问题
  2. 从 Elasticsearch 检索相关文档
  3. 调用 DeepSeek API 生成回答
  4. 支持标准 JSON 响应和流式 SSE 响应
- **流式传输实现**：
  - 接收DeepSeek API的完整JSON响应
  - 将AI回答内容分割成小块（每块约100字符）流式发送
  - 每个chunk之间有50ms延迟，模拟真实的流式体验
  - 支持实时显示AI回答内容
- **特点**：
  - 基于论坛真实内容生成回答
  - 支持实时流式输出
  - 集成DeepSeek AI的能力

## 8. 天气接口

### 8.1 获取当前天气信息

**请求：**

```http
GET /api/weather
```

**响应：**

```json
{
  "ip": "123.45.67.89",
  "city": "Beijing",
  "country": "China",
  "temperature": 22.5,
  "feels_like": 21.8,
  "humidity": 45,
  "weather": "晴朗",
  "wind_speed": 12.3,
  "updated_at": "2026-03-14 15:30:00"
}
```

**字段说明：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `ip` | string | 客户端IP地址 |
| `city` | string | 城市名称 |
| `country` | string | 国家名称 |
| `temperature` | float64 | 当前温度（摄氏度） |
| `feels_like` | float64 | 体感温度（摄氏度） |
| `humidity` | int | 湿度百分比 |
| `weather` | string | 天气状况描述 |
| `wind_speed` | float64 | 风速（km/h） |
| `updated_at` | string | 数据更新时间 |

**错误返回：**

| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 500 | `{"error": "Failed to get weather information"}` | 获取天气信息失败 |

**测试用例：**

1. 正常获取：`GET /api/weather`
2. 本地环境测试：自动获取公网IP并查询天气

### 8.2 根据IP获取天气信息

**请求：**

```http
GET /api/weather/by-ip?ip=123.45.67.89
```

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `ip` | string | 是 | IP地址 |

**响应：**

```json
{
  "ip": "123.45.67.89",
  "city": "Shanghai",
  "country": "China",
  "temperature": 25.0,
  "feels_like": 26.2,
  "humidity": 60,
  "weather": "多云",
  "wind_speed": 8.5,
  "updated_at": "2026-03-14 15:30:00"
}
```

**错误返回：**

| 状态码 | 错误信息 | 说明 |
|--------|---------|------|
| 400 | `{"error": "IP address is required"}` | IP参数缺失 |
| 500 | `{"error": "Failed to get weather information"}` | 获取天气信息失败 |

**测试用例：**

1. 正常获取：`GET /api/weather/by-ip?ip=8.8.8.8`
2. 参数缺失：`GET /api/weather/by-ip`
3. 无效IP：`GET /api/weather/by-ip?ip=invalid`

### 8.3 天气接口实现说明

- **技术栈**：
  - 使用 ipapi.co 获取IP地理位置信息
  - 使用 Open-Meteo API 获取天气数据（免费、无需API Key）
- **数据流**：
  1. 获取客户端IP地址
  2. 通过IP获取地理位置（经纬度、城市、国家）
  3. 调用 Open-Meteo API 获取天气数据
  4. 解析并返回格式化的天气信息
- **特点**：
  - 自动识别用户IP位置
  - 支持指定IP查询
  - 提供温度、湿度、风速等详细天气信息
  - 包含体感温度
  - 天气状况中文描述
  - 数据实时更新

## 9. 拉黑接口

### 8.1 拉黑用户

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

| 状态码 | 错误信息                                         | 说明     |
| --- | -------------------------------------------- | ------ |
| 400 | `{"error": "cannot block yourself"}`         | 不能拉黑自己 |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头  |
| 409 | `{"error": "already blocked"}`               | 已经拉黑   |
| 500 | `{"error": "Failed to block user"}`          | 拉黑失败   |

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
- `unblocked_at` 为 null 表示当前处于拉黑状态
- `unblocked_at` 不为 null 表示已取消拉黑
- 取消拉黑后可以重新拉黑该用户

**错误返回：**

| 状态码 | 错误信息                                         | 说明     |
| --- | -------------------------------------------- | ------ |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头  |
| 404 | `{"error": "not blocked"}`                   | 未拉黑    |
| 500 | `{"error": "Failed to unblock user"}`        | 取消拉黑失败 |

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
- `unblocked_at`: 取消拉黑时间（null 表示当前处于拉黑状态，不为 null 表示已取消拉黑）
- 黑名单列表只返回 `unblocked_at` 为 null 的用户（当前处于拉黑状态）

**错误返回：**

| 状态码 | 错误信息                                         | 说明    |
| --- | -------------------------------------------- | ----- |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 500 | `{"error": "Failed to get blocked users"}`   | 获取失败  |

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

| 状态码 | 错误信息                                         | 说明    |
| --- | -------------------------------------------- | ----- |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 500 | `{"error": "Failed to get posts"}`           | 获取失败  |

**测试用例：**

1. 正常获取：`GET /api/my/posts?page=1&page_size=10`

## 9. 收信箱

### 9.1 获取收信箱消息

当有人回复你的帖子或评论时，会收到消息通知。

**消息类型：**

- `reply_post` - 有人回复了你的帖子
- `reply_comment` - 有人回复了你的评论

**请求：**

```http
GET /api/inbox?page=1&page_size=10
Authorization: Bearer <access_token>
```

**响应：**

```json
{
  "messages": [
    {
      "post_id": "1234567890123456789",
      "comment_id": "456",
      "sender_id": "789",
      "type": "reply_comment",
      "time": 1234567890
    },
    {
      "post_id": "1234567890123456789",
      "sender_id": "789",
      "type": "reply_post",
      "time": 1234567891
    }
  ],
  "total": 2,
  "page": 1,
  "page_size": 10
}
```

**字段说明：**

| 字段           | 类型     | 说明                                  |
| ------------ | ------ | ----------------------------------- |
| `post_id`    | string | 帖子ID                                |
| `comment_id` | string | 评论ID（回复帖子时为空）                       |
| `sender_id`  | string | 回复者用户ID                             |
| `type`       | string | 消息类型：`reply_post` 或 `reply_comment` |
| `time`       | number | 消息时间戳（Unix时间）                       |

**错误返回：**

| 状态码 | 错误信息                                         | 说明    |
| --- | -------------------------------------------- | ----- |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 500 | `{"error": "Failed to get inbox"}`           | 获取失败  |

**测试用例：**

1. 正常获取：`GET /api/inbox?page=1&page_size=10`
2. 空收信箱：新用户获取收信箱

### 9.2 清空收信箱

**请求：**

```http
DELETE /api/inbox
Authorization: Bearer <access_token>
```

**响应：**

```json
{
  "message": "Inbox cleared"
}
```

**错误返回：**

| 状态码 | 错误信息                                         | 说明    |
| --- | -------------------------------------------- | ----- |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 500 | `{"error": "Failed to clear inbox"}`         | 清空失败  |

**测试用例：**

1. 正常清空：`DELETE /api/inbox`

## 10. 管理员接口

### 10.1 管理员删除帖子

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

| 状态码 | 错误信息                                         | 说明    |
| --- | -------------------------------------------- | ----- |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头 |
| 500 | `{"error": "Failed to get posts"}`           | 获取失败  |

**测试用例：**

1. 正常获取：`GET /api/my/posts?page=1&page_size=10`

## 10. 管理员接口

### 10.1 管理员删除帖子

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

| 状态码 | 错误信息                                         | 说明      |
| --- | -------------------------------------------- | ------- |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头   |
| 403 | `{"error": "Admin access required"}`         | 需要管理员权限 |
| 404 | `{"error": "Post not found"}`                | 帖子不存在   |
| 500 | `{"error": "Failed to delete post"}`         | 删除失败    |

**测试用例：**

1. 管理员删除：`DELETE /api/admin/posts/1`（使用管理员 token）
2. 普通用户尝试：使用普通用户 token 访问
3. 不存在的帖子：`DELETE /api/admin/posts/999`

### 10.2 管理员删除评论

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

| 状态码 | 错误信息                                         | 说明      |
| --- | -------------------------------------------- | ------- |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头   |
| 403 | `{"error": "Admin access required"}`         | 需要管理员权限 |
| 404 | `{"error": "Comment not found"}`             | 评论不存在   |
| 500 | `{"error": "Failed to delete comment"}`      | 删除失败    |

**测试用例：**

1. 管理员删除：`DELETE /api/admin/comments/1`（使用管理员 token）
2. 普通用户尝试：使用普通用户 token 访问
3. 不存在的评论：`DELETE /api/admin/comments/999`

### 10.3 管理员查看所有评论

**请求：**

```http
GET /api/admin/comments?page=1&page_size=10
Authorization: Bearer <admin_access_token>
```

**响应：**

```json
{
  "comments": [
    {
      "id": "1234567890123456789",
      "post_id": "1234567890123456789",
      "user_id": "1234567890123456789",
      "content": "Great post!",
      "is_deleted": false,
      "created_at": "2023-01-01T00:00:00Z",
      "updated_at": "2023-01-01T00:00:00Z",
      "user": {
        "id": "1234567890123456789",
        "email": "user@example.com",
        "nickname": "User",
        "avatar": ""
      },
      "post": {
        "id": "1234567890123456789",
        "title": "Hello World",
        "content": "This is a test post"
      }
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 10
}
```

**参数说明：**

- `page`：页码，默认 1
- `page_size`：每页数量，默认 10

**错误返回：**

| 状态码 | 错误信息                                         | 说明      |
| --- | -------------------------------------------- | ------- |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头   |
| 403 | `{"error": "Admin access required"}`         | 需要管理员权限 |
| 500 | `{"error": "Failed to get comments"}`        | 获取失败    |

**测试用例：**

1. 管理员查看：`GET /api/admin/comments?page=1&page_size=10`（使用管理员 token）
2. 普通用户尝试：使用普通用户 token 访问
3. 分页查询：`GET /api/admin/comments?page=2&page_size=5`

### 10.4 禁言用户

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

| 状态码 | 错误信息                                         | 说明      |
| --- | -------------------------------------------- | ------- |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头   |
| 403 | `{"error": "Admin access required"}`         | 需要管理员权限 |
| 404 | `{"error": "User not found"}`                | 用户不存在   |
| 500 | `{"error": "Failed to ban user"}`            | 禁言失败    |

**测试用例：**

1. 管理员禁言：`PUT /api/admin/users/2/ban`（使用管理员 token）
2. 普通用户尝试：使用普通用户 token 访问
3. 不存在的用户：`PUT /api/admin/users/999/ban`

### 10.5 解除禁言

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

| 状态码 | 错误信息                                         | 说明      |
| --- | -------------------------------------------- | ------- |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头   |
| 403 | `{"error": "Admin access required"}`         | 需要管理员权限 |
| 404 | `{"error": "User not found"}`                | 用户不存在   |
| 500 | `{"error": "Failed to unban user"}`          | 解除禁言失败  |

**测试用例：**

1. 管理员解禁：`PUT /api/admin/users/2/unban`（使用管理员 token）
2. 普通用户尝试：使用普通用户 token 访问
3. 不存在的用户：`PUT /api/admin/users/999/unban`

### 10.6 查看所有用户

**请求：**

```http
GET /api/admin/users?page=1&page_size=10
Authorization: Bearer <admin_access_token>
```

**响应：**

```json
{
  "users": [
    {
      "id": "1234567890123456789",
      "email": "user@example.com",
      "nickname": "User",
      "bio": "This is my bio",
      "avatar": "/uploads/avatar_1234567890123456789_1234567890.jpg",
      "status": 1,
      "is_admin": false,
      "is_verified": true,
      "created_at": "2023-01-01T00:00:00Z",
      "last_login_at": "2023-01-01T00:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 10
}
```

**参数说明：**

- `page`：页码，默认 1
- `page_size`：每页数量，默认 10

**错误返回：**

| 状态码 | 错误信息                                         | 说明      |
| --- | -------------------------------------------- | ------- |
| 401 | `{"error": "Authorization header required"}` | 缺少认证头   |
| 403 | `{"error": "Admin access required"}`         | 需要管理员权限 |
| 500 | `{"error": "Failed to get users"}`           | 获取失败    |

**测试用例：**

1. 管理员查看：`GET /api/admin/users?page=1&page_size=10`（使用管理员 token）
2. 普通用户尝试：使用普通用户 token 访问
3. 分页查询：`GET /api/admin/users?page=2&page_size=5`

## 11. 中间件说明

### 11.1 AuthRequired 中间件

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

### 11.2 AdminRequired 中间件

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

## 12. 错误代码汇总

| 状态码 | 含义     | 常见场景          |
| --- | ------ | ------------- |
| 200 | 成功     | 请求成功处理        |
| 201 | 创建成功   | 资源创建成功        |
| 400 | 请求参数错误 | 缺少必填字段、格式错误   |
| 401 | 未授权    | 缺少认证头、令牌无效或过期 |
| 403 | 禁止访问   | 无权限操作、需要管理员权限 |
| 404 | 资源不存在  | 帖子/评论/用户不存在   |
| 409 | 冲突     | 资源已存在（重复操作）   |
| 429 | 请求过于频繁 | 验证码重复发送       |
| 500 | 服务器错误  | 内部错误、数据库操作失败  |

## 13. GitHub Action 自动部署

### 13.1 部署配置

本项目使用 GitHub Actions 实现自动部署到云服务器。部署流程如下：

1. **代码推送**：当代码推送到 `main` 分支时触发部署
2. **构建应用**：在 Ubuntu 环境中构建 Go 应用
3. **部署到服务器**：通过 SSH 连接到云服务器并部署
4. **启动服务**：停止旧服务，启动新服务

### 13.2 配置步骤

1. **在 GitHub 仓库中设置以下 Secrets**：
   - `SERVER_HOST`：云服务器 IP 地址
   - `SERVER_USER`：服务器登录用户名
   - `SERVER_PASSWORD`：服务器登录密码
   - `SERVER_PORT`：SSH 端口（默认 22）

2. **服务器准备**：
   - 确保服务器已安装 Go 1.20 或更高版本
   - 确保服务器已开放 SSH 端口
   - 创建部署目录：`/opt/bbsDemo`
   - 确保服务器有足够的权限

3. **配置文件**：
   - 在服务器上创建配置文件 `/opt/bbsDemo/config.yaml`
   - 根据实际环境配置数据库、Redis、邮件等信息

### 13.3 部署流程

1. **代码推送**：`git push origin main`
2. **触发构建**：GitHub Actions 自动开始构建过程
3. **构建应用**：编译 Go 代码生成可执行文件
4. **部署到服务器**：通过 SSH 上传文件到服务器
5. **启动服务**：停止旧服务并启动新服务
6. **验证部署**：检查服务是否正常运行

### 13.4 部署日志

可以在 GitHub 仓库的 **Actions** 标签页查看部署日志，了解部署过程的详细信息。

### 13.5 注意事项

- 部署过程中服务会短暂中断（约 5-10 秒）
- 确保服务器有足够的磁盘空间
- 确保服务器网络连接稳定
- 定期备份配置文件和数据库
- 部署前建议在测试环境验证代码

### 13.6 手动部署（备选方案）

如果自动部署失败，可以使用以下步骤手动部署：

1. **构建应用**：
   ```bash
   go build -o bbsDemo
   ```

2. **上传文件**：
   ```bash
   scp bbsDemo user@server:/opt/bbsDemo/
   ```

3. **启动服务**：
   ```bash
   ssh user@server "cd /opt/bbsDemo && pkill -f 'bbsDemo' && nohup ./bbsDemo > app.log 2>&1 &"
   ```

