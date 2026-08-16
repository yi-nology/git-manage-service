# 快速开始

本指南将帮助你在 5 分钟内完成 Git Manage Service 的安装和基本配置。

![文档截图](/git-manage-service/images/docs/docs-getting-started.png)

## 1. 下载安装

### 方式一：下载预编译二进制（推荐）

从 [Releases](https://github.com/yi-nology/git-manage-service/releases) 页面下载适合你系统的版本：

| 平台 | 架构 | 文件名 |
|------|------|--------|
| Linux | AMD64 | `git-manage-service-linux-amd64.tar.gz` |
| Linux | ARM64 | `git-manage-service-linux-arm64.tar.gz` |
| macOS | Intel | `git-manage-service-darwin-amd64.tar.gz` |
| macOS | Apple Silicon | `git-manage-service-darwin-arm64.tar.gz` |
| Windows | AMD64 | `git-manage-service-windows-amd64.exe.zip` |
| Windows | ARM64 | `git-manage-service-windows-arm64.exe.zip` |

### 方式二：Docker

```bash
docker pull ghcr.io/yi-nology/git-manage-service:latest
```

### 方式三：从源码编译

```bash
git clone https://github.com/yi-nology/git-manage-service.git
cd git-manage-service
make build-full
```

## 2. 启动服务

### Linux / macOS

```bash
# 解压
tar -xzf git-manage-service-*.tar.gz

# 添加执行权限
chmod +x git-manage-service-*

# 设置凭证加密密钥（SyncV2 必需；首次设置后请固定保存，丢失将无法解密已存凭证）
export ENCRYPTION_KEY=$(openssl rand -hex 32)

# 运行
./git-manage-service-*
```

### Windows

```powershell
# 解压 zip 文件
# 双击运行或在命令行中执行
$env:ENCRYPTION_KEY = "your-encryption-key"   # 凭证加密密钥，首次设置后请固定保存
.\git-manage-service-windows-amd64.exe
```

### Docker

```bash
docker run -d \
  --name git-manage-service \
  -p 12345:12345 \
  -v ./data:/app/data \
  ghcr.io/yi-nology/git-manage-service:latest
```

## 3. 访问界面

浏览器打开: [http://localhost:12345](http://localhost:12345)

![首页](/git-manage-service/images/homepage.png)

## 4. 基本配置

### 4.1 添加仓库

1. 点击顶部导航 **「本地仓库」**
2. 点击 **「注册仓库」** 按钮
3. 输入仓库 **名称**、**本地路径** 和远程 URL
4. 点击保存

![注册仓库](/git-manage-service/images/repo-register.png)

### 4.2 配置 SSH 密钥（如需访问私有仓库）

1. 点击顶部导航 **「设置」**
2. 进入 **「SSH 密钥」** 页面
3. 点击 **「添加密钥」** 按钮
4. 粘贴你的 SSH 私钥内容
5. 保存后，在仓库配置中选择该密钥

![SSH 密钥管理](/git-manage-service/images/ssh-keys.png)

### 4.3 创建同步任务

1. 进入 **「本地仓库」** → 打开目标仓库详情页
2. 点击左侧功能菜单 **「同步任务」** 标签（也可直接访问 `/sync` 查看全局任务列表）
3. 点击 **「新建任务」** 按钮，配置同步规则：
   - **源仓库 / 源分支**: 选择已注册的仓库，如 `main`
   - **目标仓库 / 目标分支**: 如备份仓库 / `main`
   - **Cron 表达式**: 如 `0 */2 * * *`（每 2 小时同步），留空则仅手动触发
4. 保存并启用

![新建同步任务](/git-manage-service/images/sync-task-create.png)

### 4.4 配置通知（可选）

1. 点击顶部导航 **「设置」**
2. 进入 **「通知渠道」** 页面
3. 添加通知渠道（钉钉/企微/飞书等）
4. 配置触发事件和消息模板

![通知渠道配置](/git-manage-service/images/notification-channel.png)

## 5. 验证功能

### 手动执行同步

在同步任务列表中，点击操作列的运行按钮手动触发一次同步。

### 查看执行日志

1. 点击顶部导航 **「审计日志」**
2. 按操作类型筛选
3. 查看详细的执行记录

![审计日志](/git-manage-service/images/audit-log.png)

## 下一步

- 📘 [功能指南](/features/repo) - 了解所有功能的详细用法
- 📦 [部署方案](/deployment/binary) - 生产环境部署指南
- ⚙️ [配置参考](/configuration) - 完整的配置项说明
- 🔌 [API 文档](/api) - HTTP API 接口参考

## 常见问题

### 端口被占用？

修改配置文件 `conf/config.yaml`：

```yaml
server:
  port: 8080  # 改为其他端口
```

### 无法访问私有仓库？

确保：
1. SSH 密钥已正确添加
2. 仓库配置中选择了正确的密钥
3. 密钥有访问目标仓库的权限

### 同步失败？

查看审计日志中的错误信息，常见原因：
- 网络问题
- 权限不足
- 分支冲突
- Remote 配置错误

### 启动日志出现 "ENCRYPTION_KEY environment variable is required"？

同步引擎（SyncV2）的凭证加密需要 `ENCRYPTION_KEY` 环境变量：

```bash
export ENCRYPTION_KEY=$(openssl rand -hex 32)   # 首次生成后请固定保存并复用
```

缺失该变量时同步任务相关接口会返回 500。源码运行可用 `./control.sh start` 自动处理。
