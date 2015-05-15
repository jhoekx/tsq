package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jhoekx/tsq"
)

type SleepTask struct{}

func (t *SleepTask) Run(arguments interface{}) (data interface{}, err error) {
	args, ok := arguments.(map[string]interface{})
	if !ok {
		err = errors.New("arguments required")
		return
	}
	duration, ok := args["duration"]
	if !ok {
		err = errors.New("duration argument required")
		return
	}
	length, ok := duration.(float64)
	if !ok {
		err = errors.New("duration float64 argument required, was " + fmt.Sprintf("%T", duration))
		return
	}
	time.Sleep(time.Duration(length) * time.Second)
	return
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

func main() {
	cleanInterval := 1 * time.Minute
	maxAge := 2 * time.Minute
	qConfig := tsq.Config{
		JobStore: tsq.NewCleanedMemoryStore(cleanInterval, maxAge),
	}
	q := qConfig.NewQueue()

	q.Define("sleep", &SleepTask{})

	cmd := tsq.CommandTask{"sleep", []string{"5"}}
	q.Define("sleep-5", &cmd)

	echo := tsq.CommandTask{"echo", []string{"pong"}}
	q.Define("ping", &echo)

	fail := tsq.CommandTask{"false", []string{""}}
	q.Define("fail", &fail)

	_, err := q.Submit("sleep-5", nil)
	if err != nil {
		log.Fatalln(err)
	}

	q.Start()

	http.Handle("/tsq/", logRequests(tsq.ServeQueue("/tsq/", q)))
	log.Fatalln(http.ListenAndServe(":8000", nil))
}
