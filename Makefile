GIT_TAG := $(shell git describe --tags --always)
GIT_COMMIT := $(shell git rev-parse --short HEAD)
LDFLAGS := "-X main.GitTag=${GIT_TAG} -X main.GitCommit=${GIT_COMMIT}"
DIST := $(CURDIR)/dist
DOCKER_USER := $(shell printenv DOCKER_USER)
DOCKER_PASSWORD := $(shell printenv DOCKER_PASSWORD)
TRAVIS := $(shell printenv TRAVIS)

# configuration adjustments for golangci-lint
GOLANGCI_LINT_DISABLED_LINTERS := "" # disabling typecheck, because it currently (06.09.2019) fails with Go 1.13
# Rules for directory list as input for the golangci-lint program
LINT_DIRS := $(DIRS) $(foreach dir,$(REC_DIRS),$(dir)/...)

all: build docker push

fmt:
	go fmt ./pkg/... ./cmd/...

vet:
	go vet ./pkg/... ./cmd/...

lint:
	@golangci-lint run -D $(GOLANGCI_LINT_DISABLED_LINTERS) $(LINT_DIRS)

# Build kube-tasks binary
build: fmt vet lint
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) -o bin/kube-tasks cmd/kube-tasks.go

# Build kube-tasks docker image
docker: lint fmt vet
	cp bin/kube-tasks kube-tasks
	docker build -t iwilltry42/kube-tasks:latest .
	rm kube-tasks


# Push will only happen in travis ci
push:
ifdef TRAVIS
ifdef DOCKER_USER
ifdef DOCKER_PASSWORD
	docker login -u $(DOCKER_USER) -p $(DOCKER_PASSWORD)
	docker push iwilltry42/kube-tasks:latest
endif
endif
endif

dist:
	mkdir -p $(DIST)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) -o kube-tasks cmd/kube-tasks.go
	tar -zcvf $(DIST)/kube-tasks-linux-$(GIT_TAG).tgz kube-tasks
	rm kube-tasks
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags $(LDFLAGS) -o kube-tasks cmd/kube-tasks.go
	tar -zcvf $(DIST)/kube-tasks-macos-$(GIT_TAG).tgz kube-tasks
	rm kube-tasks
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags $(LDFLAGS) -o kube-tasks.exe cmd/kube-tasks.go
	tar -zcvf $(DIST)/kube-tasks-windows-$(GIT_TAG).tgz kube-tasks.exe
	rm kube-tasks.exe