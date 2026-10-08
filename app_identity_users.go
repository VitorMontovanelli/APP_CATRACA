package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

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
	users := []User{}
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

func (a *App) ListarAuditLogs() ([]AuditLog, error) {
	logs := []AuditLog{}
	err := a.db.Order("timestamp DESC").Limit(200).Find(&logs).Error
	return logs, err
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
