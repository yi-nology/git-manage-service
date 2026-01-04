# MySQL Deployment

This configuration runs the Git Manage Service with a dedicated MySQL database.

## Usage

1.  **Start the service:**
    ```bash
    docker-compose up -d
    ```

2.  **Database Info:**
    *   **User**: `gituser`
    *   **Password**: `gitpassword`
    *   **Database**: `git_manage`
    *   **Host**: `mysql` (internal network)

3.  **Data Persistence:**
    *   MySQL data is stored in a named volume `mysql_data`.
    *   Git repositories are stored in `./repos`.
