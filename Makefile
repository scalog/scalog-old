DOCKER_IMAGE = scalog-data

build:
	go build

build-docker:
	eval $(minikube docker-env) && docker build . -t "$(DOCKER_IMAGE)"

deploy:
	kubectl create -f data/k8s/namespace.yaml &&
	kubectl create -f data/k8s/rbac.yaml &&
	kubectl create -f data/k8s/volumes.yaml &&
	kubectl create -f data/k8s/service.yaml &&
	kubectl create -f data/k8s/controller.yaml &&

refresh-image:
	kubectl delete -f data/k8s/controller.yaml &&
	kubectl create -f data/k8s/controller.yaml
