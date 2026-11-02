package inbound

import (
	"sync"
	"time"

	"catraca-app/internal/models"
	"gorm.io/gorm"
)

type SessionManager struct {
	mu         sync.RWMutex
	db         *gorm.DB
	actorCargo string
	actorID    uint
}

func NewSessionManager(db *gorm.DB) *SessionManager {
	return &SessionManager{db: db}
}

func (s *SessionManager) SetActor(id uint, cargo string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.actorID = id
	s.actorCargo = cargo
}

func (s *SessionManager) Actor() (uint, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.actorID, s.actorCargo
}

func (s *SessionManager) ActorCargo() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.actorCargo
}

func (s *SessionManager) ActorID() uint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.actorID
}

func (s *SessionManager) IsAdmin() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.actorCargo == "super_admin" || s.actorCargo == "admin"
}

func (s *SessionManager) RegistrarAudit(action string, targetID uint) {
	s.RegistrarAuditComDetalhes(action, targetID, "")
}

func (s *SessionManager) RegistrarAuditComDetalhes(action string, targetID uint, details string) {
	s.mu.RLock()
	actorID := s.actorID
	s.mu.RUnlock()

	s.db.Create(&models.AuditLog{
		ActorID:   actorID,
		Action:    action,
		TargetID:  &targetID,
		Details:   details,
		Timestamp: time.Now(),
	})
}
