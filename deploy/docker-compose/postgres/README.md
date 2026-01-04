# PostgreSQL Deployment

This configuration runs the Git Manage Service with a dedicated PostgreSQL database.

## Usage

1.  **Start the service:**
    ```bash
    docker-compose up -d
    ```

2.  **Database Info:**
    *   **User**: `gituser`
    *   **Password**: `gitpassword`
    *   **Database**: `git_manage`
    *   **Host**: `postgres` (internal network)

3.  **Data Persistence:**
    *   PostgreSQL data is stored in a named volume `postgres_data`.
    *   Git repositories are stored in `./repos`.
