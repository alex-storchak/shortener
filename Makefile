.PHONY: help
.PHONY: generate-mocks test test-static
.PHONY: build
.PHONY: generate
.PHONY: clean clean-bin clean-mocks clean-easyjson
.PHONY: format
.PHONY: doc

BUILD_VCS ?= true
GODOC_PORT ?= 8081

format:
	find . -type f -name '*.go' \
		-not -name '*_easyjson.go' \
		-not -path '*/mocks/*' \
		-exec goimports -local "github.com/alex-storchak/shortener" -w {} \;


doc:
	godoc -http=":$(GODOC_PORT)"

build: generate generate-mocks
	cd cmd/shortener && go build -buildvcs=$(BUILD_VCS) -o shortener

generate-mocks:
	mockery

generate: generate-mocks
	go generate ./...

clean-mocks:
	@echo "Removing generated mocks..."
	# Находим каталоги 'mocks' и удаляем mock_*.go внутри
	@find . -type d -name mocks -prune -exec sh -c ' \
		for d; do \
			find "$$d" -maxdepth 1 -type f -name "mock_*.go" -print -delete; \
			rmdir "$$d" 2>/dev/null || true; \
		done' sh {} +
	@echo "Done."

clean-easyjson:
	@echo "Removing generated easyjson files..."
	@find . -type f -name "*_easyjson.go" -print -delete
	@echo "Done."

clean-bin:
	rm -f cmd/shortener/shortener

clean: clean-mocks clean-easyjson clean-bin

test: generate-mocks
	go test ./...

test-static: generate-mocks
	go vet -vettool=$(which statictest) ./...

test-shortener: build
	./autotest/run_all.sh

help:
	@echo ""
	@echo "Usage: make <target>"
	@echo "  make test-shortener"
	@echo ""
	@echo "Targets:"
	@echo "  doc                Run godoc server. Default port: $(GODOC_PORT)"
	@echo ""
	@echo "  format             Format code with goimports. Excluded files: mockery mocks, easyjson"
	@echo ""
	@echo "  build              Build shortener binary"
	@echo ""
	@echo "  generate           Go generate"
	@echo ""
	@echo "  generate-mocks     Generate mocks for interfaces"
	@echo ""
	@echo "  clean              Clean up mocks and binary"
	@echo "  clean-mocks        Clean up mocks"
	@echo "  clean-easyjson     Clean up easyjson files"
	@echo ""
	@echo "  clean-bin          Clean up binary"
	@echo ""
	@echo "  test               Run all tests"
	@echo "  test-static        Run required autotest statictest (binary file \"statictest\" must be available from PATH)"
	@echo "  test-shortener     Run required autotest shortenertest (requires binary \"./autotest/run_all.sh\" script that not present in repo)"
	@echo ""