APP_NAME       := mytoolkit
VERSION        := $(shell cat VERSION)
IMAGE          := $(APP_NAME)
TAG            := $(VERSION)
PLATFORMS      := linux/amd64,linux/arm64
NAMESPACE      := $(APP_NAME)
KIND_CLUSTER   := kind-multinodes
KUBE_CONTEXT   := kind-$(KIND_CLUSTER)
SRC            := app
BIN_DIR        := bin
LDFLAGS        := -X github.com/aeciopires/mytoolkit/internal/version.Version=$(VERSION)

REQUIRED_TOOLS := go git docker helm kubectl kind golangci-lint helm-docs

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Build the mytoolkit binary into bin/ (version from ./VERSION)
	mkdir -p $(BIN_DIR)
	cd $(SRC) && go build -ldflags="$(LDFLAGS)" -o ../$(BIN_DIR)/$(APP_NAME) ./cmd/mytoolkit

.PHONY: run
run: ## Run the web server locally (go run)
	cd $(SRC) && go run ./cmd/mytoolkit serve

.PHONY: test
test: ## Run unit tests
	cd $(SRC) && go test ./...

.PHONY: test-verbose
test-verbose: ## Run unit tests with verbose output
	cd $(SRC) && go test -v ./...

.PHONY: coverage
coverage: ## Run tests with coverage report
	cd $(SRC) && go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out

.PHONY: lint
lint: ## Run golangci-lint
	cd $(SRC) && golangci-lint run

.PHONY: fmt
fmt: ## Format Go source
	cd $(SRC) && gofmt -s -w .

.PHONY: vet
vet: ## Run go vet
	cd $(SRC) && go vet ./...

.PHONY: check-tools
check-tools: ## Verify required development/runtime tools are installed
	@missing=""; \
	for tool in $(REQUIRED_TOOLS); do \
		if command -v $$tool >/dev/null 2>&1; then \
			echo "[OK]      $$tool"; \
		else \
			echo "[MISSING] $$tool"; \
			missing="$$missing $$tool"; \
		fi; \
	done; \
	if command -v docker >/dev/null 2>&1 && ! docker compose version >/dev/null 2>&1; then \
		echo "[MISSING] docker compose (plugin)"; \
		missing="$$missing docker-compose-plugin"; \
	fi; \
	if [ -n "$$missing" ]; then \
		echo ""; echo "Missing tools:$$missing"; \
		exit 1; \
	fi; \
	echo ""; echo "All required tools are installed."

.PHONY: deps-check
deps-check: ## Verify the Go module graph is tidy (go.mod/go.sum hygiene)
	cd $(SRC) && go mod verify && go mod tidy -diff

.PHONY: docker-build
docker-build: ## Build a local single-platform Docker image (version from ./VERSION)
	docker build --build-arg VERSION=$(VERSION) -t $(IMAGE):$(TAG) -t $(IMAGE):latest .

.PHONY: docker-buildx
docker-buildx: ## Build (and validate) a multi-arch image for linux/amd64 + linux/arm64
	docker buildx build --platform $(PLATFORMS) --build-arg VERSION=$(VERSION) -t $(IMAGE):$(TAG) .

.PHONY: docker-run
docker-run: ## Run the local Docker image on :8080
	docker run --rm -p 8080:8080 $(IMAGE):$(TAG)

.PHONY: docker-push
docker-push: ## Prompt for Docker Hub credentials and push a multi-arch (amd64+arm64) image tagged $(VERSION)
	@read -p "Docker Hub username: " DOCKERHUB_USER; \
	read -s -p "Docker Hub password or access token: " DOCKERHUB_PASSWORD; echo; \
	read -p "Docker Hub repository (e.g. yourname/mytoolkit): " DOCKERHUB_REPO; \
	echo "$$DOCKERHUB_PASSWORD" | docker login --username "$$DOCKERHUB_USER" --password-stdin; \
	docker buildx build --platform $(PLATFORMS) --build-arg VERSION=$(VERSION) -t "$$DOCKERHUB_REPO:$(VERSION)" -t "$$DOCKERHUB_REPO:latest" --push .; \
	docker logout >/dev/null

.PHONY: compose-up
compose-up: ## Start the app via docker compose
	docker compose up --build

.PHONY: compose-down
compose-down: ## Stop the app started via docker compose
	docker compose down

.PHONY: helm-lint
helm-lint: ## Lint the Helm chart
	helm lint helm/$(APP_NAME)

.PHONY: helm-template
helm-template: ## Render the Helm chart locally
	helm template $(APP_NAME) helm/$(APP_NAME) --set ingress.enabled=true --set autoscaling.enabled=true

.PHONY: helm-docs
helm-docs: ## Regenerate helm/mytoolkit/README.md from README.md.gotmpl + values.yaml comments
	helm-docs --chart-search-root=helm

.PHONY: kind-load
kind-load: ## Load the local Docker image into the kind-multinodes cluster
	kind load docker-image $(IMAGE):$(TAG) --name $(KIND_CLUSTER)

.PHONY: helm-install
helm-install: ## helm upgrade --install against $(KUBE_CONTEXT) / namespace $(NAMESPACE)
	helm upgrade --install $(APP_NAME) helm/$(APP_NAME) \
		--namespace $(NAMESPACE) --create-namespace \
		--kube-context $(KUBE_CONTEXT)

.PHONY: helm-uninstall
helm-uninstall: ## helm uninstall from $(KUBE_CONTEXT) / namespace $(NAMESPACE)
	helm uninstall $(APP_NAME) --namespace $(NAMESPACE) --kube-context $(KUBE_CONTEXT)

.PHONY: helm-test
helm-test: ## Run helm test (hits /healthz) against the installed release
	helm test $(APP_NAME) --namespace $(NAMESPACE) --kube-context $(KUBE_CONTEXT)

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BIN_DIR) $(SRC)/coverage.out
