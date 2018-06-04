BUILD=go build
CLEAN=go clean
TEST=go test
DEPS=glide

PACKAGE=rlimiter


all: clean deps build

build:
	@echo "Building $(PACKAGE)"
	@$(BUILD) -o $(PACKAGE) -v
	@echo "You should find $(PACKAGE) executable in local dir"
	@echo "To run it: \033[1mmake run\033[0m"
	@echo "Consult \033[1mREADME.md\033[0m for more information"

test:
	$(TEST) ./... -cover

clean:
	@echo "Cleaning up..."
	@$(CLEAN)
	@rm -f $(PACKAGE)

run: build
	./$(PACKAGE)

deps:
	@echo "Installing dependencies"
	@$(DEPS) install

develop:
	ag -l | entr -s 'make build'

testing:
	ag -l | entr -s 'make test'
