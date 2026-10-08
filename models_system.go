package main

type Setting struct {
	Key   string `gorm:"primaryKey;size:100" json:"key"`
	Value string `gorm:"size:2000" json:"value"`
}

type BackupConfig struct {
	Token      string `json:"token"`
	ChatID     string `json:"chat_id"`
	AutoBackup bool   `json:"auto_backup"`
	LastBackup string `json:"last_backup"`
	UpdatedAt  string `json:"updated_at"`
}
