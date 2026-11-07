package models

import "time"

// DueNotification é um aviso persistente de vencimento de fatura. Cada fatura
// pode gerar no máximo um aviso "due" e um aviso "overdue", garantido pelo
// índice único composto.
type DueNotification struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	InvoiceID   uint       `gorm:"not null;uniqueIndex:idx_due_notification" json:"invoice_id"`
	Kind        string     `gorm:"size:20;not null;uniqueIndex:idx_due_notification" json:"kind"`
	StudentID   uint       `gorm:"index;not null" json:"student_id"`
	StudentName string     `gorm:"size:255;not null" json:"student_name"`
	AmountCents int        `gorm:"not null" json:"amount_cents"`
	DueDate     string     `gorm:"size:10;not null" json:"due_date"`
	Message     string     `gorm:"size:500;not null" json:"message"`
	ReadAt      *time.Time `json:"read_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// AvisoVencimentoConfig guarda a antecedência (em dias) usada para avisar o
// administrador antes do vencimento das faturas.
type AvisoVencimentoConfig struct {
	DiasAntecedencia int `json:"dias_antecedencia"`
}

// FaturaCobranca é uma fatura em aberto exibida no Painel de Cobranças, com o
// prazo já calculado em dias (negativo quando vencida).
type FaturaCobranca struct {
	InvoiceID     uint    `json:"invoice_id"`
	StudentID     uint    `json:"student_id"`
	Nome          string  `json:"nome"`
	CPF           string  `json:"cpf"`
	PlanoNome     string  `json:"plano_nome"`
	PlanoPreco    int     `json:"plano_preco"`
	ValorDevido   int     `json:"valor_devido"`
	DueDate       string  `json:"due_date"`
	DiasRestantes int     `json:"dias_restantes"`
	NoPrazoAviso  bool    `json:"no_prazo_aviso"`
	TotalEmAberto int     `json:"total_em_aberto"`
	Comprovante   *string `json:"comprovante"`
	StudentAtivo  bool    `json:"student_ativo"`
}

// FaturaVencendo é um item do card "Vencendo em breve" da Visão Geral.
type FaturaVencendo struct {
	InvoiceID     uint   `json:"invoice_id"`
	StudentID     uint   `json:"student_id"`
	StudentName   string `json:"student_name"`
	AmountCents   int    `json:"amount_cents"`
	DueDate       string `json:"due_date"`
	DiasRestantes int    `json:"dias_restantes"`
}
