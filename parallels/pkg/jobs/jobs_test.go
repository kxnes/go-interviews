package jobs

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

var (
	err      = errors.New("something went wrong")
	division = func(a, b int) (int, error) {
		if b == 0 {
			return 0, err
		}
		return a / b, nil
	}
)

func TestUndefinedOperationError(t *testing.T) {
	var err error = &UndefinedOperationError{}

	got := err.Error()
	want := "undefined operation"

	if got != want {
		t.Errorf("invalid message: got = %q, want = %q", got, want)
	}
}

func TestJobProcess(t *testing.T) {
	type want struct {
		Output int
		err    error
	}
	cases := []struct {
		name string
		job  Job
		want want
		fn   func(int, int) (int, error)
	}{
		{"zero division (check error)", Job{a: 12, b: 0}, want{err: err}, division},
		{"zero division (valid calc)", Job{a: 0, b: 3}, want{}, division},
		{
			"any other function", // test error and result do not depend
			Job{a: 12, b: 3},
			want{Output: 33, err: err},
			func(a, b int) (int, error) {
				return a + b*b + a, err
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			j := &Job{a: c.job.a, b: c.job.b, err: c.job.err}

			j.process(c.fn)

			if j.Output != c.want.Output {
				t.Errorf("invalid result: got = %v, want = %v", j.Output, c.want.Output)
			}

			if j.err != c.want.err {
				t.Errorf("invalid error: got = %v, want = %v", j.err, c.want.err)
			}
		})
	}
}

const testdata = "../../test/testdata"

func TestUnmarshal(t *testing.T) {
	type want struct {
		jobs []Job
		err  error
	}
	cases := []struct {
		name string
		path string
		want want
	}{
		{"valid JSON", filepath.Join(testdata, "valid.json"), want{jobs: []Job{
			{ID: 1, a: 16, b: 4,},
			{ID: 2, a: 128, b: 16,},
			{ID: 3, a: 8, b: 9,},
			{ID: 4, a: 7, b: 3,},
			{ID: 5, a: 8, b: 0,},
			{ID: 6, a: 7, b: 3,},
		}}},
		{"invalid JSON", filepath.Join(testdata, "invalid.json"), want{err: &json.SyntaxError{}}},
		{"file not exist", filepath.Join(testdata, "notexist.json"), want{err: &os.PathError{}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Unmarshal(c.path)

			if !reflect.DeepEqual(got, c.want.jobs) {
				t.Errorf("invalid result: got = %v, want = %v", got, c.want.jobs)
			}

			if c.want.err == nil && err != nil || c.want.err != nil && !errors.As(err, &c.want.err) {
				t.Errorf("invalid error: got = %v, want = %v", err, c.want.err)
			}
		})
	}
}

const testResult = "../../test/testresult"

func TestMarshal(t *testing.T) {
	if _, err := os.Stat(testResult); os.IsNotExist(err) {
		err = os.MkdirAll(testResult, 0755)
		if err != nil {
			panic(err)
		}
	}

	cases := []struct {
		name string
		path string
		jobs []Job
		err  error
	}{
		{"1", filepath.Join(testResult, ""), nil, &os.PathError{}},
		{"1", filepath.Join(testResult, "empty.json"), []Job{}, nil},
		{"1", filepath.Join(testResult, "valid.json"), []Job{{ID: 1, Output: 5}, {ID: 2, Output: 42}}, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := Marshal(c.path, c.jobs)

			if c.err == nil && err != nil || c.err != nil && !errors.As(err, &c.err) {
				t.Errorf("invalid error: got = %v, want = %v", err, c.err)
			}

			var got []Job

			f, _ := os.Open(c.path)
			b, _ := ioutil.ReadAll(f)
			_ = json.Unmarshal(b, &got)

			if !reflect.DeepEqual(got, c.jobs) {
				t.Errorf("invalid result: got = %v, want = %v", got, c.jobs)
			}
		})
	}
}

func TestProcessErr(t *testing.T) {
	type want struct {
		jobs []Job
		err  error
	}
	cases := []struct {
		name string
		jobs []Job
		want want
		op   func(int, int) (int, error)
	}{
		{
			"operation not exist",
			[]Job{{a: 12, b: 3}},
			want{err: &UndefinedOperationError{}},
			nil,
		},
		{
			"processed with errors", // test skip failed jobs in result
			[]Job{
				{a: 12, b: 0},
				{a: 4, b: 2},
			},
			want{jobs: []Job{{Output: 2, a: 4, b: 2}}},
			division,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Process(c.op, c.jobs)

			if !reflect.DeepEqual(got, c.want.jobs) {
				t.Errorf("invalid result: got = %v, want = %v", got, c.want.jobs)
			}

			if err != c.want.err {
				t.Errorf("invalid error: got = %v, want = %v", err, c.want.err)
			}
		})
	}
}

// Separate test because of concurrent execution. We don't know which job will execute faster.
func TestProcess(t *testing.T) {
	cases := []struct {
		name string
		want []Job
		op   func(int, int) (int, error)
	}{
		{
			"addition",
			[]Job{
				{ID: 1, Output: 15, a: 12, b: 3},
				{ID: 2, Output: 26, a: 23, b: 3},
				{ID: 3, Output: 7, a: 10, b: -3},
			},
			func(a, b int) (int, error) { return a + b, nil },
		},
		{
			"subtraction",
			[]Job{
				{ID: 1, Output: 9, a: 12, b: 3},
				{ID: 2, Output: 20, a: 23, b: 3},
				{ID: 3, Output: 13, a: 10, b: -3},
			},
			func(a, b int) (int, error) { return a - b, nil },
		},
		{
			"multiplication",
			[]Job{
				{ID: 1, Output: 36, a: 12, b: 3},
				{ID: 2, Output: 69, a: 23, b: 3},
				{ID: 3, Output: -30, a: 10, b: -3},
			},
			func(a, b int) (int, error) { return a * b, nil },
		},
		{
			"division",
			[]Job{
				{ID: 1, Output: 4, a: 12, b: 3},
				{ID: 2, Output: 7, a: 23, b: 3},
				{ID: 3, Output: -3, a: 10, b: -3},
			},
			division,
		},
	}

	jobs := []Job{
		{ID: 1, a: 12, b: 3},
		{ID: 2, a: 23, b: 3},
		{ID: 3, a: 10, b: -3},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			start := time.Now()

			got, err := Process(c.op, jobs)

			// Adding small delay for testing because anything can happen with Go scheduler or CPU loading.
			if time.Since(start) > time.Second+500*time.Millisecond {
				t.Errorf("process is not concurrent")
			}

			if err != nil {
				panic(err) // unexpected behavior, there should be no errors
			}

			// Because of concurrent. Process returns another order of done jobs.
			for _, j := range got {
				var want Job

				switch j.ID {
				case 1:
					want = c.want[0]
				case 2:
					want = c.want[1]
				case 3:
					want = c.want[2]
				}

				if j.Output != want.Output || j.a != want.a || j.b != want.b {
					t.Errorf("invalid result for id = %d: got = %v, want = %v", j.ID, j, c.want)
				}
			}
		})
	}
}

func TestComposeSmoke(t *testing.T) {
	in := filepath.Join(testdata, "valid.json")
	out := filepath.Join(testResult, "smoke.json")
	op := func(a, b int) (int, error) { return a + b, nil }

	err := Compose(in, out, op)
	if err != nil {
		panic(err) // unexpected behavior, there should be no errors
	}

	want := []Job{
		{ID: 1, Output: 20},
		{ID: 2, Output: 144},
		{ID: 3, Output: 17},
		{ID: 4, Output: 10},
		{ID: 5, Output: 8},
		{ID: 6, Output: 10},
	}

	var got []Job
	f, _ := os.Open(out)
	b, _ := ioutil.ReadAll(f)
	_ = json.Unmarshal(b, &got)

	for _, w := range want {
		var valid bool

		for _, g := range got {
			if w == g {
				valid = true
				break
			}
		}

		if !valid {
			t.Errorf("invalid result: compare content of %q and\n%v", out, want)
		}
	}
}
