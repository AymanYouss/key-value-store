# Persistent Key-Value Store

This is a simple persistent key-value store with an HTTP API, inspired by the LSM tree model. It provides endpoints for getting, setting, and deleting key-value pairs.

## Overview

The key features of this key-value store include:

- **GET Endpoint**: Retrieve the value of a key or print 'Key not found' if the key is not present.

- **SET Endpoint**: Set a key-value pair by sending a JSON-encoded request in the body.

- **DELETE Endpoint**: Delete a key from the key-value store and return the existing value if it exists.

## Architecture

The key-value store follows the LSM tree model:

- Write operations are first written to the memtable (a sorted map of key-value pairs) and appended to the Write Ahead Log (WAL) for crash safety.

- Periodically, the contents of the memtable are flushed to the disk as an SST file (Sorted String Table).

## Getting Started

To get started with the key-value store, follow these steps:

1. Clone the repository:

    ```bash
    git clone https://github.com/your-username/key-value-store.git
    ```

2. Build and run the project:

    ```bash
    cd key-value-store
    go build
    .
    ```

3. Access the key-value store via the provided HTTP endpoints:

    - GET: http://localhost:8080/get?key=keyName
    - POST: http://localhost:8080/set
    - DELETE: http://localhost:8080/del?key=keyName

## Usage

### GET Endpoint

Retrieve the value of a key:

```bash
curl http://localhost:8080/get?key=keyName
```

### SET Endpoint
```bash
curl -X POST -H "Content-Type: application/json" -d '{"key": "keyName", "value": "someValue"}' http://localhost:8080/set
```

### DELETE Endpoint
```bash
curl -X DELETE http://localhost:8080/del?key=keyName
```


