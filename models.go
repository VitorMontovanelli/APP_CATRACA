package main

import (
	"time"

	"gorm.io/gorm"
)

// ==== Catraca Virtual (Novo) ====

type Aluno struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Nome           string    `gorm:"size:255;not null" json:"nome"`
	CPF            string    `gorm:"size:14;uniqueIndex" json:"cpf"`
	Status         bool      `gorm:"default:true" json:"status"`
	VencimentoPlano time.Time `gorm:"not null" json:"vencimento_plano"`
	IDBiometriaMock int      `gorm:"uniqueIndex;default:0" json:"id_biometria_mock"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type RegistroAcesso struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	AlunoID   uint      `gorm:"index;not null" json:"aluno_id"`
	Aluno     Aluno     `gorm:"foreignKey:AlunoID" json:"-"`
	DataHora  time.Time `gorm:"autoCreateTime" json:"data_hora"`
	Liberado  bool      `gorm:"not null" json:"liberado"`
	Motivo    string    `gorm:"size:255" json:"motivo"`
}

type ResultadoAcesso struct {
	Nome     string `json:"nome"`
	Liberado bool   `json:"liberado"`
	Mensagem string `json:"mensagem"`
}

// ==== Agenda Calendário ====

// Agendamento representa a reserva de uma vaga para um aluno em um dia e turno.
type Agendamento struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	Data       string     `gorm:"size:10;index;not null" json:"data"` // formato YYYY-MM-DD
	StudentID  *uint      `gorm:"index" json:"student_id"`            // nil quando o aluno não está cadastrado
	Nome       string     `gorm:"size:255;not null" json:"nome"`
	Turno      string     `gorm:"size:10;not null;check:turno IN ('manha','tarde','noite')" json:"turno"`
	Telefone   *string    `gorm:"size:20" json:"telefone"`
	Observacao *string    `gorm:"size:500" json:"observacao"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

// DiaAgenda resume a ocupação de um dia do calendário.
type DiaAgenda struct {
	Data       string `json:"data"`
	Total      int    `json:"total"`
	Capacidade int    `json:"capacidade"`
}

// CapacidadeDia guarda a capacidade de vagas customizada de um dia específico.
type CapacidadeDia struct {
	Data       string `gorm:"primaryKey;size:10" json:"data"` // formato YYYY-MM-DD
	Capacidade int    `gorm:"not null" json:"capacidade"`
}

type InadimplenteReport struct {
	StudentID        uint    `json:"student_id"`
	Nome             string  `json:"nome"`
	CPF              string  `json:"cpf"`
	Telefone         *string `json:"telefone"`
	Email            *string `json:"email"`
	PlanoNome        string  `json:"plano_nome"`
	PlanoPreco       int     `json:"plano_preco"`
	StatusPlano      string  `json:"status_plano"`
	InvoiceID        uint    `json:"invoice_id"`
	ValorDevido      int     `json:"valor_devido"`
	DataVencimento   string  `json:"data_vencimento"`
	DiasVencido      int     `json:"dias_vencido"`
	UltimoPagamento  *string `json:"ultimo_pagamento"`
	TotalEmAberto    int     `json:"total_em_aberto"`
}

// ==== Usuários ====

type User struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"size:255;not null" json:"name"`
	Email         string    `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Senha         string    `gorm:"size:255;not null" json:"-"`
	Cargo         string    `gorm:"size:20;not null;check:cargo IN ('super_admin','admin')" json:"cargo"`
	RegistradorID *uint     `json:"registrador_id"`
	Ativo         bool      `gorm:"default:true" json:"ativo"`
	Foto          *string   `gorm:"size:500000" json:"foto"`
	Permissoes    string    `gorm:"size:1000;default:'home,alunos,planos,financeiro,cobranca,metodos_pagamento,usuarios,logs'" json:"permissoes"`
	CriadoEm      time.Time `gorm:"autoCreateTime" json:"criado_em"`
}

// ==== Alunos (Students - legado) ====

type Student struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Nome            string    `gorm:"size:255;not null" json:"nome"`
	CPF             string    `gorm:"size:14;uniqueIndex" json:"cpf"`
	DataNascimento  *string   `gorm:"size:10" json:"data_nascimento"`
	Telefone        *string   `gorm:"size:20" json:"telefone"`
	Email           *string   `gorm:"size:255" json:"email"`
	FormaPagamentoID *uint    `json:"forma_pagamento_id"`
	DataEntrada     time.Time `gorm:"autoCreateTime" json:"data_entrada"`
	Observacao      *string   `gorm:"size:500" json:"observacao"`
	Ativo           bool      `gorm:"default:true" json:"ativo"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type StudentComPlano struct {
	Student
	PlanoNome        *string `json:"plano_nome"`
	PlanoPreco       *int    `json:"plano_preco"`
	PlanoStatus      *string `json:"plano_status"`
	StudentPlanID    *uint   `json:"student_plan_id"`
	PaymentMethod    *string `json:"payment_method"`
	DueDay           *int    `json:"due_day"`
	UltimaFatura     *string `json:"ultima_fatura"`
	FaturaValor      *int    `json:"fatura_valor"`
	FaturaVencimento *string `json:"fatura_vencimento"`
}

type Plan struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	Name            string     `gorm:"size:255;not null" json:"name"`
	Description     string     `gorm:"size:500" json:"description"`
	DurationDays    int        `gorm:"not null" json:"duration_days"`
	PriceCents      int        `gorm:"not null" json:"price_cents"`
	GracePeriodDays int        `gorm:"default:5" json:"grace_period_days"`
	Active          bool       `gorm:"default:true" json:"active"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
}

type PaymentMethod struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"size:255;not null" json:"name"`
	Type          string    `gorm:"size:20;not null;check:type IN ('pix','credit_card','boleto','debit_card','wallet','cash')" json:"type"`
	Enabled       bool      `gorm:"default:true" json:"enabled"`
	FeePercent    float64   `gorm:"default:0" json:"fee_percent"`
	FeeFixedCents int       `gorm:"default:0" json:"fee_fixed_cents"`
	CreatedAt     time.Time `json:"created_at"`
}

type PaymentGateway struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	Type      string    `gorm:"size:20;not null;check:type IN ('mercadopago','picpay','pagseguro','asaas','stone','cielo','vindi','pagar_me','iugu','galax_pay')" json:"type"`
	Enabled   bool      `gorm:"default:true" json:"enabled"`
	Config    *string   `gorm:"size:2000" json:"config"`
	CreatedAt time.Time `json:"created_at"`
}

type StudentPlan struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	StudentID       uint       `gorm:"index;not null" json:"student_id"`
	PlanID          uint       `gorm:"not null" json:"plan_id"`
	Status          string     `gorm:"size:20;default:active;check:status IN ('active','overdue','suspended','cancelled','expired')" json:"status"`
	PaymentMethodID *uint      `json:"payment_method_id"`
	DueDay          int        `gorm:"default:5" json:"due_day"`
	StartDate       time.Time  `gorm:"autoCreateTime" json:"start_date"`
	EndDate         *time.Time `json:"end_date"`
	CancelledAt     *time.Time `json:"cancelled_at"`
	Notes           *string    `gorm:"size:500" json:"notes"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
	StudentName     string     `gorm:"-" json:"student_name"`
	PlanName        string     `gorm:"-" json:"plan_name"`
	PlanPriceCents  int        `gorm:"-" json:"plan_price_cents"`
}

type Invoice struct {
	ID                   uint       `gorm:"primaryKey" json:"id"`
	StudentPlanID        uint       `gorm:"not null" json:"student_plan_id"`
	StudentID            uint       `gorm:"index;not null" json:"student_id"`
	PlanID               uint       `gorm:"not null" json:"plan_id"`
	AmountCents          int        `gorm:"not null" json:"amount_cents"`
	Status               string     `gorm:"size:20;default:pending;check:status IN ('pending','paid','overdue','cancelled','refunded')" json:"status"`
	PaymentMethodID      *uint      `json:"payment_method_id"`
	DueDate              string     `gorm:"size:10;not null" json:"due_date"`
	PaidAt               *time.Time `json:"paid_at"`
	PaidAmountCents      *int       `json:"paid_amount_cents"`
	GatewayTransactionID *string    `gorm:"size:255" json:"gateway_transaction_id"`
	PixQrCode            *string    `gorm:"size:2000" json:"pix_qr_code"`
	PixBrCode            *string    `gorm:"size:500" json:"pix_br_code"`
	Comprovante          *string    `gorm:"size:500" json:"comprovante"`
	Notes                *string    `gorm:"size:1000" json:"notes"`
	CreatedAt            time.Time  `json:"created_at"`
	StudentName          string     `gorm:"-" json:"student_name"`
	PlanName             string     `gorm:"-" json:"plan_name"`
	PaymentMethodName    string     `gorm:"-" json:"payment_method_name"`
}

type AccessLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    *uint     `json:"user_id"`
	Timestamp time.Time `gorm:"autoCreateTime" json:"timestamp"`
	Resultado string    `gorm:"size:20;not null;check:resultado IN ('liberado','negado')" json:"resultado"`
	Motivo    string    `gorm:"size:255" json:"motivo"`
}

type AuditLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ActorID   uint      `gorm:"not null" json:"actor_id"`
	Action    string    `gorm:"size:100;not null" json:"action"`
	TargetID  *uint     `json:"target_id"`
	Details   string    `gorm:"size:500" json:"details"`
	Timestamp time.Time `gorm:"autoCreateTime" json:"timestamp"`
}

// ==== Configurações & Backup Telegram ====

type Setting struct {
	Key   string `gorm:"primaryKey;size:100" json:"key"`
	Value string `gorm:"size:2000" json:"value"`
}

type BackupConfig struct {
	Token      string `json:"token"`
	ChatID     string `json:"chat_id"`
	AutoBackup bool   `json:"auto_backup"`
	LastBackup string `json:"last_backup"`
	UpdatedAt  string `json:"updated_at"`
}

// BeforeCreate hook para alunos da catraca: gera vencimento padrão (30 dias) se não definido
func (a *Aluno) BeforeCreate(tx *gorm.DB) error {
	if a.VencimentoPlano.IsZero() {
		a.VencimentoPlano = time.Now().AddDate(0, 1, 0)
	}
	return nil
}
