package service

import (
	"context"

	"catraca-app/internal/student/domain"
	"catraca-app/internal/student/ports"
)

type StudentService struct {
	repo   ports.StudentRepository
	camera ports.CameraPort
}

func NewStudentService(repo ports.StudentRepository, camera ports.CameraPort) *StudentService {
	return &StudentService{
		repo:   repo,
		camera: camera,
	}
}

func (s *StudentService) CriarStudent(ctx context.Context, student *domain.Student) error {
	if err := student.Validate(); err != nil {
		return err
	}
	student.Ativo = true
	return s.repo.Create(ctx, student)
}

func (s *StudentService) BuscarStudent(ctx context.Context, id uint) (*domain.Student, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *StudentService) ListarStudents(ctx context.Context) ([]domain.Student, error) {
	return s.repo.ListAll(ctx)
}

func (s *StudentService) CapturarFoto(ctx context.Context) (string, error) {
	if s.camera == nil {
		return "", domain.ErrPhotoTooLarge
	}
	return s.camera.CaptureFrame(ctx)
}

func (s *StudentService) SalvarFoto(ctx context.Context, id uint, base64Photo string) error {
	// Limite defensivo: 5MB em base64 = ~7M caracteres
	if len(base64Photo) > 7*1024*1024 {
		return domain.ErrPhotoTooLarge
	}
	return s.repo.SavePhoto(ctx, id, base64Photo)
}
