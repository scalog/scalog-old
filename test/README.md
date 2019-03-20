# Scalog Testing

This directory includes a series of end to end tests for ensuring that this Scalog implementation works as expected.

### Motivation

While unit tests that guarantee the correct behavior of each individual layer would have been preferrable, we currently have no infrastructure for spoofing requests and responses from different layers and different replicas. For the time being, we will rely on end to end tests for verifying correctness.

### Running the tests

Change your directory to the test directory. To run a specific test, just use `go test TEST_NAME`. If golang is cacheing the results, just run `go test TEST_NAME -count=1` to force the test rerun.