package main

import "fmt"

func (a *App) CriarStudentComPlano(nome, cpf, dataNascimento, telefone, telefoneUrgencia, email string, formaPagamentoID, planID, dueDay uint) (*Student, error) {
	if a.actorCargo != "super_admin" && a.actorCargo != "admin" {
		return nil, fmt.Errorf("permissão negada")
	}

	tx := a.db.Begin()

	student := Student{
		Nome:             nome,
		CPF:              cpf,
		DataNascimento:   strPtr(dataNascimento),
		Telefone:         strPtr(telefone),
		TelefoneUrgencia: strPtr(telefoneUrgencia),
		Email:            strPtr(email),
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
