# Chat Auth Service

This is a simple chat authentication service that use OAuth2 to authenticate users. The service is written in Go and
uses Postgres for store user data and Redis for session management.

## Getting Started

1. Create a `.env` file in the root of the project with the following variables:
    ```
    PORT=8080
    APP_ENV=local
    DB_HOST=localhost
    DB_PORT=5444
    DB_DATABASE=auth
    DB_USERNAME=<DB_USERNAME>
    DB_PASSWORD=<DB_PASSWORD>
    DB_SCHEMA=public
    SONAR_TOKEN=<SONAR_TOKEN> # Optional
    GOOGLE_APPLICATION_KEY=<GOOGLE_APPLICATION_KEY>
    GOOGLE_APPLICATION_SECRET=<GOOGLE_APPLICATION_SECRET>
    GITHUB_APPLICATION_KEY=<GITHUB_APPLICATION_KEY>
    GITHUB_APPLICATION_SECRET=<GITHUB_APPLICATION_SECRET>
    SESSION_SECRET=<SESSION_SECRET> # Any random string with 32 characters
    SESSION_MANAGER_SECRET=<SESSION_MANAGER> # Any random string with 32 characters
    REDIS_HOST=<REDIS_HOST> i use redislabs.com
    REDIS_PORT=<REDIS_PORT>
    REDIS_PASSWORD=<REDIS_PASSWORD>
    SESSION_TIMEOUT=<SESSION_TIMEOUT> # in seconds '3600s'
    ```

2. Run the following command to start the Postgres and Redis containers
    ```bash
    make docker-run
    ```

3. Run the following command to start the application
    ```bash
    make run
    ```

4. Using a browser navigate to `localhost:8080/login/google?username=<USERNAME>` to authenticate with Google.

   Obs.: github provider is not working.

5. After authenticating, will return a token that can be used to authenticate with the chat service, that is not
   implemented yet.

   Obs.: The token is not a JWT token, it is a random key that is stored in Redis.
   The value of the key is the user information, like roles and permissions.

6. Refresh and Logout endpoints still are in development.

## MakeFile

run all make commands with clean tests
```bash
make all build
```

build the application
```bash
make build
```

run the application
```bash
make run
```

Create DB container
```bash
make docker-run
```

Shutdown DB container
```bash
make docker-down
```

live reload the application
```bash
make watch
```

run the test suite
```bash
make test
```

clean up binary from the last build
```bash
make clean
```