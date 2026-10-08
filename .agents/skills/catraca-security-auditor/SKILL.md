---
name: catraca-security-auditor
description: "Auditoria de segurança, conformidade OWASP Desktop para Wails, sanitização de IPC, isolamento de hardware (V4L2/webcam), autenticação/sessão, e proteção de banco SQLite."
---

# Catraca Security Auditor: OWASP Desktop & Hardware Isolation

Habilidade especializada para auditoria e garantia de segurança do software desktop `APP_CATRACA`.

## 🛡️ Checklist de Segurança Obrigatório

### 1. Comunicação IPC e Bindings Wails (Frontend -> Go)
- **Nunca confiar nos parâmetros de entrada vindos do frontend**: Validar rigorosamente IDs numéricos, tamanhos de strings e payloads JSON.
- **Tratamento Seguro de Erros**: Nunca vazar stack traces técnicos ou mensagens internas do SQLite para o usuário. Retorne erros de domínio amigáveis (ex: `ErrStudentNotFound`, `ErrUnauthorized`).
- **Validação de Tipos**: Use structs fortemente tipadas com validação explícita antes de repassar aos use cases.

### 2. Execução de Processos e Acesso a Hardware (Webcam / Catraca)
- **Zero Command Injection**:
  - **PROIBIDO** concatenar strings em comandos de shell (ex: `exec.Command("sh", "-c", "ffmpeg ... " + param)`).
  - Sempre use `exec.Command("ffmpeg", "-f", "v4l2", "-i", device, ...)` com argumentos separados e caminho estrito.
  - O parâmetro `device` deve ser restrito a uma whitelist estrita (ex: `/dev/video[0-9]+`).
- **Path Traversal**:
  - Salvar fotos de webcam somente em diretórios pré-configurados e canônicos (`filepath.Clean`).
  - Proibir nomes de arquivos com `../`, caracteres nulos ou extensões arbitrárias. Aceitar apenas `.jpg` ou `.png`.

### 3. Autenticação, Senhas e Sessões
- **Hash de Senha**: Utilize estritamente `golang.org/x/crypto/bcrypt` com custo mínimo 12 para todas as senhas de usuários e administradores.
- **Zero Hardcoded Secrets**: Credenciais de banco, tokens de bot Telegram ou chaves de gateway de pagamento devem ser carregadas de variáveis de ambiente (`.env`) ou chaveiro local seguro do SO, nunca hardcoded em código de repositório.
- **Auditoria de Ações Sensíveis**: Registrar todas as operações de desbloqueio manual de catraca, alteração de status de fatura e exclusão de cadastros na tabela `AuditLog`.

### 4. Proteção de Dados e SQLite
- **Consultas Parametrizadas**: Sempre use queries parametrizadas (GORM ou SQL puro). Proibido `db.Raw("... WHERE id = " + id)`.
- **Criptografia de Backups**: Backups gerados para envio externo (Telegram/Google Drive) devem ter opção de compressão protegida por senha ou chave GPG.
- **Permissões de Arquivo**: O banco `catraca.db` deve ter permissões estritas `0600` (apenas leitura/escrita pelo usuário do sistema operacional que executa o software).
