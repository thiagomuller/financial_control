# ─── Variables ────────────────────────────────────────────────────────────────
IMAGE_NAME   := my-finances-api
IMAGE_TAG    := scan
FULL_IMAGE   := $(IMAGE_NAME):$(IMAGE_TAG)
TAR_PATH     := /tmp/$(IMAGE_NAME).tar
PODMAN       := distrobox-host-exec podman

# ─── Default target ───────────────────────────────────────────────────────────
.DEFAULT_GOAL := help

.PHONY: help lint trivy trivy-image test build all

help: ## Show available targets
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*##"}; {printf "  %-14s %s\n", $$1, $$2}'

# ─── Targets ──────────────────────────────────────────────────────────────────
lint: ## Run static analysis with go vet
	go vet ./...

test: ## Run unit tests
	go test ./...

build: ## Build all Go packages
	go build ./...

trivy: ## Scan source tree for HIGH/CRITICAL vulnerabilities (filesystem)
	trivy fs --exit-code 0 --severity HIGH,CRITICAL .

trivy-image: ## Build image, save to tar, scan with trivy, then clean up
	$(PODMAN) build -t $(FULL_IMAGE) .
	$(PODMAN) save -o $(TAR_PATH) $(FULL_IMAGE)
	trivy image --input $(TAR_PATH) --exit-code 0 --severity HIGH,CRITICAL
	rm -f $(TAR_PATH)

all: lint test trivy ## Run lint → test → trivy (CI pipeline)
