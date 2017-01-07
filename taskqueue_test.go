package tsq

import (
	"errors"
	"testing"
	"time"
)

type TestTask struct{}

type TestRun struct {
	started    chan bool
	forward    chan bool
	finished   chan bool
	shouldFail bool
	shouldWait bool
}

func NewTestRun() (run *TestRun) {
	run = &TestRun{
		started:    make(chan bool, 1),
		forward:    make(chan bool, 1),
		finished:   make(chan bool, 1),
		shouldFail: false,
		shouldWait: false,
	}
	return
}

func (tsk *TestTask) Run(args interface{}) (data interface{}, err error) {
	run := args.(*TestRun)
	run.started <- true
	if run.shouldWait {
		<-run.forward
	}
	if run.shouldFail {
		run.finished <- true
		err = errors.New("ERROR")
		return
	}
	run.finished <- true
	data = "DATA"
	return
}

func (job *TestRun) WaitForStart(t *testing.T) {
	select {
	case <-job.started:
	case <-time.After(1 * time.Second):
		t.Error("wait for start timeout")
	}
}

func (job *TestRun) WaitForFinish(t *testing.T) {
	select {
	case <-job.finished:
	case <-time.After(1 * time.Second):
		t.Error("wait for finish timeout")
	}
}

func NewTestQueue() (tsq *TaskQueue) {
	tsq = New()
	tsk := &TestTask{}
	tsq.Define("test", tsk)
	tsq.Start()
	return
}

func TestUndefinedTask(t *testing.T) {
	tsq := NewTestQueue()
	_, err := tsq.Submit("notest", make([]interface{}, 1))
	if err.Error() != "Unknown task: notest" {
		t.Fail()
	}
}

func TestSuccess(t *testing.T) {
	tsq := NewTestQueue()
	run := NewTestRun()
	job, _ := tsq.Submit("test", run)
	run.WaitForFinish(t)
	if job.Status != JOB_SUCCESS {
		t.Fail()
	}
}

func TestFailure(t *testing.T) {
	tsq := NewTestQueue()
	run := NewTestRun()
	run.shouldFail = true
	job, _ := tsq.Submit("test", run)
	run.WaitForFinish(t)
	if job.Status != JOB_FAILURE {
		t.Fail()
	}
}

func TestRunning(t *testing.T) {
	tsq := NewTestQueue()
	run := NewTestRun()
	run.shouldWait = true
	job, _ := tsq.Submit("test", run)
	run.WaitForStart(t)
	if job.Status != JOB_RUNNING {
		t.Error("failed")
	}
}

func TestReturnValue(t *testing.T) {
	tsq := NewTestQueue()
	run := NewTestRun()
	job, _ := tsq.Submit("test", run)
	run.WaitForFinish(t)
	if job.Result.(string) != "DATA" {
		t.Fail()
	}
}

func TestJobsAreSerialized(t *testing.T) {
	tsq := NewTestQueue()
	run1 := NewTestRun()
	run1.shouldWait = true
	run2 := NewTestRun()
	tsq.Submit("test", run1)
	job2, _ := tsq.Submit("test", run2)
	run1.WaitForStart(t)
	if job2.Status != JOB_PENDING {
		t.Fail()
	}
	run1.forward <- true
	run2.WaitForFinish(t)
}

func TestJobsAreStored(t *testing.T) {
	tsq := NewTestQueue()
	run := NewTestRun()
	job, _ := tsq.Submit("test", run)
	run.WaitForFinish(t)
	res, _ := tsq.GetJob(job.UUID)
	if res != job {
		t.Fail()
	}
}

func TestGetUnknownJob(t *testing.T) {
	tsq := NewTestQueue()
	run := NewTestRun()
	tsq.Submit("test", run)
	run.WaitForFinish(t)
	_, err := tsq.GetJob("foo")
	if err.Error() != "Job foo not found" {
		t.Fail()
	}
}
