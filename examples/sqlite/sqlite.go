package main

import (
	"log"
	"net/http"

	"github.com/jhoekx/tsq"
)

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

func main() {
	qConfig := tsq.Config{
		JobStore: tsq.NewSQLiteStore(),
	}
	q := qConfig.NewQueue()

	echo := tsq.CommandTask{"echo", []string{"pong"}}
	q.Define("ping", &echo)

	fail := tsq.CommandTask{"false", []string{""}}
	q.Define("fail", &fail)

	q.Start()

	http.Handle("/tsq/", logRequests(tsq.ServeQueue("/tsq/", q)))
	log.Fatalln(http.ListenAndServe(":8000", nil))
}
