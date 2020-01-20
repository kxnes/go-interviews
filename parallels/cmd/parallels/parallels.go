package main

import (
	"flag"
	"fmt"
	"github.com/kxnes/go-interviews/parallels/pkg/jobs"
	"os"
)

func promptError(w string, err error) {
	fmt.Printf("Error occurred when %s data: %q\n", w, err)
	os.Exit(1)
}

func promptUsage(arg, val string) {
	if val != "" {
		return
	}

	fmt.Printf("argument %q is not defined\n", "-"+arg)
	flag.Usage()
	os.Exit(1)
}

type zeroDivision int

func (e zeroDivision) Error() string {
	return fmt.Sprintf("division by zero in %d / 0", int(e))
}

var operations = map[string]func(int, int) (int, error){
	"+": func(a, b int) (int, error) { return a + b, nil },
	"-": func(a, b int) (int, error) { return a - b, nil },
	"*": func(a, b int) (int, error) { return a * b, nil },
	"/": func(a, b int) (int, error) {
		if b == 0 {
			return 0, zeroDivision(a)
		}
		return a / b, nil
	},
}

func main() {
	var in, out, op string

	flag.StringVar(&in, "in", "", "input JSON filename with jobs")
	flag.StringVar(&op, "op", "", "operation that will be perform on input data")
	flag.StringVar(&out, "out", "out.json", "output JSON filename for done jobs")
	flag.Parse()

	promptUsage("in", in)
	promptUsage("op", op)

	j, err := jobs.Unmarshal(in)

	if err != nil {
		promptError("reading", err)
	}

	if _, ok := operations[op]; !ok {
		promptError("preparing", fmt.Errorf("undefined operation %s", op))
	}

	j, err = jobs.Process(operations[op], j)
	if err != nil {
		promptError("processing", err)
	}

	err = jobs.Marshal(out, j)
	if err != nil {
		promptError("writing", err)
	}
}
