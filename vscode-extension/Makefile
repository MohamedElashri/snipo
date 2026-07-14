.PHONY: all build package install clean

VSCE = npx vsce

all: build

build:
	npm install
	npm run compile

package: build
	$(VSCE) package --allow-star-activation

install: package
	code --install-extension snipo-0.0.1.vsix --force

clean:
	rm -rf out
	rm -f *.vsix
	rm -rf node_modules

help:
	@echo "Available commands:"
	@echo "  build     - Build the extension"
	@echo "  package   - Package the extension"
	@echo "  install   - Install the extension"
	@echo "  clean     - Clean the extension"
	@echo "  help      - Show this help message"