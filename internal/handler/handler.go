package handler

import (
	"github.com/nickkcj/orbit-backend/internal/service"
	"github.com/nickkcj/orbit-backend/internal/worker"
)

type Handler struct {
	services   *service.Services
	taskClient *worker.TaskClient
}

func New(services *service.Services, taskClient *worker.TaskClient) *Handler {
	return &Handler{
		services:   services,
		taskClient: taskClient,
	}
}
