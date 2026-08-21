package moderation

import (
	"log/slog"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/reports"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/reviews"
)

// Service coordinates admin moderation workflows.
type Service struct {
	reviews *reviews.Service
	reports *reports.Service
}

// NewService creates a moderation service.
func NewService(reviews *reviews.Service, reports *reports.Service) *Service {
	return &Service{
		reviews: reviews,
		reports: reports,
	}
}

// Reviews returns the reviews service for admin moderation handlers.
func (s *Service) Reviews() *reviews.Service {
	return s.reviews
}

// Reports returns the reports service for admin moderation handlers.
func (s *Service) Reports() *reports.Service {
	return s.reports
}

// Handler serves admin moderation endpoints.
type Handler struct {
	svc *Service
	log *slog.Logger
}

// NewHandler creates a moderation handler.
func NewHandler(svc *Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}
