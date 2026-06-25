# Transit-Pulse

A tracker for all the MTA public subway lines in NYC.

<video controls width="800">
    <source src="assets/demo" type="video/webm">
</video>

## Features

- Update with the GTFS-RT information provided in protobuf format from the [MTA API](https://mta.info/developers)
- Track last known location of subway trainsets every 30 seconds

## Setup

Developing and deploying both rely on container technology, either `docker` or `podman`.

### Development

1. Clone the repository
2. Deploy the `postgres` and `valkey` instances via the script found in `backend/infra/compose.yml`
3. Create/modify a `.env` file modeled after `.env-example` in `backend`
4. Install dependencies and start both frontend and backend

### Deployment

A `compose.yml` file for deploying has been provided. Unfortunately, the container images for the app have not been uploaded to any image libraries. Therefore, you will have to build them locally, which you can do with the following command.

```
docker compose up -d --build
```
