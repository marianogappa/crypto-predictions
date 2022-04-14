# crypto-predictions

Crypto Predictions is a state machine-based engine for tracking crypto-related predictions across social media posts.

https://user-images.githubusercontent.com/1078546/158062428-09e54090-a627-4a3b-9948-41411f1b7191.mp4

## Architecture

![Crypto Predictions Diagram](https://user-images.githubusercontent.com/1078546/158062284-a27b5e31-86a2-4514-9fcc-a60d38230db3.png)

The engine maintains a database of "predictions", which are state machines that track crypto-related predictions like "Bitcoin will hit 100k this year" against historical and live market data from crypto exchanges like Binance and others.

It is shipped as a single binary (Back Office static assets are embedded) which runs all components by default, but can be configured via flags to run individual components separately.

## Getting started

- Download latest binary from [here](https://github.com/marianogappa/crypto-predictions/releases/latest).
- Have an addressable postgres instance, e.g. `brew install postgresql && brew services start postgresql`
- Note that, by default, crypto-predictions will try to connect to `postgres://your_user:@localhost:5432/your_user?sslmode=disable`, which will work on a default install, but this will add tables to a database with your username, and you may already have something else there. If you don't want that, just set the adequate environment variables below to customise that behaviour. It is on you to create a different database, but the binary will run migrations on it.
- There are only two required envs: `PREDICTIONS_TWITTER_BEARER_TOKEN` & `PREDICTIONS_YOUTUBE_API_KEY`, which are the minimum credentials from Twitter API & Youtube API to be able to fetch metadata for creating predictions. You can follow Twitter & Youtube's instructions to get these.
- Run the binary with the envs: `PREDICTIONS_TWITTER_BEARER_TOKEN=value1 PREDICTIONS_YOUTUBE_API_KEY=value2 ./crypto-predictions`
- The binary logs structured JSON logs, so you might want to run like this: `./crypto-predictions 2>&1 | jq .` for pretty-printed output.
- Logs will tell you the Postgres instance it connected to, where the API is listening, where the BackOffice can be browsed and where you can see Swagger docs for the API.

### Configuration

Crypto Predictions is configured via enviroment variables. The following variables are supported:

#### Social network API credentials

- `PREDICTIONS_TWITTER_BEARER_TOKEN`: required if running API or BackOffice, so that metadata can be fetched for Predictions & Accounts.
- `PREDICTIONS_YOUTUBE_API_KEY`: required if running API or BackOffice, so that metadata can be fetched for Predictions & Accounts.

#### Database configuration

- `PREDICTIONS_POSTGRES_USER`: defaults to current user.
- `PREDICTIONS_POSTGRES_PASS`: defaults to empty string.
- `PREDICTIONS_POSTGRES_PORT`: defaults to 5432.
- `PREDICTIONS_POSTGRES_DATABASE`: defaults to current user.
- `PREDICTIONS_POSTGRES_HOST`: defaults to localhost.

#### Components configuration

- `PREDICTIONS_API_PORT`: defaults to 2345. In the special case of running BackOffice but not API, setting this or the URL is required.
- `PREDICTIONS_API_URL`: defaults to localhost:2345. In the special case of running BackOffice but not API, setting this or the PORT is required.
- `PREDICTIONS_BACKOFFICE_PORT`: defaults to 1234.
- `PREDICTIONS_DAEMON_DURATION`: defaults to 60 seconds. The format is as described here: https://pkg.go.dev/time#ParseDuration.
- `PREDICTIONS_DEBUG`: set to any value to enable debugging logs.

## Components

**API**

The API is responsible for CRUDing predictions and related entities against the storage component.

**Daemon**

The Daemon is a background process that continuously queries crypto exchanges for historical and live market data, and updates the predictions' state.

**Back Office**

The Back Office is a UI interface for maintaining the predictions database. Check the summary video for an example of how it works.

**Storage**

The Storage component is an interface to a Postgres database for storing the predictions state machines and related entities. A memory implementation also exists for testing. It can easily be hooked up on main.go if you want to test the engine without having a Postgres instance deployed.

**Market**

The Market component is responsible for querying all supported exchanges for historical and live market data.

**Metadata Fetcher**

The Metadata Fetcher component is responsible for querying social media post metadata (e.g. when a post was created, which user posted, the post text) from Twitter/Youtube, so that predictions can be created without the need to manually input these fields.

**Website**

A website is planned, as the interface for the main users to the engine.
