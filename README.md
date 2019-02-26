# scalog

## Setup
The following assumes you're running a Unix based machine (Linux/Mac). These are also written by David in retrospected after he's installed everything so some steps might be missing.

### Installing Go
1. Find the right download for your system [here](https://golang.org/dl/).
2. Set your `$GOPATH` by adding this to your `~/.bashrc`:
    ```console
    export GOPATH="$HOME/go"
    PATH=$PATH:$GOPATH/bin
    ```
    Refresh your environment variables and prepare for installing `dep` (for dependency management) by running:
     ```console
     source ~/.bashrc
     mkdir ~/go/bin
     mkdir ~/go/src
     mkdir ~/go/pkg
     ```
3. Install `dep` by following instructions [here](https://github.com/golang/dep).


### Cloning
Run the following to clone this repository into the right place:
```console
cd ~/go/src
mkdir -p github.com/scalog
cd github.com/scalog
git clone https://github.com/scalog/scalog.git
```

## Updating dependencies
Run `dep ensure` in the main directory. This should update your `Gopkg.lock` and `Gopkg.toml` files.