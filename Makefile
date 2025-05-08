export minikube_ip=$(shell minikube ip)

start: run

run:
	go run main.go --master_url=https://$(minikube_ip):8443 --grep=nginx

deps:
	go mod download
	go mod verify
	go mod tidy

minikube:
	minikube start
	minikube node list

up:
	minikube kubectl -- apply -f contrib/two_nginx.yaml

down:
	minikube kubectl -- delete deployment nginx-deployment
	minikube kubectl -- delete service nginx-service

test:
	curl -v http://$(minikube_ip):31080
