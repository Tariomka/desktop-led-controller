BIN_DIR = bin
ifdef OS
	RM = del /s /q
	EXE_NAME = desktop_app.exe
else
	RM = rm -rf
	EXE_NAME = desktop_app
endif

run: build
	@./$(BIN_DIR)/$(EXE_NAME)

build: create
	@echo Staring to build executable, please wait...
	@go build -o ./$(BIN_DIR)/$(EXE_NAME) main.go
	@echo Executable built successfully.

create:
	@if not exist $(BIN_DIR) mkdir $(BIN_DIR)

clean:
	@$(RM) $(BIN_DIR)

tests:
	@go test ./test/...

tests_verbose:
	@go test -v ./test/...