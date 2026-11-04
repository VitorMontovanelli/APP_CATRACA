package inbound

import (
	"strings"
	"testing"

	"catraca-app/internal/models"
	"gorm.io/gorm"
)

func setupStudentTestDB(t *testing.T) (*StudentHandler, *StudentCrudHandler, *StudentPlansHandler) {
	t.Helper()
	database := setupTestDB(t)
	if err := database.AutoMigrate(&models.Student{}, &models.StudentPlan{}); err != nil {
		t.Fatalf("failed to migrate students: %v", err)
	}
	session := NewSessionManager(database)
	session.SetActor(1, "admin")
	crud := NewStudentCrudHandler(database, session)
	plans := NewStudentPlansHandler(database, session)
	return NewStudentHandler(database, session), crud, plans
}

func countStudents(t *testing.T, db *gorm.DB, where string, args ...interface{}) int64 {
	t.Helper()
	var total int64
	if err := db.Raw("SELECT COUNT(*) FROM students WHERE "+where, args...).Scan(&total).Error; err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	return total
}

func TestCriarAlunoSemCPFGravaNull(t *testing.T) {
	h, _, plans := setupStudentTestDB(t)

	if _, err := plans.CriarStudentComPlano("Ana Silva", "", "", "(11) 99999-1111", "", "", 0, 0, 0); err != nil {
		t.Fatalf("primeiro aluno sem CPF deveria salvar: %v", err)
	}
	if _, err := plans.CriarStudentComPlano("Bia Souza", "", "", "(11) 99999-2222", "", "", 0, 0, 0); err != nil {
		t.Fatalf("segundo aluno sem CPF deveria salvar: %v", err)
	}

	if got := countStudents(t, h.StudentCrudHandler.db, "cpf IS NULL"); got != 2 {
		t.Fatalf("esperava 2 CPFs NULL, encontrou %d", got)
	}
	if got := countStudents(t, h.StudentCrudHandler.db, "cpf = ''"); got != 0 {
		t.Fatalf("não deveria existir CPF vazio, encontrou %d", got)
	}
}

func TestCriarAlunoExigeNomeETelefone(t *testing.T) {
	_, _, plans := setupStudentTestDB(t)

	if _, err := plans.CriarStudentComPlano("", "", "", "(11) 99999-1111", "", "", 0, 0, 0); err == nil {
		t.Fatal("esperava erro para nome vazio")
	}
	if _, err := plans.CriarStudentComPlano("Ana", "", "", "", "", "", 0, 0, 0); err == nil {
		t.Fatal("esperava erro para telefone vazio")
	}
}

func TestCriarAlunoComCPFDuplicadoMostraNome(t *testing.T) {
	_, crud, plans := setupStudentTestDB(t)

	if _, err := crud.CriarStudent("Carlos Lima", "147.089.317-78", "", "(11) 99999-1111", "", 0); err != nil {
		t.Fatalf("primeiro aluno deveria salvar: %v", err)
	}

	_, err := plans.CriarStudentComPlano("Outra Pessoa", "14708931778", "", "(11) 99999-2222", "", "", 0, 0, 0)
	if err == nil {
		t.Fatal("esperava erro de CPF duplicado")
	}
	want := "Este CPF já está cadastrado para Carlos Lima"
	if err.Error() != want {
		t.Fatalf("mensagem = %q, esperava %q", err.Error(), want)
	}
}

func TestAtualizarAlunoLimpandoCPFGravaNull(t *testing.T) {
	h, crud, _ := setupStudentTestDB(t)

	created, err := crud.CriarStudent("Ana Silva", "111.444.777-35", "2000-01-01", "(11) 99999-1111", "", 0)
	if err != nil {
		t.Fatalf("não deveria falhar ao criar: %v", err)
	}

	if _, err := crud.AtualizarStudent(created.ID, "Ana Silva", "", "", "(11) 99999-1111", "", "", 0); err != nil {
		t.Fatalf("não deveria falhar ao limpar o CPF: %v", err)
	}
	if got := countStudents(t, h.StudentCrudHandler.db, "id = ? AND cpf IS NULL", created.ID); got != 1 {
		t.Fatalf("esperava CPF NULL após edição, encontrou %d", got)
	}
	if got := countStudents(t, h.StudentCrudHandler.db, "id = ? AND data_nascimento IS NULL", created.ID); got != 1 {
		t.Fatalf("esperava data de nascimento NULL após edição, encontrou %d", got)
	}
}

func TestAtualizarAlunoNaoConflitaComProprioCPF(t *testing.T) {
	_, crud, _ := setupStudentTestDB(t)

	created, err := crud.CriarStudent("Ana Silva", "111.444.777-35", "", "(11) 99999-1111", "", 0)
	if err != nil {
		t.Fatalf("não deveria falhar ao criar: %v", err)
	}

	if _, err := crud.AtualizarStudent(created.ID, "Ana Silva editada", "11144477735", "", "(11) 99999-1111", "", "", 0); err != nil {
		t.Fatalf("editar mantendo o próprio CPF não deveria conflitar: %v", err)
	}
}

func TestErroCPFDuplicadoEhAmigavel(t *testing.T) {
	_, crud, _ := setupStudentTestDB(t)

	if _, err := crud.CriarStudent("Carlos Lima", "147.089.317-78", "", "(11) 99999-1111", "", 0); err != nil {
		t.Fatalf("primeiro aluno deveria salvar: %v", err)
	}
	_, err := crud.CriarStudent("Outra Pessoa", "14708931778", "", "(11) 99999-2222", "", 0)
	if err == nil || !strings.HasPrefix(err.Error(), "Este CPF já está cadastrado para ") {
		t.Fatalf("esperava mensagem amigável de CPF duplicado, veio: %v", err)
	}
}
