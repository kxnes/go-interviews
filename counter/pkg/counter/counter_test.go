package counter

import (
	"bufio"
	"io"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
)

const (
	testdata    = "../../test/testdata/example.txt"
	testdataLen = 20
	testURL     = "https://golang.org"
	eps         = 1000 // latency in milliseconds (depends from connection)
)

func TestNewCounterPositive(t *testing.T) {
	got := NewCounter("go", 1)
	want := &Counter{
		http:    http.Client{Timeout: timeout * time.Second},
		workers: 1,
		pattern: regexp.MustCompile(`\bgo\b`),
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf(`NewCounter() got = "%v", want = "%v"`, got, want)
	}
}

func TestNewCounterNegative(t *testing.T) {
	cases := []struct {
		name    string
		workers int
	}{
		{
			name:    "zero",
			workers: 0,
		},
		{
			name:    "negative",
			workers: -1,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			defer func() {
				err := recover()
				if err == nil {
					t.Error("NewCounter() not panic")
					return
				}

				want := "non-positive workers count for NewCounter"
				if err.(error).Error() != want {
					t.Errorf("NewCounter() panic = %q, want = %q", err, want)
				}
			}()

			NewCounter("word", c.workers)
		})
	}
}

func TestCounterInputStream(t *testing.T) {
	type args struct {
		r  io.Reader
		in chan string
	}
	cases := []struct {
		name    string
		args    args
		dataLen int
	}{
		{
			name: "buffered in chan",
			args: args{
				in: make(chan string, 4),
			},
			dataLen: testdataLen,
		},
		{
			name: "unbuffered in chan",
			args: args{
				in: make(chan string),
			},
			dataLen: testdataLen,
		},
		{
			name: "empty reader",
			args: args{
				r:  strings.NewReader(""),
				in: make(chan string),
			},
			dataLen: 0,
		},
	}

	counter := Counter{}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.args.r == nil {
				f, err := os.Open(testdata)
				if err != nil {
					t.Fatalf("%q issues = %q", testdata, err)
				}
				c.args.r = bufio.NewReader(f)
			}

			go counter.inputStream(c.args.r, c.args.in)

			var cnt int
			for got := range c.args.in {
				cnt++
				if got != testURL {
					t.Errorf("inputStream() data = %q, want = %q", got, testURL)
					return
				}
			}

			if cnt != c.dataLen {
				t.Errorf(`inputStream() read = "%d", want = "%d"`, cnt, c.dataLen)
			}
		})
	}
}

func TestCounterCountOne(t *testing.T) {
	cases := []struct {
		name    string
		counter *Counter
		url     string
		want    int
		wantErr bool
	}{
		{
			name: "smoke 1",
			counter: &Counter{
				http:    http.Client{Timeout: timeout * time.Second},
				pattern: regexp.MustCompile(`\bgo\b`),
			},
			url:     testURL,
			want:    9,
			wantErr: false,
		},
		{
			name: "error",
			counter: &Counter{
				http:    http.Client{Timeout: 1},
				pattern: regexp.MustCompile(`\bgo\b`),
			},
			url:     testURL,
			want:    0,
			wantErr: true,
		},
		{
			name: "smoke 2",
			counter: &Counter{
				http:    http.Client{Timeout: timeout * time.Second},
				pattern: regexp.MustCompile(`\bmemes\b`),
			},
			url:     testURL,
			want:    0,
			wantErr: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := c.counter.Count(c.url)
			if (err != nil) != c.wantErr {
				t.Errorf(`Count() error = %v, want = "%t"`, err, c.wantErr)
				return
			}
			if got != c.want {
				t.Errorf(`Count() got = "%d", want = "%d"`, got, c.want)
			}
		})
	}
}

// ---- ATTENTION
// This test are flaky because of connection speed and go scheduler.
// But it is demonstrate the main idea of parallel execution Count().
func TestCounterCountAll(t *testing.T) {
	cases := []struct {
		name    string
		counter *Counter
		timing  int64
		wantErr bool
	}{
		{
			name: "5 workers",
			counter: &Counter{
				http:    http.Client{Timeout: timeout * time.Second},
				workers: 5,
				pattern: regexp.MustCompile(`\bgo\b`),
			},
			timing:  1000,
			wantErr: false,
		},
		{
			name: "1 worker",
			counter: &Counter{
				http:    http.Client{Timeout: timeout * time.Second},
				workers: 1,
				pattern: regexp.MustCompile(`\bgo\b`),
			},
			timing:  5000,
			wantErr: false,
		},
		{
			name: "errors",
			counter: &Counter{
				http:    http.Client{Timeout: 1},
				workers: 1,
				pattern: regexp.MustCompile(`\bgo\b`),
			},
			timing:  0,
			wantErr: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f, err := os.Open(testdata)
			if err != nil {
				t.Fatalf("%q issues = %q", testdata, err)
			}

			in := make(chan string, testdataLen)

			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				in <- scanner.Text()
			}
			close(in)

			out := make(chan Amount, testdataLen)

			var want Amount
			if c.wantErr {
				want = Amount{URL: testURL, Val: 0}
			} else {
				want = Amount{URL: testURL, Val: 9, Err: nil}
			}

			go c.counter.countAll(in, out)

			start := time.Now().UnixNano()
			for got := range out {
				if (got.Err != nil) != c.wantErr {
					t.Errorf(`countAll() error = %v, want = "%t"`, err, c.wantErr)
					return
				}

				if got.URL != want.URL || got.Val != want.Val {
					t.Errorf(`countAll() got = "%v", want = "%v"`, got, want)
					return
				}
			}

			delta := (time.Now().UnixNano() - start) / int64(time.Millisecond)
			if !(c.timing-eps < delta && delta < c.timing+eps) {
				t.Errorf("countAll() concurrent = %d, want %d±%d", delta, c.timing, eps)
			}
		})
	}
}

// ---- ATTENTION
// This is the same as testing StreamBuf() because here uses Stream() that
// pass zero (in, out) capacity inside StreamBuf() and only do this.
func TestCounterStreamSmoke(t *testing.T) {
	f, err := os.Open(testdata)
	if err != nil {
		t.Fatalf("%q issues = %q", testdata, err)
	}

	c := NewCounter("go", 5)
	want := Amount{URL: testURL, Val: 9}

	var cnt int
	for got := range c.Stream(f) {
		cnt++
		if !reflect.DeepEqual(got, want) {
			t.Errorf(`Stream() got = "%v", want = "%v"`, got, want)
		}
	}

	if cnt != testdataLen {
		t.Errorf(`Stream() result = "%d", want = "%d"`, cnt, testdataLen)
	}
}
