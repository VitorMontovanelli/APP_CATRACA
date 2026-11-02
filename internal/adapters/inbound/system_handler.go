package inbound

import (
	"context"

	"gorm.io/gorm"
)

type SystemHandler struct {
	*SystemBackupHandler
	*SystemMetricsHandler
	*SystemTelegramHandler
}

func NewSystemHandler(db *gorm.DB, session *SessionManager) *SystemHandler {
	backup := NewSystemBackupHandler(db, session)
	metrics := NewSystemMetricsHandler(db, session)
	telegram := NewSystemTelegramHandler(db, session, backup)

	return &SystemHandler{
		SystemBackupHandler:   backup,
		SystemMetricsHandler:  metrics,
		SystemTelegramHandler: telegram,
	}
}

func (h *SystemHandler) SetContext(ctx context.Context) {
	h.SystemBackupHandler.SetContext(ctx)
	h.SystemMetricsHandler.SetContext(ctx)
	h.SystemTelegramHandler.SetContext(ctx)
}
