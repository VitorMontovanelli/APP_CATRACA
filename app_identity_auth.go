package main

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

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
