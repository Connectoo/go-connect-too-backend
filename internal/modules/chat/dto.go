package chat

import (
	"time"

	"github.com/google/uuid"
)

// ConversationResponse is the API conversation payload.
type ConversationResponse struct {
	ID         uuid.UUID  `json:"id"`
	CustomerID uuid.UUID  `json:"customer_id"`
	EmployeeID uuid.UUID  `json:"employee_id"`
	BookingID  *uuid.UUID `json:"booking_id,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// MessageResponse is the API chat message payload.
type MessageResponse struct {
	ID             uuid.UUID  `json:"id"`
	ConversationID uuid.UUID  `json:"conversation_id"`
	SenderID       uuid.UUID  `json:"sender_id"`
	Message        string     `json:"message"`
	AttachmentURL  *string    `json:"attachment_url,omitempty"`
	ContentType    *string    `json:"content_type,omitempty"`
	ReadAt         *time.Time `json:"read_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// SendMessageRequest creates a chat message.
type SendMessageRequest struct {
	Message       string  `json:"message"`
	AttachmentURL *string `json:"attachment_url,omitempty"`
	ContentType   *string `json:"content_type,omitempty"`
}

func toConversationResponse(item *Conversation) ConversationResponse {
	return ConversationResponse{
		ID:         item.ID,
		CustomerID: item.CustomerID,
		EmployeeID: item.EmployeeID,
		BookingID:  item.BookingID,
		CreatedAt:  item.CreatedAt,
		UpdatedAt:  item.UpdatedAt,
	}
}

func toConversationResponses(items []Conversation) []ConversationResponse {
	out := make([]ConversationResponse, 0, len(items))
	for i := range items {
		out = append(out, toConversationResponse(&items[i]))
	}
	return out
}

func toMessageResponse(item *Message) MessageResponse {
	return MessageResponse{
		ID:             item.ID,
		ConversationID: item.ConversationID,
		SenderID:       item.SenderID,
		Message:        item.Message,
		AttachmentURL:  item.AttachmentURL,
		ContentType:    item.ContentType,
		ReadAt:         item.ReadAt,
		CreatedAt:      item.CreatedAt,
	}
}

func toMessageResponses(items []Message) []MessageResponse {
	out := make([]MessageResponse, 0, len(items))
	for i := range items {
		out = append(out, toMessageResponse(&items[i]))
	}
	return out
}
