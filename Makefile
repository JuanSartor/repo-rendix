.PHONY: backend frontend install dev

## Instala dependencias de Go y Node
install:
	cd backend && go mod tidy
	cd frontend && npm install

## Corre el backend Go
backend:
	cd backend && go run cmd/main.go

## Corre el frontend React
frontend:
	cd frontend && npm run dev

## Corre ambos en paralelo (requiere make >= 4.x o usar dos terminales)
dev:
	@echo "Abrí dos terminales y corré:"
	@echo "  make backend"
	@echo "  make frontend"
