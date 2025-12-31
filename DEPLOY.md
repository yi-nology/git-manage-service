# Git Manage Service Deployment Guide

## Docker Deployment (Recommended)

### Prerequisites

- Docker installed
- Docker Compose installed

### Quick Start

1. **Clone the repository** (if not already done).

2. **Configure Environment**:
   ```bash
   cp .env.example .env
   # Edit .env to set your secrets
   ```

3. **Prepare Directories**:
   ```bash
   # Create data directory for database persistence
   mkdir -p data
   
   # Create repos directory (or use existing path)
   mkdir -p repos
   ```

4. **Start the Service**:
   ```bash
   docker-compose up -d
   ```

5. **Access the Application**:
   Open browser at `http://localhost:8080`

### Volume Configuration

The `docker-compose.yml` defines three key volumes:

1. **Database Storage** (`./data`):
   - Maps to `/app/data` in container.
   - Persists the SQLite database (`git_sync.db`).

2. **SSH Keys** (`~/.ssh`):
   - Maps host's SSH keys to `/root/.ssh` in container (Read-Only).
   - Allows the tool to use your existing SSH keys for Git authentication.
   - **Note**: Ensure your host SSH keys have correct permissions.

3. **Repositories** (`./repos`):
   - Maps to `/repos` in container.
   - This is where you should clone/store your git repositories.
   - When registering a repo in the UI, use the container path (e.g., `/repos/my-project`).

### SSH Authentication Setup

To ensure the container can access private repositories via SSH:

1. **Host Keys**: The default configuration mounts `~/.ssh` from host to `/root/.ssh` in container.
   - Ensure your `id_rsa` (or other key) exists on host.
   - Ensure `known_hosts` contains the git server fingerprints (e.g., GitHub, GitLab).

2. **Permissions**:
   Docker typically mounts with root permissions. If you encounter permission issues:
   ```bash
   # On Host
   chmod 600 ~/.ssh/id_rsa
   ```

3. **Known Hosts**:
   If the container complains about "Host key verification failed", you can manually add the host key in the container or ensure your host's `known_hosts` is populated.
   
   *Auto-scan inside container:*
   ```bash
   docker exec -it git-manage-service ssh-keyscan github.com >> /root/.ssh/known_hosts
   ```

### Troubleshooting

- **Database Errors**: Check write permissions on `./data` directory.
- **Git Errors**: Check `docker logs git-manage-service` for detailed git output (enable Debug Mode in UI settings).
- **Timezone**: Default is `Asia/Shanghai`. Change `TZ` in `docker-compose.yml` if needed.
