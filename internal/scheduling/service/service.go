package service

import (
	"context"

	"catraca-app/internal/scheduling/domain"
	"catraca-app/internal/scheduling/ports"
)

type ScheduleService struct {
	repo ports.ScheduleRepository
}

func NewScheduleService(repo ports.ScheduleRepository) *ScheduleService {
	return &ScheduleService{repo: repo}
}

func (s *ScheduleService) AgendarVaga(ctx context.Context, ag *domain.Agendamento) error {
	if !domain.ValidarTurno(ag.Turno) {
		return domain.ErrTurnoInvalido
	}

	capacidade, err := s.repo.ObterCapacidadeDia(ctx, ag.Data)
	if err != nil {
		capacidade = domain.CapacidadeDiaPadrao
	}

	totalAtual, err := s.repo.ContarAgendamentosDia(ctx, ag.Data)
	if err != nil {
		return err
	}

	if totalAtual >= capacidade {
		return domain.ErrCapacidadeEsgotada
	}

	return s.repo.CriarAgendamento(ctx, ag)
}

func (s *ScheduleService) AgendarIndividual(ctx context.Context, ag *domain.AgendamentoIndividual) error {
	if !domain.ValidarPratica(ag.Pratica) {
		return domain.ErrPraticaInvalida
	}

	limite, err := s.repo.ObterLimiteIndividual(ctx)
	if err != nil || limite <= 0 {
		limite = 5
	}

	totalDia, err := s.repo.ContarIndividuaisDia(ctx, ag.Data)
	if err != nil {
		return err
	}
	if totalDia >= limite {
		return domain.ErrLimiteIndividualAtingido
	}

	ocupado, err := s.repo.ExisteHorarioIndividual(ctx, ag.Data, ag.Hora)
	if err != nil {
		return err
	}
	if ocupado {
		return domain.ErrHorarioOcupado
	}

	return s.repo.CriarAgendamentoIndividual(ctx, ag)
}
