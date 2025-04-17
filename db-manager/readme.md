# Run with 
```bash
go run ./... superconfig.json 0
```
From inside `db-manager`, where 0 is the node index. That is, the service will be started with the configuration specified in the 0th element of the config.json object.

## Test with (in a separate terminal)
```bash
curl localhost:8081/api/v1/heartbeat
```

# Setting up DB

## Install posgres 
https://www.postgresql.org/download/

## Creating the registry[x] database

```bash
psql -U postgres
```

```sql
CREATE DATABASE registry0;
```

# Testing

Tests are in `*_test.go` files. First start the db manager and then execute the tests in another terminal. This isn't technically the proper way to do it, but it is easy and it works.