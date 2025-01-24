GOLANGCI_LINT_CACHE?=/tmp/praktikum-golangci-lint-cache

.PHONY: golangci-lint-run
golangci-lint-run: _golangci-lint-rm-unformatted-report

.PHONY: _golangci-lint-reports-mkdir
_golangci-lint-reports-mkdir:
	mkdir -p ./golangci-lint

.PHONY: _golangci-lint-run
_golangci-lint-run: _golangci-lint-reports-mkdir
	-docker run --rm \
    -v $(shell pwd):/app \
    -v $(GOLANGCI_LINT_CACHE):/root/.cache \
    -w /app \
    golangci/golangci-lint:v1.63.4 \
        golangci-lint run \
            -c .golangci.yml \
	> ./golangci-lint/report-unformatted.json

.PHONY: _golangci-lint-format-report
_golangci-lint-format-report: _golangci-lint-run
	cat ./golangci-lint/report-unformatted.json | jq > ./golangci-lint/report.json

.PHONY: _golangci-lint-rm-unformatted-report
_golangci-lint-rm-unformatted-report: _golangci-lint-format-report
	rm ./golangci-lint/report-unformatted.json

.PHONY: golangci-lint-clean
golangci-lint-clean:
	sudo rm -rf ./golangci-lint


# Имя исполняемого файла
BINARY_NAME=gmloyalty

# Цель по умолчанию
.PHONY: all
all: build

# Установка зависимостей
.PHONY: deps
deps: 
	@echo "==> Installing dependencies..." 
	@go mod tidy

# Компиляция проекта
.PHONY: build
build: deps 
	@echo "==> Building the project..." 
	@go build -o ./cmd/gophermart/$(BINARY_NAME) ./cmd/gophermart/main.go

# Запуск тестов
.PHONY: test
test: 
	@echo "==> Running tests..." 
	@go test ./...

# Очистка
.PHONY: clean
clean: 
	@echo "==> Cleaning up..." 
	@rm -f ./cmd/gophermart/$(BINARY_NAME)

# Запуск проекта (конфигурирование через переменные окружения)
.PHONY: run
run: build 
	@echo "==> Running the project..." 
	@./cmd/gophermart/$(BINARY_NAME)

.PHONY: compose-run
compose-run: 
	docker-compose up 