# Scalog: A Scalable, Totally Ordered Shared Log

[![Build Status](https://travis-ci.org/scalog/scalog.svg?branch=master)](https://travis-ci.org/scalog/scalog)

Scalog is an ongoing research venture, striving to create a high throughput and easily reconfigurable totally ordered shared log.

## Quickstart

The easiest way to bootstrap a bare metal computing cluster with scalog is:

1. Clone this repository on the desired machines
2. cd into this repository and then run `cd deploy`
3. Run the bootstrapping scripts by executing `chmod +x bootstrap-scalog.sh && ./bootstrap-scalog` or `bash bootstrap-scalog.sh`

After following the prompts from the scripts, the bootstrapping script should install `kubeadm` and `docker`,
bootstrap a kubernetes cluster, and then start a small sample instance of Scalog on that cluster.

If for whatever reason you would like to see the outputs of each command executed in the script, run `debug-bootstrap-scalog.sh` instead.

### Kubernetes Dashboard

In most cases, it is convinient to utilize Kubernetes's built in dashboard to view the status of the Scalog cluster, as opposed to using
the terminal `kubectl` interface. To activate the kubernetes dashboard, run:

```
kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v1.10.1/src/deploy/recommended/kubernetes-dashboard.yaml
```

This instruction loads the image for running the kubernetes dashboard. At this point, the dashboard is only accessible
internally. We need to expose the cluster to external traffic. Run `kubectl -n kube-system edit service kubernetes-dashboard`
to edit the service. Change `type: ClusterIP` to `type: NodePort` and save the file. This should update the existing service
to route traffic from an external port to the dashboard service.

You can find this port number by running `kubectl -n kube-system get service kubernetes-dashboard`. Access to the kubernetes dashboard
is therefore done by using the IP address base (found from `kubectl cluster-info`) and using the port given by running the aforementioned
command. Historically, if using Google Chrome, a popup window will appear describing an unsafe connection -- just bypass it. If
nothing appears by querying the IP address and proper port, try using https.

To access the dashboard, Kubernetes requires login credentials. As a quick and dirty way to generate an admin token, we simple create
a new user. Make a new file (and name it, for example, `admin.yaml`), and populate it with the following:

```
apiVersion: v1
kind: ServiceAccount
metadata:
  name: admin-user
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: admin-user
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: admin-user
  namespace: kube-system
```

Save and exit the file. Then run `kubectl create -f admin.yaml` or whatever you have named the file. This should create a user account
with administrative priveledges to most portions of the cluster. With this, we can generate an access token for the dashboard. Get
the token by running `kubectl -n kube-system describe secret $(kubectl -n kube-system get secret | grep admin-user | awk '{print $1}')`.

## Developing on Scalog
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
