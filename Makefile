start: run

run:
	go run main.go --master_url=https://192.168.39.107:8443

deps:
	go mod download
	go mod verify
	go mod tidy
