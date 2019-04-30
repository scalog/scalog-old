#!/bin/bash
echo "Downloading Scalog packages..."
mkdir scalog
mkdir scalog/crds

wget -O scalog/namespace.yaml https://raw.githubusercontent.com/scalog/scalog/master/data/k8s/namespace.yaml
wget -O scalog/volumes.yaml https://raw.githubusercontent.com/scalog/scalog/master/data/k8s/volumes.yaml
wget -O scalog/rbac.yaml https://raw.githubusercontent.com/scalog/scalog/master/data/k8s/rbac.yaml

wget -O scalog/crds/scalog_v1alpha1_scalogservice_cr.yaml https://raw.githubusercontent.com/scalog/scalog-operator/master/deploy/crds/scalog_v1alpha1_scalogservice_cr.yaml
wget -O scalog/crds/scalog_v1alpha1_scalogservice_crd.yaml https://raw.githubusercontent.com/scalog/scalog-operator/master/deploy/crds/scalog_v1alpha1_scalogservice_crd.yaml
wget -O scalog/operator.yaml https://raw.githubusercontent.com/scalog/scalog-operator/master/deploy/operator.yaml
echo "Successfully downloaded Scalog packages"

kubectl create -f scalog/namespace.yaml
kubectl create -f scalog/volumes.yaml
kubectl create -f scalog/rbac.yaml

kubectl create -f scalog/crds/scalog_v1alpha1_scalogservice_crd.yaml
kubectl create -f scalog/operator.yaml
kubectl create -f scalog/crds/scalog_v1alpha1_scalogservice_cr.yaml

echo "Scalog is now running :^)"