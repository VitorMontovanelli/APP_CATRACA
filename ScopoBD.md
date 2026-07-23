# ScopoBD — Catraca-App

## Hierarquia de Acesso

```
ADMIN SUPREMO
├── Acesso a tudo
└── Cria usuário ADMIN

ADMIN
├── Cria usuário NORMAL
└── Acesso parcial (telas limitadas)

ALUNO (Usuário Normal)
└── Acesso limitado
```

---

## Estrutura do Projeto

```
catraca-app/
│
├── main.go                    # Entry point Wails (Go)
├── app.go                     # App struct + bindings (CriarUsuario, ListarUsuarios)
├── db.go                      # Inicialização do SQLite + criação das tabelas
├── go.mod / go.sum            # Módulo Go
├── wails.json                 # Configuração Wails
├── .gitignore
├── ScopoBD.md
│
├── frontend/                  # Frontend React + Vite + TypeScript
│   ├── index.html
│   ├── package.json
│   ├── vite.config.ts
│   ├── tsconfig.json
│   ├── public/
│   │   └── vite.svg
│   ├── src/
│   │   ├── main.tsx           # Entry point React
│   │   ├── App.tsx            # Componente raiz
│   │   ├── App.css            # Estilos globais
│   │   ├── Login.tsx          # Tela de login
│   │   ├── Login.css          # Estilos do login
│   │   ├── vite-env.d.ts
│   │   └── assets/
│   │       └── react.svg
│   ├── dist/                  # Build de produção (gerado)
│   └── node_modules/
│
├── build/                     # Artefatos de build Wails
│   ├── appicon.png
│   ├── bin/
│   │   └── catraca-app.exe    # Binário compilado
│   └── windows/
│
└── .opencode/                 # Configuração opencode
    ├── skills/
    └── package.json
```

---

## Banco de Dados (SQLite)

### Tabela: `users`

| Campo            | Tipo         | Descrição                                    |
|------------------|--------------|----------------------------------------------|
| id               | INTEGER (PK) | Único                                        |
| name             | TEXT         | Nome do usuário                              |
| email            | TEXT         | Email                                        |
| senha            | TEXT         | Senha                                        |
| cargo            | TEXT (enum)  | `super_admin` / `admin` / `aluno`            |
| registrador_id   | INTEGER (FK) | ID de quem criou (rastreabilidade)           |
| ativo            | BOOLEAN      | Ativar / desativar                           |
| criado_em        | TEXT         | Data de criação                              |

### Tabela: `alunos` (dados específicos do aluno)

| Campo            | Tipo         | Descrição                                      |
|------------------|--------------|--------------------------------------------------|
| user_id          | INTEGER (FK) | FK → users.id                                   |
| plano            | TEXT         | Plano contratado                                |
| status           | TEXT (enum)  | `ativo` / `inadimplente` / `suspenso`           |
| data_vencimento  | TEXT         | Vencimento do plano                             |
| biometria_id     | INTEGER (FK) | FK → biometria.id                               |

### Tabela: `biometria`

| Campo       | Tipo         | Descrição                                        |
|-------------|--------------|--------------------------------------------------|
| id          | INTEGER (PK) | ID da digital                                    |
| user_id     | INTEGER (FK) | FK → users.id                                    |
| template    | BLOB         | Dado biométrico (template, nunca imagem crua)    |
| enrolled_at | TEXT         | Data do cadastro                                 |

### Tabela: `catraca` (dispositivo físico)

| Campo     | Tipo         | Descrição                                |
|-----------|--------------|------------------------------------------|
| id        | INTEGER (PK) | ID                                        |
| ip        | TEXT         | IP do dispositivo                        |
| serial    | TEXT         | Serial do dispositivo                    |
| status    | TEXT (enum)  | `online` / `offline`                     |
| last_seen | TEXT         | Timestamp da última vez online           |

### Tabela: `access_logs` (registro de acesso)

| Campo     | Tipo         | Descrição                                  |
|-----------|--------------|--------------------------------------------|
| id        | INTEGER (PK) | ID                                          |
| user_id   | INTEGER (FK) | FK → users.id (quem tentou passar)         |
| timestamp | TEXT         | Data/hora da tentativa                     |
| resultado | TEXT (enum)  | `liberado` / `negado`                     |
| motivo    | TEXT         | Ex: "mensalidade vencida", "digital não reconhecida" |

### Tabela: `audit_logs` (rastreabilidade administrativa)

| Campo     | Tipo         | Descrição                                    |
|-----------|--------------|----------------------------------------------|
| id        | INTEGER (PK) | ID                                            |
| actor_id  | INTEGER (FK) | FK → users.id (quem fez a ação)              |
| action    | TEXT         | Ex: "criou aluno", "removeu admin"           |
| target_id | INTEGER      | ID do alvo da ação                           |
| timestamp | TEXT         | Data/hora                                    |

> Nota: A tabela `audit_logs` permite rastrear quantas vezes os alunos frequentam por dia, auxiliando na definição de "cliente ativo" da academia.

---

## Tecnologias

| Camada     | Tecnologia               |
|------------|--------------------------|
| Frontend   | React 19 + TypeScript    |
| Build      | Vite 7                   |
| Backend    | Go + Wails 2             |
| Database   | SQLite (via modernc.org/sqlite) |
| Desktop    | Wails v2 (WebView2)      |
