.PHONY: build clean

GO          ?=  go
BIN_FILE    =   fs-exporter

build:
	@echo ">> make binary"
	@$(GO) build -gcflags=all="-N -l" -o "${BIN_FILE}" github.com/microyahoo/fs_exporter

clean:
	rm -rf $(BIN_FILE)
