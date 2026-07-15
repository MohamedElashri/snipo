.PHONY: all build test package install clean help

VSCE = npx vsce

all: build

build:
	npm install
	npm run esbuild

test:
	npm run test

package: build
	$(VSCE) package --allow-star-activation

install: package
	code --install-extension *.vsix --force

clean:
	rm -rf out
	rm -f *.vsix
	rm -rf node_modules

help:
	@echo "Available commands:"
	@echo "  build     - Install dependencies and bundle the extension (esbuild)"
	@echo "  test      - Run headless unit tests"
	@echo "  package   - Package the extension into a .vsix file"
	@echo "  install   - Package and install the extension into VS Code"
	@echo "  clean     - Remove build artifacts and node_modules"
	@echo "  help      - Show this help message"