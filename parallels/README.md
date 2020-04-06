Parallels
=========

> Taken from: somewhere from the internet.

Description
-----------

Write command-line program that takes JSON (any random content) content and execute it in parallel.

Input JSON file is a file with test data in which format errors were deliberately made. 
Your program must perform operation on this partially valid data with any of `+`, `-`, `*`, `/`. 
And must doing it in parallel (concurrency).

Solution
--------

Command-line arguments:

 - `-in` (type `string`): input JSON filename with jobs
 - `-op` (type `string`): operation that will be perform on input data
 - `-out` (type `string`): output JSON filename for done jobs (default `out.json`)

Taste it!

Example
-------

```bash
$ go run cmd/parallels/parallels.go -in test/testdata/valid.json -op +
# result (in out.json by default)
$ head -n 5 out.json
# [
#   {
#     "Job": 5,
#     "Output": 8
#   },
# ...
```
