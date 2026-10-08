package service_test

import (
	"context"
	"errors"
	"testing"

	"catraca-app/internal/scheduling/domain"
	"catraca-app/internal/scheduling/service"
)

type mockScheduleRepo struct {
	capacidades map[string]int
	agendas     map[string][]*domain.Agendamento
	individuais map[string][]*domain.AgendamentoIndividual
}

func newMockScheduleRepo() *mockScheduleRepo {
	return &mockScheduleRepo{
		capacidades: make(map[string]int),
		agendas:     make(map[string][]*domain.Agendamento),
		individuais: make(map[string][]*domain.AgendamentoIndividual),
	}
}

func (m *mockScheduleRepo) ObterCapacidadeDia(ctx context.Context, data string) (int, error) {
	if c, ok := m.capacidades[data]; ok {
		return c, nil
	}
	return domain.CapacidadeDiaPadrao, nil
}

func (m *mockScheduleRepo) DefinirCapacidadeDia(ctx context.Context, data string, cap int) error {
	m.capacidades[data] = cap
	return nil
}

func (m *mockScheduleRepo) ContarAgendamentosDia(ctx context.Context, data string) (int, error) {
	return len(m.agendas[data]), nil
}

func (m *mockScheduleRepo) CriarAgendamento(ctx context.Context, ag *domain.Agendamento) error {
	m.agendas[ag.Data] = append(m.agendas[ag.Data], ag)
	return nil
}

func (m *mockScheduleRepo) RemoverAgendamento(ctx context.Context, id uint) error {
	return nil
}

func (m *mockScheduleRepo) ObterLimiteIndividual(ctx context.Context) (int, error) {
	return 2, nil
}

func (m *mockScheduleRepo) ContarIndividuaisDia(ctx context.Context, data string) (int, error) {
	return len(m.individuais[data]), nil
}

func (m *mockScheduleRepo) ExisteHorarioIndividual(ctx context.Context, data, hora string) (bool, error) {
	for _, item := range m.individuais[data] {
		if item.Hora == hora {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockScheduleRepo) CriarAgendamentoIndividual(ctx context.Context, ag *domain.AgendamentoIndividual) error {
	m.individuais[ag.Data] = append(m.individuais[ag.Data], ag)
	return nil
}

func (m *mockScheduleRepo) RemoverAgendamentoIndividual(ctx context.Context, id uint) error {
	return nil
}

func TestScheduleService(t *testing.T) {
	repo := newMockScheduleRepo()
	repo.capacidades["2026-10-10"] = 2
	svc := service.NewScheduleService(repo)
	ctx := context.Background()

	// 1. Agendamento com turno inválido
	err := svc.AgendarVaga(ctx, &domain.Agendamento{
		Data:  "2026-10-10",
		Nome:  "Aluno 1",
		Turno: "madrugada",
	})
	if !errors.Is(err, domain.ErrTurnoInvalido) {
		t.Errorf("esperado ErrTurnoInvalido, recebido %v", err)
	}

	// 2. Preenchendo até o limite
	_ = svc.AgendarVaga(ctx, &domain.Agendamento{Data: "2026-10-10", Nome: "A1", Turno: "manha"})
	_ = svc.AgendarVaga(ctx, &domain.Agendamento{Data: "2026-10-10", Nome: "A2", Turno: "tarde"})

	// 3. Capacidade esgotada
	err = svc.AgendarVaga(ctx, &domain.Agendamento{
		Data:  "2026-10-10",
		Nome:  "A3 Excedente",
		Turno: "noite",
	})
	if !errors.Is(err, domain.ErrCapacidadeEsgotada) {
		t.Errorf("esperado ErrCapacidadeEsgotada, recebido %v", err)
	}

	// 4. Agendamento individual com horário duplicado
	_ = svc.AgendarIndividual(ctx, &domain.AgendamentoIndividual{
		Data: "2026-10-10", Hora: "09:00", Pratica: "ventosaterapia",
	})
	err = svc.AgendarIndividual(ctx, &domain.AgendamentoIndividual{
		Data: "2026-10-10", Hora: "09:00", Pratica: "personal_trainer",
	})
	if !errors.Is(err, domain.ErrHorarioOcupado) {
		t.Errorf("esperado ErrHorarioOcupado, recebido %v", err)
	}
}
