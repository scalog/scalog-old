#!/bin/bash
echo "Bootstrapping kubernetes cluster"
kubeadm init --pod-network-cidr=10.244.0.0/16

# Create Pod Network Layer
kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/a70459be0084506e4ec919aa1c114638878db11b/Documentation/kube-flannel.yml

# Change permissions, so kubectl can run as a normal user
sudo cp /etc/kubernetes/admin.conf $HOME/
sudo chown $(id -u):$(id -g) $HOME/admin.conf
export KUBECONFIG=$HOME/admin.conf
