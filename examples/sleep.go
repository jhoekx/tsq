package main

import (
	"errors"
	"fmt"
	"log"
	"time"
)

import "github.com/jhoekx/tsq"

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

func main() {
	server := tsq.New()

	server.Define("sleep", &SleepTask{})

	cmd := tsq.CommandTask{"sleep", []string{"5"}}
	server.Define("sleep-5", &cmd)

	echo := tsq.CommandTask{"echo", []string{"pong"}}
	server.Define("ping", &echo)

	fail := tsq.CommandTask{"false", []string{""}}
	server.Define("fail", &fail)

	_, err := server.TaskQueue.Submit("sleep-5", nil)
	if err != nil {
		log.Fatalln(err)
	}

	server.Start()
}
