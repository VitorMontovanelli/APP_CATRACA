# Agendamento-App

**Versão:** 1.0.0

Sistema desktop de controle de acesso biométrico e gestão financeira para academias, construído com Wails 2, Go e React.

## Funcionalidades

- **Agendamento** — matricula nas aulas, acesso as aulas e disponibilidades das aulas
- **Alunos** — cadastro completo com planos, vencimentos e histórico
- **Financeiro** — faturas, confirmação/cancelamento de pagamento, upload de comprovantes (PDF via base64)
- **Cobrança** — lista de faturas vencidas com upload de comprovante
- **Planos** — gerenciamento de planos (mensal, trimestral, semestral, anual) com alunos vinculados
- **Logs** — auditoria de ações e histórico de acesso à catraca
- **Usuários** — controle de permissões hierárquico (super_admin / admin)

## Stack

| Camada    | Tecnologia                                 |
|-----------|--------------------------------------------|
| Frontend  | React 19 + TypeScript + Vite 7             |
| Backend   | Go 1.26 + GORM                             |
| Desktop   | Wails 2 (WebView2)                         |
| Banco     | SQLite via glebarez/sqlite (pure Go, CGO-free) |

## Requisitos

- Go 1.26+
- Node.js 20+
- Wails CLI 2.13+
- WebView2 Runtime (Windows 10/11 incluso)

## Desenvolvimento

```bash
wails dev
```

## Build

```bash
wails build
```

> **Nota:** Em máquinas com política de Controle de Aplicativo (WDAC), `wailsbuildings.exe` pode ser bloqueado. Use `wails build -skipbindings` como alternativa (as bindings já estão geradas em `frontend/wailsjs/`).

## Estrutura do projeto

```
agendamento-app/
├── app.go             # Lógica de negócio (Go)
├── db.go              # Inicialização do banco e seed
├── models.go          # Modelos GORM
├── version.go         # Constante de versão
├── main.go            # Entrypoint Wails
├── frontend/
│   ├── src/
│   │   ├── App.tsx          # Roteamento e layout
│   │   ├── App.css          # Estilos globais
│   │   └── pages/           # Páginas React
│   │       ├── Home.tsx
│   │       ├── Login.tsx
│   │       ├── Alunos.tsx
│   │       ├── Financeiro.tsx
│   │       ├── Cobranca.tsx
│   │       ├── Planos.tsx
│   │       ├── Logs.tsx
│   │       └── Usuarios.tsx
│   └── wailsjs/             # Bindings geradas (Go → JS)
└── build/                   # Artefatos de build
```

## Licença

Proprietário — todos os direitos reservados.
