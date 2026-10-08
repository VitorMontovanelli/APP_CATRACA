package main

import (
	"time"
)

type Plan struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	Name             string     `gorm:"size:255;not null" json:"name"`
	Description      string     `gorm:"size:500" json:"description"`
	DurationDays     int        `gorm:"not null" json:"duration_days"`
	PriceCents       int        `gorm:"not null" json:"price_cents"`
	PrecoCartaoCents *int       `gorm:"default:null" json:"preco_cartao_cents"`
	GracePeriodDays  int        `gorm:"default:5" json:"grace_period_days"`
	Active           bool       `gorm:"default:true" json:"active"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        *time.Time `json:"updated_at"`
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
	Type      string    `gorm:"size:20;not null" json:"type"`
	Enabled   bool      `gorm:"default:true" json:"enabled"`
	Config    *string   `gorm:"size:2000" json:"config"`
	CreatedAt time.Time `json:"created_at"`
}

type StudentPlan struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	StudentID       uint       `gorm:"index;not null" json:"student_id"`
	PlanID          uint       `gorm:"not null" json:"plan_id"`
	Status          string     `gorm:"size:20;default:active" json:"status"`
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
	Status               string     `gorm:"size:20;default:pending" json:"status"`
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

type InadimplenteReport struct {
	StudentID       uint    `json:"student_id"`
	Nome            string  `json:"nome"`
	CPF             string  `json:"cpf"`
	Telefone        *string `json:"telefone"`
	Email           *string `json:"email"`
	PlanoNome       string  `json:"plano_nome"`
	PlanoPreco      int     `json:"plano_preco"`
	StatusPlano     string  `json:"status_plano"`
	InvoiceID       uint    `json:"invoice_id"`
	ValorDevido     int     `json:"valor_devido"`
	DataVencimento  string  `json:"data_vencimento"`
	DiasVencido     int     `json:"dias_vencido"`
	UltimoPagamento *string `json:"ultimo_pagamento"`
	TotalEmAberto   int     `json:"total_em_aberto"`
}
