# scalog

## Setup
The following assumes you're running a Unix based machine (Linux/Mac). These are also written by David in retrospect after he's installed everything so some steps might be missing.

### Installing Go
1. Find the right download for your system [here](https://golang.org/dl/).
2. Set your `$GOPATH` by adding this to your `~/.bashrc` (`~/.bash_profile` on MacOS). Note that this is different from `$GOROOT`, where Go is installed:
    ```sh
    export GOPATH="$HOME/go"
    PATH=$PATH:$GOPATH/bin
    ```
    Refresh your environment variables:
    - Windows, Linux:
      ```sh
       source ~/.bashrc
      ```
    - MacOS:
      ```sh
       source ~/.bash_profile
      ```
    Prepare for installing `dep` (for dependency management) by running:
     ```sh
     mkdir ~/go
     mkdir ~/go/bin
     mkdir ~/go/src
     mkdir ~/go/pkg
     ```
3. Install `dep` by executing:
    ```sh
    curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
    ```
    Based on the instructions [here](https://github.com/golang/dep).


### Cloning
Run the following to clone this repository into the right place:
```sh
cd ~/go/src
mkdir -p github.com/scalog
cd github.com/scalog
git clone https://github.com/scalog/scalog.git
```

## Updating dependencies
Run `dep ensure` in `~/go/src/github.com/scalog/scalog`. This should update your `Gopkg.lock` and `Gopkg.toml` files.

## ProtoBuf
### Installing
In order to convert `.proto` files into `.pb.go` files for serializing & deserializing messages, we need to install ProtoBuf for Golang.
1. Follow instructions for your architecture to install ProtoBuf for C++ [here](https://github.com/protocolbuffers/protobuf/blob/master/src/README.md). Be sure to download the file whose name starts with `protobuf-cpp-`.
2. Install ProtoBuf for Golang by running:
    ```sh
    go get -u github.com/golang/protobuf/protoc-gen-go
    ```
### Generating .pb.go files
Run (in scalog's main directory)
```sh
./genpb.sh
```

## Running Scalog
This project is a work in progress. While the kubernetes operator is being implemented, Scalog can be run using a series of scripts. 
