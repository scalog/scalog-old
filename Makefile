DOCKER_ORDER_IMAGE = scalog-order
DOCKER_DATA_IMAGE = scalog-data

default:
	@echo Scalog script operator

build:
	go build

docker-minikube:
	eval $(minikube docker-env)

build-data: docker-minikube
	docker build . -t "$(DOCKER_DATA_IMAGE)" --build-arg image_type=data

build-order: docker-minikube
	docker build . -t "$(DOCKER_ORDER_IMAGE)" --build-arg image_type=order

deploy: build-data build-order
	kubectl create -f data/k8s/namespace.yaml && \
	kubectl create -f data/k8s/rbac.yaml && \
	kubectl create -f data/k8s/volumes.yaml && \
	kubectl create -f data/k8s/service.yaml && \
	kubectl create -f data/k8s/controller.yaml && \
	kubectl create -f order/k8s/service.yaml && \
	kubectl create -f order/k8s/deployment.yaml

refresh-image:
	kubectl delete -f data/k8s/controller.yaml && \
	kubectl create -f data/k8s/controller.yaml && \
	kubectl delete -f order/k8s/deployment.yaml && \
	kubectl create -f order/k8s/deployment.yaml
