# Single task queue

## Getting Started
```sh
[jeroen@jhoekx-laptop tsq]$ pwd
/home/jeroen/dev/go/src/github.com/jhoekx/tsq
[jeroen@jhoekx-laptop tsq]$ go get github.com/mattn/go-sqlite3
[jeroen@jhoekx-laptop tsq]$ go install github.com/mattn/go-sqlite3
[jeroen@jhoekx-laptop tsq]$ go test
PASS
ok      github.com/jhoekx/tsq   0.003s
[jeroen@jhoekx-laptop tsq]$ cd examples/sqlite/
[jeroen@jhoekx-laptop sqlite]$ go run sqlite.go
2017/01/13 11:10:02 Running migration: V1__001_CreateJobDB
[jeroen@jhoekx-laptop sqlite]$ curl -X GET http://localhost:8000/tsq/
[{"name":"tasks","href":"/tsq/tasks/"},{"name":"jobs","href":"/tsq/jobs/"}]
```
