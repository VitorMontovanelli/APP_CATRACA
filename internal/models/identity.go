package models

import (
	"time"
)

// DefaultAdminPermissoes define as telas liberadas por padrão para novos administradores.
const DefaultAdminPermissoes = "home,agenda,alunos,planos,financeiro,cobranca,metodos_pagamento,usuarios,configuracoes"

type User struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"size:255;not null" json:"name"`
	Email         string    `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Senha         string    `gorm:"size:255;not null" json:"-"`
	Cargo         string    `gorm:"size:20;not null;check:cargo IN ('super_admin','admin')" json:"cargo"`
	RegistradorID *uint     `json:"registrador_id"`
	Ativo         bool      `gorm:"default:true" json:"ativo"`
	Foto          *string   `gorm:"size:500000" json:"foto"`
	Permissoes    string    `gorm:"size:1000;default:'home,agenda,alunos,planos,financeiro,cobranca,metodos_pagamento,usuarios,configuracoes'" json:"permissoes"`
	CriadoEm      time.Time `gorm:"autoCreateTime" json:"criado_em"`
}

type AuditLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ActorID   uint      `gorm:"not null" json:"actor_id"`
	Action    string    `gorm:"size:100;not null" json:"action"`
	TargetID  *uint     `json:"target_id"`
	Details   string    `gorm:"size:500" json:"details"`
	Timestamp time.Time `gorm:"autoCreateTime" json:"timestamp"`
}
