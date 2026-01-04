# SQLite Deployment

This configuration runs the Git Manage Service with an embedded SQLite database.

## Usage

1.  **Start the service:**
    ```bash
    docker-compose up -d
    ```

2.  **Data Persistence:**
    *   Database file is stored in `./data/git_sync.db`.
    *   Git repositories are stored in `./repos`.

3.  **Configuration:**
    *   Default port: `8080` (API), `8888` (RPC).
    *   SSH keys are mounted read-only from `~/.ssh` on the host.
