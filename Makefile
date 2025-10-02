.PHONY: run lint build tidy

# Executa a aplicação local
run:
	go run ./backend/cmd/server/main.go

# Executa o linting (requer golangci-lint instalado)
lint:
	golangci-lint run ./...

# Compila o binário
build:
	go build -o bin/app ./backend/cmd/server/main.go

# Atualiza dependências do Go
tidy:
	go mod tidy