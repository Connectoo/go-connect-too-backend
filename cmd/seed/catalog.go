package main

import "github.com/google/uuid"

const (
	demoPassword           = "Demo123!"
	demoEmailDomain        = "yopmail.com"
	legacyDemoEmailDomain  = "demo.go-connect.local"
)

// Catalog collects every seeded entity for Excel export.
type Catalog struct {
	Overview      []OverviewRow
	Users         []UserRow
	Categories    []CategoryRow
	Employees     []EmployeeRow
	Services      []ServiceRow
	Customers     []CustomerRow
	Addresses     []AddressRow
	Availability  []AvailabilityRow
	KYC           []KYCRow
	Subscriptions []SubscriptionRow
	Payments      []PaymentRow
	Bookings      []BookingRow
	Reviews       []ReviewRow
	Notifications []NotificationRow
	Chat          []ChatRow
	Support       []SupportRow
	Reports       []ReportRow
}

type OverviewRow struct {
	Section string
	Detail  string
}

type UserRow struct {
	ID        uuid.UUID
	Name      string
	Email     string
	Phone     string
	Role      string
	Status    string
	Password  string
	Portal    string
	PortalURL string
}

type CategoryRow struct {
	ID          uuid.UUID
	Name        string
	Description string
	IsActive    bool
}

type EmployeeRow struct {
	UserID             uuid.UUID
	ProfileID          uuid.UUID
	Name               string
	Email              string
	DisplayName        string
	Phone              string
	Bio                string
	Location           string
	VerificationStatus string
	ExperienceYears    int
	Skills             string
	Languages          string
	AverageRating      string
	TotalReviews       int
}

type ServiceRow struct {
	ID              uuid.UUID
	EmployeeName    string
	Category        string
	Title           string
	Description     string
	PriceINR        string
	DurationMinutes int
	IsActive        bool
}

type CustomerRow struct {
	UserID    uuid.UUID
	ProfileID uuid.UUID
	Name      string
	Email     string
	Phone     string
}

type AddressRow struct {
	CustomerName string
	Label        string
	AddressLine  string
	City         string
	State        string
	Pincode      string
	IsDefault    bool
}

type AvailabilityRow struct {
	EmployeeName string
	DayOfWeek    string
	StartTime    string
	EndTime      string
}

type KYCRow struct {
	EmployeeName string
	Status       string
	IDProof      string
	AddressProof string
	Notes        string
}

type SubscriptionRow struct {
	EmployeeName string
	Plan         string
	Status       string
	StartsAt     string
	ExpiresAt    string
}

type PaymentRow struct {
	EmployeeName string
	Plan         string
	AmountINR    string
	Status       string
	Provider     string
}

type BookingRow struct {
	ID            uuid.UUID
	CustomerName  string
	EmployeeName  string
	ServiceTitle  string
	Date          string
	StartTime     string
	EndTime       string
	Status        string
	AmountINR     string
	CustomerNotes string
}

type ReviewRow struct {
	BookingID    uuid.UUID
	CustomerName string
	EmployeeName string
	Rating       int
	Comment      string
	Status       string
	HasReply     bool
}

type NotificationRow struct {
	UserName string
	Type     string
	Title    string
	Body     string
	Read     bool
}

type ChatRow struct {
	CustomerName  string
	EmployeeName  string
	BookingRef    string
	MessageCount  int
	SampleMessage string
}

type SupportRow struct {
	CustomerName string
	Subject      string
	Status       string
	Priority     string
	Messages     int
}

type ReportRow struct {
	ReporterName string
	ReportedName string
	Reason       string
	Status       string
}
