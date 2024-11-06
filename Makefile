build:
	env GOOS=linux go build -o bin/main main.go
run: build
	./bin/main