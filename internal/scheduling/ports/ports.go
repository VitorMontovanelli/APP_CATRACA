package ports

import (
	"context"

	"catraca-app/internal/scheduling/domain"
)

type ScheduleRepository interface {
	ObterCapacidadeDia(ctx context.Context, data string) (int, error)
	DefinirCapacidadeDia(ctx context.Context, data string, cap int) error
	ContarAgendamentosDia(ctx context.Context, data string) (int, error)
	CriarAgendamento(ctx context.Context, ag *domain.Agendamento) error
	RemoverAgendamento(ctx context.Context, id uint) error
	ObterLimiteIndividual(ctx context.Context) (int, error)
	ContarIndividuaisDia(ctx context.Context, data string) (int, error)
	ExisteHorarioIndividual(ctx context.Context, data, hora string) (bool, error)
	CriarAgendamentoIndividual(ctx context.Context, ag *domain.AgendamentoIndividual) error
	RemoverAgendamentoIndividual(ctx context.Context, id uint) error
}

type ScheduleServicePort interface {
	AgendarVaga(ctx context.Context, ag *domain.Agendamento) error
	AgendarIndividual(ctx context.Context, ag *domain.AgendamentoIndividual) error
}
