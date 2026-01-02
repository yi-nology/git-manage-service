# 部署指南

本指南涵盖了 Git Manage Service 的 Docker 和 Kubernetes 部署流程。

## 目录结构

```
deploy/
├── config.yaml          # 应用主配置文件
├── CONFIG_GUIDE.md      # 配置文件详细说明
├── .env                 # 环境变量文件 (敏感信息)
├── docker-compose.yml   # Docker Compose 编排文件
└── k8s/                 # Kubernetes 资源清单
    ├── configmap.yaml
    ├── secret.yaml
    ├── mysql.yaml
    ├── deployment.yaml
    └── service.yaml
```

---

## 1. 本地 Docker 部署

适用于开发测试或单机部署环境。

### 步骤

1. **准备配置文件**
   - 检查 `deploy/config.yaml`，根据需要修改配置（详见 `CONFIG_GUIDE.md`）。
   - 检查 `deploy/.env`，设置数据库密码等敏感信息。

2. **启动服务**
   进入 `deploy` 目录并运行 Docker Compose：
   ```bash
   cd deploy
   docker-compose up -d
   ```
   
   该命令会自动：
   - 构建应用镜像（使用根目录下的 Dockerfile）。
   - 启动 MySQL 数据库容器。
   - 启动应用容器，并连接到 MySQL。

3. **验证部署**
   - 访问 `http://localhost:8080` 确认服务运行正常。
   - 查看日志：`docker-compose logs -f app`。

### 环境变量说明 (.env)

| 变量名 | 默认值 | 说明 |
| :--- | :--- | :--- |
| `APP_PORT` | `8080` | 应用对外暴露端口 |
| `DB_TYPE` | `mysql` | 数据库类型 (mysql/sqlite/postgres) |
| `DB_PASSWORD` | - | 数据库连接密码 |
| `WEBHOOK_SECRET` | - | Webhook 签名密钥 |

---

## 2. Kubernetes 集群部署

适用于生产环境的高可用部署。

### 步骤

1. **创建 ConfigMap 和 Secret**
   ```bash
   kubectl apply -f deploy/k8s/configmap.yaml
   kubectl apply -f deploy/k8s/secret.yaml
   ```
   *注意：生产环境中，建议使用 SealedSecrets 或其他密钥管理工具管理 Secret。*

2. **部署数据库 (可选)**
   如果使用集群外部的数据库，请跳过此步并修改 ConfigMap 中的数据库地址。
   ```bash
   kubectl apply -f deploy/k8s/mysql.yaml
   ```

3. **部署应用**
   ```bash
   kubectl apply -f deploy/k8s/deployment.yaml
   kubectl apply -f deploy/k8s/service.yaml
   ```

### 常见问题排查

**Q1: Pod 启动失败，状态为 CrashLoopBackOff**
- 查看日志：`kubectl logs -f <pod-name>`
- 检查数据库连接配置是否正确（Host, Port, User, Password）。
- 确认数据库服务是否已就绪。

**Q2: 无法挂载 SSH 密钥**
- `deployment.yaml` 中使用了 `hostPath` 挂载 `/root/.ssh`。这依赖于节点上存在该路径。
- **解决方案**：建议将 SSH 私钥创建为 Kubernetes Secret，并挂载到 Pod 中，而不是依赖宿主机文件。

**Q3: 配置文件未生效**
- 确认 ConfigMap 已更新，并且 Pod 已重启（ConfigMap 挂载通常需要重启 Pod 才能加载最新更改，或等待 kubelet 同步）。

---

## 3. 多环境支持

- **开发环境**：直接使用 `docker-compose.yml`，配合 `DB_TYPE=sqlite` 可快速启动。
- **生产环境**：
  - 建议使用 Kubernetes 部署。
  - 将 `config.yaml` 中的 `debug` 设为 `false`。
  - 数据库密码等敏感信息**必须**通过环境变量或 Secret 注入，不要写在 `config.yaml` 明文中。
