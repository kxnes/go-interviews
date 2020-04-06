URL Counter
===========

> Taken from: Mail.ru Group.

Description
-----------

Write command-line program that counts some word from the bunch of URLs.

The process receives strings containing the URLs. Each URL needs to be pulled and counted 
the number of occurrences of the string `q` in the response. 
At the end of the work, the application displays the total number of `q` lines found in all URLs.

Constraints
-----------

 - The entered URL should begin to be processed immediately and in parallel with the next.
 URLs should be processed in parallel, but no more than `k` at a time. 
 URL handlers should not generate extra goroutines, i.e. if `k` = 1000 and no URLs are being processed, 
 1000 goroutines should not be created.

 - Need to do without global variables and use only standard libraries.

Solution
--------

Command-line arguments:

 - `-k` (type `int`): the number of goroutines processing urls (default `5`)
 - `-q` (type `string`): the search word (default `go`)

Taste it!

Example
-------

As a PIPE:

```bash
$ cat test/testdata/example.txt | go run cmd/counter/counter.go -k 3
# Count for https://golang.org: 9
# Count for https://golang.org: 9
# ...
```

As the program arguments:

```bash
$ go run cmd/counter/counter.go -q for https://golang.com https://google.com
# Count for https://google.com: 1
# Count for https://golang.com: 3
```
