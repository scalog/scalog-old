# Scalog Client

A Go Client library for the [Scalog](https://github.com/scalog/scalog) Project.

[![Build Status](https://travis-ci.org/scalog/scalog-client.svg?branch=master)](https://travis-ci.org/scalog/scalog-client)

## Setup

The following assumes [Go](https://golang.org/) and [dep](https://github.com/golang/dep) are already installed and configured on your machine.

Navigate to the project root directory.

Rename the file `config.yaml.template` to `config.yaml`.

In `config.yaml`, set the IP and port on which the Scalog discovery service is running.

```
discovery-address:
  ip:   "127.0.0.1" // Set the IP
  port: 8000        // Set the port
```

Run the below command in the root directory to download the dependencies and build the project.

```
dep ensure
go build
```

## Command Line Interface

Start the command line interface by running the below command in the root directory. Once started, run the `help` command to see the available commands.

```
./scalog-client it
help
```

## Benchmarking

Run the below command in the root directory to perform benchmark tests for a single client. Optionally, specify additional flags `--num` and `--size` for the number of append operations and the size of each append operation, respectively.

```
./scalog-client bench
```

## Testing

Run the below command in the root directory to perform end-to-end testing.

```
./scalog-client test
```

## Example Usage

```go
import "github.com/scalog/scalog-client/client"

func main() {
  client := client.NewClient()
  resp, err := client.Append("Hello, World!")
  if err != nil {
      fmt.Println(err.Error())
  }
  fmt.Println(fmt.Sprintf("Global sequence number %d assigned to record", resp))
}
```
