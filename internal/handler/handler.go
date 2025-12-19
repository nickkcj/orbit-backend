package handler

import (
	"github.com/nickkcj/orbit-backend/internal/service"
)

type Handler struct {
	services *service.Services
}

func New(services *service.Services) *Handler {
	return &Handler{
		services: services,
	}
}
