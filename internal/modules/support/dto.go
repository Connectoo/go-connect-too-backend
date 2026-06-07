package support

import (
	"time"

	"github.com/google/uuid"
)

type CreateTicketRequest struct {
	Subject string `json:"subject"`
	Message string `json:"message"`
}

type UpdateTicketRequest struct {
	Status   *string `json:"status,omitempty"`
	Priority *string `json:"priority,omitempty"`
}

type AddMessageRequest struct {
	Message string `json:"message"`
}

type TicketResponse struct {
	ID         uuid.UUID `json:"id"`
	CustomerID uuid.UUID `json:"customer_id"`
	Subject    string    `json:"subject"`
	Status     string    `json:"status"`
	Priority   string    `json:"priority"`
	CreatedAt  string    `json:"created_at"`
	UpdatedAt  string    `json:"updated_at"`
}

type MessageResponse struct {
	ID        uuid.UUID `json:"id"`
	TicketID  uuid.UUID `json:"ticket_id"`
	SenderID  uuid.UUID `json:"sender_id"`
	Message   string    `json:"message"`
	IsStaff   bool      `json:"is_staff"`
	CreatedAt string    `json:"created_at"`
}

type TicketDetailResponse struct {
	TicketResponse
	Messages []MessageResponse `json:"messages"`
}

func toTicketResponse(ticket *Ticket) *TicketResponse {
	return &TicketResponse{
		ID:         ticket.ID,
		CustomerID: ticket.CustomerID,
		Subject:    ticket.Subject,
		Status:     ticket.Status,
		Priority:   ticket.Priority,
		CreatedAt:  ticket.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:  ticket.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toMessageResponse(message *Message) MessageResponse {
	return MessageResponse{
		ID:        message.ID,
		TicketID:  message.TicketID,
		SenderID:  message.SenderID,
		Message:   message.Message,
		IsStaff:   message.IsStaff,
		CreatedAt: message.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func toTicketResponses(items []Ticket) []TicketResponse {
	out := make([]TicketResponse, 0, len(items))
	for i := range items {
		out = append(out, *toTicketResponse(&items[i]))
	}
	return out
}
