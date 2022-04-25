# pluteo

`pluteo` is the central server receiving, archiving and serving all data available through [skybi.dev](skybi.dev)

## Running in production (Docker)

To run `pluteo` in production, we strongly recommend simply using our stable Docker images released to GHCR.

To make the setup more straightforward, an [example `docker-compose.yml`](docker-compose.example.yml) is provided.

## Running in development (local)

To run a local development copy, simply clone the commit/branch you want to run and create a `.env` file in the same
directory you run the `go run` or binary command.

See the [example `.env`](.env.example) for a quick overview.

## Configuration variables

| Environment variable           | Type            | Default                 | Description                                                                                                            |
|--------------------------------|-----------------|-------------------------|------------------------------------------------------------------------------------------------------------------------|
| `SB_ENVIRONMENT`               | `prod` or `dev` | `prod`                  | Whether the server starts in development or production mode                                                            |
| `SB_POSTGRES_DSN`              | `PSQL DSN`      | `<none>`                | The PostgreSQL connection string to use                                                                                |
| `SB_PORTAL_API_LISTEN_ADDRESS` | `URI`           | `:8081`                 | The URI the portal API listens to                                                                                      |
| `SB_PORTAL_API_BASE_ADDRESS`   | `URL`           | `http://localhost:8081` | The absolute base address the portal API will be accessible from (used for session cookies)                            |
| `SB_PORTAL_API_ALLOWED_ORIGIN` | `URL`           | `http://localhost:3000` | The content of the `Access-Control-Allow-Origin` CORS header for the portal API (used for portal frontend deployments) |
| `SB_OIDC_PROVIDER_URL`         | `URL`           | `<none>`                | The OIDC provider URL to use for SSO                                                                                   |
| `SB_OIDC_CLIENT_ID`            | `string`        | `<none>`                | The client ID used to connect to the OIDC provider                                                                     |
| `SB_OIDC_CLIENT_SECRET`        | `string`        | `<none>`                | The client secret used to connect to the OIDC provider                                                                 |
| `SB_DATA_API_LISTEN_ADDRESS`   | `URI`           | `:8082`                 | The URI the data API listens to                                                                                        |
