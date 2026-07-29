package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type App struct {
	ctx        context.Context
	db         *gorm.DB
	actorCargo string
	actorID    uint
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.db = initDB()
}

// =============== Catraca Virtual ===============

// CadastrarAluno cria um novo aluno para a catraca com vencimento calculado automaticamente.
func (a *App) CadastrarAluno(nome, cpf string, diasValidade int, idBiometria int) error {
	if a.actorCargo != "super_admin" && a.actorCargo != "admin" {
		return fmt.Errorf("permissão negada")
	}
	vencimento := time.Now().AddDate(0, 0, diasValidade)
	aluno := Aluno{
		Nome:            nome,
		CPF:             cpf,
		Status:          true,
		VencimentoPlano: vencimento,
		IDBiometriaMock: idBiometria,
	}
	if err := a.db.Create(&aluno).Error; err != nil {
		return err
	}
	a.registrarAudit("cadastrou_aluno_catraca", aluno.ID)
	return nil
}

// VerificarAcessoAluno simula a leitura da digital na catraca e retorna o resultado.
func (a *App) VerificarAcessoAluno(idBiometria int) (ResultadoAcesso, error) {
	var aluno Aluno
	err := a.db.Where("id_biometria_mock = ?", idBiometria).First(&aluno).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			registro := RegistroAcesso{
				AlunoID:  0,
				DataHora: time.Now(),
				Liberado: false,
				Motivo:   "Aluno não encontrado",
			}
			a.db.Create(&registro)
			return ResultadoAcesso{
				Nome:     "",
				Liberado: false,
				Mensagem: "Aluno não encontrado",
			}, nil
		}
		return ResultadoAcesso{}, err
	}

	if !aluno.Status {
		registro := RegistroAcesso{
			AlunoID:  aluno.ID,
			Liberado: false,
			Motivo:   "Aluno inativo",
		}
		a.db.Create(&registro)
		return ResultadoAcesso{
			Nome:     aluno.Nome,
			Liberado: false,
			Mensagem: "Aluno inativo",
		}, nil
	}

	if time.Now().After(aluno.VencimentoPlano) {
		registro := RegistroAcesso{
			AlunoID:  aluno.ID,
			Liberado: false,
			Motivo:   "Plano Vencido",
		}
		a.db.Create(&registro)
		return ResultadoAcesso{
			Nome:     aluno.Nome,
			Liberado: false,
			Mensagem: "Plano Vencido",
		}, nil
	}

	registro := RegistroAcesso{
		AlunoID:  aluno.ID,
		Liberado: true,
		Motivo:   "Acesso Liberado",
	}
	a.db.Create(&registro)
	return ResultadoAcesso{
		Nome:     aluno.Nome,
		Liberado: true,
		Mensagem: "Acesso Liberado",
	}, nil
}

// =============== Alunos (Students CRUD) ===============

func (a *App) ListarStudents() ([]Student, error) {
	var list []Student
	err := a.db.Order("nome").Find(&list).Error
	return list, err
}

func (a *App) CriarStudent(nome, cpf, dataNascimento, telefone, email string, formaPagamentoID uint) (*Student, error) {
	if a.actorCargo != "super_admin" && a.actorCargo != "admin" {
		return nil, fmt.Errorf("permissão negada")
	}
	student := Student{
		Nome:            nome,
		CPF:             cpf,
		DataNascimento:  strPtr(dataNascimento),
		Telefone:        strPtr(telefone),
		Email:           strPtr(email),
		FormaPagamentoID: uintPtr(formaPagamentoID),
	}
	if err := a.db.Create(&student).Error; err != nil {
		return nil, err
	}
	a.registrarAudit("criou_aluno", student.ID)
	return a.BuscarStudent(student.ID)
}

func (a *App) BuscarStudent(id uint) (*Student, error) {
	var s Student
	err := a.db.First(&s, id).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (a *App) CriarStudentComPlano(nome, cpf, dataNascimento, telefone, email string, formaPagamentoID, planID, dueDay uint) (*Student, error) {
	if a.actorCargo != "super_admin" && a.actorCargo != "admin" {
		return nil, fmt.Errorf("permissão negada")
	}

	tx := a.db.Begin()

	student := Student{
		Nome:            nome,
		CPF:             cpf,
		DataNascimento:  strPtr(dataNascimento),
		Telefone:        strPtr(telefone),
		Email:           strPtr(email),
		FormaPagamentoID: uintPtr(formaPagamentoID),
	}
	if err := tx.Create(&student).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if planID > 0 {
		sp := StudentPlan{
			StudentID:       student.ID,
			PlanID:          planID,
			DueDay:          int(dueDay),
			PaymentMethodID: uintPtr(formaPagamentoID),
		}
		if err := tx.Create(&sp).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	a.registrarAudit("criou_aluno", student.ID)
	return a.BuscarStudent(student.ID)
}

func (a *App) AtualizarStudent(id uint, nome, cpf, dataNascimento, telefone, email string, formaPagamentoID uint) (*Student, error) {
	if a.actorCargo != "super_admin" && a.actorCargo != "admin" {
		return nil, fmt.Errorf("permissão negada")
	}
	student, err := a.BuscarStudent(id)
	if err != nil {
		return nil, err
	}
	student.Nome = nome
	student.CPF = cpf
	student.DataNascimento = strPtr(dataNascimento)
	student.Telefone = strPtr(telefone)
	student.Email = strPtr(email)
	student.FormaPagamentoID = uintPtr(formaPagamentoID)

	if err := a.db.Save(student).Error; err != nil {
		return nil, err
	}
	a.registrarAudit("editou_aluno", id)
	return a.BuscarStudent(id)
}

func (a *App) AtivarStudent(id uint, ativo bool) error {
	if a.actorCargo != "super_admin" && a.actorCargo != "admin" {
		return fmt.Errorf("permissão negada")
	}
	err := a.db.Model(&Student{}).Where("id = ?", id).Update("ativo", ativo).Error
	if err == nil {
		action := "desativou_aluno"
		if ativo {
			action = "ativou_aluno"
		}
		a.registrarAudit(action, id)
	}
	return err
}

func (a *App) ListarStudentsComPlanos() ([]StudentComPlano, error) {
	type rawResult struct {
		ID               uint
		Nome             string
		CPF              string
		DataNascimento   *string
		Telefone         *string
		Email            *string
		FormaPagamentoID *uint
		DataEntrada      string
		Observacao       *string
		Ativo            bool
		CreatedAt        string
		UpdatedAt        *string
		PlanoNome        *string
		PlanoPreco       *int
		PlanoStatus      *string
		StudentPlanID    *uint
		PaymentMethod    *string
		DueDay           *int
		UltimaFatura     *string
		FaturaValor      *int
		FaturaVencimento *string
	}

	var raw []rawResult
	err := a.db.Raw(`
		SELECT
			s.id, s.nome, s.cpf, s.data_nascimento, s.telefone, s.email,
			s.forma_pagamento_id, s.data_entrada, s.observacao, s.ativo,
			s.created_at, s.updated_at,
			p.name, p.price_cents, sp.status,
			sp.id, pm.name, sp.due_day,
			i.status, i.amount_cents, i.due_date
		FROM students s
		LEFT JOIN student_plans sp ON sp.student_id = s.id AND sp.status = 'active'
		LEFT JOIN plans p ON p.id = sp.plan_id
		LEFT JOIN payment_methods pm ON pm.id = COALESCE(sp.payment_method_id, s.forma_pagamento_id)
		LEFT JOIN invoices i ON i.id = (
			SELECT id FROM invoices WHERE student_id = s.id ORDER BY due_date DESC LIMIT 1
		)
		ORDER BY s.nome
	`).Scan(&raw).Error
	if err != nil {
		return nil, err
	}

	list := make([]StudentComPlano, len(raw))
	for i, r := range raw {
		de := time.Time{}
		ca := time.Time{}
		var ua time.Time
		if t, err := time.Parse(time.RFC3339, r.DataEntrada); err == nil {
			de = t
		}
		if t, err := time.Parse(time.RFC3339, r.CreatedAt); err == nil {
			ca = t
		}
		if r.UpdatedAt != nil {
			if t, err := time.Parse(time.RFC3339, *r.UpdatedAt); err == nil {
				ua = t
			}
		}

		list[i] = StudentComPlano{
			Student: Student{
				ID:               r.ID,
				Nome:             r.Nome,
				CPF:              r.CPF,
				DataNascimento:   r.DataNascimento,
				Telefone:         r.Telefone,
				Email:            r.Email,
				FormaPagamentoID: r.FormaPagamentoID,
				DataEntrada:      de,
				Observacao:       r.Observacao,
				Ativo:            r.Ativo,
				CreatedAt:        ca,
				UpdatedAt:        ua,
			},
			PlanoNome:        r.PlanoNome,
			PlanoPreco:       r.PlanoPreco,
			PlanoStatus:      r.PlanoStatus,
			StudentPlanID:    r.StudentPlanID,
			PaymentMethod:    r.PaymentMethod,
			DueDay:           r.DueDay,
			UltimaFatura:     r.UltimaFatura,
			FaturaValor:      r.FaturaValor,
			FaturaVencimento: r.FaturaVencimento,
		}
	}
	return list, nil
}

// =============== Login & Usuários ===============

func (a *App) Login(email, senha string) (*User, error) {
	var u User
	err := a.db.Where("email = ?", email).First(&u).Error
	if err != nil {
		return nil, fmt.Errorf("email ou senha inválidos")
	}

	if strings.HasPrefix(u.Senha, "$2a$") || strings.HasPrefix(u.Senha, "$2y$") {
		// Validar usando bcrypt
		if err := bcrypt.CompareHashAndPassword([]byte(u.Senha), []byte(senha)); err != nil {
			return nil, fmt.Errorf("email ou senha inválidos")
		}
	} else {
		// Senha legada em texto plano
		if u.Senha != senha {
			return nil, fmt.Errorf("email ou senha inválidos")
		}
		// Migrar de forma transparente para bcrypt
		hashed, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
		if err == nil {
			a.db.Model(&u).Update("senha", string(hashed))
		}
	}

	if !u.Ativo {
		return nil, fmt.Errorf("usuário desativado")
	}
	a.actorCargo = u.Cargo
	a.actorID = u.ID
	a.registrarAuditComDetalhes("login", u.ID, fmt.Sprintf("email=%s, cargo=%s", email, u.Cargo))
	return &u, nil
}

func (a *App) CriarUsuario(name, email, senha, cargo string, registradorID uint) (*User, error) {
	if cargo != "super_admin" && cargo != "admin" {
		return nil, fmt.Errorf("cargo inválido: %s", cargo)
	}
	if !podeCriar(a.actorCargo, cargo) {
		return nil, fmt.Errorf("você não tem permissão para criar usuário com cargo %s", cargo)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar hash da senha: %w", err)
	}

	user := User{
		Name:          name,
		Email:         email,
		Senha:         string(hashed),
		Cargo:         cargo,
		RegistradorID: uintPtr(registradorID),
	}
	if err := a.db.Create(&user).Error; err != nil {
		return nil, err
	}
	a.registrarAudit("criou_usuario", user.ID)
	return a.BuscarUsuario(user.ID)
}

func (a *App) ListarUsuarios() ([]User, error) {
	var users []User
	var err error
	if a.actorCargo == "super_admin" {
		err = a.db.Select("id, name, email, cargo, registrador_id, ativo, criado_em").Order("id").Find(&users).Error
	} else {
		// admin só pode ver a si mesmo e usuários criados por ele (excluindo super_admins)
		err = a.db.Select("id, name, email, cargo, registrador_id, ativo, criado_em").
			Where("(id = ? OR registrador_id = ?) AND cargo != ?", a.actorID, a.actorID, "super_admin").
			Order("id").Find(&users).Error
	}
	return users, err
}

func (a *App) BuscarUsuario(id uint) (*User, error) {
	var u User
	err := a.db.First(&u, id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (a *App) AtualizarUsuario(id uint, name, email, cargo string) (*User, error) {
	if cargo != "super_admin" && cargo != "admin" {
		return nil, fmt.Errorf("cargo inválido: %s", cargo)
	}
	u, err := a.BuscarUsuario(id)
	if err != nil {
		return nil, err
	}
	if !podeGerenciar(a.actorCargo, u.Cargo) {
		return nil, fmt.Errorf("você não tem permissão para alterar este usuário")
	}
	u.Name = name
	u.Email = email
	u.Cargo = cargo
	if err := a.db.Save(u).Error; err != nil {
		return nil, err
	}
	a.registrarAudit("editou_usuario", id)
	return a.BuscarUsuario(id)
}

func (a *App) AtivarUsuario(id uint, ativo bool) error {
	u, err := a.BuscarUsuario(id)
	if err != nil {
		return err
	}
	if !podeGerenciar(a.actorCargo, u.Cargo) {
		return fmt.Errorf("você não tem permissão para %s este usuário", ativarDesativarLabel(ativo))
	}
	err = a.db.Model(&User{}).Where("id = ?", id).Update("ativo", ativo).Error
	if err != nil {
		return err
	}
	a.registrarAudit(ativarDesativarAction(ativo), id)
	return nil
}

func (a *App) DeletarUsuario(id uint) error {
	u, err := a.BuscarUsuario(id)
	if err != nil {
		return err
	}
	if !podeGerenciar(a.actorCargo, u.Cargo) {
		return fmt.Errorf("você não tem permissão para deletar este usuário")
	}
	if err := a.db.Delete(&u).Error; err != nil {
		return err
	}
	a.registrarAudit("deletou_usuario", id)
	return nil
}

func (a *App) MeuPerfil() (*User, error) {
	return a.BuscarUsuario(a.actorID)
}

func (a *App) AppVersion() string {
	return Version
}

func (a *App) AlterarSenha(senhaAtual, novaSenha string) error {
	var u User
	if err := a.db.First(&u, a.actorID).Error; err != nil {
		return err
	}

	if strings.HasPrefix(u.Senha, "$2a$") || strings.HasPrefix(u.Senha, "$2y$") {
		if err := bcrypt.CompareHashAndPassword([]byte(u.Senha), []byte(senhaAtual)); err != nil {
			return fmt.Errorf("senha atual incorreta")
		}
	} else {
		if u.Senha != senhaAtual {
			return fmt.Errorf("senha atual incorreta")
		}
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(novaSenha), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("erro ao criptografar nova senha: %w", err)
	}

	err = a.db.Model(&User{}).Where("id = ?", a.actorID).Update("senha", string(hashed)).Error
	if err == nil {
		a.registrarAudit("alterou_senha", a.actorID)
	}
	return err
}

func (a *App) AlterarFoto(fotoBase64 string) error {
	err := a.db.Model(&User{}).Where("id = ?", a.actorID).Update("foto", fotoBase64).Error
	if err == nil {
		a.registrarAudit("alterou_foto", a.actorID)
	}
	return err
}

func podeCriar(actorCargo, targetCargo string) bool {
	return actorCargo == "super_admin" && targetCargo == "admin"
}

func podeGerenciar(actorCargo, targetCargo string) bool {
	return actorCargo == "super_admin"
}

func ativarDesativarLabel(ativo bool) string {
	if ativo {
		return "ativar"
	}
	return "desativar"
}

func ativarDesativarAction(ativo bool) string {
	if ativo {
		return "ativou_usuario"
	}
	return "desativou_usuario"
}

func (a *App) registrarAudit(action string, targetID uint) {
	a.registrarAuditComDetalhes(action, targetID, "")
}

func (a *App) registrarAuditComDetalhes(action string, targetID uint, details string) {
	a.db.Exec(
		"INSERT INTO audit_logs (actor_id, action, target_id, details) VALUES (?, ?, ?, ?)",
		a.actorID, action, targetID, details,
	)
}

// =============== Planos (CRUD) ===============

func (a *App) ListarPlanos() ([]Plan, error) {
	var list []Plan
	err := a.db.Order("price_cents").Find(&list).Error
	return list, err
}

func (a *App) CriarPlano(name, description string, durationDays, priceCents, gracePeriodDays int) (*Plan, error) {
	if a.actorCargo != "super_admin" && a.actorCargo != "admin" {
		return nil, fmt.Errorf("permissão negada")
	}
	plan := Plan{
		Name:            name,
		Description:     description,
		DurationDays:    durationDays,
		PriceCents:      priceCents,
		GracePeriodDays: gracePeriodDays,
	}
	if err := a.db.Create(&plan).Error; err != nil {
		return nil, err
	}
	a.registrarAudit("criou_plano", plan.ID)
	return a.BuscarPlano(plan.ID)
}

func (a *App) BuscarPlano(id uint) (*Plan, error) {
	var p Plan
	err := a.db.First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (a *App) AtualizarPlano(id uint, name, description string, durationDays, priceCents, gracePeriodDays int) (*Plan, error) {
	if a.actorCargo != "super_admin" && a.actorCargo != "admin" {
		return nil, fmt.Errorf("permissão negada")
	}
	plan, err := a.BuscarPlano(id)
	if err != nil {
		return nil, err
	}
	plan.Name = name
	plan.Description = description
	plan.DurationDays = durationDays
	plan.PriceCents = priceCents
	plan.GracePeriodDays = gracePeriodDays
	if err := a.db.Save(plan).Error; err != nil {
		return nil, err
	}
	a.registrarAudit("editou_plano", id)
	return a.BuscarPlano(id)
}

func (a *App) AtivarPlano(id uint, ativo bool) error {
	if a.actorCargo != "super_admin" && a.actorCargo != "admin" {
		return fmt.Errorf("permissão negada")
	}
	err := a.db.Model(&Plan{}).Where("id = ?", id).Update("active", ativo).Error
	if err == nil {
		action := "desativou_plano"
		if ativo {
			action = "ativou_plano"
		}
		a.registrarAudit(action, id)
	}
	return err
}

func (a *App) DeletarPlano(id uint) error {
	if a.actorCargo != "super_admin" && a.actorCargo != "admin" {
		return fmt.Errorf("permissão negada")
	}
	n := a.db.Delete(&Plan{}, id).RowsAffected
	if n == 0 {
		return fmt.Errorf("plano não encontrado")
	}
	a.registrarAudit("deletou_plano", id)
	return nil
}

type AlunoPorPlano struct {
	StudentID     uint   `json:"student_id"`
	Nome          string `json:"nome"`
	CPF           string `json:"cpf"`
	Status        string `json:"status"`
	StartDate     string `json:"start_date"`
}

func (a *App) ListarAlunosPorPlano(planID uint, page, pageSize int) ([]AlunoPorPlano, int, error) {
	var total int64
	a.db.Model(&StudentPlan{}).Where("plan_id = ?", planID).Count(&total)

	var list []AlunoPorPlano
	err := a.db.Raw(`
		SELECT s.id, s.nome, s.cpf, sp.status, sp.start_date
		FROM student_plans sp
		JOIN students s ON s.id = sp.student_id
		WHERE sp.plan_id = ?
		ORDER BY sp.start_date DESC
		LIMIT ? OFFSET ?
	`, planID, pageSize, (page-1)*pageSize).Scan(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, int(total), nil
}

// =============== Métodos de Pagamento ===============

func (a *App) ListarPaymentMethods() ([]PaymentMethod, error) {
	var list []PaymentMethod
	err := a.db.Order("name").Find(&list).Error
	return list, err
}

func (a *App) CriarPaymentMethod(name, mtype string, feePercent float64, feeFixedCents int) (*PaymentMethod, error) {
	if a.actorCargo != "super_admin" {
		return nil, fmt.Errorf("permissão negada")
	}
	pm := PaymentMethod{
		Name:          name,
		Type:          mtype,
		FeePercent:    feePercent,
		FeeFixedCents: feeFixedCents,
	}
	if err := a.db.Create(&pm).Error; err != nil {
		return nil, err
	}
	a.registrarAudit("criou_metodo_pagamento", pm.ID)
	return &pm, nil
}

func (a *App) AtivarPaymentMethod(id uint, enabled bool) error {
	if a.actorCargo != "super_admin" {
		return fmt.Errorf("permissão negada")
	}
	err := a.db.Model(&PaymentMethod{}).Where("id = ?", id).Update("enabled", enabled).Error
	if err == nil {
		action := "desativou_metodo_pagamento"
		if enabled {
			action = "ativou_metodo_pagamento"
		}
		a.registrarAudit(action, id)
	}
	return err
}

// =============== Planos dos Alunos (StudentPlans) ===============

func (a *App) ListarStudentPlans() ([]StudentPlan, error) {
	type spRaw struct {
		StudentPlan
		StudentName    string
		PlanName       string
		PlanPriceCents int
	}
	var raw []spRaw
	err := a.db.Raw(`
		SELECT sp.id, sp.student_id, sp.plan_id, sp.status, sp.payment_method_id,
			sp.due_day, sp.start_date, sp.end_date, sp.cancelled_at, sp.notes,
			sp.created_at, sp.updated_at,
			s.nome AS student_name, p.name AS plan_name, p.price_cents AS plan_price_cents
		FROM student_plans sp
		JOIN students s ON s.id = sp.student_id
		JOIN plans p ON p.id = sp.plan_id
		ORDER BY sp.start_date DESC
	`).Scan(&raw).Error
	if err != nil {
		return nil, err
	}
	list := make([]StudentPlan, len(raw))
	for i, r := range raw {
		list[i] = r.StudentPlan
		list[i].StudentName = r.StudentName
		list[i].PlanName = r.PlanName
		list[i].PlanPriceCents = r.PlanPriceCents
	}
	return list, nil
}

func (a *App) CriarStudentPlan(studentID, planID uint, dueDay int) (*StudentPlan, error) {
	sp := StudentPlan{
		StudentID: studentID,
		PlanID:    planID,
		DueDay:    dueDay,
	}
	if err := a.db.Create(&sp).Error; err != nil {
		return nil, err
	}
	a.registrarAudit("criou_plano_aluno", sp.ID)
	return a.BuscarStudentPlan(sp.ID)
}

func (a *App) BuscarStudentPlan(id uint) (*StudentPlan, error) {
	type spRaw struct {
		StudentPlan
		StudentName    string
		PlanName       string
		PlanPriceCents int
	}
	var raw spRaw
	err := a.db.Raw(`
		SELECT sp.id, sp.student_id, sp.plan_id, sp.status, sp.payment_method_id,
			sp.due_day, sp.start_date, sp.end_date, sp.cancelled_at, sp.notes,
			sp.created_at, sp.updated_at,
			s.nome AS student_name, p.name AS plan_name, p.price_cents AS plan_price_cents
		FROM student_plans sp
		JOIN students s ON s.id = sp.student_id
		JOIN plans p ON p.id = sp.plan_id
		WHERE sp.id = ?
	`, id).Scan(&raw).Error
	if err != nil {
		return nil, err
	}
	raw.StudentPlan.StudentName = raw.StudentName
	raw.StudentPlan.PlanName = raw.PlanName
	raw.StudentPlan.PlanPriceCents = raw.PlanPriceCents
	return &raw.StudentPlan, nil
}

func (a *App) CancelarStudentPlan(id uint) error {
	err := a.db.Model(&StudentPlan{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       "cancelled",
		"cancelled_at": time.Now(),
	}).Error
	if err == nil {
		a.registrarAudit("cancelou_plano_aluno", id)
	}
	return err
}

func (a *App) AtualizarPlanoAluno(studentID, planID, formaPagamentoID uint, dueDay int) error {
	if a.actorCargo != "super_admin" && a.actorCargo != "admin" {
		return fmt.Errorf("permissão negada")
	}

	var current StudentPlan
	hasCurrent := a.db.Where("student_id = ? AND status = 'active'", studentID).First(&current).Error == nil

	if planID == 0 {
		if hasCurrent {
			return a.db.Model(&current).Updates(map[string]interface{}{
				"status":       "cancelled",
				"cancelled_at": time.Now(),
			}).Error
		}
		return nil
	}

	if hasCurrent && current.PlanID == planID {
		return a.db.Model(&current).Updates(map[string]interface{}{
			"payment_method_id": uintPtr(formaPagamentoID),
			"due_day":           dueDay,
		}).Error
	}

	if hasCurrent {
		a.db.Model(&current).Updates(map[string]interface{}{
			"status":       "cancelled",
			"cancelled_at": time.Now(),
		})
	}

	sp := StudentPlan{
		StudentID:       studentID,
		PlanID:          planID,
		DueDay:          dueDay,
		PaymentMethodID: uintPtr(formaPagamentoID),
	}
	if err := a.db.Create(&sp).Error; err != nil {
		return err
	}
	a.registrarAuditComDetalhes("vinculou_plano_aluno", sp.ID,
		fmt.Sprintf("aluno=%d, plano=%d, forma_pagamento=%d, due_day=%d", studentID, planID, formaPagamentoID, dueDay))
	return nil
}

// =============== Faturas (Invoices) ===============

func (a *App) GerarInvoice(studentPlanID uint) (*Invoice, error) {
	type spMini struct {
		ID             uint
		StudentID      uint
		PlanID         uint
		PlanPriceCents int
		Status         string
	}
	var sp spMini
	err := a.db.Raw(`
		SELECT sp.id, sp.student_id, sp.plan_id, p.price_cents, sp.status
		FROM student_plans sp
		JOIN plans p ON p.id = sp.plan_id
		WHERE sp.id = ?
	`, studentPlanID).Scan(&sp).Error
	if err != nil {
		return nil, err
	}
	if sp.Status != "active" {
		return nil, fmt.Errorf("plano do aluno não está ativo")
	}

	inv := Invoice{
		StudentPlanID: studentPlanID,
		StudentID:     sp.StudentID,
		PlanID:        sp.PlanID,
		AmountCents:   sp.PlanPriceCents,
		DueDate:       time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
	}
	if err := a.db.Create(&inv).Error; err != nil {
		return nil, err
	}
	a.registrarAudit("gerou_fatura", inv.ID)
	return a.BuscarInvoice(inv.ID)
}

func (a *App) BuscarInvoice(id uint) (*Invoice, error) {
	type invRaw struct {
		Invoice
		StudentName       string
		PlanName          string
		PaymentMethodName *string
	}
	var raw invRaw
	err := a.db.Raw(`
		SELECT i.id, i.student_plan_id, i.student_id, i.plan_id, i.amount_cents,
			i.status, i.payment_method_id, i.due_date, i.paid_at, i.paid_amount_cents,
			i.gateway_transaction_id, i.pix_qr_code, i.pix_br_code, i.notes, i.created_at,
			s.nome AS student_name, p.name AS plan_name, pm.name AS payment_method_name
		FROM invoices i
		JOIN students s ON s.id = i.student_id
		JOIN plans p ON p.id = i.plan_id
		LEFT JOIN payment_methods pm ON pm.id = i.payment_method_id
		WHERE i.id = ?
	`, id).Scan(&raw).Error
	if err != nil {
		return nil, err
	}
	raw.Invoice.StudentName = raw.StudentName
	raw.Invoice.PlanName = raw.PlanName
	if raw.PaymentMethodName != nil {
		raw.Invoice.PaymentMethodName = *raw.PaymentMethodName
	}
	return &raw.Invoice, nil
}

func (a *App) ListarInvoices() ([]Invoice, error) {
	type invRaw struct {
		Invoice
		StudentName       string
		PlanName          string
		PaymentMethodName string
	}
	var raw []invRaw
	err := a.db.Raw(`
		SELECT i.id, i.student_plan_id, i.student_id, i.plan_id, i.amount_cents,
			i.status, i.payment_method_id, i.due_date, i.paid_at, i.paid_amount_cents,
			i.gateway_transaction_id, i.pix_qr_code, i.pix_br_code, i.notes, i.created_at,
			s.nome AS student_name, p.name AS plan_name, COALESCE(pm.name, '') AS payment_method_name
		FROM invoices i
		JOIN students s ON s.id = i.student_id
		JOIN plans p ON p.id = i.plan_id
		LEFT JOIN payment_methods pm ON pm.id = i.payment_method_id
		ORDER BY i.due_date DESC
	`).Scan(&raw).Error
	if err != nil {
		return nil, err
	}
	list := make([]Invoice, len(raw))
	for i, r := range raw {
		list[i] = r.Invoice
		list[i].StudentName = r.StudentName
		list[i].PlanName = r.PlanName
		list[i].PaymentMethodName = r.PaymentMethodName
	}
	return list, nil
}

func (a *App) ConfirmarPagamento(invoiceID uint, paymentMethodID uint, paidAmountCents int) error {
	userName := a.actorCargo
	err := a.db.Exec(
		"UPDATE invoices SET status='paid', payment_method_id=?, paid_at=datetime('now'), paid_amount_cents=?, notes=COALESCE(notes||' | ','')||'Confirmado por '||? WHERE id=?",
		paymentMethodID, paidAmountCents, userName, invoiceID,
	).Error
	if err == nil {
		a.registrarAuditComDetalhes("confirmou_pagamento", invoiceID,
			fmt.Sprintf("valor=%d, metodo=%d", paidAmountCents, paymentMethodID))
		a.gerarProximaFatura(invoiceID)
	}
	return err
}

func (a *App) CancelarInvoice(invoiceID uint) error {
	err := a.db.Model(&Invoice{}).Where("id = ?", invoiceID).Update("status", "cancelled").Error
	if err == nil {
		a.registrarAudit("cancelou_fatura", invoiceID)
	}
	return err
}

func (a *App) DeletarInvoice(invoiceID uint) error {
	if a.actorCargo != "super_admin" && a.actorCargo != "admin" {
		return fmt.Errorf("permissão negada")
	}
	n := a.db.Delete(&Invoice{}, invoiceID).RowsAffected
	if n == 0 {
		return fmt.Errorf("fatura não encontrada")
	}
	a.registrarAudit("deletou_fatura", invoiceID)
	return nil
}

// =============== Gateways ===============

func (a *App) ListarPaymentGateways() ([]PaymentGateway, error) {
	var list []PaymentGateway
	err := a.db.Order("name").Find(&list).Error
	return list, err
}

func (a *App) AtivarPaymentGateway(id uint, enabled bool) error {
	if a.actorCargo != "super_admin" {
		return fmt.Errorf("permissão negada")
	}
	err := a.db.Model(&PaymentGateway{}).Where("id = ?", id).Update("enabled", enabled).Error
	if err == nil {
		action := "desativou_gateway"
		if enabled {
			action = "ativou_gateway"
		}
		a.registrarAudit(action, id)
	}
	return err
}

func (a *App) AtualizarConfigGateway(id uint, config string) error {
	if a.actorCargo != "super_admin" {
		return fmt.Errorf("permissão negada")
	}
	err := a.db.Model(&PaymentGateway{}).Where("id = ?", id).Update("config", config).Error
	if err == nil {
		a.registrarAudit("configurou_gateway", id)
	}
	return err
}

// =============== AccessLogs ===============

type CobrancaFatura struct {
	ID          uint    `json:"id"`
	AmountCents int     `json:"amount_cents"`
	Status      string  `json:"status"`
	DueDate     string  `json:"due_date"`
	PaidAt      *string `json:"paid_at"`
	Comprovante *string `json:"comprovante"`
	DiasVencido int     `json:"dias_vencido"`
}

type CobrancaAluno struct {
	StudentID   uint             `json:"student_id"`
	Nome        string           `json:"nome"`
	CPF         string           `json:"cpf"`
	PlanoNome   string           `json:"plano_nome"`
	PlanoPreco  int              `json:"plano_preco"`
	StatusPlano string           `json:"status_plano"`
	DueDay      int              `json:"due_day"`
	Faturas     []CobrancaFatura `json:"faturas"`
}

func (a *App) ListarCobranca() ([]CobrancaAluno, error) {
	type alunoRaw struct {
		StudentID   uint
		Nome        string
		CPF         string
		PlanoNome   string
		PlanoPreco  int
		StatusPlano string
		DueDay      int
	}
	var alunos []alunoRaw
	err := a.db.Raw(`
		SELECT s.id AS student_id, s.nome, s.cpf, p.name AS plano_nome,
			p.price_cents AS plano_preco, sp.status AS status_plano, sp.due_day AS due_day
		FROM students s
		JOIN student_plans sp ON sp.student_id = s.id AND sp.status IN ('active','overdue','expired')
		JOIN plans p ON p.id = sp.plan_id
		ORDER BY s.nome
	`).Scan(&alunos).Error
	if err != nil {
		return nil, err
	}

	result := make([]CobrancaAluno, len(alunos))
	for i, al := range alunos {
		var faturas []CobrancaFatura
		a.db.Raw(`
			SELECT id, amount_cents, status, due_date,
				paid_at, comprovante,
				CAST(julianday('now') - julianday(due_date) AS INTEGER) AS dias_vencido
			FROM invoices
			WHERE student_plan_id = (SELECT id FROM student_plans WHERE student_id = ? ORDER BY id DESC LIMIT 1)
			ORDER BY due_date DESC
		`, al.StudentID).Scan(&faturas)

		if faturas == nil {
			faturas = []CobrancaFatura{}
		}
		result[i] = CobrancaAluno{
			StudentID:   al.StudentID,
			Nome:        al.Nome,
			CPF:         al.CPF,
			PlanoNome:   al.PlanoNome,
			PlanoPreco:  al.PlanoPreco,
			StatusPlano: al.StatusPlano,
			DueDay:      al.DueDay,
			Faturas:     faturas,
		}
	}
	return result, nil
}

func (a *App) ListarInadimplentes() ([]InadimplenteReport, error) {
	type rawResult struct {
		StudentID      uint
		Nome           string
		CPF            string
		Telefone       *string
		Email          *string
		PlanoNome      string
		PlanoPreco     int
		StatusPlano    string
		InvoiceID      uint
		ValorDevido    int
		DataVencimento string
		DiasVencido    int
		UltimoPagamento *string
		TotalEmAberto  int
	}

	var raw []rawResult
	err := a.db.Raw(`
		SELECT
			s.id AS student_id,
			s.nome,
			s.cpf,
			s.telefone,
			s.email,
			p.name AS plano_nome,
			p.price_cents AS plano_preco,
			sp.status AS status_plano,
			i.id AS invoice_id,
			i.amount_cents AS valor_devido,
			i.due_date AS data_vencimento,
			CAST(julianday('now') - julianday(i.due_date) AS INTEGER) AS dias_vencido,
			ult.ultimo_pagamento,
			aberto.total AS total_em_aberto
		FROM students s
		JOIN student_plans sp ON sp.student_id = s.id AND sp.status IN ('active','overdue','expired')
		JOIN plans p ON p.id = sp.plan_id
		JOIN invoices i ON i.student_plan_id = sp.id AND i.status IN ('pending','overdue') AND i.due_date < date('now')
		LEFT JOIN (
			SELECT student_id, MAX(paid_at) as ultimo_pagamento
			FROM invoices WHERE status = 'paid' GROUP BY student_id
		) ult ON ult.student_id = s.id
		JOIN (
			SELECT sp2.id as spid, SUM(i2.amount_cents) as total
			FROM student_plans sp2
			JOIN invoices i2 ON i2.student_plan_id = sp2.id AND i2.status IN ('pending','overdue')
			GROUP BY sp2.id
		) aberto ON aberto.spid = sp.id
		ORDER BY i.due_date ASC
	`).Scan(&raw).Error
	if err != nil {
		return nil, err
	}

	list := make([]InadimplenteReport, len(raw))
	for i, r := range raw {
		list[i] = InadimplenteReport{
			StudentID:      r.StudentID,
			Nome:           r.Nome,
			CPF:            r.CPF,
			Telefone:       r.Telefone,
			Email:          r.Email,
			PlanoNome:      r.PlanoNome,
			PlanoPreco:     r.PlanoPreco,
			StatusPlano:    r.StatusPlano,
			InvoiceID:      r.InvoiceID,
			ValorDevido:    r.ValorDevido,
			DataVencimento: r.DataVencimento,
			DiasVencido:    r.DiasVencido,
			UltimoPagamento: r.UltimoPagamento,
			TotalEmAberto:  r.TotalEmAberto,
		}
	}
	return list, nil
}

func (a *App) SalvarComprovante(invoiceID uint, base64Data, fileName string) error {
	dir := "comprovantes"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	if ext != ".pdf" {
		return fmt.Errorf("formato de arquivo inválido: apenas comprovantes em formato PDF são permitidos")
	}
	dest := filepath.Join(dir, fmt.Sprintf("invoice_%d%s", invoiceID, ext))

	_, raw, found := strings.Cut(base64Data, ",")
	if !found {
		raw = base64Data
	}

	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return fmt.Errorf("erro ao decodificar arquivo: %w", err)
	}

	if err := os.WriteFile(dest, data, 0644); err != nil {
		return fmt.Errorf("erro ao salvar arquivo: %w", err)
	}

	now := time.Now()
	err = a.db.Exec(
		"UPDATE invoices SET comprovante = ?, status = 'paid', paid_at = ? WHERE id = ?",
		dest, now, invoiceID,
	).Error
	if err != nil {
		return fmt.Errorf("erro ao atualizar fatura: %w", err)
	}

	a.registrarAuditComDetalhes("confirmou_pagamento", invoiceID,
		fmt.Sprintf("comprovante=%s, invoice=%d", dest, invoiceID))
	a.gerarProximaFatura(invoiceID)
	return nil
}

func (a *App) CriarInvoiceTeste() error {
	plano := Plan{
		Name:            "Plano Eduardo",
		Description:     "Plano anual do Eduardo",
		DurationDays:    365,
		PriceCents:      120000,
		GracePeriodDays: 10,
		Active:          true,
	}
	if err := a.db.Create(&plano).Error; err != nil {
		return fmt.Errorf("erro ao criar plano: %w", err)
	}

	aluno := Student{
		Nome:   "Eduardo Teste",
		CPF:    fmt.Sprintf("888%09d", plano.ID),
		Ativo:  true,
	}
	if err := a.db.Create(&aluno).Error; err != nil {
		return fmt.Errorf("erro ao criar aluno: %w", err)
	}

	sp := StudentPlan{
		StudentID: aluno.ID,
		PlanID:    plano.ID,
		DueDay:    15,
		Status:    "active",
		StartDate: time.Now().AddDate(0, -1, 0),
	}
	if err := a.db.Create(&sp).Error; err != nil {
		return fmt.Errorf("erro ao criar vínculo: %w", err)
	}

	inv := Invoice{
		StudentPlanID: sp.ID,
		StudentID:     aluno.ID,
		PlanID:        plano.ID,
		AmountCents:   plano.PriceCents,
		Status:        "overdue",
		DueDate:       time.Now().AddDate(0, 0, -5).Format("2006-01-02"),
	}
	if err := a.db.Create(&inv).Error; err != nil {
		return fmt.Errorf("erro ao criar fatura: %w", err)
	}

	a.registrarAudit("criou_teste", inv.ID)
	return nil
}

func (a *App) DeletarDadosTeste() error {
	if a.actorCargo != "super_admin" && a.actorCargo != "admin" {
		return fmt.Errorf("permissão negada")
	}

	a.db.Exec("DELETE FROM invoices WHERE student_id IN (SELECT id FROM students WHERE nome LIKE '%Teste%' OR nome LIKE '%Eduardo%')")
	a.db.Exec("DELETE FROM student_plans WHERE student_id IN (SELECT id FROM students WHERE nome LIKE '%Teste%' OR nome LIKE '%Eduardo%')")
	a.db.Exec("DELETE FROM students WHERE nome LIKE '%Teste%' OR nome LIKE '%Eduardo%'")
	a.db.Exec("DELETE FROM plans WHERE name LIKE '%Teste%' OR name LIKE '%Eduardo%'")

	a.registrarAudit("deletou_dados_teste", 0)
	return nil
}

func (a *App) ListarAccessLogs() ([]AccessLog, error) {
	var logs []AccessLog
	err := a.db.Order("timestamp DESC").Limit(100).Find(&logs).Error
	return logs, err
}

func (a *App) ListarAuditLogs() ([]AuditLog, error) {
	var logs []AuditLog
	err := a.db.Order("timestamp DESC").Limit(200).Find(&logs).Error
	return logs, err
}

// =============== Utilitários ===============

func (a *App) gerarProximaFatura(invoiceID uint) {
	type invData struct {
		StudentPlanID uint
		StudentID     uint
		PlanID        uint
		AmountCents   int
	}
	var cur invData
	err := a.db.Raw("SELECT student_plan_id, student_id, plan_id, amount_cents FROM invoices WHERE id = ?", invoiceID).Scan(&cur).Error
	if err != nil || cur.StudentPlanID == 0 {
		return
	}

	var dueDay int
	a.db.Raw("SELECT due_day FROM student_plans WHERE id = ?", cur.StudentPlanID).Scan(&dueDay)
	if dueDay == 0 {
		dueDay = 5
	}

	nextDue := nextDueDate(time.Now(), dueDay)

	var count int64
	a.db.Model(&Invoice{}).Where("student_plan_id = ? AND due_date = ?", cur.StudentPlanID, nextDue).Count(&count)
	if count > 0 {
		return
	}

	inv := Invoice{
		StudentPlanID: cur.StudentPlanID,
		StudentID:     cur.StudentID,
		PlanID:        cur.PlanID,
		AmountCents:   cur.AmountCents,
		Status:        "pending",
		DueDate:       nextDue,
	}
	if err := a.db.Create(&inv).Error; err == nil {
		a.registrarAudit("gerou_proxima_fatura", inv.ID)
	}
}

func nextDueDate(from time.Time, day int) string {
	next := time.Date(from.Year(), from.Month(), day, 0, 0, 0, 0, from.Location())
	if next.Before(from) || next.Equal(from) {
		next = next.AddDate(0, 1, 0)
	}
	return next.Format("2006-01-02")
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func uintPtr(n uint) *uint {
	if n == 0 {
		return nil
	}
	return &n
}

type PlanoContador struct {
	Nome string `json:"nome"`
	Qtd  int64  `json:"qtd"`
}

type MetricasHome struct {
	TotalUsuarios      int64           `json:"total_usuarios"`
	TotalAdmins        int64           `json:"total_admins"`
	TotalAlunos        int64           `json:"total_alunos"`
	TotalFaturadoCents int64           `json:"total_faturado_cents"`
	TotalAtrasoCents   int64           `json:"total_atraso_cents"`
	TotalPendenteCents int64           `json:"total_pendente_cents"`
	PlanosDistribuicao []PlanoContador `json:"planos_distribuicao"`
}

func (a *App) ObterMetricasHome() (*MetricasHome, error) {
	var stats MetricasHome

	// Total usuários
	a.db.Model(&User{}).Count(&stats.TotalUsuarios)

	// Admins
	a.db.Model(&User{}).Where("cargo = ? OR cargo = ?", "super_admin", "admin").Count(&stats.TotalAdmins)

	// Alunos matriculados
	a.db.Model(&Student{}).Count(&stats.TotalAlunos)

	// Total Faturado
	a.db.Model(&Invoice{}).Where("status = ?", "paid").Select("COALESCE(SUM(amount_cents), 0)").Scan(&stats.TotalFaturadoCents)

	// Total Atraso
	a.db.Model(&Invoice{}).Where("status = ?", "overdue").Select("COALESCE(SUM(amount_cents), 0)").Scan(&stats.TotalAtrasoCents)

	// Total Pendente
	a.db.Model(&Invoice{}).Where("status = ?", "pending").Select("COALESCE(SUM(amount_cents), 0)").Scan(&stats.TotalPendenteCents)

	// Distribuição de planos
	type result struct {
		Name string
		Qtd  int64
	}
	var dist []result
	err := a.db.Raw(`
		SELECT p.name, COUNT(sp.id) as qtd
		FROM plans p
		LEFT JOIN student_plans sp ON sp.plan_id = p.id AND sp.status = 'active'
		GROUP BY p.id
	`).Scan(&dist).Error
	if err == nil {
		stats.PlanosDistribuicao = make([]PlanoContador, len(dist))
		for i, r := range dist {
			stats.PlanosDistribuicao[i] = PlanoContador{
				Nome: r.Name,
				Qtd:  r.Qtd,
			}
		}
	}

	return &stats, nil
}

type AcessoEstatistica struct {
	Label    string `json:"label"`
	Liberado int    `json:"liberado"`
	Negado   int    `json:"negado"`
}

func (a *App) ObterEstatisticasAcesso(filtro string) ([]AcessoEstatistica, error) {
	var list []AcessoEstatistica
	var query string
	var since time.Time

	now := time.Now()

	switch filtro {
	case "dia":
		// Últimos 7 dias, agrupados por dia
		since = now.AddDate(0, 0, -6) // 7 dias incluindo hoje
		query = `
			SELECT strftime('%d/%m', data_hora) as label,
			       SUM(CASE WHEN liberado = 1 THEN 1 ELSE 0 END) as liberado,
			       SUM(CASE WHEN liberado = 0 THEN 1 ELSE 0 END) as negado
			FROM registro_acessos
			WHERE data_hora >= ?
			GROUP BY strftime('%Y-%m-%d', data_hora)
			ORDER BY data_hora ASC
		`
	case "semana":
		// Últimas 4 semanas, agrupadas por semana
		since = now.AddDate(0, 0, -27) // 28 dias
		query = `
			SELECT 'Sem ' || strftime('%W', data_hora) as label,
			       SUM(CASE WHEN liberado = 1 THEN 1 ELSE 0 END) as liberado,
			       SUM(CASE WHEN liberado = 0 THEN 1 ELSE 0 END) as negado
			FROM registro_acessos
			WHERE data_hora >= ?
			GROUP BY strftime('%Y-%W', data_hora)
			ORDER BY data_hora ASC
		`
	case "mes":
		// Últimos 6 meses, agrupados por mês
		since = now.AddDate(0, -5, 0) // 6 meses
		query = `
			SELECT strftime('%m/%Y', data_hora) as label,
			       SUM(CASE WHEN liberado = 1 THEN 1 ELSE 0 END) as liberado,
			       SUM(CASE WHEN liberado = 0 THEN 1 ELSE 0 END) as negado
			FROM registro_acessos
			WHERE data_hora >= ?
			GROUP BY strftime('%Y-%m', data_hora)
			ORDER BY data_hora ASC
		`
	default:
		return nil, fmt.Errorf("filtro inválido: %s", filtro)
	}

	err := a.db.Raw(query, since).Scan(&list).Error
	return list, err
}

func (a *App) SalvarPermissoesUsuario(userID uint, permissoes string) error {
	if a.actorCargo != "super_admin" {
		return fmt.Errorf("permissão negada: apenas super_admin pode alterar permissões")
	}
	err := a.db.Model(&User{}).Where("id = ? AND cargo != ?", userID, "super_admin").Update("permissoes", permissoes).Error
	if err == nil {
		a.registrarAuditComDetalhes("alterou_permissoes", userID, fmt.Sprintf("permissoes=%s", permissoes))
	}
	return err
}
