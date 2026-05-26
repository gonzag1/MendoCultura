package handlers

import (
	"MendoCultura/config"

	"gorm.io/gorm"
)

type Server struct {
	db     *gorm.DB
	config config.Config
}

func NewServer(db *gorm.DB, cfg config.Config) *Server {
	return &Server{db: db, config: cfg}
}
