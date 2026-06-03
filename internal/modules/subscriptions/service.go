package subscriptions

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/payments"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/pagination"
)

const maxPlanNameLength = 100

type EmployeeProfileStore interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*employees.Profile, error)
}

type AdminStore interface {
	ListAdmin(ctx context.Context, filter AdminListFilter) ([]EmployeeSubscription, int, error)
}

type Store interface {
	ListActivePlans(ctx context.Context) ([]Plan, error)
	GetPlanByID(ctx context.Context, id uuid.UUID) (*Plan, error)
	CreatePlan(ctx context.Context, plan *Plan) (*Plan, error)
	UpdatePlan(ctx context.Context, id uuid.UUID, plan *Plan, at time.Time) (*Plan, error)
	CurrentByEmployeeID(ctx context.Context, employeeID uuid.UUID, at time.Time) (*EmployeeSubscription, error)
	ListAllSubscriptions(ctx context.Context) ([]EmployeeSubscription, error)
}

type PaymentOrderCreator interface {
	CreateSubscriptionOrder(ctx context.Context, employeeID uuid.UUID, plan payments.PlanSnapshot) (*payments.CreateSubscriptionOrderResponse, error)
}

type Service struct {
	profiles EmployeeProfileStore
	store    Store
	payments PaymentOrderCreator
	now      func() time.Time
}

func NewService(profiles EmployeeProfileStore, store Store, payments PaymentOrderCreator) *Service {
	return &Service{profiles: profiles, store: store, payments: payments, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) ListPlans(ctx context.Context) ([]PlanResponse, error) {
	plans, err := s.store.ListActivePlans(ctx)
	if err != nil {
		return nil, err
	}
	return toPlanResponses(plans), nil
}

func (s *Service) CreateOrder(ctx context.Context, userID uuid.UUID, req CreateOrderRequest) (*payments.CreateSubscriptionOrderResponse, error) {
	if req.PlanID == uuid.Nil {
		return nil, fmt.Errorf("%w: plan_id is required", ErrValidation)
	}
	profile, err := s.profiles.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	plan, err := s.store.GetPlanByID(ctx, req.PlanID)
	if err != nil {
		return nil, err
	}
	if !plan.IsActive {
		return nil, ErrPlanInactive
	}
	if plan.Price == 0 {
		return nil, fmt.Errorf("%w: free trial does not require payment order", ErrValidation)
	}
	return s.payments.CreateSubscriptionOrder(ctx, profile.ID, payments.PlanSnapshot{ID: plan.ID, Name: plan.Name, Price: plan.Price, Currency: plan.Currency, DurationDays: plan.DurationDays})
}

func (s *Service) Current(ctx context.Context, userID uuid.UUID) (*SubscriptionResponse, error) {
	profile, err := s.profiles.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	sub, err := s.store.CurrentByEmployeeID(ctx, profile.ID, s.now())
	if err != nil {
		return nil, err
	}
	return toSubscriptionResponse(sub), nil
}

func (s *Service) CreatePlan(ctx context.Context, req CreatePlanRequest) (*PlanResponse, error) {
	plan, err := validatePlanRequest(req)
	if err != nil {
		return nil, err
	}
	at := s.now()
	plan.ID = uuid.New()
	plan.CreatedAt = at
	plan.UpdatedAt = at
	created, err := s.store.CreatePlan(ctx, plan)
	if err != nil {
		return nil, err
	}
	return toPlanResponse(created), nil
}

func (s *Service) UpdatePlan(ctx context.Context, id uuid.UUID, req UpdatePlanRequest) (*PlanResponse, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("%w: id is required", ErrValidation)
	}
	plan, err := validatePlanRequest(req)
	if err != nil {
		return nil, err
	}
	updated, err := s.store.UpdatePlan(ctx, id, plan, s.now())
	if err != nil {
		return nil, err
	}
	return toPlanResponse(updated), nil
}

func (s *Service) ListSubscriptions(ctx context.Context, status string, page pagination.Params) (pagination.Result[SubscriptionResponse], error) {
	if adminStore, ok := s.store.(AdminStore); ok {
		items, total, err := adminStore.ListAdmin(ctx, AdminListFilter{
			Status: status,
			Offset: page.Offset(),
			Limit:  page.Limit,
		})
		if err != nil {
			return pagination.Result[SubscriptionResponse]{}, err
		}
		out := make([]SubscriptionResponse, 0, len(items))
		for i := range items {
			out = append(out, *toSubscriptionResponse(&items[i]))
		}
		return pagination.NewResult(out, page, total), nil
	}

	items, err := s.store.ListAllSubscriptions(ctx)
	if err != nil {
		return pagination.Result[SubscriptionResponse]{}, err
	}
	out := make([]SubscriptionResponse, 0, len(items))
	for i := range items {
		out = append(out, *toSubscriptionResponse(&items[i]))
	}
	return pagination.NewResult(out, page, len(out)), nil
}

func validatePlanRequest(req CreatePlanRequest) (*Plan, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrValidation)
	}
	if utf8.RuneCountInString(name) > maxPlanNameLength {
		return nil, fmt.Errorf("%w: name is too long", ErrValidation)
	}
	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if len(currency) != 3 {
		return nil, fmt.Errorf("%w: currency must be ISO 4217 code", ErrValidation)
	}
	if req.Price < 0 || req.DurationDays <= 0 || req.ServiceLimit < -1 {
		return nil, fmt.Errorf("%w: invalid plan values", ErrValidation)
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	return &Plan{Name: name, Price: req.Price, Currency: currency, DurationDays: req.DurationDays, ServiceLimit: req.ServiceLimit, IsFeaturedAllowed: req.IsFeaturedAllowed, IsPriorityAllowed: req.IsPriorityAllowed, IsActive: isActive}, nil
}

func toPlanResponses(plans []Plan) []PlanResponse {
	out := make([]PlanResponse, 0, len(plans))
	for i := range plans {
		out = append(out, *toPlanResponse(&plans[i]))
	}
	return out
}
func toPlanResponse(plan *Plan) *PlanResponse {
	if plan == nil {
		return nil
	}
	return &PlanResponse{ID: plan.ID, Name: plan.Name, Price: plan.Price, Currency: plan.Currency, DurationDays: plan.DurationDays, ServiceLimit: plan.ServiceLimit, IsFeaturedAllowed: plan.IsFeaturedAllowed, IsPriorityAllowed: plan.IsPriorityAllowed, IsActive: plan.IsActive, CreatedAt: plan.CreatedAt.UTC().Format(time.RFC3339), UpdatedAt: plan.UpdatedAt.UTC().Format(time.RFC3339)}
}
func toSubscriptionResponse(sub *EmployeeSubscription) *SubscriptionResponse {
	var startsAt, expiresAt *string
	if sub.StartsAt != nil {
		value := sub.StartsAt.UTC().Format(time.RFC3339)
		startsAt = &value
	}
	if sub.ExpiresAt != nil {
		value := sub.ExpiresAt.UTC().Format(time.RFC3339)
		expiresAt = &value
	}
	return &SubscriptionResponse{ID: sub.ID, EmployeeID: sub.EmployeeID, PlanID: sub.PlanID, PlanName: sub.PlanName, Status: sub.Status, StartsAt: startsAt, ExpiresAt: expiresAt, Plan: toPlanResponse(sub.Plan), CreatedAt: sub.CreatedAt.UTC().Format(time.RFC3339), UpdatedAt: sub.UpdatedAt.UTC().Format(time.RFC3339)}
}
