# 认证与授权

## 概述

NexusGate 采用 JWT + RBAC 认证模型。用户登录后获取 JWT Token，后续请求通过 `Authorization: Bearer <token>` 携带身份信息。系统支持三种角色：admin、operator、viewer。

## 源码文件

| 文件 | 说明 |
|------|------|
| `server/internal/handler/auth.go` | 登录、用户 CRUD、审计日志 |
| `server/internal/handler/middleware/auth.go` | JWT 解析 & 角色鉴权中间件 |
| `server/internal/model/user.go` | User、AuditLog 模型 |
| `server/internal/store/seed.go` | 自动创建默认管理员 |
| `web/src/views/Login.vue` | 登录页面 |
| `web/src/views/Users.vue` | 用户管理页面 |
| `web/src/views/Audit.vue` | 审计日志页面 |

## 数据模型

### User

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uint | PK | 主键 |
| username | string | unique, not null | 用户名 |
| password | string | not null | bcrypt 哈希 |
| role | string | default: viewer | admin / operator / viewer |
| email | string | - | 邮箱 |
| created_at | time | auto | 创建时间 |
| updated_at | time | auto | 更新时间 |
| deleted_at | time | soft delete | 软删除 |

### AuditLog

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uint | PK | 主键 |
| user_id | uint | index | 操作用户 ID |
| username | string | - | 操作用户名 |
| action | string | index | 操作类型 (login/create/delete/reboot...) |
| resource | string | - | 操作对象 |
| detail | string | text | 详情描述 |
| ip | string | - | 来源 IP |
| created_at | time | auto | 操作时间 |

## 角色权限

| 角色 | 设备管理 | 配置下发 | 用户管理 | 审计日志 | 系统设置 |
|------|----------|----------|----------|----------|----------|
| admin | 全部 | 全部 | 全部 | 查看 | 全部 |
| operator | 全部 | 全部 | 禁止 | 禁止 | 禁止 |
| viewer | 只读 | 禁止 | 禁止 | 禁止 | 禁止 |

## API 接口

### POST /api/v1/auth/login (公开)

```json
// 请求
{ "username": "admin", "password": "admin123" }

// 响应
{
  "token": "eyJhbGciOi...",
  "user": { "id": 1, "username": "admin", "role": "admin" }
}
```

- 密码校验：bcrypt.CompareHashAndPassword
- Token 签名算法：HS256
- Token 有效期：24 小时
- Token Payload：`{ user_id, username, role, exp, iat }` (基于 jwt.RegisteredClaims)

### GET /api/v1/users (admin)

返回所有用户列表。密码字段不输出 (json:"-")。

### POST /api/v1/users (admin)

```json
{ "username": "ops1", "password": "securepass", "role": "operator", "email": "ops@example.com" }
```

- 密码最少 8 位
- 密码存储为 bcrypt hash (cost=10)

### DELETE /api/v1/users/:id (admin)

软删除用户。

### GET /api/v1/audit-logs (admin)

返回最近 200 条审计日志，按时间倒序。

## JWT 中间件

```
请求 → JWTAuth 中间件 → 解析 Bearer Token → 写入 gin.Context
  ├─ c.Set("user_id", claims.UserID)
  ├─ c.Set("username", claims.Username)
  └─ c.Set("role", claims.Role)
```

### RequireRole 中间件

```
请求 → RequireRole("admin") → 检查 c.Get("role") == "admin"
  ├─ 通过 → next()
  └─ 拒绝 → 403 Forbidden
```

## 自动种子

首次启动时，若 users 表为空，自动创建：
- 用户名：admin
- 密码：admin123
- 角色：admin

## 前端认证流程

1. 用户提交登录表单 → 调用 `login()` API
2. 成功 → 存储 token 和 username 到 localStorage
3. Axios 拦截器自动在每次请求 Header 添加 `Authorization: Bearer {token}`
4. 收到 401 响应 → 清除 token → 跳转到 /login
5. 路由守卫：非 /login 路由检查 localStorage 中是否有 token
