BINARY_NAME=goshort
BUILD_DIR=build

ifdef ComSpec
    RM = del /Q /F
    MKDIR = if not exist $(subst /,\,$(1)) mkdir $(subst /,\,$(1))
    EXE_EXT=.exe
    SET_ENV = set
else
    RM = rm -rf
    MKDIR = mkdir -p $(1)
    EXE_EXT=
    SET_ENV = export
endif

build-windows-amd64:
	$(call MKDIR,$(BUILD_DIR)\windows-amd64)
	$(SET_ENV) CGO_ENABLED=0 GOOS=windows GOARCH=amd64 && go build -o $(BUILD_DIR)/windows-amd64/$(BINARY_NAME)$(EXE_EXT)

build-windows-arm64:
	$(call MKDIR,$(BUILD_DIR)\windows-arm64)
	$(SET_ENV) CGO_ENABLED=0 GOOS=windows GOARCH=arm64 && go build -o $(BUILD_DIR)/windows-arm64/$(BINARY_NAME)$(EXE_EXT)

build-linux-amd64:
	$(call MKDIR,$(BUILD_DIR)/linux-amd64)
	$(SET_ENV) CGO_ENABLED=0 GOOS=linux GOARCH=amd64 && go build -o $(BUILD_DIR)/linux-amd64/$(BINARY_NAME)

build-linux-arm64:
	$(call MKDIR,$(BUILD_DIR)/linux-arm64)
	$(SET_ENV) CGO_ENABLED=0 GOOS=linux GOARCH=arm64 && go build -o $(BUILD_DIR)/linux-arm64/$(BINARY_NAME)

build-linux:
	$(call MKDIR,$(BUILD_DIR))
	$(MAKE) build-linux-amd64
	$(MAKE) build-linux-arm64

build-macos-amd64:
	$(call MKDIR,$(BUILD_DIR)/macos-amd64)
	$(SET_ENV) CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 && go build -o $(BUILD_DIR)/macos-amd64/$(BINARY_NAME)

build-macos-arm64:
	$(call MKDIR,$(BUILD_DIR)/macos-arm64)
	$(SET_ENV) CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 && go build -o $(BUILD_DIR)/macos-arm64/$(BINARY_NAME)

build-macos:
	$(call MKDIR,$(BUILD_DIR))
	$(MAKE) build-macos-amd64
	$(MAKE) build-macos-arm64

all: build-windows-amd64 build-windows-arm64 build-linux build-macos

clean:
	$(RM) $(subst /,\,$(BUILD_DIR))

.PHONY: build-windows-amd64 build-windows-arm64 build-linux-amd64 build-linux-arm64 build-linux build-macos-amd64 build-macos-arm64 build-macos all clean
