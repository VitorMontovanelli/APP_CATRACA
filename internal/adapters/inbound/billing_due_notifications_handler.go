package inbound

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"catraca-app/internal/models"
	toast "git.sr.ht/~jackmordaunt/go-toast/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	chaveAvisoVencimentoDias  = "aviso_vencimento_dias"
	avisoVencimentoDiasPadrao = 1
	avisoVencimentoDiasMaximo = 30
	faturasVencendoJanelaDias = 7
)

type BillingDueNotificationsHandler struct {
	ctx     context.Context
	db      *gorm.DB
	session *SessionManager
}

func NewBillingDueNotificationsHandler(db *gorm.DB, session *SessionManager) *BillingDueNotificationsHandler {
	return &BillingDueNotificationsHandler{db: db, session: session}
}

func (h *BillingDueNotificationsHandler) SetContext(ctx context.Context) {
	h.ctx = ctx
}

func (h *BillingDueNotificationsHandler) validarAdmin() error {
	if !h.session.IsAdmin() {
		return fmt.Errorf("permissão negada")
	}
	return nil
}

func (h *BillingDueNotificationsHandler) diasAntecedencia() int {
	var valor string
	h.db.Model(&models.Setting{}).Where("key = ?", chaveAvisoVencimentoDias).Pluck("value", &valor)
	dias, err := strconv.Atoi(strings.TrimSpace(valor))
	if err != nil || dias < 0 {
		return avisoVencimentoDiasPadrao
	}
	if dias > avisoVencimentoDiasMaximo {
		return avisoVencimentoDiasMaximo
	}
	return dias
}

func (h *BillingDueNotificationsHandler) ObterConfigAvisoVencimento() (*models.AvisoVencimentoConfig, error) {
	if err := h.validarAdmin(); err != nil {
		return nil, err
	}
	return &models.AvisoVencimentoConfig{DiasAntecedencia: h.diasAntecedencia()}, nil
}

func (h *BillingDueNotificationsHandler) SalvarConfigAvisoVencimento(dias int) error {
	if err := h.validarAdmin(); err != nil {
		return err
	}
	if dias < 0 || dias > avisoVencimentoDiasMaximo {
		return fmt.Errorf("a antecedência deve estar entre 0 e %d dias", avisoVencimentoDiasMaximo)
	}
	err := h.db.Exec(
		"INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value",
		chaveAvisoVencimentoDias, strconv.Itoa(dias),
	).Error
	if err == nil {
		h.session.RegistrarAuditComDetalhes("configurou_aviso_vencimento", 0, fmt.Sprintf("dias=%d", dias))
	}
	return err
}

// VerificarAvisosVencimento cria os avisos que ainda não existem e devolve a
// lista atual de avisos ativos. É idempotente: pode ser chamada ao abrir o app
// e a cada hora sem duplicar notificações.
func (h *BillingDueNotificationsHandler) VerificarAvisosVencimento() ([]models.DueNotification, error) {
	if err := h.validarAdmin(); err != nil {
		return nil, err
	}
	if _, err := h.gerarAvisosPendentes(); err != nil {
		return nil, err
	}
	return h.ListarAvisosVencimento()
}

func (h *BillingDueNotificationsHandler) gerarAvisosPendentes() (int, error) {
	antecedencia := h.diasAntecedencia()
	hoje := dataSemFuso(time.Now())

	type candidato struct {
		InvoiceID   uint
		StudentID   uint
		StudentName string
		AmountCents int
		DueDate     string
	}
	var pendentes []candidato
	err := h.db.Raw(`
		SELECT i.id AS invoice_id, i.student_id, s.nome AS student_name,
		       i.amount_cents, i.due_date
		FROM invoices i
		JOIN students s ON s.id = i.student_id
		WHERE i.status IN ('pending', 'overdue') AND s.ativo = 1
		ORDER BY i.due_date ASC, i.id ASC`).Scan(&pendentes).Error
	if err != nil {
		return 0, err
	}

	criados := 0
	for _, inv := range pendentes {
		vencimento, err := parseData(inv.DueDate)
		if err != nil {
			continue
		}
		restantes := diasEntre(vencimento, hoje)

		var kind, mensagem string
		switch {
		case restantes < 0:
			kind = "overdue"
			atraso := -restantes
			if atraso == 1 {
				mensagem = fmt.Sprintf("%s — %s venceu há 1 dia, %s",
					inv.StudentName, formatarReais(inv.AmountCents), formatarDiaMes(inv.DueDate))
			} else {
				mensagem = fmt.Sprintf("%s — %s venceu há %d dias, %s",
					inv.StudentName, formatarReais(inv.AmountCents), atraso, formatarDiaMes(inv.DueDate))
			}
		case restantes <= antecedencia:
			abertura := ajustarParaDiaUtil(vencimento.AddDate(0, 0, -antecedencia))
			if hoje.Before(abertura) {
				continue
			}
			kind = "due"
			mensagem = fmt.Sprintf("%s — %s %s, %s",
				inv.StudentName, formatarReais(inv.AmountCents),
				descricaoVencimento(restantes), formatarDiaMes(inv.DueDate))
		default:
			continue
		}

		notificacao := models.DueNotification{
			InvoiceID:   inv.InvoiceID,
			Kind:        kind,
			StudentID:   inv.StudentID,
			StudentName: inv.StudentName,
			AmountCents: inv.AmountCents,
			DueDate:     inv.DueDate,
			Message:     mensagem,
		}
		resultado := h.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&notificacao)
		if resultado.Error != nil {
			continue
		}
		if resultado.RowsAffected > 0 {
			criados++
			// Quando a fatura vence, o aviso de "vence hoje/amanhã" deixa de
			// fazer sentido: mantém apenas o aviso de atraso.
			if kind == "overdue" {
				h.db.Where("invoice_id = ? AND kind = ?", inv.InvoiceID, "due").Delete(&models.DueNotification{})
			}
		}
	}

	if criados > 0 {
		go notificarWindows(criados)
	}
	return criados, nil
}

// ListarAvisosVencimento devolve somente avisos de faturas ainda em aberto e
// de alunos ativos. Faturas pagas, canceladas ou de alunos desativados somem
// automaticamente da lista.
func (h *BillingDueNotificationsHandler) ListarAvisosVencimento() ([]models.DueNotification, error) {
	if err := h.validarAdmin(); err != nil {
		return nil, err
	}
	hoje := dataSemFuso(time.Now()).Format("2006-01-02")
	var lista []models.DueNotification
	err := h.db.Raw(`
		SELECT dn.id, dn.invoice_id, dn.kind, dn.student_id, dn.student_name,
		       dn.amount_cents, dn.due_date, dn.message, dn.read_at, dn.created_at
		FROM due_notifications dn
		JOIN invoices i ON i.id = dn.invoice_id
		JOIN students s ON s.id = dn.student_id
		WHERE i.status IN ('pending', 'overdue') AND s.ativo = 1
		  AND NOT (dn.kind = 'due' AND i.due_date < ?)
		ORDER BY CASE WHEN dn.read_at IS NULL THEN 0 ELSE 1 END,
		         dn.created_at DESC, dn.id DESC`, hoje).Scan(&lista).Error
	if err != nil {
		return nil, err
	}
	if lista == nil {
		lista = []models.DueNotification{}
	}
	return lista, nil
}

func (h *BillingDueNotificationsHandler) MarcarAvisoComoLido(id uint) error {
	if err := h.validarAdmin(); err != nil {
		return err
	}
	return h.db.Model(&models.DueNotification{}).
		Where("id = ? AND read_at IS NULL", id).
		Update("read_at", time.Now()).Error
}

func (h *BillingDueNotificationsHandler) MarcarTodosAvisosComoLidos() error {
	if err := h.validarAdmin(); err != nil {
		return err
	}
	return h.db.Model(&models.DueNotification{}).
		Where("read_at IS NULL").
		Update("read_at", time.Now()).Error
}

// ListarFaturasVencendo alimenta o card "Vencendo em breve" da Visão Geral
// com as faturas em aberto dos próximos 7 dias.
func (h *BillingDueNotificationsHandler) ListarFaturasVencendo() ([]models.FaturaVencendo, error) {
	if err := h.validarAdmin(); err != nil {
		return nil, err
	}
	hoje := dataSemFuso(time.Now())
	limite := hoje.AddDate(0, 0, faturasVencendoJanelaDias)

	type linha struct {
		InvoiceID   uint
		StudentID   uint
		StudentName string
		AmountCents int
		DueDate     string
	}
	var rows []linha
	err := h.db.Raw(`
		SELECT i.id AS invoice_id, i.student_id, s.nome AS student_name,
		       i.amount_cents, i.due_date
		FROM invoices i
		JOIN students s ON s.id = i.student_id
		WHERE i.status IN ('pending', 'overdue') AND s.ativo = 1
		  AND i.due_date >= ? AND i.due_date <= ?
		ORDER BY i.due_date ASC, i.id ASC`,
		hoje.Format("2006-01-02"), limite.Format("2006-01-02")).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	lista := make([]models.FaturaVencendo, 0, len(rows))
	for _, r := range rows {
		vencimento, err := parseData(r.DueDate)
		if err != nil {
			continue
		}
		lista = append(lista, models.FaturaVencendo{
			InvoiceID:     r.InvoiceID,
			StudentID:     r.StudentID,
			StudentName:   r.StudentName,
			AmountCents:   r.AmountCents,
			DueDate:       r.DueDate,
			DiasRestantes: diasEntre(vencimento, hoje),
		})
	}
	return lista, nil
}

func notificarWindows(quantidade int) {
	n := toast.Notification{
		AppID: "CatracaVMD",
		Title: "Avisos de vencimento",
		Body:  fmt.Sprintf("%d fatura(s) precisam de atenção. Abra o CatracaVMD para ver os detalhes.", quantidade),
	}
	_ = n.Push()
}

func dataSemFuso(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func parseData(valor string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", strings.TrimSpace(valor))
	if err != nil {
		return time.Time{}, err
	}
	return dataSemFuso(t), nil
}

func diasEntre(a, b time.Time) int {
	return int(dataSemFuso(a).Sub(dataSemFuso(b)).Hours() / 24)
}

// ajustarParaDiaUtil antecipa sábado/domingo para a sexta anterior, já que o
// estúdio funciona de segunda a sexta.
func ajustarParaDiaUtil(t time.Time) time.Time {
	switch t.Weekday() {
	case time.Saturday:
		return t.AddDate(0, 0, -1)
	case time.Sunday:
		return t.AddDate(0, 0, -2)
	}
	return t
}

func descricaoVencimento(dias int) string {
	switch dias {
	case 0:
		return "vence hoje"
	case 1:
		return "vence amanhã"
	default:
		return fmt.Sprintf("vence em %d dias", dias)
	}
}

func formatarDiaMes(data string) string {
	if len(data) >= 10 {
		return data[8:10] + "/" + data[5:7]
	}
	return data
}

func formatarReais(cents int) string {
	sinal := ""
	if cents < 0 {
		sinal = "-"
		cents = -cents
	}
	inteiros := strconv.Itoa(cents / 100)
	var comSeparador strings.Builder
	for i, digito := range inteiros {
		if i > 0 && (len(inteiros)-i)%3 == 0 {
			comSeparador.WriteByte('.')
		}
		comSeparador.WriteRune(digito)
	}
	return fmt.Sprintf("%sR$ %s,%02d", sinal, comSeparador.String(), cents%100)
}
