#!/bin/bash
echo "Bootstrapping kubernetes cluster"
modprobe br_netfilter
sudo swapoff -a 

kubeadm init --pod-network-cidr=10.244.0.0/16

# Create Pod Network Layer
kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/a70459be0084506e4ec919aa1c114638878db11b/Documentation/kube-flannel.yml