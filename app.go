package main

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type App struct {
	ctx        context.Context
	db         *gorm.DB
	actorCargo string
	actorID    uint
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.db = initDB()
}

func ehAdmin(cargo string) bool {
	return cargo == "super_admin" || cargo == "admin"
}

func (a *App) registrarAudit(action string, targetID uint) {
	a.registrarAuditComDetalhes(action, targetID, "")
}

func (a *App) registrarAuditComDetalhes(action string, targetID uint, details string) {
	a.db.Create(&AuditLog{
		ActorID:   a.actorID,
		Action:    action,
		TargetID:  &targetID,
		Details:   details,
		Timestamp: time.Now(),
	})
}

func (a *App) onShutdown() {
	config, err := a.ObterConfigTelegram()
	if err != nil || config == nil || !config.AutoBackup {
		return
	}
	_, _ = a.enviarBackupTelegram()
}

func (a *App) AppVersion() string {
	return "2.1.0"
}
