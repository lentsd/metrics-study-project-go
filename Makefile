AIR ?= $(shell go env GOPATH)/bin/air

.PHONY: dev
dev:
	@test -x "$(AIR)" || { echo 'Air не найден. Установи: go install github.com/air-verse/air@latest'; exit 1; }
	"$(AIR)" --build.cmd "go build -o ./cmd/server/server ./cmd/server" --build.entrypoint "./cmd/server/server"
