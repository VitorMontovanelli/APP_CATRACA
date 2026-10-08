package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"catraca-app/internal/student/domain"
	"catraca-app/internal/student/service"
)

type mockStudentRepo struct {
	students map[uint]*domain.Student
	photos   map[uint]string
	nextID   uint
}

func newMockStudentRepo() *mockStudentRepo {
	return &mockStudentRepo{
		students: make(map[uint]*domain.Student),
		photos:   make(map[uint]string),
		nextID:   1,
	}
}

func (m *mockStudentRepo) FindByID(ctx context.Context, id uint) (*domain.Student, error) {
	if s, ok := m.students[id]; ok {
		return s, nil
	}
	return nil, errors.New("not found")
}

func (m *mockStudentRepo) ListAll(ctx context.Context) ([]domain.Student, error) {
	var list []domain.Student
	for _, s := range m.students {
		list = append(list, *s)
	}
	return list, nil
}

func (m *mockStudentRepo) Create(ctx context.Context, s *domain.Student) error {
	s.ID = m.nextID
	m.nextID++
	m.students[s.ID] = s
	return nil
}

func (m *mockStudentRepo) Update(ctx context.Context, s *domain.Student) error {
	m.students[s.ID] = s
	return nil
}

func (m *mockStudentRepo) SetActive(ctx context.Context, id uint, active bool) error {
	if s, ok := m.students[id]; ok {
		s.Ativo = active
		return nil
	}
	return errors.New("not found")
}

func (m *mockStudentRepo) SavePhoto(ctx context.Context, id uint, base64Photo string) error {
	m.photos[id] = base64Photo
	return nil
}

func (m *mockStudentRepo) GetPhoto(ctx context.Context, id uint) (string, error) {
	return m.photos[id], nil
}

func (m *mockStudentRepo) RemovePhoto(ctx context.Context, id uint) error {
	delete(m.photos, id)
	return nil
}

type mockCamera struct {
	mockData string
}

func (m *mockCamera) CaptureFrame(ctx context.Context) (string, error) {
	return m.mockData, nil
}

func TestStudentService_CriarStudent(t *testing.T) {
	repo := newMockStudentRepo()
	cam := &mockCamera{mockData: "data:image/jpeg;base64,mockphoto"}
	svc := service.NewStudentService(repo, cam)
	ctx := context.Background()

	// 1. Erro CPF inválido
	err := svc.CriarStudent(ctx, &domain.Student{
		Nome: "João Silva",
		CPF:  "123.abc",
	})
	if !errors.Is(err, domain.ErrInvalidCPF) {
		t.Errorf("esperado ErrInvalidCPF, recebido %v", err)
	}

	// 2. Erro Nome vazio
	err = svc.CriarStudent(ctx, &domain.Student{
		Nome: "",
		CPF:  "123.456.789-00",
	})
	if !errors.Is(err, domain.ErrEmptyStudentName) {
		t.Errorf("esperado ErrEmptyStudentName, recebido %v", err)
	}

	// 3. Sucesso
	validStudent := &domain.Student{
		Nome: "Mariana Souza",
		CPF:  "111.222.333-44",
	}
	err = svc.CriarStudent(ctx, validStudent)
	if err != nil {
		t.Fatalf("erro ao criar aluno válido: %v", err)
	}
	if validStudent.ID == 0 {
		t.Errorf("esperado ID gerado > 0, recebido %d", validStudent.ID)
	}

	// 4. Captura de Foto Mock
	foto, err := svc.CapturarFoto(ctx)
	if err != nil || foto != "data:image/jpeg;base64,mockphoto" {
		t.Errorf("esperado foto mock capturada, recebido: %s (err: %v)", foto, err)
	}

	// 5. Foto gigante deve dar erro
	giganticPhoto := strings.Repeat("A", 8*1024*1024)
	err = svc.SalvarFoto(ctx, validStudent.ID, giganticPhoto)
	if !errors.Is(err, domain.ErrPhotoTooLarge) {
		t.Errorf("esperado ErrPhotoTooLarge, recebido %v", err)
	}
}
