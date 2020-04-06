// Package jobs contains the instruments to work with parallel execution JSON jobs.
package jobs

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"
)

// UndefinedOperationError - undefined operation error.
type UndefinedOperationError struct{}

func (e *UndefinedOperationError) Error() string {
	return "undefined operation"
}

// Job represents the job entity.
type Job struct {
	// The `ID` and `Output` are export fields for output and interaction purpose.
	ID     int `json:"Job"`
	Output int
	// The `a` and `b` are unexported because `Output` must depend only of the result `Job.process`.
	// They properly can be set only by `Unmarshal` function.
	a, b int
	err  error
}

// process does the main processing for `Job`. Applying `fn` to `a` and `b`.
func (j *Job) process(fn func(int, int) (int, error)) {
	// emulate long time execution
	time.Sleep(time.Second)
	// end emulate long time execution

	var err error

	j.Output, err = fn(j.a, j.b)
	if err != nil {
		j.err = err
	}
}

// Unmarshal opens file by `path` and parses the JSON-encoded data.
// Result will be (parsed array of jobs, error).
//
// Ever JSON job must be synchronized with `Job` type.
// Unmarshal will collect only jobs with properly type (`Job`).
func Unmarshal(path string) ([]Job, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	//noinspection GoUnhandledErrorResult
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	jsonData := make([]map[string]interface{}, 0)

	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		return nil, err
	}

	var (
		id   int
		jobs = make([]Job, 0)
	)

	for _, d := range jsonData {
		arg1, ok1 := d["arg1"]
		arg2, ok2 := d["arg2"]

		if !ok1 || !ok2 {
			continue // Not enough keys.
		}

		arg1Val, ok1 := arg1.(float64)
		arg2Val, ok2 := arg2.(float64)

		if !ok1 || !ok2 {
			continue // Invalid type of keys.
		}

		id++

		jobs = append(jobs, Job{a: int(arg1Val), b: int(arg2Val), ID: id})
	}

	return jobs, nil
}

// Marshal stores the JSON encoding of `jobs` into the output file `path`.
//
// Under hood uses `json.MarshalIndent` for pretty print.
func Marshal(path string, jobs []Job) error {
	data, err := json.MarshalIndent(jobs, "", "  ")
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	_, err = f.Write(data)
	if err != nil {
		return err
	}

	return nil
}

// Process run concurrently all incoming `Job`.
func Process(fn func(int, int) (int, error), jobs []Job) ([]Job, error) {
	if fn == nil {
		return nil, &UndefinedOperationError{}
	}

	processed := make(chan Job)

	for _, j := range jobs {
		go func(j Job) {
			j.process(fn)
			processed <- j
		}(j)
	}

	valid := make([]Job, 0)

	for range jobs {
		j := <-processed
		if j.err != nil {
			continue
		}

		valid = append(valid, j)
	}

	return valid, nil
}

// Compose aggregates Unmarshal, Process and Marshall into one function.
func Compose(in, out string, fn func(int, int) (int, error)) error {
	j, err := Unmarshal(in)
	if err != nil {
		return err
	}

	j, err = Process(fn, j)
	if err != nil {
		return err
	}

	err = Marshal(out, j)
	if err != nil {
		return err
	}

	return nil
}
