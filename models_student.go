package main

import (
	"time"
)

type Student struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	Nome             string    `gorm:"size:255;not null" json:"nome"`
	CPF              string    `gorm:"size:14;uniqueIndex" json:"cpf"`
	DataNascimento   *string   `gorm:"size:10" json:"data_nascimento"`
	Telefone         *string   `gorm:"size:20" json:"telefone"`
	TelefoneUrgencia *string   `gorm:"size:20" json:"telefone_urgencia"`
	Email            *string   `gorm:"size:255" json:"email"`
	LaudoMedico      *string   `gorm:"size:500" json:"laudo_medico"`
	Foto             *string   `gorm:"size:500" json:"foto"`
	FormaPagamentoID *uint     `json:"forma_pagamento_id"`
	DataEntrada      time.Time `gorm:"autoCreateTime" json:"data_entrada"`
	Observacao       *string   `gorm:"size:500" json:"observacao"`
	Ativo            bool      `gorm:"default:true" json:"ativo"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
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
