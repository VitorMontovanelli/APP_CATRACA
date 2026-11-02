package inbound

import (
	"context"

	"gorm.io/gorm"
)

type BillingHandler struct {
	*BillingPlansHandler
	*BillingInvoicesHandler
	*BillingRecurringHandler
	*BillingReportsHandler
}

func NewBillingHandler(db *gorm.DB, session *SessionManager) *BillingHandler {
	recurring := NewBillingRecurringHandler(db, session)
	plans := NewBillingPlansHandler(db, session)
	invoices := NewBillingInvoicesHandler(db, session, recurring)
	reports := NewBillingReportsHandler(db, session)

	return &BillingHandler{
		BillingPlansHandler:     plans,
		BillingInvoicesHandler:  invoices,
		BillingRecurringHandler: recurring,
		BillingReportsHandler:   reports,
	}
}

func (h *BillingHandler) SetContext(ctx context.Context) {
	h.BillingPlansHandler.SetContext(ctx)
	h.BillingInvoicesHandler.SetContext(ctx)
	h.BillingRecurringHandler.SetContext(ctx)
	h.BillingReportsHandler.SetContext(ctx)
}
