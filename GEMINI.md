# Diretrizes do Agente - APP_CATRACA

## 🧠 Instruções Arquiteturais & Engenharia de Software

### 1. Limite Estrito de Linhas por Arquivo (Zero Monólitos)
- **Máximo de 200 linhas** por arquivo `.go`. Arquivos com mais de 200 linhas devem ser decompostos imediatamente.
- **Máximo de 150 linhas** por componente React/TSX (`frontend/src/`).
- **Máximo de 40 linhas** por função ou método (Princípio da Responsabilidade Única - SRP).
- Proibido concentrar múltiplas entidades ou regras de negócio em um único arquivo (anti-pattern *God Object*).

### 2. Arquitetura Hexagonal (Ports & Adapters) + DDD
A base de código deve respeitar rigorosamente a separação em camadas:
- **`internal/domain/` (Domínio Puro)**:
  - Contém Entidades, Value Objects e Erros de Domínio.
  - **Zero dependências externas**: proibido importar `gorm`, `wails`, `net/http` ou banco de dados.
- **`internal/ports/` (Contratos de Interface)**:
  - Define interfaces de entrada (`Inbound Ports` / Casos de Uso) e saída (`Outbound Ports` / Repositórios e Hardware).
- **`internal/application/` (Casos de Uso)**:
  - Implementa a orquestração do fluxo de negócio. Cada caso de uso deve residir em seu próprio arquivo (`.go`).
- **`internal/adapters/inbound/` (Entrada)**:
  - Controladores Wails IPC, CLI ou handlers que apenas traduzem requisições e delegam para os casos de uso.
- **`internal/adapters/outbound/` (Saída / Infraestrutura)**:
  - Repositórios GORM/SQLite, driver de webcam (`ffmpeg`/V4L2), cliente Telegram e integrações de hardware.

### 3. Bounded Contexts (Contextos Delimitados)
Toda funcionalidade pertence a um dos seguintes contextos:
1. **Access**: Controle de acesso à catraca, leitura biométrica e logs de passagem.
2. **Student**: Cadastro de alunos, fotos de identificação e laudos.
3. **Billing**: Planos, faturas recorrentes, pagamentos e inadimplência.
4. **Scheduling**: Agenda de aulas, limites diários de capacidade e agendamentos individuais.
5. **Identity**: Autenticação, usuários administrativos e RBAC (Super Admin / Admin).
6. **System**: Backups automatizados, logs de auditoria e integrações de sistema.

### 4. Testes Unitários & TDD
- **Regra de Cobertura**: Nenhum Use Case ou regra de negócio de domínio pode ser adicionado sem seu respectivo teste unitário (`*_test.go`).
- **Isolamento Completo**: Testes unitários devem utilizar mocks das portas (repositórios em memória), sem depender de banco de dados SQLite físico ou câmera física conectada.
- **Testes de Repositório**: Testes de integração de banco devem rodar com SQLite `:memory:`.

### 5. Blindagem de Segurança (OWASP Desktop)
- **RBAC Estrito**: Todas as operações administrativas devem validar a permissão e o cargo do usuário autenticado (`a.actorCargo`).
- **Parametrização SQL**: Proibida concatenação de strings em queries do GORM (evitar SQL Injection).
- **Sanitização de Hardware**: Parâmetros passados para subprocessos (`ffmpeg`) devem validar caminhos estritos (`/dev/video*`).
- **Proteção de Senhas**: Senhas devem ser hasheadas com `bcrypt.DefaultCost`. O hash nunca deve ser exposto na API Wails (`json:"-"`).
