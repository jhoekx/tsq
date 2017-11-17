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
[jeroen@jhoekx-laptop sqlite]$ curl -X POST http://localhost:8000/tsq/tasks/sleep-5/
{"uuid":"85724738-97aa-404d-8007-add0c6bec1cf","name":"sleep-5","status":"PENDING","arguments":null,"result":null,"created":"2017-11-17T21:26:22.723086282+01:00","updated":"2017-11-17T21:26:22.723086282+01:00","href":"/tsq/jobs/85724738-97aa-404d-8007-add0c6bec1cf/"}
[jeroen@jhoekx-laptop sqlite]$ curl -X POST http://localhost:8000/tsq/tasks/sleep-5/?jobTimeoutSeconds=10
{"uuid":"6cb844cb-261c-454d-a52f-1c53f3394175","name":"sleep-5","status":"SUCCESS","arguments":"","result":{"stderr":"","stdout":""},"created":"2017-11-17T21:26:05.283245685+01:00","updated":"2017-11-17T21:26:10.306581887+01:00","href":"/tsq/jobs/6cb844cb-261c-454d-a52f-1c53f3394175/"}
[jeroen@jhoekx-laptop examples]$ curl -X POST http://localhost:8000/tsq/tasks/sleep-5/?jobTimeoutSeconds=1
HTTP 504 Timed out waiting for job f76344fd-ce62-49f5-b628-515d759321bc
```
