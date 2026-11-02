package main

import (
	"context"

	"catraca-app/internal/adapters/inbound"
	"catraca-app/internal/infra/db"
	"gorm.io/gorm"
)

type App struct {
	ctx context.Context
	db  *gorm.DB

	*inbound.AccessHandler
	*inbound.BillingHandler
	*inbound.IdentityHandler
	*inbound.ScheduleHandler
	*inbound.StudentHandler
	*inbound.SystemHandler
}

func NewApp() *App {
	database := db.InitDB()
	session := inbound.NewSessionManager(database)

	return &App{
		db:              database,
		AccessHandler:   inbound.NewAccessHandler(database, session),
		BillingHandler:  inbound.NewBillingHandler(database, session),
		IdentityHandler: inbound.NewIdentityHandler(database, session),
		ScheduleHandler: inbound.NewScheduleHandler(database, session),
		StudentHandler:  inbound.NewStudentHandler(database, session),
		SystemHandler:   inbound.NewSystemHandler(database, session),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.AccessHandler.SetContext(ctx)
	a.BillingHandler.SetContext(ctx)
	a.IdentityHandler.SetContext(ctx)
	a.ScheduleHandler.SetContext(ctx)
	a.StudentHandler.SetContext(ctx)
	a.SystemHandler.SetContext(ctx)
}

func (a *App) onShutdown() {
	a.SystemHandler.OnShutdown()
}

func (a *App) AppVersion() string {
	return Version
}
