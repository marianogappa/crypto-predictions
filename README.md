# crypto-predictions

Crypto Predictions is a state machine-based engine for tracking crypto-related predictions across social media posts.

https://user-images.githubusercontent.com/1078546/158062428-09e54090-a627-4a3b-9948-41411f1b7191.mp4

## Architecture

![Crypto Predictions Diagram](https://user-images.githubusercontent.com/1078546/158062284-a27b5e31-86a2-4514-9fcc-a60d38230db3.png)

The engine maintains a database of "predictions", which are state machines that track crypto-related predictions like "Bitcoin will hit 100k this year" against historical and live market data from crypto exchanges like Binance and others.

It is shipped as a single binary (Back Office static assets are embedded) which runs all components by default, but can be configured via flags to run individual components separately.

### Components

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
