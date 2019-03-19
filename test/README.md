# Scalog Testing

This directory includes a series of end to end tests for ensuring that this Scalog implementation works as expected.

### Motivation

While unit tests that guarantee the correct behavior of each individual layer would have been preferrable, we currently have no infrastructure for spoofing requests and responses from different layers and different replicas. For the time being, we will rely on end to end tests for verifying correctness.

### Running the tests

To run the tests, you should spin up several scalog nodes on your local machine. It's pretty nasty right now, but here's how you do it for now:

1. Run a order layer node with `./scalog order --localRun --port 1337`
2. Run a data layer node with `export NAME=asdf-0-0 && ./scalog data --localRun --port 8080`
3. Run another data layer node with `export NAME=asdf-0-1 && ./scalog data --localRun --port 8081`
4. To run the current test cases, run `go test test/sanity_test.go`