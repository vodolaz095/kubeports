export app=kubeports
export majorVersion=0
export minorVersion=1

export arch=$(shell uname)-$(shell uname -m)
export gittip=$(shell git log --format='%h' -n 1)
export subver=$(shell hostname)_on_$(shell date -u '+%Y-%m-%d_%I:%M:%S%p')
export patchVersion=$(shell git log --format='%h' | wc -l)
export ver=$(majorVersion).$(minorVersion).$(patchVersion).$(gittip)-$(arch)

clean:
	rm -f build/$(app)

run: start

vuln:
	which govulncheck
	govulncheck ./...

start:
	go run main.go --master_url=https://$(shell minikube ip):8443

start/nginx:
	go run main.go --master_url=https://$(shell minikube ip):8443 --grep=nginx

start/jaeger:
	go run main.go --master_url=https://$(shell minikube ip):8443 --grep=nginx


deps:
	go mod download
	go mod verify
	go mod tidy

minikube:
	minikube start
	minikube node list

up/nginx:
	minikube kubectl -- apply -f contrib/two_nginx.yaml

up/jaeger:
	minikube kubectl -- apply -f contrib/jaeger.yaml

down/nginx:
	minikube kubectl -- delete -f contrib/two_nginx.yaml

down/jaeger:
	minikube kubectl -- delete -f contrib/jaeger.yaml

test:
	curl -v http://$(minikube_ip):31080

build: clean
# https://www.reddit.com/r/golang/comments/10te58n/error_loading_shared_library_libresolvso2_no_such/
	CGO_ENABLED=0 go build -ldflags "-X main.Subversion=$(subver) -X main.Version=$(ver)" -o build/$(app) main.go

binary:
	./build/kubeports --master_url=https://$(shell minikube ip):8443 --grep=nginx
