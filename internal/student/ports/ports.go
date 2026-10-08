package ports

import (
	"context"

	"catraca-app/internal/student/domain"
)

// CameraPort define a porta de saída para captura de imagens de hardware (webcam).
type CameraPort interface {
	CaptureFrame(ctx context.Context) (string, error)
}

// StudentRepository define a porta de saída para persistência de alunos.
type StudentRepository interface {
	FindByID(ctx context.Context, id uint) (*domain.Student, error)
	ListAll(ctx context.Context) ([]domain.Student, error)
	Create(ctx context.Context, s *domain.Student) error
	Update(ctx context.Context, s *domain.Student) error
	SetActive(ctx context.Context, id uint, active bool) error
	SavePhoto(ctx context.Context, id uint, base64Photo string) error
	GetPhoto(ctx context.Context, id uint) (string, error)
	RemovePhoto(ctx context.Context, id uint) error
}

// StudentServicePort define a porta de entrada para os casos de uso de alunos.
type StudentServicePort interface {
	CriarStudent(ctx context.Context, s *domain.Student) error
	BuscarStudent(ctx context.Context, id uint) (*domain.Student, error)
	ListarStudents(ctx context.Context) ([]domain.Student, error)
	CapturarFoto(ctx context.Context) (string, error)
	SalvarFoto(ctx context.Context, id uint, base64Photo string) error
}
