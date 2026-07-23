# Catraca-App

Sistema desktop de controle de acesso e gestão financeira para academias.

## Funcionalidades

- **Catraca** — liberação/negação de acesso por biometria
- **Alunos** — cadastro, planos, vencimentos
- **Financeiro** — faturas, confirmação de pagamento, comprovantes
- **Cobrança** — lista de alunos com faturas pendentes/pagas, upload de PDF
- **Planos** — gerenciamento de planos com lista de alunos vinculados
- **Logs** — auditoria e histórico de acesso
- **Usuários** — controle de permissões (super_admin / admin)

## Stack

| Camada    | Tecnologia                     |
|-----------|--------------------------------|
| Frontend  | React 19 + TypeScript + Vite 7 |
| Backend   | Go 1.26                        |
| Desktop   | Wails 2 (WebView2)             |
| Banco     | SQLite (glebarez/sqlite)       |

## Requisitos

- Go 1.26+
- Node.js 20+
- Wails CLI 2.13+
- WebView2 (Windows 10/11)

## Desenvolvimento

```bash
wails dev
```

## Build

```bash
wails build
```

> **Nota:** Em máquinas com política de Controle de Aplicativo (WDAC), o binding
> pode ser bloqueado. Use `wails build -skipbindings` como alternativa.

## Versão

```
v1.0.0
```

## Credencial padrão

| Email              | Senha    | Cargo       |
|--------------------|----------|-------------|
| admin@catraca.com  | admin123 | super_admin |
