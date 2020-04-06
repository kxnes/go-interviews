// Package counter contains the instruments to work
// with counting words from URLs or from io.Reader.
package counter

import (
	"bufio"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"sync"
	"time"
)

const (
	delta   = 1
	timeout = 10
)

// semaphore uses as a simple version of Semaphore sync primitive.
type semaphore struct {
	wg    sync.WaitGroup
	queue chan int
}

// add adds `delta` (check `const` block above) to `WaitGroup` and
// puts it to inner semaphore `queue`.
func (s *semaphore) add() {
	s.wg.Add(delta)
	s.queue <- delta
}

// done decrements negative `delta` (check `const` block) to `WaitGroup` and
// release it from inner semaphore `queue`.
func (s *semaphore) done() {
	<-s.queue
	s.wg.Done()
}

// wait blocks until the `WaitGroup` counter is zero and closes inner semaphore `queue`.
func (s *semaphore) wait() {
	s.wg.Wait()
	close(s.queue)
}

type (
	// Amount represents the single result of `Stream()` and `StreamBuf()`.
	Amount struct {
		URL string // checking URL
		Val int    // the amount of all occurrences
		Err error  // possible error on URL checking
	}
	// Counter is the main type that
	Counter struct {
		http    http.Client    // main client to make connections
		workers int            // size of parallel-working goroutines
		pattern *regexp.Regexp // searching pattern (\b + "word" + \b)
	}
)

// inputStream puts the `io.Reader` payload line-by-line to `in` channel for next processing.
func (c *Counter) inputStream(r io.Reader, in chan string) {
	scan := bufio.NewScanner(r)
	for scan.Scan() {
		in <- scan.Text()
	}
	close(in)
}

// countAll processing all inside coming `in` channel strings to `out` channel results.
// Here placed the main logic of parallel searching for `Stream()` and `StreamBuf()`.
func (c *Counter) countAll(in chan string, out chan Amount) {
	sem := semaphore{queue: make(chan int, c.workers)}

	for url := range in {
		sem.add()

		go func(url string) {
			val, err := c.Count(url)
			out <- Amount{URL: url, Val: val, Err: err}

			sem.done()
		}(url)
	}

	sem.wait()
	close(out)
}

// Count counts the single amount of all occurrences `pattern` for given `url`.
func (c *Counter) Count(url string) (int, error) {
	resp, err := c.http.Get(url)
	if err != nil {
		return 0, err
	}

	defer func() { _ = resp.Body.Close() }()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	return len(c.pattern.FindAll(data, -1)), nil
}

// Stream represents the unbuffered version of `StreamBuf()`.
func (c *Counter) Stream(r io.Reader) <-chan Amount {
	return c.StreamBuf(r, 0)
}

// StreamBuf processes incoming data thru `io.Reader` into return `Amount` channel as soon as possible.
// It is depends on `in` and `out` channels capacity `cap` and the num of `Counter` `workers`.
// The returned channel will be closed if all data processed and never stops if some errors occurred.
// For each URL errors will be stored in the single `Amount`.
func (c *Counter) StreamBuf(r io.Reader, cap int) <-chan Amount {
	var (
		in  = make(chan string, cap)
		out = make(chan Amount, cap)
	)

	go c.inputStream(r, in)
	go c.countAll(in, out)

	return out
}

// NewCounter returns a new copy of `Counter`.
func NewCounter(word string, workers int) *Counter {
	if workers <= 0 {
		panic(errors.New("non-positive workers count for NewCounter"))
	}

	return &Counter{
		http:    http.Client{Timeout: timeout * time.Second},
		workers: workers,
		pattern: regexp.MustCompile(`\b` + word + `\b`),
	}
}
