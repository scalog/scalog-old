#!/bin/bash
echo "Downloading Scalog packages..."
mkdir operator
mkdir operator/crds

wget -O operator/crds/scalog_v1alpha1_scalogservice_cr.yaml https://raw.githubusercontent.com/scalog/scalog-operator/master/deploy/crds/scalog_v1alpha1_scalogservice_cr.yaml
wget -O operator/crds/scalog_v1alpha1_scalogservice_crd.yaml https://raw.githubusercontent.com/scalog/scalog-operator/master/deploy/crds/scalog_v1alpha1_scalogservice_crd.yaml
wget -O operator/operator.yaml https://raw.githubusercontent.com/scalog/scalog-operator/master/deploy/operator.yaml
wget -O operator/op-rbac.yaml https://raw.githubusercontent.com/scalog/scalog-operator/master/deploy/rbac.yaml
wget -O operator/op-role.yaml https://raw.githubusercontent.com/scalog/scalog-operator/master/deploy/role.yaml
wget -O operator/op-rb.yaml https://raw.githubusercontent.com/scalog/scalog-operator/master/deploy/role_binding.yaml
wget -O operator/op-sa.yaml https://raw.githubusercontent.com/scalog/scalog-operator/master/deploy/service_account.yaml

echo "Successfully downloaded Scalog packages"

kubectl create -f ../k8s/namespace.yaml
kubectl create -f ../k8s/volumes.yaml
kubectl create -f ../k8s/rbac.yaml

kubectl create -f operator/op-sa.yaml
kubectl create -f operator/op-role.yaml
kubectl create -f operator/op-rb.yaml
kubectl create -f operator/op-rbac.yaml
kubectl create -f operator/crds/scalog_v1alpha1_scalogservice_crd.yaml
kubectl create -f operator/operator.yaml
kubectl create -f operator/crds/scalog_v1alpha1_scalogservice_cr.yaml

echo "Scalog is now running :^)"