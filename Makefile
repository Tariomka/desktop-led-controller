run: build
	@./bin/led_server

build: create
	@go build -o ./bin/led_server ./src/main.go

create:
	@mkdir -p bin

clean:
	@rm -rf bin

test_:
	@go test ./test/...

test_v:
	@go test -v ./test/...