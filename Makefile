BIN_NAME := tw
INSTALL_DIR := /usr/local/bin

.PHONY: build install clean

build:
	go build -o $(BIN_NAME) .

install: build
	@mkdir -p $(INSTALL_DIR)
	cp $(BIN_NAME) $(INSTALL_DIR)/$(BIN_NAME)
	@echo "Installed $(BIN_NAME) to $(INSTALL_DIR)/$(BIN_NAME)"

clean:
	rm -f $(BIN_NAME)
