package workers

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/bookings"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/chat"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/customers"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/events"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/notifications"
	ws "github.com/MustafaKheda/go-connect-too-backend/internal/modules/websocket"
)

// CustomerProfileStore resolves customer profiles by id.
type CustomerProfileStore interface {
	GetByID(ctx context.Context, id uuid.UUID) (*customers.Profile, error)
}

// EmployeeProfileStore resolves employee profiles by id.
type EmployeeProfileStore interface {
	GetByID(ctx context.Context, id uuid.UUID) (*employees.Profile, error)
}

// NotificationCreator stores in-app notifications.
type NotificationCreator interface {
	Create(ctx context.Context, input notifications.CreateInput) (*notifications.NotificationResponse, error)
}

// PushSender delivers push notifications.
type PushSender interface {
	SendToUser(ctx context.Context, userID string, message notifications.PushMessage) error
}

// ChatConversationEnsurer creates conversations for bookings.
type ChatConversationEnsurer interface {
	EnsureConversationForBooking(ctx context.Context, bookingID, customerID, employeeID uuid.UUID) (*chat.Conversation, error)
}

// RealtimeBroadcaster pushes live updates to connected clients.
type RealtimeBroadcaster interface {
	SendToUser(userID uuid.UUID, payload []byte)
}

// EmailSender delivers transactional email.
type EmailSender interface {
	Send(ctx context.Context, message notifications.EmailMessage) error
}

// UserEmailLookup resolves user emails by id.
type UserEmailLookup interface {
	GetByID(ctx context.Context, id uuid.UUID) (email string, err error)
}

// NotificationWorker handles platform events and delivers notifications.
type NotificationWorker struct {
	customers CustomerProfileStore
	employees EmployeeProfileStore
	users     UserEmailLookup
	notifier  NotificationCreator
	push      PushSender
	email     EmailSender
	chat      ChatConversationEnsurer
	hub       RealtimeBroadcaster
	log       *slog.Logger
}

// NewNotificationWorker creates a notification worker.
func NewNotificationWorker(
	customers CustomerProfileStore,
	employees EmployeeProfileStore,
	notifier NotificationCreator,
	push PushSender,
	chatSvc ChatConversationEnsurer,
	hub RealtimeBroadcaster,
	log *slog.Logger,
) *NotificationWorker {
	return NewNotificationWorkerWithEmail(customers, employees, nil, notifier, push, nil, chatSvc, hub, log)
}

// NewNotificationWorkerWithEmail creates a notification worker with optional email delivery.
func NewNotificationWorkerWithEmail(
	customers CustomerProfileStore,
	employees EmployeeProfileStore,
	users UserEmailLookup,
	notifier NotificationCreator,
	push PushSender,
	email EmailSender,
	chatSvc ChatConversationEnsurer,
	hub RealtimeBroadcaster,
	log *slog.Logger,
) *NotificationWorker {
	if push == nil {
		push = notifications.NoopPushProvider{}
	}
	if email == nil {
		email = notifications.NoopEmailProvider{}
	}
	return &NotificationWorker{
		customers: customers,
		employees: employees,
		users:     users,
		notifier:  notifier,
		push:      push,
		email:     email,
		chat:      chatSvc,
		hub:       hub,
		log:       log,
	}
}

// Register subscribes the worker to platform events.
func (w *NotificationWorker) Register(dispatcher *events.Dispatcher) {
	types := []events.Type{
		events.TypeBookingCreated,
		events.TypeBookingAccepted,
		events.TypeBookingRejected,
		events.TypeBookingCancelled,
		events.TypeBookingCompleted,
		events.TypeMessageSent,
		events.TypePaymentSuccess,
		events.TypeSubscriptionExpiring,
		events.TypeKYCApproved,
		events.TypeKYCRejected,
	}
	for _, eventType := range types {
		dispatcher.Subscribe(eventType, w.Handle)
	}
}

// Handle processes a platform event.
func (w *NotificationWorker) Handle(ctx context.Context, event events.Event) {
	switch event.Type {
	case events.TypeBookingCreated,
		events.TypeBookingAccepted,
		events.TypeBookingRejected,
		events.TypeBookingCancelled,
		events.TypeBookingCompleted:
		w.handleBookingEvent(ctx, event)
	case events.TypeMessageSent:
		w.handleMessageSent(ctx, event)
	default:
		w.handleGenericEvent(ctx, event)
	}
}

func (w *NotificationWorker) handleBookingEvent(ctx context.Context, event events.Event) {
	bookingID, ok := event.UUIDPayload("booking_id")
	if !ok {
		w.log.Error("booking event missing booking_id", slog.String("type", string(event.Type)))
		return
	}
	customerProfileID, ok := event.UUIDPayload("customer_id")
	if !ok {
		w.log.Error("booking event missing customer_id", slog.String("type", string(event.Type)))
		return
	}
	employeeProfileID, ok := event.UUIDPayload("employee_id")
	if !ok {
		w.log.Error("booking event missing employee_id", slog.String("type", string(event.Type)))
		return
	}

	if event.Type == events.TypeBookingCreated && w.chat != nil {
		if _, err := w.chat.EnsureConversationForBooking(ctx, bookingID, customerProfileID, employeeProfileID); err != nil {
			w.log.Error("ensure chat conversation failed", slog.String("error", err.Error()))
		}
	}

	customer, err := w.customers.GetByID(ctx, customerProfileID)
	if err != nil {
		w.log.Error("resolve customer profile failed", slog.String("error", err.Error()))
		return
	}
	employee, err := w.employees.GetByID(ctx, employeeProfileID)
	if err != nil {
		w.log.Error("resolve employee profile failed", slog.String("error", err.Error()))
		return
	}

	title, body := bookingCopy(event.Type, event.StringPayload("status"))
	data := map[string]any{
		"booking_id": bookingID.String(),
		"status":     event.StringPayload("status"),
	}

	w.deliverToUser(ctx, customer.UserID, string(event.Type), title, body, data)
	w.deliverToUser(ctx, employee.UserID, string(event.Type), title, body, data)
	w.sendBookingEmail(ctx, customer.UserID, title, body)
	w.sendBookingEmail(ctx, employee.UserID, title, body)
}

func (w *NotificationWorker) handleMessageSent(ctx context.Context, event events.Event) {
	recipientID, ok := event.UUIDPayload("recipient_user_id")
	if !ok {
		w.log.Error("message.sent missing recipient_user_id")
		return
	}

	title := "New message"
	body := event.StringPayload("message")
	data := map[string]any{
		"conversation_id": event.StringPayload("conversation_id"),
		"message_id":      event.StringPayload("message_id"),
		"sender_id":       event.StringPayload("sender_id"),
	}

	w.deliverToUser(ctx, recipientID, string(events.TypeMessageSent), title, body, data)

	if w.hub != nil {
		payload, err := ws.Encode(ws.Message{
			Type:    ws.MessageTypeChatMessage,
			Payload: data,
		})
		if err != nil {
			w.log.Error("encode chat websocket message failed", slog.String("error", err.Error()))
			return
		}
		w.hub.SendToUser(recipientID, payload)
	}
}

func (w *NotificationWorker) handleGenericEvent(ctx context.Context, event events.Event) {
	userID, ok := event.UUIDPayload("user_id")
	if !ok {
		return
	}

	title, body := genericCopy(event.Type)
	data := map[string]any{}
	for key, value := range event.Payload {
		data[key] = value
	}
	w.deliverToUser(ctx, userID, string(event.Type), title, body, data)
}

func (w *NotificationWorker) deliverToUser(
	ctx context.Context,
	userID uuid.UUID,
	eventType, title, body string,
	data map[string]any,
) {
	created, err := w.notifier.Create(ctx, notifications.CreateInput{
		UserID: userID,
		Type:   eventType,
		Title:  title,
		Body:   body,
		Data:   data,
	})
	if err != nil {
		w.log.Error("create notification failed", slog.String("error", err.Error()))
		return
	}

	if w.hub != nil {
		payload, err := ws.Encode(ws.Message{
			Type: ws.MessageTypeNotification,
			Payload: map[string]any{
				"id":         created.ID.String(),
				"type":       created.Type,
				"title":      created.Title,
				"body":       created.Body,
				"data":       created.Data,
				"created_at": created.CreatedAt,
			},
		})
		if err != nil {
			w.log.Error("encode notification websocket message failed", slog.String("error", err.Error()))
		} else {
			w.hub.SendToUser(userID, payload)
		}
	}

	pushData := map[string]string{"type": eventType}
	for key, value := range data {
		pushData[key] = fmt.Sprint(value)
	}
	if err := w.push.SendToUser(ctx, userID.String(), notifications.PushMessage{
		Title: title,
		Body:  body,
		Data:  pushData,
	}); err != nil {
		w.log.Error("push notification failed", slog.String("error", err.Error()))
	}
}

func (w *NotificationWorker) sendBookingEmail(ctx context.Context, userID uuid.UUID, title, body string) {
	if w.email == nil || w.users == nil {
		return
	}
	email, err := w.users.GetByID(ctx, userID)
	if err != nil || strings.TrimSpace(email) == "" {
		return
	}
	if err := w.email.Send(ctx, notifications.EmailMessage{
		To:      email,
		Subject: title,
		Body:    body,
	}); err != nil {
		w.log.Error("booking email failed", slog.String("error", err.Error()))
	}
}

func bookingCopy(eventType events.Type, status string) (string, string) {
	switch eventType {
	case events.TypeBookingCreated:
		return "Booking created", "Your booking request was submitted."
	case events.TypeBookingAccepted:
		return "Booking accepted", "Your booking was accepted."
	case events.TypeBookingRejected:
		return "Booking rejected", "Your booking was rejected."
	case events.TypeBookingCancelled:
		return "Booking cancelled", "Your booking was cancelled."
	case events.TypeBookingCompleted:
		return "Booking completed", "Your booking was marked completed."
	default:
		return "Booking update", fmt.Sprintf("Booking status is now %s.", status)
	}
}

func genericCopy(eventType events.Type) (string, string) {
	switch eventType {
	case events.TypePaymentSuccess:
		return "Payment successful", "Your payment was processed successfully."
	case events.TypeSubscriptionExpiring:
		return "Subscription expiring", "Your subscription is expiring soon."
	case events.TypeKYCApproved:
		return "KYC approved", "Your verification documents were approved."
	case events.TypeKYCRejected:
		return "KYC rejected", "Your verification documents were rejected."
	default:
		return "Notification", "You have a new notification."
	}
}

// BookingPublisher bridges booking module events into the platform dispatcher.
type BookingPublisher struct {
	dispatcher *events.Dispatcher
}

// NewBookingPublisher creates a booking event publisher.
func NewBookingPublisher(dispatcher *events.Dispatcher) *BookingPublisher {
	return &BookingPublisher{dispatcher: dispatcher}
}

// Publish implements bookings.EventPublisher.
func (p *BookingPublisher) Publish(ctx context.Context, event bookings.BookingEvent) {
	p.dispatcher.Publish(ctx, events.Event{
		Type: events.Type(event.Type),
		Payload: map[string]any{
			"booking_id":  event.BookingID.String(),
			"customer_id": event.CustomerID.String(),
			"employee_id": event.EmployeeID.String(),
			"status":      event.Status,
		},
	})
}
