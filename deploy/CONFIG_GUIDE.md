# Git Manage Service 配置指南

本文档详细说明了 `config.yaml` 配置文件中的各项配置，帮助您更好地管理和部署服务。

## 1. 服务器配置 (server)

配置 HTTP 服务器的基本参数。

| 配置项 | 类型 | 默认值 | 必填 | 说明 |
| :--- | :--- | :--- | :--- | :--- |
| `port` | int | `8080` | 是 | 服务监听的端口号。请确保该端口未被占用。 |

**示例：**
```yaml
server:
  port: 8080
```

---

## 2. 数据库配置 (database)

配置数据持久化存储。支持 SQLite, MySQL, PostgreSQL。

| 配置项 | 类型 | 默认值 | 必填 | 说明 |
| :--- | :--- | :--- | :--- | :--- |
| `type` | string | `sqlite` | 是 | 数据库类型。可选值：`sqlite`, `mysql`, `postgres`。 |
| `path` | string | `git_sync.db` | 否 | SQLite 数据库文件路径。仅当 `type` 为 `sqlite` 时有效。建议使用绝对路径。 |
| `host` | string | - | 否 | 数据库服务器主机地址。MySQL/PGSQL 必填。 |
| `port` | int | - | 否 | 数据库端口。MySQL 默认 `3306`，PGSQL 默认 `5432`。 |
| `user` | string | - | 否 | 数据库用户名。 |
| `password` | string | - | 否 | 数据库密码。建议通过环境变量注入，不要直接提交到代码库。 |
| `dbname` | string | - | 否 | 数据库名称。 |
| `dsn` | string | - | 否 | 自定义 DSN 连接字符串。如果提供，将覆盖上述 host/port/user 等字段自动生成的 DSN。 |

**示例 (SQLite):**
```yaml
database:
  type: sqlite
  path: data/git_sync.db
```

**示例 (MySQL):**
```yaml
database:
  type: mysql
  host: localhost
  port: 3306
  user: root
  password: secure_password
  dbname: git_manage
```

---

## 3. Webhook 配置 (webhook)

配置外部系统回调（Webhook）的安全策略。

| 配置项 | 类型 | 默认值 | 必填 | 说明 |
| :--- | :--- | :--- | :--- | :--- |
| `secret` | string | `my-secret-key` | 是 | 用于验证 Webhook 请求签名的密钥。务必修改此值以保证安全。 |
| `rate_limit` | int | `100` | 否 | 每分钟允许的请求次数限制。 |
| `ip_whitelist` | list | `[]` | 否 | 允许访问 Webhook 接口的 IP 白名单。留空表示不限制 IP。 |

**示例：**
```yaml
webhook:
  secret: "change_me_to_something_secure"
  rate_limit: 60
  ip_whitelist:
    - "192.168.1.100"
    - "10.0.0.5"
```

---

## 4. 调试模式 (debug)

| 配置项 | 类型 | 默认值 | 必填 | 说明 |
| :--- | :--- | :--- | :--- | :--- |
| `debug` | bool | `false` | 否 | 是否开启调试模式。开启后可能会输出更多日志信息。 |

**示例：**
```yaml
debug: true
```

---

## 最佳实践

1. **不要直接在 git 中提交包含密码的 config.yaml**。
   - 生产环境建议使用环境变量覆盖敏感配置（如 `DB_PASSWORD` 覆盖 config 中的 password）。
   - 或者使用 Kubernetes Secret / Docker Secrets 管理配置文件。
2. **区分环境**。
   - 开发环境使用 `debug: true` 和 `sqlite`。
   - 生产环境建议使用 `mysql` 或 `postgres` 并关闭 `debug`。
