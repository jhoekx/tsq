package tsq

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Server struct {
	TaskQueue *TaskQueue
	Router    *mux.Router
}

func New() (server *Server) {
	server = &Server{}
	server.Router = mux.NewRouter()

	queue := &TaskQueue{}
	queue.Tasks = make(map[string]Runner)
	queue.jobQueue = make(chan *Job, 10)
	queue.Jobs = make([]*Job, 0, 10)
	queue.CleanInterval = time.Duration(24) * time.Hour
	queue.MaxAge = time.Duration(7*24) * time.Hour
	server.TaskQueue = queue

	return
}

func (s *Server) Define(name string, r Runner) {
	s.TaskQueue.Define(name, r)
}

func (s *Server) Start() {
	s.TaskQueue.Start()

	s.Router.HandleFunc("/", jsonResponse(s.listServices))
	s.Router.HandleFunc("/tasks/", jsonResponse(s.listDefinedTasks)).Name("tasks")
	s.Router.HandleFunc("/tasks/{name}/", jsonResponse(s.submitTask)).Methods("POST").Name("submitTask")
	s.Router.HandleFunc("/jobs/", jsonResponse(s.listJobs)).Name("jobs")
	s.Router.HandleFunc("/jobs/{uuid}/", jsonResponse(s.getJobStatus)).Name("job")

	http.Handle("/", s.Router)
	log.Fatalln(http.ListenAndServe(":8080", nil))
}

type NameRef struct {
	Name string `json:"name"`
	Href string `json:"href"`
}

func (s *Server) listServices(w http.ResponseWriter, r *http.Request) (data interface{}, err error) {
	tasksUrl, err := s.Router.Get("tasks").URL()
	if err != nil {
		return
	}
	jobsUrl, err := s.Router.Get("jobs").URL()
	if err != nil {
		return
	}
	services := []NameRef{
		NameRef{"tasks", tasksUrl.String()},
		NameRef{"jobs", jobsUrl.String()},
	}
	data = services
	return
}

func (s *Server) listDefinedTasks(w http.ResponseWriter, r *http.Request) (data interface{}, err error) {
	tasks := make([]NameRef, 0, len(s.TaskQueue.Tasks))
	for key := range s.TaskQueue.Tasks {
		taskUrl, err := s.Router.Get("submitTask").URL("name", key)
		if err != nil {
			return tasks, err
		}
		tasks = append(tasks, NameRef{key, taskUrl.String()})
	}
	data = tasks
	return
}

func (s *Server) submitTask(w http.ResponseWriter, r *http.Request) (data interface{}, err error) {
	name := mux.Vars(r)["name"]

	var arguments interface{}
	if r.Header.Get("Content-Type") == "application/json" && r.ContentLength != 0 {
		err = json.NewDecoder(r.Body).Decode(&arguments)
		if err != nil {
			return
		}
	}

	job, err := s.TaskQueue.Submit(name, arguments)
	if err != nil {
		err = &httpError{404, err}
		return
	}
	url, err := s.Router.Get("job").URL("uuid", job.UUID)
	data = WebJob{job, url.String()}
	return
}

type WebJob struct {
	*Job
	Href string `json:"href"`
}

func (s *Server) listJobs(w http.ResponseWriter, r *http.Request) (data interface{}, err error) {
	jobs := make([]interface{}, 0, len(s.TaskQueue.Jobs))
	for _, job := range s.TaskQueue.Jobs {
		url, err := s.Router.Get("job").URL("uuid", job.UUID)
		if err != nil {
			return data, err
		}
		jobs = append(jobs, WebJob{job, url.String()})
	}
	data = jobs
	return
}

func (s *Server) getJobStatus(w http.ResponseWriter, r *http.Request) (data interface{}, err error) {
	uuid := mux.Vars(r)["uuid"]
	job, err := s.TaskQueue.GetJob(uuid)
	if err != nil {
		err = &httpError{404, err}
		return
	}

	url, err := s.Router.Get("job").URL("uuid", job.UUID)
	if err != nil {
		return data, err
	}
	data = WebJob{job, url.String()}
	return
}

type httpError struct {
	Status int
	Err    error
}

func (e *httpError) Error() string {
	return "HTTP " + strconv.Itoa(e.Status) + " " + e.Err.Error()
}

type httpHandler func(http.ResponseWriter, *http.Request) (interface{}, error)

func jsonResponse(fn httpHandler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := fn(w, r)
		if e, ok := err.(*httpError); ok {
			http.Error(w, e.Error(), e.Status)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		err = json.NewEncoder(w).Encode(data)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
		return
	}
}
