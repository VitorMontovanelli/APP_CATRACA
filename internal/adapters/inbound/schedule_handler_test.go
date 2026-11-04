package inbound

import (
	"context"
	"testing"

	"catraca-app/internal/models"
	"catraca-app/internal/scheduling/domain"
)

func setupScheduleTest(t *testing.T) (*ScheduleHandler, *SessionManager) {
	t.Helper()

	database := setupTestDB(t)
	if err := database.AutoMigrate(
		&models.Agendamento{},
		&models.CapacidadeDia{},
		&models.Student{},
	); err != nil {
		t.Fatalf("failed to migrate schedule models: %v", err)
	}

	session := NewSessionManager(database)
	session.SetActor(1, "admin")

	handler := NewScheduleHandler(database, session)
	handler.SetContext(context.Background())
	return handler, session
}

func TestObterConfiguracaoAgendaUsaPadraoUnico(t *testing.T) {
	handler, _ := setupScheduleTest(t)

	cfg, err := handler.ObterConfiguracaoAgenda()
	if err != nil {
		t.Fatalf("ObterConfiguracaoAgenda: %v", err)
	}
	if cfg.CapacidadePadrao != domain.CapacidadeDiaPadrao {
		t.Fatalf("capacidade padrão = %d, esperado %d", cfg.CapacidadePadrao, domain.CapacidadeDiaPadrao)
	}
	if cfg.CapacidadePadrao != 48 {
		t.Fatalf("capacidade padrão = %d, esperado 48", cfg.CapacidadePadrao)
	}
	if len(cfg.Horarios) != len(domain.HorariosAula) {
		t.Fatalf("horários = %v, esperado %v", cfg.Horarios, domain.HorariosAula)
	}
	for i, h := range domain.HorariosAula {
		if cfg.Horarios[i] != h {
			t.Fatalf("horário %d = %q, esperado %q", i, cfg.Horarios[i], h)
		}
	}
}

func TestAdicionarAgendamentoComHorarioCriaDiaComCapacidadePadrao(t *testing.T) {
	handler, _ := setupScheduleTest(t)

	ag, err := handler.AdicionarAgendamentoComHorario("2026-11-03", 0, "Maria Teste", "06:00")
	if err != nil {
		t.Fatalf("AdicionarAgendamentoComHorario: %v", err)
	}
	if ag.Hora != "06:00" {
		t.Fatalf("hora = %q, esperado 06:00", ag.Hora)
	}
	if ag.Turno != "manha" {
		t.Fatalf("turno = %q, esperado manha", ag.Turno)
	}

	capacidade := handler.capacidadeDoDia("2026-11-03")
	if capacidade != domain.CapacidadeDiaPadrao {
		t.Fatalf("capacidade do dia = %d, esperado %d", capacidade, domain.CapacidadeDiaPadrao)
	}

	list, err := handler.ListarAgendamentosMes(2026, 11)
	if err != nil {
		t.Fatalf("ListarAgendamentosMes: %v", err)
	}
	if len(list) != 1 || list[0].Total != 1 || list[0].Capacidade != 48 {
		t.Fatalf("mês = %+v, esperado 2026-11-03 com 1/48", list)
	}
}

func TestAdicionarAgendamentoComHorarioRejeitaFimDeSemana(t *testing.T) {
	handler, _ := setupScheduleTest(t)

	if _, err := handler.AdicionarAgendamentoComHorario("2026-11-07", 0, "Sábado", "08:00"); err == nil {
		t.Fatal("esperava erro ao agendar no sábado")
	}
	if _, err := handler.AdicionarAgendamentoComHorario("2026-11-08", 0, "Domingo", "08:00"); err == nil {
		t.Fatal("esperava erro ao agendar no domingo")
	}
}

func TestDefinirCapacidadeDiaRejeitaFimDeSemana(t *testing.T) {
	handler, _ := setupScheduleTest(t)

	if err := handler.DefinirCapacidadeDia("2026-11-07", 20); err == nil {
		t.Fatal("esperava erro ao definir capacidade no sábado")
	}
}

func TestAtualizarHorarioAgendamento(t *testing.T) {
	handler, _ := setupScheduleTest(t)

	ag, err := handler.AdicionarAgendamentoComHorario("2026-11-03", 0, "Maria Teste", "06:00")
	if err != nil {
		t.Fatalf("AdicionarAgendamentoComHorario: %v", err)
	}

	atualizado, err := handler.AtualizarHorarioAgendamento(ag.ID, "15:00")
	if err != nil {
		t.Fatalf("AtualizarHorarioAgendamento: %v", err)
	}
	if atualizado.Hora != "15:00" || atualizado.Turno != "tarde" {
		t.Fatalf("agendamento atualizado = %+v, esperado 15:00/tarde", atualizado)
	}
}

func TestListarAgendamentosMesUsaCapacidadePadraoSemRegistro(t *testing.T) {
	handler, _ := setupScheduleTest(t)

	if err := handler.db.Create(&models.Agendamento{
		Data:  "2026-11-04",
		Nome:  "Aluno Sem Registro",
		Turno: "manha",
		Hora:  "07:00",
	}).Error; err != nil {
		t.Fatalf("criar agendamento: %v", err)
	}

	list, err := handler.ListarAgendamentosMes(2026, 11)
	if err != nil {
		t.Fatalf("ListarAgendamentosMes: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("mês = %+v, esperado um dia", list)
	}
	if list[0].Capacidade != 48 {
		t.Fatalf("capacidade = %d, esperado 48", list[0].Capacidade)
	}
}
