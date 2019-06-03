#!/bin/bash

RED="\033[0;31m"
GREEN="\033[1;32m"
YELLOW="\033[1;33m"
NC="\033[0m"

echo -e "Scalog Install\n"
echo "This script is meant to bootstrap a scalog cluster quickly on a cluster of machines. Within this cluster, we expect one machine to be used to bootstrap the kubernetes cluster, and all other machines should simply join this cluster."

read -p $'\e[33mIs this the primary machine? (yes/no):\e[0m' isPrimary
while ! [[ "$isPrimary" =~ ^(yes|no)$ ]]
do
  read -p $'\e[33mIs this the primary machine? (yes/no):\e[0m' isPrimary
  echo -e "${RED}Enter a valid string${NC}"
done

echo "Installing docker on this local machine"

# Install Docker CE
## Set up the repository:
### Install packages to allow apt to use a repository over HTTPS
sudo apt-get update > /dev/null 2>&1  && sudo apt-get install -y apt-transport-https ca-certificates curl software-properties-common > /dev/null 2>&1
if [ $? -ne 0 ]
then
  echo -e "${RED}Failed to install apt-transport-https ca-certificates curl software-properties-common${NC}"
  exit 1
fi

### Add Docker’s official GPG key
(curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -) > /dev/null 2>&1
if [ $? -ne 0 ]
then
  echo -e "${RED}Failed to add Docker's GPG key${NC}"
  exit 1
fi

### Add Docker apt repository.
sudo add-apt-repository \
  "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) \
  stable"
if [ $? -ne 0 ]
then
  echo -e "${RED}Failed to add Docker's apt repository${NC}"
  exit 1
fi

## Install Docker CE.
sudo apt-get update > /dev/null 2>&1 && sudo apt-get install -y docker-ce=18.06.2~ce~3-0~ubuntu > /dev/null 2>&1
if [ $? -ne 0 ]
then
  echo -e "${RED}Failed to install Docker CE${NC}"
  exit 1
fi

echo -e "${GREEN}Successfully installed Docker${NC}"
echo "Installing kubeadm on this local machine..."

sudo apt-get update > /dev/null 2>&1 && sudo apt-get install -y apt-transport-https curl > /dev/null 2>&1
if [ $? -ne 0 ]
then
  echo -e "${RED}Failed to install apt-transport-https curl${NC}"
  exit 1
fi

(curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -) > /dev/null 2>&1
if [ $? -ne 0 ]
then
  echo -e "${RED}Failed to add kubeadm apt-key${NC}"
  exit 1
fi

(echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee -a /etc/apt/sources.list.d/kubernetes.list) > /dev/null 2>&1
if [ $? -ne 0 ]
then
  echo -e "${RED}Failed to set kubenretes repository${NC}"
  exit 1
fi

sudo apt-get update > /dev/null 2>&1 && sudo apt-get install -y kubelet kubeadm kubectl > /dev/null 2>&1 && sudo apt-mark hold kubelet kubeadm kubectl > /dev/null 2>&1
if [ $? -ne 0 ]
then
  echo -e "${RED}Failed to install kubeadm${NC}"
  exit 1
fi

modprobe br_netfilter
if [ $? -ne 0 ]
then
  echo -e "${RED}Failed modprobe br_netfilter${NC}"
  exit 1
fi

sudo swapoff -a
if [ $? -ne 0 ]
then
  echo -e "${RED}Failed to turn swap off${NC}"
  exit 1
fi

echo -e "${GREEN}Successfully installed kubeadm!!!${NC}\n"

if [ $isPrimary == "yes" ]
then
  echo "Bootstrapping kubernetes cluster..."

  kubeout=$(sudo kubeadm init --pod-network-cidr=10.244.0.0/16)
  if [ $? -ne 0 ]
  then
    echo -e "${RED}Failed to bootstrap kubernetes cluster${NC}"
    exit 1
  fi

  mkdir -p $HOME/.kube && sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config && sudo chown $(id -u):$(id -g) $HOME/.kube/config
  if [ $? -ne 0 ]
  then
    echo -e "${RED}Failed to export KUBECONFIG${NC}"
    exit 1
  fi

  # Create Pod Network Layer
  kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/a70459be0084506e4ec919aa1c114638878db11b/Documentation/kube-flannel.yml > /dev/null 2>&1
  if [ $? -ne 0 ]
  then
    echo -e "${RED}Failed to create pod network layer for kubernetes${NC}"
    exit 1
  fi

  echo -e $kubeout
  echo -e "${GREEN}Kubernetes cluster successfully bootstrapped!${NC}"
  echo -e "${YELLOW}Copy paste the kubeadm script output into the other servers in this cluster. It should look something like kubeadm join 130.127.133.78:6443 --token nrorf3.vdp46o3fa75ykvch \ --discovery-token-ca-cert-hash sha256:aeab241dce8a6d3eb9e0b6dc98aa39342357f9b3f204d281bb87e526afb4691a"

  read -p $'\e[33mPress enter when the other nodes have joined the cluster:\e[0m'

  echo "Starting Scalog..."
  echo -e "${YELLOW}Downloading Scalog packages...${NC}"
  mkdir operator
  mkdir operator/crds

  wget -O operator/crds/scalog_v1alpha1_scalogservice_cr.yaml https://raw.githubusercontent.com/scalog/scalog-operator/master/deploy/crds/scalog_v1alpha1_scalogservice_cr.yaml
  wget -O operator/crds/scalog_v1alpha1_scalogservice_crd.yaml https://raw.githubusercontent.com/scalog/scalog-operator/master/deploy/crds/scalog_v1alpha1_scalogservice_crd.yaml
  wget -O operator/operator.yaml https://raw.githubusercontent.com/scalog/scalog-operator/master/deploy/operator.yaml
  wget -O operator/op-rbac.yaml https://raw.githubusercontent.com/scalog/scalog-operator/master/deploy/rbac.yaml
  wget -O operator/op-role.yaml https://raw.githubusercontent.com/scalog/scalog-operator/master/deploy/role.yaml
  wget -O operator/op-rb.yaml https://raw.githubusercontent.com/scalog/scalog-operator/master/deploy/role_binding.yaml
  wget -O operator/op-sa.yaml https://raw.githubusercontent.com/scalog/scalog-operator/master/deploy/service_account.yaml

  echo -e "${GREEN}Successfully downloaded Scalog packages${NC}"
  echo "Starting Scalog Services..."

  kubectl create -f k8s/namespace.yaml
  kubectl create -f k8s/volumes.yaml
  kubectl create -f k8s/rbac.yaml

  kubectl create -f operator/op-sa.yaml
  kubectl create -f operator/op-role.yaml
  kubectl create -f operator/op-rb.yaml
  kubectl create -f operator/op-rbac.yaml
  kubectl create -f operator/crds/scalog_v1alpha1_scalogservice_crd.yaml
  kubectl create -f operator/operator.yaml
  kubectl create -f operator/crds/scalog_v1alpha1_scalogservice_cr.yaml

else
  # Follower
  echo -e "${YELLOW}Copy and paste the kubeadm join script received from the primary machine. It should look something like: kubeadm join 130.127.133.78:6443 --token nrorf3.vdp46o3fa75ykvch --discovery-token-ca-cert-hash sha256:aeab241dce8a6d3eb9e0b6dc98aa39342357f9b3f204d281bb87e526afb4691a. It helps if you put sudo before it.${NC}"
fi

echo -e "${GREEN}Success!${NC}"
