// +build

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/kxnes/go-interviews/counter/pkg/counter"
)

func main() {
	k := flag.Int("k", 5, "the number of goroutines processing urls")
	q := flag.String("q", "go", "the search word")
	flag.Parse()

	var r io.Reader

	stat, err := os.Stdin.Stat()
	if err != nil {
		log.Fatalln("stdin get stat error", err)
	}

	switch {
	case stat.Mode()&os.ModeNamedPipe != 0:
		r = os.Stdin
	case flag.NArg() != 0:
		r = strings.NewReader(strings.Join(flag.Args(), "\n"))
	default:
		log.Fatalln("reader is not defined")
	}

	c := counter.NewCounter(*q, *k)

	for res := range c.Stream(r) {
		if res.Err != nil {
			fmt.Printf("Error for %s: %v\n", res.URL, res.Err)
		} else {
			fmt.Printf("Count for %s: %d\n", res.URL, res.Val)
		}
	}
}
