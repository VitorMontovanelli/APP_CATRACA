package inbound

import (
	"context"
	"fmt"

	"catraca-app/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type IdentityUsersHandler struct {
	ctx     context.Context
	db      *gorm.DB
	session *SessionManager
}

func NewIdentityUsersHandler(db *gorm.DB, session *SessionManager) *IdentityUsersHandler {
	return &IdentityUsersHandler{db: db, session: session}
}

func (h *IdentityUsersHandler) SetContext(ctx context.Context) {
	h.ctx = ctx
}

func uintPtr(v uint) *uint {
	return &v
}

func podeCriar(actorCargo, targetCargo string) bool {
	return actorCargo == "super_admin" && targetCargo == "admin"
}

func podeGerenciar(actorCargo, _ string) bool {
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

func (h *IdentityUsersHandler) CriarUsuario(name, email, senha, cargo string, registradorID uint) (*models.User, error) {
	if cargo != "super_admin" && cargo != "admin" {
		return nil, fmt.Errorf("cargo inválido: %s", cargo)
	}
	if !podeCriar(h.session.ActorCargo(), cargo) {
		return nil, fmt.Errorf("você não tem permissão para criar usuário com cargo %s", cargo)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar hash da senha: %w", err)
	}

	user := models.User{
		Name:          name,
		Email:         email,
		Senha:         string(hashed),
		Cargo:         cargo,
		RegistradorID: uintPtr(registradorID),
		Permissoes:    models.DefaultAdminPermissoes,
	}
	if err := h.db.Create(&user).Error; err != nil {
		return nil, err
	}
	h.session.RegistrarAudit("criou_usuario", user.ID)
	return h.BuscarUsuario(user.ID)
}

func (h *IdentityUsersHandler) ListarUsuarios() ([]models.User, error) {
	users := []models.User{}
	actorCargo := h.session.ActorCargo()
	actorID := h.session.ActorID()
	var err error
	if actorCargo == "super_admin" {
		err = h.db.Select("id, name, email, cargo, registrador_id, ativo, permissoes, criado_em").Order("id").Find(&users).Error
	} else {
		err = h.db.Select("id, name, email, cargo, registrador_id, ativo, permissoes, criado_em").
			Where("(id = ? OR registrador_id = ?) AND cargo != ?", actorID, actorID, "super_admin").
			Order("id").Find(&users).Error
	}
	return users, err
}

func (h *IdentityUsersHandler) BuscarUsuario(id uint) (*models.User, error) {
	var u models.User
	if err := h.db.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (h *IdentityUsersHandler) AtualizarUsuario(id uint, name, email, cargo string) (*models.User, error) {
	if cargo != "super_admin" && cargo != "admin" {
		return nil, fmt.Errorf("cargo inválido: %s", cargo)
	}
	u, err := h.BuscarUsuario(id)
	if err != nil {
		return nil, err
	}
	if !podeGerenciar(h.session.ActorCargo(), u.Cargo) {
		return nil, fmt.Errorf("você não tem permissão para alterar este usuário")
	}
	u.Name = name
	u.Email = email
	u.Cargo = cargo
	if err := h.db.Save(u).Error; err != nil {
		return nil, err
	}
	h.session.RegistrarAudit("editou_usuario", id)
	return h.BuscarUsuario(id)
}

func (h *IdentityUsersHandler) AtivarUsuario(id uint, ativo bool) error {
	u, err := h.BuscarUsuario(id)
	if err != nil {
		return err
	}
	if !podeGerenciar(h.session.ActorCargo(), u.Cargo) {
		return fmt.Errorf("você não tem permissão para %s este usuário", ativarDesativarLabel(ativo))
	}
	if err := h.db.Model(&models.User{}).Where("id = ?", id).Update("ativo", ativo).Error; err != nil {
		return err
	}
	h.session.RegistrarAudit(ativarDesativarAction(ativo), id)
	return nil
}

func (h *IdentityUsersHandler) DeletarUsuario(id uint) error {
	u, err := h.BuscarUsuario(id)
	if err != nil {
		return err
	}
	if !podeGerenciar(h.session.ActorCargo(), u.Cargo) {
		return fmt.Errorf("você não tem permissão para deletar este usuário")
	}
	if err := h.db.Delete(&u).Error; err != nil {
		return err
	}
	h.session.RegistrarAudit("deletou_usuario", id)
	return nil
}

func (h *IdentityUsersHandler) MeuPerfil() (*models.User, error) {
	return h.BuscarUsuario(h.session.ActorID())
}

func (h *IdentityUsersHandler) ListarAuditLogs() ([]models.AuditLog, error) {
	logs := []models.AuditLog{}
	err := h.db.Order("timestamp DESC").Limit(200).Find(&logs).Error
	return logs, err
}

func (h *IdentityUsersHandler) SalvarPermissoesUsuario(userID uint, permissoes string) error {
	if h.session.ActorCargo() != "super_admin" {
		return fmt.Errorf("permissão negada: apenas super_admin pode alterar permissões")
	}
	err := h.db.Model(&models.User{}).Where("id = ? AND cargo != ?", userID, "super_admin").Update("permissoes", permissoes).Error
	if err == nil {
		h.session.RegistrarAuditComDetalhes("alterou_permissoes", userID, fmt.Sprintf("permissoes=%s", permissoes))
	}
	return err
}

func (h *IdentityUsersHandler) RedefinirSenhaUsuario(userID uint, novaSenha string) error {
	actorID := h.session.ActorID()
	if h.session.ActorCargo() != "super_admin" && actorID != userID {
		return fmt.Errorf("permissão negada para alterar a senha")
	}
	if len(novaSenha) < 4 {
		return fmt.Errorf("a senha deve ter no mínimo 4 caracteres")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(novaSenha), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("erro ao gerar hash da senha: %w", err)
	}
	if err := h.db.Model(&models.User{}).Where("id = ?", userID).Update("senha", string(hashed)).Error; err != nil {
		return err
	}
	h.session.RegistrarAuditComDetalhes("redefiniu_senha", userID, fmt.Sprintf("por_actor=%d", actorID))
	return nil
}
