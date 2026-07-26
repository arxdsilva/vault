.PHONY: help \
	backend backend-build backend-vet backend-test backend-lint backend-windows \
	frontend frontend-install frontend-typecheck frontend-test frontend-build \
	test dev package clean

help:
	@echo "Backend (Go):"
	@echo "  make backend            - vet + test + build"
	@echo "  make backend-build      - go build ./..."
	@echo "  make backend-vet        - go vet ./..."
	@echo "  make backend-test       - go test -race -cover ./..."
	@echo "  make backend-lint       - golangci-lint run (if installed)"
	@echo "  make backend-windows    - cross-compile check (GOOS=windows)"
	@echo ""
	@echo "Frontend (React/TS):"
	@echo "  make frontend           - install + typecheck + test + build"
	@echo "  make frontend-install   - npm install"
	@echo "  make frontend-typecheck - tsc --noEmit"
	@echo "  make frontend-test      - vitest run"
	@echo "  make frontend-build     - vite build"
	@echo ""
	@echo "Combined:"
	@echo "  make test               - backend + frontend"
	@echo "  make dev                - wails3 dev"
	@echo "  make package            - wails3 package (see also: wails3 task package:dmg)"
	@echo "  make clean              - remove build artifacts"

## Backend (Go) ##############################################################

backend: backend-vet backend-test backend-build

backend-build:
	go build ./...

backend-vet:
	go vet ./...

backend-test:
	go test -race -cover ./...

backend-lint:
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run || \
		echo "golangci-lint not installed; skipping (see README.md)"

backend-windows:
	GOOS=windows GOARCH=amd64 go build ./...
	GOOS=windows GOARCH=amd64 go vet ./...

## Frontend (React/TS) ########################################################

frontend: frontend-install frontend-typecheck frontend-test frontend-build

frontend-install:
	cd frontend && npm install

frontend-typecheck:
	cd frontend && npx tsc --noEmit

frontend-test:
	cd frontend && npx vitest run

frontend-build:
	cd frontend && npm run build

## Combined ###################################################################

test: backend frontend

dev:
	wails3 dev

package:
	wails3 package

clean:
	rm -rf bin frontend/dist
