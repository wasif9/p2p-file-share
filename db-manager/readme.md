# Run with 
```bash
go run ./main.go config.json
```

or 
```bash
go run ./... config.json
```

# Run with Docker
Assuming at root directory P2P-FILE-SHARE and navigate to db-manager folder first by running 
```bash
cd db-manager
```
then run
```bash
docker-compose down -v
docker-compose up --build -d
```

## Test with (in a separate terminal)
```bash
curl localhost:8081/api/v1/heartbeat
```

# Setting up DB

## Install posgres 
https://www.postgresql.org/download/

## Creating the registry[x] database

```bash
psql -u postgres
```

```sql
CREATE DATABASE registry0;
```

# Testing

Tests are in `*_test.go` files. First start the db manager and then execute the tests in another terminal. This isn't technically the proper way to do it, but it is easy and it works.