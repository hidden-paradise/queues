CLUSTER_NAME = mycluster
REGISTRY_PORT = 5500
NAMESPACE = queues-learning
REGISTRY_DOCKER=localhost:$(REGISTRY_PORT)
REGISTRY_K8S=myregistry.localhost:$(REGISTRY_PORT)
REDIS_QUEUE="learning-queues"

.PHONY: all create-cluster build-images push-images deploy-registry deploy-deps deploy-producer deploy-consumer deploy delete-cluster

all: deploy

create-cluster:
	k3d cluster create $(CLUSTER_NAME) \
		--api-port 127.0.0.1:6443 \
		--port 8080:80@loadbalancer \
		--port 8443:443@loadbalancer \
		--registry-create $(REGISTRY_K8S) \
		--k3s-arg "--kubelet-arg=resolv-conf=/etc/resolv.conf@server:0" \
		--wait

build-consumer:
	docker build -t $(REGISTRY_DOCKER)/consumer:latest . --file consumer/Dockerfile
	docker push $(REGISTRY_DOCKER)/consumer:latest

build-producer:
	docker build -t $(REGISTRY_DOCKER)/producer:latest . --file producer/Dockerfile
	docker push $(REGISTRY_DOCKER)/producer:latest

create-namespace:
	kubectl create namespace $(NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -

deploy-deps: create-namespace
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
	helm repo update

	helm upgrade --install prometheus prometheus-community/kube-prometheus-stack \
      --namespace queues-learning

	helm upgrade --install redis bitnami/redis \
		--namespace $(NAMESPACE) \
		--set architecture=standalone \
		--set auth.enabled=false \
		--set metrics.enabled=true \
		--set metrics.serviceMonitor.enabled=true \
		--set metrics.serviceMonitor.namespace=queues-learning \
		--set metrics.serviceMonitor.interval=15s \
		--set metrics.serviceMonitor.scrapeTimeout=5s \
		--set metrics.serviceMonitor.additionalLabels.release=prometheus \
		--set metrics.extraEnvVars[0].name=REDIS_EXPORTER_CHECK_SINGLE_KEYS \
		--set metrics.extraEnvVars[0].value=$(REDIS_QUEUE)


deploy-producer: build-producer
	helm upgrade --install producer ./helm/producer \
		--namespace $(NAMESPACE) \
		--set image.repository=$(REGISTRY_K8S)/producer \
		--set image.tag=latest \
  		--set redisAddr="redis.queues-learning.svc.cluster.local:6379" \
  		--set env.QUEUE_NAME=$(REDIS_QUEUE)

deploy-consumer: build-consumer
	helm upgrade --install consumer ./helm/consumer \
		--namespace $(NAMESPACE) \
		--set image.repository=$(REGISTRY_K8S)/consumer \
		--set image.tag=latest \
		--set redisAddr="redis.queues-learning.svc.cluster.local:6379" \
		--set env.QUEUE_NAME=$(REDIS_QUEUE)

deploy: deploy-deps deploy-producer deploy-consumer

delete-cluster:
	k3d cluster delete $(CLUSTER_NAME)