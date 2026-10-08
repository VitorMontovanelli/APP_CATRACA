---
name: catraca-architect
description: "Guia e regras operacionais para decompor o monólito APP_CATRACA em Arquitetura Hexagonal (Ports & Adapters) e DDD, garantindo limite estrito de 200 linhas por arquivo Go e 150 linhas por componente React."
---

# Catraca Architect: Hexagonal DDD & Clean Code

Habilidade especializada para guiar a refatoração do software de controle de acesso desktop (Wails + Go + React) migrando do monólito `app.go` para uma arquitetura desacoplada, manutenível e extensível.

## 🎯 Princípios Norteadores (Baseados na Biblioteca Técnica)
- **Hexagonal Architecture (Alistair Cockburn / Robert C. Martin)**:
  - O núcleo de domínio (**Domain**) não conhece banco de dados (GORM/SQL), interfaces gráficas (Wails/IPC), nem bibliotecas de hardware.
  - A comunicação externa ocorre estritamente via **Portas (Interfaces Go)** e **Adaptadores**.
- **Domain-Driven Design (Eric Evans / Vaughn Vernon)**:
  - Separação em **Bounded Contexts**:
    1. `access`: Regras de autorização de catraca, leitura de cartão/biometria, log de eventos.
    2. `student`: Cadastro de aluno, fotos, plano associado, validação cadastral.
    3. `billing`: Faturas, métodos de pagamento, inadimplência, webhook/gateway.
    4. `scheduling`: Agenda individual, capacidade por dia/horário, reservas.
    5. `identity`: Usuários administradores, autenticação, hash de senha, logs de auditoria.
    6. `system`: Backup, status do hardware (catraca serial/webcam), notificações Telegram.
- **Limite Estrito de Linhas (Zero Monólitos)**:
  - Arquivos `.go`: **Máximo de 200 linhas**.
  - Funções/Métodos: **Máximo de 40 linhas**.
  - Componentes React (`.tsx`): **Máximo de 150 linhas**.

## 📂 Estrutura Padrão de Diretórios por Contexto
```
internal/<context>/
├── domain/            # Entidades, Value Objects e Erros de Negócio (Puro Go, ZERO deps)
│   ├── entity.go
│   └── errors.go
├── ports/             # Interfaces de entrada e saída
│   ├── repository.go  # Porta de saída (ex: StudentRepository)
│   └── service.go     # Porta de entrada (Casos de uso / Use Cases)
├── service/           # Implementação dos casos de uso (Regras da aplicação)
│   └── service.go
└── adapters/
    ├── storage/       # Adaptador de saída (GORM / SQLite / SQL puro)
    │   └── sqlite_repo.go
    └── ipc/           # Adaptador de entrada (Wails bindings / handlers)
        └── handler.go
```

## 📋 Protocolo de Refatoração Passo a Passo
1. **Identificar o Subdomínio**: Ao refatorar uma função de `app.go`, isole a entidade de domínio e sua interface de repositório.
2. **Criar a Porta**: Defina a interface em `ports/`.
3. **Mover a Lógica de Negócio**: Coloque a regra pura em `domain/` ou `service/`.
4. **Implementar o Adaptador de Dados**: Escreva a persistência em `adapters/storage/`.
5. **Criar o Handler Wails**: Crie um handler conciso (<180 linhas) em `adapters/ipc/` que apenas valida entrada, chama o use case e retorna o resultado para o frontend.
6. **Validar Linhas**: Execute `scripts/check_architecture.sh` para certificar que nenhum arquivo ultrapassou 200 linhas.
