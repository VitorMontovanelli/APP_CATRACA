package inbound

import (
	"context"

	"gorm.io/gorm"
)

type StudentHandler struct {
	*StudentCrudHandler
	*StudentPlansHandler
	*StudentLaudosHandler
	*StudentPhotosHandler
}

func NewStudentHandler(db *gorm.DB, session *SessionManager) *StudentHandler {
	return &StudentHandler{
		StudentCrudHandler:   NewStudentCrudHandler(db, session),
		StudentPlansHandler:  NewStudentPlansHandler(db, session),
		StudentLaudosHandler: NewStudentLaudosHandler(db, session),
		StudentPhotosHandler: NewStudentPhotosHandler(db, session),
	}
}

func (h *StudentHandler) SetContext(ctx context.Context) {
	h.StudentCrudHandler.SetContext(ctx)
	h.StudentPlansHandler.SetContext(ctx)
	h.StudentLaudosHandler.SetContext(ctx)
	h.StudentPhotosHandler.SetContext(ctx)
}
