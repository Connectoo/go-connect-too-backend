package subscriptions

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/payments"
)

type mockProfiles struct{ profile *employees.Profile }

func (m mockProfiles) GetByUserID(context.Context, uuid.UUID) (*employees.Profile, error) {
	return m.profile, nil
}

type mockStore struct{ plan *Plan }

func (m mockStore) ListActivePlans(context.Context) ([]Plan, error)       { return []Plan{*m.plan}, nil }
func (m mockStore) GetPlanByID(context.Context, uuid.UUID) (*Plan, error) { return m.plan, nil }
func (m mockStore) CreatePlan(context.Context, *Plan) (*Plan, error)      { return nil, nil }
func (m mockStore) UpdatePlan(context.Context, uuid.UUID, *Plan, time.Time) (*Plan, error) {
	return nil, nil
}
func (m mockStore) CurrentByEmployeeID(context.Context, uuid.UUID, time.Time) (*EmployeeSubscription, error) {
	return nil, ErrNotFound
}
func (m mockStore) ListAllSubscriptions(context.Context) ([]EmployeeSubscription, error) {
	return nil, nil
}

type mockPayments struct{ called bool }

func (m *mockPayments) CreateSubscriptionOrder(context.Context, uuid.UUID, payments.PlanSnapshot) (*payments.CreateSubscriptionOrderResponse, error) {
	m.called = true
	return &payments.CreateSubscriptionOrderResponse{}, nil
}

func TestService_CreateOrder_rejectsFreeTrialPaymentOrder(t *testing.T) {
	pay := &mockPayments{}
	svc := NewService(mockProfiles{profile: &employees.Profile{ID: uuid.New()}}, mockStore{plan: &Plan{ID: uuid.New(), Price: 0, IsActive: true}}, pay)

	_, err := svc.CreateOrder(context.Background(), uuid.New(), CreateOrderRequest{PlanID: uuid.New()})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("CreateOrder() error = %v, want ErrValidation", err)
	}
	if pay.called {
		t.Fatal("payment order creator called for free trial")
	}
}
