package handlers

import (
	"MendoCultura/config"
	"MendoCultura/services"
	"net/http"

	"gorm.io/gorm"
)

type Server struct {
	db       *gorm.DB
	config   config.Config
	mpClient *http.Client
}

func (s *Server) mercadoPago() services.MercadoPago {
	return services.MercadoPago{Token: s.config.MPAccessToken, Client: s.mpClient}
}

func NewServer(db *gorm.DB, cfg config.Config) *Server {
	return &Server{db: db, config: cfg}
}
