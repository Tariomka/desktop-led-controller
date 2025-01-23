BIN_DIR = bin
RAYLIB = raylib.dll
ifdef OS
	RM = del /s /q
	CP = copy
	EXE_NAME = desktop_app.exe
else
	RM = rm -rf
	CP = cp
	EXE_NAME = desktop_app
endif

run: build copy_raylib
	@./$(BIN_DIR)/$(EXE_NAME)

run_no_build:
	@./$(BIN_DIR)/$(EXE_NAME)

build: create
	@echo Staring to build executable, please wait...
	@go build -o ./$(BIN_DIR)/$(EXE_NAME) -ldflags "-H=windowsgui" main.go
	@echo Executable built successfully.

copy_raylib: create
	@if not exist ./$(BIN_DIR)/$(RAYLIB) $(CP) $(RAYLIB) $(BIN_DIR)

create:
	@if not exist $(BIN_DIR) mkdir $(BIN_DIR)

clean:
	@$(RM) $(BIN_DIR)

tests:
	@go test ./test/...

tests_verbose:
	@go test -v ./test/...

# remove later
placeholder_run: placeholder_build
	@./$(BIN_DIR)/placeholder.exe

placeholder_build:
	@echo Staring to build executable, please wait...
	@go build -o ./$(BIN_DIR)/placeholder.exe ./rpi_placeholder/main.go
	@echo Executable built successfully.