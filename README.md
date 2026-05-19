# S3 mock

Minimal S3-compatible storage server for local development and testing.

Objects are stored directly on the filesystem and exposed through a small HTTP API.

## Features

- Pure Go with zero dependencies
- Objects stored as files
- Docker-ready

## Supported endpoints

```http
GET  /path/to/object
HEAD /path/to/object
PUT  /path/to/object
```

For example, the following request:

```http
PUT /images/avatar.png
```

stores file at:

```text
./storage/images/avatar.png
```

## Usage

Server available at `http://localhost:8080`.

### Run locally

```bash
go run .
```

### Run with Docker

```bash
make build
make run
```

### Persist files locally

After building the image with `make build`:

```bash
docker run \
  --rm \
  -p 8080:8080 \
  -v $PWD/data:/app/storage \
  s3-mock
```

## Development

```bash
make test # run tests
make coverage-html # open HTML coverage report
make lint # run Go linters
make lint-dockerfile # lint Dockerfile
make check # run all checks
```
