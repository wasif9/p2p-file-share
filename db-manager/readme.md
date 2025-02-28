## Run with 
```bash
go run ./main.go
```

or 
```bash
go run ./...
```

## Test with (in a separate terminal)
```bash
curl localhost:8080/foo
```

Or just go to http://localhost:8080/whatever 🙂

# Setting up DB

## Install posgres 
https://www.postgresql.org/download/

## Creating the registry[x] database

```bash
psql -u postgres
```

```sql
CREATE DATABASE registry0;