#!/bin/bash

echo "Bootstrapping kubernetes cluster..."

# Initiate kubernetes cluster
sudo kubeadm init --pod-network-cidr=10.244.0.0/16

# Change permissions, so kubectl can run as a normal user
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config

# Create Pod Network Layer
kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/a70459be0084506e4ec919aa1c114638878db11b/Documentation/kube-flannel.yml
