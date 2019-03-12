DOCKER_ORDER_IMAGE = scalog-order
DOCKER_DATA_IMAGE = scalog-data

default:
	@echo Scalog script operator. Make sure that you run eval \$\(minikube docker-env\) prior to running any of these commands.

dep:
	dep ensure -v

build: dep
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o scalog .

docker-data: build
	docker build . -f Dockerfile.data -t "$(DOCKER_DATA_IMAGE)"

docker-order: build
	docker build . -f Dockerfile.order -t "$(DOCKER_ORDER_IMAGE)"

docker-build: build
	docker build . -f Dockerfile.data -t "$(DOCKER_DATA_IMAGE)"
	docker build . -f Dockerfile.order -t "$(DOCKER_ORDER_IMAGE)"

deploy: docker-build
	kubectl create -f data/k8s/namespace.yaml && \
	kubectl create -f data/k8s/rbac.yaml && \
	kubectl create -f data/k8s/volumes.yaml && \
	kubectl create -f data/k8s/service.yaml && \
	kubectl create -f data/k8s/controller.yaml && \
	kubectl create -f order/k8s/service.yaml && \
	kubectl create -f order/k8s/deployment.yaml

refresh-image: refresh-data
	kubectl delete -f order/k8s/deployment.yaml && \
	kubectl create -f order/k8s/deployment.yaml

refresh-data:
	kubectl delete -f data/k8s/controller.yaml && \
	kubectl create -f data/k8s/controller.yaml

destroy:
	kubectl delete -f order/k8s/deployment.yaml && \
	kubectl delete -f order/k8s/service.yaml && \
	kubectl delete -f data/k8s/controller.yaml && \
	kubectl delete -f data/k8s/service.yaml && \
	kubectl delete -f data/k8s/volumes.yaml && \
	kubectl delete -f data/k8s/rbac.yaml && \
	kubectl delete -f data/k8s/namespace.yaml