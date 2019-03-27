package tsq

import (
	"encoding/json"
	"errors"
	"mime"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type server struct {
	taskQueue *TaskQueue
	router    *mux.Router
}

func ServeQueue(baseURL string, q *TaskQueue) http.Handler {
	svr := &server{
		router:    mux.NewRouter().PathPrefix(baseURL).Subrouter(),
		taskQueue: q,
	}
	svr.registerRoutes()

	return svr.router
}

func (s *server) registerRoutes() {
	s.router.HandleFunc("/", jsonResponse(s.listServices))
	s.router.HandleFunc("/tasks/", jsonResponse(s.listDefinedTasks)).Name("tasks")
	s.router.HandleFunc("/tasks/{name}/", jsonResponse(s.submitTask)).Methods("POST").Name("submitTask")
	s.router.HandleFunc("/jobs/", jsonResponse(s.listJobs)).Name("jobs")
	s.router.HandleFunc("/jobs/{uuid}/", jsonResponse(s.getJobStatus)).Name("job")
}

type NameRef struct {
	Name string `json:"name"`
	Href string `json:"href"`
}

func (s *server) listServices(w http.ResponseWriter, r *http.Request) (data interface{}, err error) {
	tasksUrl, err := s.router.Get("tasks").URL()
	if err != nil {
		return
	}
	jobsUrl, err := s.router.Get("jobs").URL()
	if err != nil {
		return
	}
	services := []NameRef{
		{"tasks", tasksUrl.String()},
		{"jobs", jobsUrl.String()},
	}
	data = services
	return
}

func (s *server) listDefinedTasks(w http.ResponseWriter, r *http.Request) (data interface{}, err error) {
	tasks := make([]NameRef, 0, len(s.taskQueue.tasks))
	for key := range s.taskQueue.tasks {
		taskUrl, err := s.router.Get("submitTask").URL("name", key)
		if err != nil {
			return tasks, err
		}
		tasks = append(tasks, NameRef{key, taskUrl.String()})
	}
	data = tasks
	return
}

func (s *server) submitTask(w http.ResponseWriter, r *http.Request) (data interface{}, err error) {
	name := mux.Vars(r)["name"]
	timeout, err := getTimeout(r)
	if err != nil {
		return
	}

	var arguments interface{}
	if r.Header.Get("Content-Type") != "" {
		mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			return nil, err
		}
		if mt == "application/json" && r.ContentLength != 0 {
			err = json.NewDecoder(r.Body).Decode(&arguments)
			if err != nil {
				return nil, err
			}
		}
	}

	job, err := s.taskQueue.Submit(name, arguments)
	if err != nil {
		err = &httpError{404, err}
		return
	}

	if timeout > 0 {
		job, err = waitForJob(s.taskQueue, job.UUID, time.Duration(timeout)*time.Second)
		if err != nil {
			return
		}
	}

	url, err := s.router.Get("job").URL("uuid", job.UUID)
	data = WebJob{job, url.String()}
	return
}

type WebJob struct {
	*Job
	Href string `json:"href"`
}

func (s *server) listJobs(w http.ResponseWriter, r *http.Request) (data interface{}, err error) {
	storedJobs, err := s.taskQueue.jobStore.GetJobs()
	if err != nil {
		return
	}
	jobs := make([]interface{}, 0, len(storedJobs))
	for _, job := range storedJobs {
		url, err := s.router.Get("job").URL("uuid", job.UUID)
		if err != nil {
			return data, err
		}
		jobs = append(jobs, WebJob{job, url.String()})
	}
	data = jobs
	return
}

func (s *server) getJobStatus(w http.ResponseWriter, r *http.Request) (data interface{}, err error) {
	uuid := mux.Vars(r)["uuid"]
	job, err := s.taskQueue.GetJob(uuid)
	if err != nil {
		err = &httpError{404, err}
		return
	}

	url, err := s.router.Get("job").URL("uuid", job.UUID)
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
		var err error
		data, err := fn(w, r)
		if e, ok := err.(*httpError); ok {
			http.Error(w, e.Error(), e.Status)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		renderJSON(w, data)
	}
}

func renderJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func getTimeout(r *http.Request) (timeout int, err error) {
	timeoutParam := r.URL.Query().Get("jobTimeoutSeconds")
	if len(timeoutParam) == 0 {
		return
	}
	timeout, err = strconv.Atoi(timeoutParam)
	return
}

func waitForJob(taskQueue *TaskQueue, uuid string, timeout time.Duration) (job *Job, err error) {
	tick := time.Tick(500 * time.Millisecond)
	stop := time.After(timeout)

	for {
		select {
		case <-stop:
			err = &httpError{504, errors.New("Timed out waiting for job " + job.UUID)}
			return
		case <-tick:
			job, err = taskQueue.GetJob(uuid)
			if err != nil {
				return
			}
			if job.Status == JOB_SUCCESS || job.Status == JOB_FAILURE {
				return
			}
		}
	}
}
