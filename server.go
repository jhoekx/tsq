package tsq

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type server struct {
	taskQueue *TaskQueue
	router    *mux.Router
}

func ServeQueue(baseURL string, q *TaskQueue) http.Handler {
	svr := &server{
		router:    mux.NewRouter().Path(baseURL).Subrouter(),
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

	var arguments interface{}
	if r.Header.Get("Content-Type") == "application/json" && r.ContentLength != 0 {
		err = json.NewDecoder(r.Body).Decode(&arguments)
		if err != nil {
			return
		}
	}

	job, err := s.taskQueue.Submit(name, arguments)
	if err != nil {
		err = &httpError{404, err}
		return
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
	jobs := make([]interface{}, 0, len(s.taskQueue.jobStore.GetJobs()))
	for _, job := range s.taskQueue.jobStore.GetJobs() {
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
