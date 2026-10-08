---
name: catraca-test-runner
description: "Execução, escrita e validação de testes unitários em Go (TDD) com mocks para portas de saída, testes de repositório em SQLite in-memory e testes de componentes frontend com cobertura mínima de 80%."
---

# Catraca Test Runner: TDD, Mocks & Testes de Integração

Habilidade especializada para criação e execução de suítes de testes automatizados no projeto `APP_CATRACA`.

## 🧪 Padrões de Teste Unitário (Go)

### 1. Testes de Casos de Uso (Service Layer) com Mocks
- Os testes unitários das regras de negócio em `internal/<context>/service/` devem rodar em **milissegundos** e sem dependência de banco de dados ou arquivos externos.
- Use mocks gerados com `github.com/stretchr/testify/mock` ou `go.uber.org/mock` para simular as portas (`ports.Repository`, `ports.PaymentGateway`, etc.).
- **Tabela de Testes (Table-Driven Tests)**: Padrão canônico do Go com structs anônimas:
  ```go
  tests := []struct {
      name      string
      input     domain.CreateStudentInput
      mockSetup func(m *mocks.StudentRepository)
      wantErr   bool
      errType   error
  }{
      // Casos de sucesso, validação e erro de persistência
  }
  ```

### 2. Testes de Repositório (Integration Tests)
- Para testar `internal/<context>/adapters/storage/sqlite_repo.go`, utilize um banco SQLite em memória:
  ```go
  db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
  ```
- Valide queries complexas, transações atômicas e restrições de unicidade (`UNIQUE constraints`).

### 3. Cobertura Mínima e Execução Rápida
- **Cobertura de Domínio e Service**: Mínimo de **80%** de code coverage.
- Para rodar toda a suíte de testes com relatório de cobertura:
  ```bash
  go test -race -covermode=atomic -coverprofile=coverage.out ./internal/...
  go tool cover -func=coverage.out
  ```

### 4. Testes de Frontend (React / TypeScript)
- Testar componentes visuais e hooks de estado com Vitest / React Testing Library.
- Mockar as chamadas globais de IPC do Wails (`window.go.main.App...`).
