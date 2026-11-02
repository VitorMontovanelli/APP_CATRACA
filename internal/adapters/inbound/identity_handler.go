package inbound

import (
	"context"

	"gorm.io/gorm"
)

type IdentityHandler struct {
	*IdentityAuthHandler
	*IdentityUsersHandler
}

func NewIdentityHandler(db *gorm.DB, session *SessionManager) *IdentityHandler {
	return &IdentityHandler{
		IdentityAuthHandler:  NewIdentityAuthHandler(db, session),
		IdentityUsersHandler: NewIdentityUsersHandler(db, session),
	}
}

func (h *IdentityHandler) SetContext(ctx context.Context) {
	h.IdentityAuthHandler.SetContext(ctx)
	h.IdentityUsersHandler.SetContext(ctx)
}
