package payments

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/pagination"
)

type EmployeeProfileStore interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*employees.Profile, error)
}

type AdminStore interface {
	ListAdmin(ctx context.Context, filter AdminListFilter) ([]Payment, int, error)
}

type Store interface {
	CreateSubscriptionOrder(ctx context.Context, employeeID uuid.UUID, plan PlanSnapshot, order *GatewayOrder, at time.Time) (*Payment, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Payment, error)
	GetByIDForEmployee(ctx context.Context, id, employeeID uuid.UUID) (*Payment, error)
	ListByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]Payment, error)
	ListAll(ctx context.Context) ([]Payment, error)
	InsertWebhookEvent(ctx context.Context, event *WebhookEvent) (bool, error)
	MarkWebhookProcessed(ctx context.Context, eventID uuid.UUID, at time.Time) error
	ActivateByProviderOrder(ctx context.Context, provider, providerOrderID, providerPaymentID string, raw []byte, at time.Time) error
	CreateRefund(ctx context.Context, refund *Refund, at time.Time) (*Refund, error)
}

type Service struct {
	profiles EmployeeProfileStore
	store    Store
	razorpay Gateway
	keyID    string
	now      func() time.Time
}

func NewService(profiles EmployeeProfileStore, store Store, razorpay Gateway, keyID string) *Service {
	return &Service{profiles: profiles, store: store, razorpay: razorpay, keyID: keyID, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) CreateSubscriptionOrder(ctx context.Context, employeeID uuid.UUID, plan PlanSnapshot) (*CreateSubscriptionOrderResponse, error) {
	if employeeID == uuid.Nil || plan.ID == uuid.Nil {
		return nil, fmt.Errorf("%w: employee_id and plan_id are required", ErrValidation)
	}
	if plan.Price <= 0 {
		return nil, fmt.Errorf("%w: paid plan amount must be greater than zero", ErrValidation)
	}
	order, err := s.razorpay.CreateOrder(ctx, CreateOrderInput{Amount: plan.Price, Currency: plan.Currency, Receipt: uuid.NewString(), Notes: map[string]string{"employee_id": employeeID.String(), "plan_id": plan.ID.String(), "plan_name": plan.Name}})
	if err != nil {
		return nil, err
	}
	payment, err := s.store.CreateSubscriptionOrder(ctx, employeeID, plan, order, s.now())
	if err != nil {
		return nil, err
	}
	return &CreateSubscriptionOrderResponse{PaymentID: payment.ID, SubscriptionID: payment.SubscriptionID, Provider: payment.Provider, ProviderOrderID: payment.ProviderOrderID, Amount: payment.Amount, Currency: payment.Currency, RazorpayKeyID: s.keyID}, nil
}

func (s *Service) ListForEmployeeUser(ctx context.Context, userID uuid.UUID) ([]PaymentResponse, error) {
	profile, err := s.profiles.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	items, err := s.store.ListByEmployeeID(ctx, profile.ID)
	if err != nil {
		return nil, err
	}
	return toResponses(items), nil
}

func (s *Service) ListAll(ctx context.Context, status string, page pagination.Params) (pagination.Result[PaymentResponse], error) {
	if adminStore, ok := s.store.(AdminStore); ok {
		items, total, err := adminStore.ListAdmin(ctx, AdminListFilter{
			Status: status,
			Offset: page.Offset(),
			Limit:  page.Limit,
		})
		if err != nil {
			return pagination.Result[PaymentResponse]{}, err
		}
		return pagination.NewResult(toResponses(items), page, total), nil
	}

	items, err := s.store.ListAll(ctx)
	if err != nil {
		return pagination.Result[PaymentResponse]{}, err
	}
	return pagination.NewResult(toResponses(items), page, len(items)), nil
}

func (s *Service) VerifySubscriptionPayment(ctx context.Context, employeeID uuid.UUID, req VerifyPaymentInput) (*PaymentResponse, error) {
	if req.PaymentID == uuid.Nil {
		return nil, fmt.Errorf("%w: payment_id is required", ErrValidation)
	}
	if strings.TrimSpace(req.ProviderOrderID) == "" || strings.TrimSpace(req.ProviderPaymentID) == "" || strings.TrimSpace(req.Signature) == "" {
		return nil, fmt.Errorf("%w: provider_order_id, provider_payment_id, and signature are required", ErrValidation)
	}
	if !s.razorpay.VerifyPayment(req.ProviderOrderID, req.ProviderPaymentID, req.Signature) {
		return nil, ErrInvalidSignature
	}

	payment, err := s.store.GetByIDForEmployee(ctx, req.PaymentID, employeeID)
	if err != nil {
		return nil, err
	}
	if payment.ProviderOrderID != req.ProviderOrderID {
		return nil, fmt.Errorf("%w: provider order mismatch", ErrValidation)
	}
	if payment.Status != StatusPending {
		return nil, ErrPaymentNotPending
	}

	raw, _ := json.Marshal(map[string]string{
		"provider_payment_id": req.ProviderPaymentID,
		"verified_by":         "client",
	})
	if err := s.store.ActivateByProviderOrder(ctx, ProviderRazorpay, req.ProviderOrderID, req.ProviderPaymentID, raw, s.now()); err != nil {
		return nil, err
	}

	updated, err := s.store.GetByID(ctx, payment.ID)
	if err != nil {
		return nil, err
	}
	return toResponse(updated), nil
}

func (s *Service) Refund(ctx context.Context, adminUserID, paymentID uuid.UUID, req RefundRequest) (*RefundResponse, error) {
	payment, err := s.store.GetByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}
	if payment.Status != StatusSuccess {
		return nil, ErrPaymentNotRefundable
	}

	amount := payment.Amount
	if req.Amount != nil {
		if *req.Amount <= 0 || *req.Amount > payment.Amount {
			return nil, fmt.Errorf("%w: invalid refund amount", ErrValidation)
		}
		amount = *req.Amount
	}

	reason := optionalRefundReason(req.Reason)
	at := s.now()
	refund := &Refund{
		ID:        uuid.New(),
		PaymentID: payment.ID,
		Amount:    amount,
		Reason:    reason,
		Status:    RefundStatusCompleted,
		CreatedBy: adminUserID,
		CreatedAt: at,
		UpdatedAt: at,
	}

	created, err := s.store.CreateRefund(ctx, refund, at)
	if err != nil {
		return nil, err
	}
	return toRefundResponse(created), nil
}

func optionalRefundReason(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func toRefundResponse(refund *Refund) *RefundResponse {
	return &RefundResponse{
		ID:        refund.ID,
		PaymentID: refund.PaymentID,
		Amount:    refund.Amount,
		Reason:    refund.Reason,
		Status:    refund.Status,
		CreatedAt: refund.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: refund.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func (s *Service) ProcessRazorpayWebhook(ctx context.Context, payload []byte, signature, eventID string) error {
	if !s.razorpay.VerifyWebhook(payload, signature) {
		return ErrInvalidSignature
	}
	event, err := parseRazorpayWebhook(payload)
	if err != nil {
		return err
	}
	if eventID == "" {
		eventID = event.ProviderEventID
	}
	if eventID == "" {
		return fmt.Errorf("%w: event_id is required", ErrValidation)
	}

	at := s.now()
	storedEvent := &WebhookEvent{ID: uuid.New(), Provider: ProviderRazorpay, EventID: eventID, EventType: event.EventType, Payload: payload, CreatedAt: at}
	inserted, err := s.store.InsertWebhookEvent(ctx, storedEvent)
	if err != nil {
		return err
	}
	if !inserted {
		return ErrDuplicateWebhook
	}

	if event.EventType != "payment.captured" {
		return nil
	}
	if err := s.store.ActivateByProviderOrder(ctx, ProviderRazorpay, event.OrderID, event.PaymentID, payload, at); err != nil {
		return err
	}
	return s.store.MarkWebhookProcessed(ctx, storedEvent.ID, at)
}

type razorpayWebhook struct{ ProviderEventID, EventType, OrderID, PaymentID string }

func parseRazorpayWebhook(payload []byte) (*razorpayWebhook, error) {
	var body struct {
		Account string `json:"account_id"`
		Event   string `json:"event"`
		Payload struct {
			Payment struct {
				Entity struct {
					ID      string `json:"id"`
					OrderID string `json:"order_id"`
				} `json:"entity"`
			} `json:"payment"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return nil, fmt.Errorf("%w: invalid webhook payload", ErrValidation)
	}
	if strings.TrimSpace(body.Event) == "" {
		return nil, fmt.Errorf("%w: event is required", ErrValidation)
	}
	return &razorpayWebhook{ProviderEventID: body.Account + ":" + body.Event + ":" + body.Payload.Payment.Entity.ID, EventType: body.Event, OrderID: body.Payload.Payment.Entity.OrderID, PaymentID: body.Payload.Payment.Entity.ID}, nil
}

func toResponses(items []Payment) []PaymentResponse {
	out := make([]PaymentResponse, 0, len(items))
	for i := range items {
		out = append(out, *toResponse(&items[i]))
	}
	return out
}

func toResponse(payment *Payment) *PaymentResponse {
	return &PaymentResponse{ID: payment.ID, EmployeeID: payment.EmployeeID, SubscriptionID: payment.SubscriptionID, Provider: payment.Provider, ProviderOrderID: payment.ProviderOrderID, ProviderPaymentID: payment.ProviderPaymentID, Amount: payment.Amount, Currency: payment.Currency, Status: payment.Status, CreatedAt: payment.CreatedAt.UTC().Format(time.RFC3339), UpdatedAt: payment.UpdatedAt.UTC().Format(time.RFC3339)}
}
