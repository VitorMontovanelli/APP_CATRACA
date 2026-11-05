package inbound

import (
	"context"
	"testing"

	"catraca-app/internal/models"
	"catraca-app/internal/scheduling/domain"
)

func setupScheduleIndividualTest(t *testing.T) (*ScheduleHandler, *ScheduleIndividualCrudHandler) {
	t.Helper()

	database := setupTestDB(t)
	if err := database.AutoMigrate(
		&models.Agendamento{},
		&models.AgendamentoIndividual{},
		&models.CapacidadeDia{},
		&models.Student{},
	); err != nil {
		t.Fatalf("failed to migrate schedule models: %v", err)
	}

	session := NewSessionManager(database)
	session.SetActor(1, "admin")

	schedule := NewScheduleHandler(database, session)
	schedule.SetContext(context.Background())

	individual := NewScheduleIndividualCrudHandler(database, session)
	individual.SetContext(context.Background())

	return schedule, individual
}

func TestCapacidadeSeparadaPorCalendario(t *testing.T) {
	schedule, _ := setupScheduleIndividualTest(t)
	data := "2026-11-02"

	if err := schedule.DefinirCapacidadeDiaCalendario(domain.CalendarioPilates, data, 30); err != nil {
		t.Fatalf("definir capacidade do pilates: %v", err)
	}
	if got := capacidadeDoCalendario(schedule.db, domain.CalendarioPilates, data); got != 30 {
		t.Fatalf("capacidade do pilates = %d, esperado 30", got)
	}
	if got := capacidadeDoCalendario(schedule.db, models.PraticaVentosaterapia, data); got != domain.CapacidadeIndividualPadrao {
		t.Fatalf("capacidade individual = %d, esperado %d", got, domain.CapacidadeIndividualPadrao)
	}

	if err := schedule.DefinirCapacidadeDiaCalendario(models.PraticaVentosaterapia, data, 4); err != nil {
		t.Fatalf("definir capacidade individual: %v", err)
	}
	if got := capacidadeDoCalendario(schedule.db, models.PraticaVentosaterapia, data); got != 4 {
		t.Fatalf("capacidade individual = %d, esperado 4", got)
	}
	if got := capacidadeDoCalendario(schedule.db, domain.CalendarioPilates, data); got != 30 {
		t.Fatalf("capacidade do pilates foi alterada para %d, esperado 30", got)
	}
	if got := capacidadeDoCalendario(schedule.db, models.PraticaLiberacaoMiofascial, data); got != domain.CapacidadeIndividualPadrao {
		t.Fatalf("capacidade de outro calendário individual = %d, esperado %d", got, domain.CapacidadeIndividualPadrao)
	}
}

func TestConfiguracaoPorCalendarioTemPadroesDistintos(t *testing.T) {
	schedule, _ := setupScheduleIndividualTest(t)

	pilates, err := schedule.ObterConfiguracaoAgendaCalendario(domain.CalendarioPilates)
	if err != nil {
		t.Fatalf("configuração do pilates: %v", err)
	}
	if pilates.CapacidadePadrao != 48 {
		t.Fatalf("capacidade padrão do pilates = %d, esperado 48", pilates.CapacidadePadrao)
	}
	if pilates.CapacidadeMaxima != 0 {
		t.Fatalf("capacidade máxima do pilates = %d, esperado sem limite", pilates.CapacidadeMaxima)
	}

	individual, err := schedule.ObterConfiguracaoAgendaCalendario(models.PraticaPersonalTrainer)
	if err != nil {
		t.Fatalf("configuração individual: %v", err)
	}
	if individual.CapacidadePadrao != 10 || individual.CapacidadeMaxima != 10 {
		t.Fatalf("padrão/máximo individual = %d/%d, esperado 10/10", individual.CapacidadePadrao, individual.CapacidadeMaxima)
	}
	if len(individual.Horarios) != len(domain.HorariosAula) {
		t.Fatalf("horários individuais = %v, esperado a grade única %v", individual.Horarios, domain.HorariosAula)
	}
}

func TestDefinirCapacidadeCalendarioRejeitaFimDeSemanaEAcimaDoLimite(t *testing.T) {
	schedule, _ := setupScheduleIndividualTest(t)

	if err := schedule.DefinirCapacidadeDiaCalendario(models.PraticaKinesioTape, "2026-11-07", 5); err == nil {
		t.Fatal("esperava erro ao definir capacidade no sábado")
	}
	if err := schedule.DefinirCapacidadeDiaCalendario(models.PraticaKinesioTape, "2026-11-02", 11); err == nil {
		t.Fatal("esperava erro ao definir capacidade individual acima de 10")
	}
	if err := schedule.DefinirCapacidadeDiaCalendario(domain.CalendarioPilates, "2026-11-02", 100); err != nil {
		t.Fatalf("o pilates não deveria ter limite máximo: %v", err)
	}
}

func TestAgendamentoIndividualRejeitaFimDeSemana(t *testing.T) {
	_, individual := setupScheduleIndividualTest(t)

	if _, err := individual.AdicionarAgendamentoIndividual("2026-11-07", 0, "Sábado", "08:00", models.PraticaVentosaterapia, ""); err == nil {
		t.Fatal("esperava erro ao agendar no sábado")
	}
	if _, err := individual.AdicionarAgendamentoIndividual("2026-11-08", 0, "Domingo", "08:00", models.PraticaVentosaterapia, ""); err == nil {
		t.Fatal("esperava erro ao agendar no domingo")
	}
}

func TestConflitoDeHorarioIndividualEPorPratica(t *testing.T) {
	_, individual := setupScheduleIndividualTest(t)
	data := "2026-11-03"

	if _, err := individual.AdicionarAgendamentoIndividual(data, 0, "Maria", "08:00", models.PraticaVentosaterapia, ""); err != nil {
		t.Fatalf("primeiro agendamento: %v", err)
	}
	if _, err := individual.AdicionarAgendamentoIndividual(data, 0, "João", "08:00", models.PraticaLiberacaoMiofascial, ""); err != nil {
		t.Fatalf("mesmo horário em outra prática deveria ser permitido: %v", err)
	}
	if _, err := individual.AdicionarAgendamentoIndividual(data, 0, "Ana", "08:00", models.PraticaVentosaterapia, ""); err == nil {
		t.Fatal("esperava erro para o mesmo horário na mesma prática")
	}
}

func TestCapacidadeIndividualLimitaAgendamentos(t *testing.T) {
	schedule, individual := setupScheduleIndividualTest(t)
	data := "2026-11-04"

	if err := schedule.DefinirCapacidadeDiaCalendario(models.PraticaPersonalTrainer, data, 1); err != nil {
		t.Fatalf("definir capacidade: %v", err)
	}
	if _, err := individual.AdicionarAgendamentoIndividual(data, 0, "Maria", "08:00", models.PraticaPersonalTrainer, ""); err != nil {
		t.Fatalf("primeiro agendamento: %v", err)
	}
	if _, err := individual.AdicionarAgendamentoIndividual(data, 0, "João", "09:00", models.PraticaPersonalTrainer, ""); err == nil {
		t.Fatal("esperava erro ao ultrapassar a capacidade do calendário")
	}
}

func TestAtualizarIndividualNaoConflitaEntrePraticas(t *testing.T) {
	_, individual := setupScheduleIndividualTest(t)
	data := "2026-11-05"

	if _, err := individual.AdicionarAgendamentoIndividual(data, 0, "Maria", "08:00", models.PraticaVentosaterapia, ""); err != nil {
		t.Fatalf("agendamento de ventosaterapia: %v", err)
	}
	ag, err := individual.AdicionarAgendamentoIndividual(data, 0, "João", "09:00", models.PraticaLiberacaoMiofascial, "")
	if err != nil {
		t.Fatalf("agendamento de liberação miofascial: %v", err)
	}

	if _, err := individual.AtualizarAgendamentoIndividual(ag.ID, "08:00", models.PraticaLiberacaoMiofascial, ""); err != nil {
		t.Fatalf("mover para o mesmo horário de outra prática deveria ser permitido: %v", err)
	}
}
