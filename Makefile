BIN_DIR = bin
RAYLIB = raylib.dll
BUILD_FLAGS = -ldflags "-H=windowsgui"
ifdef OS
	RM = del /s /q
	CP = copy
	EXE_NAME = desktop_app.exe
else
	RM = rm -rf
	CP = cp
	EXE_NAME = desktop_app
endif

run_no_flags: build_no_flags copy_raylib
	@./$(BIN_DIR)/$(EXE_NAME)

run: build copy_raylib
	@./$(BIN_DIR)/$(EXE_NAME)

run_no_build:
	@./$(BIN_DIR)/$(EXE_NAME)

# Does not work on linux because of BUILD_FLAGS
build: create
	@echo Staring to build executable, please wait...
	@go build -o ./$(BIN_DIR)/$(EXE_NAME) $(BUILD_FLAGS) main.go
	@echo Executable built successfully.

build_no_flags: create
	@echo Staring to build executable, please wait...
	@go build -o ./$(BIN_DIR)/$(EXE_NAME) main.go
	@echo Executable built successfully.

copy_raylib: create
	@if [ ! -a ./$(BIN_DIR)/$(RAYLIB) ]; then $(CP) $(RAYLIB) $(BIN_DIR); fi

create:
	@if [ ! -d $(BIN_DIR) ]; then mkdir $(BIN_DIR); fi

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