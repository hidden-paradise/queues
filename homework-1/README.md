# Homework 1

> **Task**
> 
> Create producer and consumer with medium between with your choice.


## What we have here

Producer and consumer deployed in local kubernetes cluster. 

## Why kubernetes
I want to inspect queue metrics and my own. So I added prometheus and grafana there.

Also I want to scale consumers and producers, and kubernetes allowed it to me. 


## Requirements
1. Docker
2. k3d
3. make
4. helm
5. kubectl

## How to run
I supposed docker is already installed.

Install requirements. 
```bash
brew install helm k3d kubectl make
```

Go to homework-1 directory.
```bash
cd homework-1
```

Setup cluster
```bash
make create-cluster
```

Deploy infrastructure and applications
```
make deploy
```

Do port forwarding for grafana
```bash
ubectl --context k3d-queues-homework-1 --namespace queues-homework-1 port-forward svc/prometheus-grafana 3000:80
```

Grafana credentials
> admin : prom-operator

Metrics for these applications are available through a link
```
http://localhost:3000/explore?schemaVersion=1&panes=%7B%22f12%22:%7B%22datasource%22:%22prometheus%22,%22queries%22:%5B%7B%22refId%22:%22A%22,%22expr%22:%22redis_key_size%7Bkey%3D%5C%22queues-homework-1%5C%22%7D%22,%22range%22:true,%22instant%22:true,%22datasource%22:%7B%22type%22:%22prometheus%22,%22uid%22:%22prometheus%22%7D,%22editorMode%22:%22code%22,%22legendFormat%22:%22__auto%22%7D,%7B%22refId%22:%22B%22,%22expr%22:%22rate%28producer_jobs_total%5B5s%5D%29%22,%22range%22:true,%22instant%22:true,%22datasource%22:%7B%22type%22:%22prometheus%22,%22uid%22:%22prometheus%22%7D,%22editorMode%22:%22code%22,%22legendFormat%22:%22__auto%22%7D,%7B%22refId%22:%22C%22,%22expr%22:%22rate%28consumer_jobs_total%5B$__rate_interval%5D%29%22,%22range%22:true,%22instant%22:true,%22datasource%22:%7B%22type%22:%22prometheus%22,%22uid%22:%22prometheus%22%7D,%22editorMode%22:%22code%22,%22legendFormat%22:%22__auto%22%7D,%7B%22refId%22:%22D%22,%22expr%22:%22rate%28consumer_jobs_failed_total%5B$__rate_interval%5D%29%22,%22range%22:true,%22instant%22:true,%22datasource%22:%7B%22type%22:%22prometheus%22,%22uid%22:%22prometheus%22%7D,%22editorMode%22:%22code%22,%22legendFormat%22:%22__auto%22,%22hide%22:false%7D%5D,%22range%22:%7B%22from%22:%22now-30m%22,%22to%22:%22now%22%7D%7D%7D&orgId=1
```

If link doesn't work, you could inspect them via **Explore**
```
redis_key_size{key="queues-homework-1"}
rate(producer_jobs_total[5s])
rate(consumer_jobs_total[$__rate_interval])
rate(consumer_jobs_failed_total[$__rate_interval])
```

We can scale consumer or producer 
```bash
date; kubectl --context k3d-queues-homework-1 --namespace queues-homework-1 scale deployment producer --replicas 30
date; kubectl --context k3d-queues-homework-1 --namespace queues-homework-1 scale deployment producer --replicas 1
date; kubectl --context k3d-queues-homework-1 --namespace queues-homework-1 scale deployment producer --replicas 0

date; kubectl --context k3d-queues-homework-1 --namespace queues-homework-1 scale deployment consumer --replicas 30
date; kubectl --context k3d-queues-homework-1 --namespace queues-homework-1 scale deployment consumer --replicas 1
date; kubectl --context k3d-queues-homework-1 --namespace queues-homework-1 scale deployment consumer --replicas 0
```

## How to uninstall
In homework-1 directory run the next command
```bash
make delete-cluster
```

Probably ou have to delete something via...
```bash
k3d node delete <node_name>
k3d registry delete <registry_name>
```



