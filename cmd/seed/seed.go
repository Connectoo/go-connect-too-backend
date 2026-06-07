package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/security"
	"github.com/google/uuid"
)

const (
	planStarter = "00000000-0000-0000-0000-000000000602"
)

var dayNames = []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}

type seeder struct {
	db           *sql.DB
	catalog      *Catalog
	now          time.Time
	passwordHash string
}

func mustUUID(n int) uuid.UUID {
	return uuid.MustParse(fmt.Sprintf("00000000-0000-4000-a000-%012d", n))
}

func catUUID(n int) uuid.UUID {
	return uuid.MustParse(fmt.Sprintf("00000000-0000-4000-b000-%012d", n))
}

func (s *seeder) run(ctx context.Context) error {
	hash, err := security.HashPassword(demoPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	s.passwordHash = hash

	s.catalog.Overview = []OverviewRow{
		{"Password (all demo accounts)", demoPassword},
		{"API base URL", "http://localhost:8080/api/v1"},
		{"Website", "http://localhost:3000"},
		{"Admin portal", "http://localhost:3001/login"},
		{"Customer portal", "http://localhost:3002/login"},
		{"Employee portal", "http://localhost:3003/login"},
		{"Re-run seed", "make seed"},
		{"Reset demo data", "make seed (auto-cleans @demo.go-connect.local users first)"},
	}

	if err := s.seedCategories(ctx); err != nil {
		return err
	}
	if err := s.seedAdmin(ctx); err != nil {
		return err
	}
	if err := s.seedCustomers(ctx); err != nil {
		return err
	}
	if err := s.seedEmployees(ctx); err != nil {
		return err
	}
	if err := s.seedAddresses(ctx); err != nil {
		return err
	}
	if err := s.seedAvailability(ctx); err != nil {
		return err
	}
	if err := s.seedSubscriptions(ctx); err != nil {
		return err
	}
	if err := s.seedServices(ctx); err != nil {
		return err
	}
	if err := s.seedBookings(ctx); err != nil {
		return err
	}
	if err := s.seedReviews(ctx); err != nil {
		return err
	}
	if err := s.seedNotifications(ctx); err != nil {
		return err
	}
	if err := s.seedChat(ctx); err != nil {
		return err
	}
	if err := s.seedSupport(ctx); err != nil {
		return err
	}
	if err := s.seedReports(ctx); err != nil {
		return err
	}
	if err := s.seedBadges(ctx); err != nil {
		return err
	}
	return nil
}

func (s *seeder) seedCategories(ctx context.Context) error {
	categories := []struct {
		id   uuid.UUID
		name string
		desc string
	}{
		{catUUID(1), "Home Cleaning", "Deep cleaning, regular housekeeping, move-in/out cleaning"},
		{catUUID(2), "Plumbing", "Leak repair, pipe fitting, bathroom installations"},
		{catUUID(3), "Electrical", "Wiring, switchboard repair, appliance installation"},
		{catUUID(4), "Tutoring", "School subjects, exam prep, language coaching"},
		{catUUID(5), "Beauty & Wellness", "Salon at home, spa, grooming services"},
		{catUUID(6), "Fitness Training", "Personal training, yoga, nutrition coaching"},
		{catUUID(7), "Appliance Repair", "AC, refrigerator, washing machine repair"},
		{catUUID(8), "Pest Control", "Termite, rodent, and general pest treatment"},
	}

	for _, c := range categories {
		_, err := s.db.ExecContext(ctx, `
			INSERT INTO categories (id, name, description, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, true, $4, $4)
			ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description, is_active = true, updated_at = EXCLUDED.updated_at`,
			c.id, c.name, c.desc, s.now,
		)
		if err != nil {
			return fmt.Errorf("category %s: %w", c.name, err)
		}
		s.catalog.Categories = append(s.catalog.Categories, CategoryRow{
			ID: c.id, Name: c.name, Description: c.desc, IsActive: true,
		})
	}
	return nil
}

func (s *seeder) seedAdmin(ctx context.Context) error {
	id := mustUUID(1)
	email := "admin@" + demoEmailDomain
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO users (id, name, email, phone, password_hash, role, status, email_verified_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'admin', 'active', $6, $6, $6)
		ON CONFLICT (email, role) DO UPDATE SET
			name = EXCLUDED.name, password_hash = EXCLUDED.password_hash, status = 'active',
			email_verified_at = EXCLUDED.email_verified_at, updated_at = EXCLUDED.updated_at`,
		id, "Demo Admin", email, "+919900000001", s.passwordHash, s.now,
	)
	if err != nil {
		return fmt.Errorf("admin user: %w", err)
	}
	s.addUser(id, "Demo Admin", email, "+919900000001", "admin", "Admin dashboard", "http://localhost:3001/login")
	return nil
}

func (s *seeder) seedCustomers(ctx context.Context) error {
	customers := []struct {
		num   int
		name  string
		slug  string
		phone string
	}{
		{101, "Alice Sharma", "alice", "+919900010101"},
		{102, "Priya Patel", "priya", "+919900010102"},
		{103, "Rahul Mehta", "rahul", "+919900010103"},
		{104, "Ananya Iyer", "ananya", "+919900010104"},
	}

	for _, c := range customers {
		userID := mustUUID(c.num)
		profileID := mustUUID(c.num + 100)
		email := c.slug + "@" + demoEmailDomain

		_, err := s.db.ExecContext(ctx, `
			INSERT INTO users (id, name, email, phone, password_hash, role, status, email_verified_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, 'customer', 'active', $6, $6, $6)
			ON CONFLICT (email, role) DO UPDATE SET
				name = EXCLUDED.name, phone = EXCLUDED.phone, password_hash = EXCLUDED.password_hash,
				status = 'active', email_verified_at = EXCLUDED.email_verified_at, updated_at = EXCLUDED.updated_at
			RETURNING id`,
			userID, c.name, email, c.phone, s.passwordHash, s.now,
		)
		if err != nil {
			return fmt.Errorf("customer user %s: %w", c.name, err)
		}

		_, err = s.db.ExecContext(ctx, `
			INSERT INTO customer_profiles (id, user_id, created_at, updated_at)
			VALUES ($1, $2, $3, $3)
			ON CONFLICT (user_id) DO UPDATE SET updated_at = EXCLUDED.updated_at`,
			profileID, userID, s.now,
		)
		if err != nil {
			return fmt.Errorf("customer profile %s: %w", c.name, err)
		}

		s.addUser(userID, c.name, email, c.phone, "customer", "Customer portal", "http://localhost:3002/login")
		s.catalog.Customers = append(s.catalog.Customers, CustomerRow{
			UserID: userID, ProfileID: profileID, Name: c.name, Email: email, Phone: c.phone,
		})
	}
	return nil
}

func (s *seeder) seedEmployees(ctx context.Context) error {
	employees := []struct {
		num       int
		name      string
		slug      string
		phone     string
		display   string
		bio       string
		location  string
		lat, lng  float64
		skills    []string
		languages []string
		years     int
		verify    string
		kycStatus string
		kycNote   string
	}{
		{301, "Karim Hassan", "karim", "+919900020301", "Karim Home Services",
			"Professional cleaner with 8 years experience in residential deep cleaning.",
			"Indiranagar, Bangalore", 12.9784, 77.6408,
			[]string{"Deep cleaning", "Sanitization", "Move-out cleaning"},
			[]string{"English", "Hindi", "Kannada"}, 8, "approved", "approved", "Verified Aadhaar + utility bill"},
		{302, "Sara Khan", "sara", "+919900020302", "Sara Tutors",
			"Math and science tutor for grades 6–12. IIT graduate, patient teaching style.",
			"Koramangala, Bangalore", 12.9352, 77.6245,
			[]string{"Mathematics", "Physics", "CBSE", "JEE foundation"},
			[]string{"English", "Hindi"}, 5, "approved", "approved", "Verified PAN + address proof"},
		{303, "Dev Reddy", "dev", "+919900020303", "Dev Plumbing Co.",
			"Licensed plumber specializing in bathroom and kitchen installations.",
			"HSR Layout, Bangalore", 12.9116, 77.6473,
			[]string{"Pipe repair", "Bathroom fitting", "Water heater"},
			[]string{"English", "Telugu", "Kannada"}, 6, "pending", "pending", "Awaiting admin review"},
		{304, "Meera Nair", "meera", "+919900020304", "Meera Beauty Studio",
			"Certified beautician offering salon services at home.",
			"Jayanagar, Bangalore", 12.9308, 77.5838,
			[]string{"Facial", "Hair styling", "Bridal makeup"},
			[]string{"English", "Malayalam", "Hindi"}, 4, "rejected", "rejected", "ID proof unclear — resubmit required"},
	}

	for _, e := range employees {
		userID := mustUUID(e.num)
		profileID := mustUUID(e.num + 100)
		email := e.slug + "@" + demoEmailDomain

		_, err := s.db.ExecContext(ctx, `
			INSERT INTO users (id, name, email, phone, password_hash, role, status, email_verified_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, 'employee', 'active', $6, $6, $6)
			ON CONFLICT (email, role) DO UPDATE SET
				name = EXCLUDED.name, phone = EXCLUDED.phone, password_hash = EXCLUDED.password_hash,
				status = 'active', email_verified_at = EXCLUDED.email_verified_at, updated_at = EXCLUDED.updated_at`,
			userID, e.name, email, e.phone, s.passwordHash, s.now,
		)
		if err != nil {
			return fmt.Errorf("employee user %s: %w", e.name, err)
		}

		_, err = s.db.ExecContext(ctx, `
			INSERT INTO employee_profiles (
				id, user_id, display_name, phone, bio, experience_years, location_text,
				latitude, longitude, service_area_radius_km, languages, skills,
				verification_status, average_rating, total_reviews, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 10, $10, $11, $12, NULL, 0, $13, $13)
			ON CONFLICT (user_id) DO UPDATE SET
				display_name = EXCLUDED.display_name, phone = EXCLUDED.phone, bio = EXCLUDED.bio,
				experience_years = EXCLUDED.experience_years, location_text = EXCLUDED.location_text,
				latitude = EXCLUDED.latitude, longitude = EXCLUDED.longitude,
				languages = EXCLUDED.languages, skills = EXCLUDED.skills,
				verification_status = EXCLUDED.verification_status, updated_at = EXCLUDED.updated_at`,
			profileID, userID, e.display, e.phone, e.bio, e.years, e.location,
			e.lat, e.lng, e.languages, e.skills, e.verify, s.now,
		)
		if err != nil {
			return fmt.Errorf("employee profile %s: %w", e.name, err)
		}

		kycID := mustUUID(e.num + 200)
		rejectionReason := sql.NullString{}
		if e.kycStatus == "rejected" {
			rejectionReason = sql.NullString{String: e.kycNote, Valid: true}
		}
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO employee_kyc (id, employee_id, id_proof_url, address_proof_url, status, rejection_reason, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
			ON CONFLICT (employee_id) DO UPDATE SET
				status = EXCLUDED.status, rejection_reason = EXCLUDED.rejection_reason, updated_at = EXCLUDED.updated_at`,
			kycID, profileID,
			"https://demo.go-connect.local/kyc/"+e.slug+"-id.pdf",
			"https://demo.go-connect.local/kyc/"+e.slug+"-address.pdf",
			e.kycStatus, rejectionReason, s.now,
		)
		if err != nil {
			return fmt.Errorf("employee kyc %s: %w", e.name, err)
		}

		s.addUser(userID, e.name, email, e.phone, "employee", "Employee portal", "http://localhost:3003/login")
		s.catalog.Employees = append(s.catalog.Employees, EmployeeRow{
			UserID: userID, ProfileID: profileID, Name: e.name, Email: email,
			DisplayName: e.display, Phone: e.phone, Bio: e.bio, Location: e.location,
			VerificationStatus: e.verify, ExperienceYears: e.years,
			Skills: joinComma(e.skills), Languages: joinComma(e.languages),
		})
		s.catalog.KYC = append(s.catalog.KYC, KYCRow{
			EmployeeName: e.name, Status: e.kycStatus,
			IDProof:      "https://demo.go-connect.local/kyc/" + e.slug + "-id.pdf",
			AddressProof: "https://demo.go-connect.local/kyc/" + e.slug + "-address.pdf",
			Notes:        e.kycNote,
		})
	}
	return nil
}

func (s *seeder) seedAddresses(ctx context.Context) error {
	type addrDef struct {
		customerNum int
		label       string
		line        string
		city        string
		state       string
		pincode     string
		isDefault   bool
	}
	addrs := []addrDef{
		{101, "Home", "42 MG Road, Flat 3B", "Bangalore", "Karnataka", "560001", true},
		{101, "Office", "Tech Park, Tower A, Floor 5", "Bangalore", "Karnataka", "560103", false},
		{102, "Home", "18 5th Cross, JP Nagar", "Bangalore", "Karnataka", "560078", true},
		{103, "Home", "7 Residency Road", "Bangalore", "Karnataka", "560025", true},
		{104, "Home", "22 Church Street", "Bangalore", "Karnataka", "560001", true},
		{104, "Parents", "88 Lake View Apartments", "Mysore", "Karnataka", "570001", false},
	}

	names := map[int]string{101: "Alice Sharma", 102: "Priya Patel", 103: "Rahul Mehta", 104: "Ananya Iyer"}

	for i, a := range addrs {
		userID := mustUUID(a.customerNum)
		addrID := mustUUID(800 + i)
		_, err := s.db.ExecContext(ctx, `
			INSERT INTO user_addresses (id, user_id, label, address_line, city, state, country, pincode, is_default, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, 'India', $7, $8, $9, $9)
			ON CONFLICT (id) DO NOTHING`,
			addrID, userID, a.label, a.line, a.city, a.state, a.pincode, a.isDefault, s.now,
		)
		if err != nil {
			return fmt.Errorf("address %s: %w", a.label, err)
		}
		s.catalog.Addresses = append(s.catalog.Addresses, AddressRow{
			CustomerName: names[a.customerNum], Label: a.label, AddressLine: a.line,
			City: a.city, State: a.state, Pincode: a.pincode, IsDefault: a.isDefault,
		})
	}
	return nil
}

func (s *seeder) seedAvailability(ctx context.Context) error {
	// Approved employees: Mon–Sat 09:00–18:00
	for _, empNum := range []int{301, 302} {
		profileID := mustUUID(empNum + 100)
		name := map[int]string{301: "Karim Hassan", 302: "Sara Khan"}[empNum]
		for day := 1; day <= 6; day++ {
			slotID := mustUUID(900 + empNum*10 + day)
			_, err := s.db.ExecContext(ctx, `
				INSERT INTO employee_availability (id, employee_id, day_of_week, start_time, end_time, is_available, created_at, updated_at)
				VALUES ($1, $2, $3, '09:00', '18:00', true, $4, $4)
				ON CONFLICT (id) DO NOTHING`,
				slotID, profileID, day, s.now,
			)
			if err != nil {
				return fmt.Errorf("availability %s day %d: %w", name, day, err)
			}
			s.catalog.Availability = append(s.catalog.Availability, AvailabilityRow{
				EmployeeName: name, DayOfWeek: dayNames[day], StartTime: "09:00", EndTime: "18:00",
			})
		}
	}
	return nil
}

func (s *seeder) seedSubscriptions(ctx context.Context) error {
	starts := s.now.AddDate(0, 0, -10)
	expires := s.now.AddDate(0, 0, 20)

	for i, empNum := range []int{301, 302} {
		profileID := mustUUID(empNum + 100)
		name := map[int]string{301: "Karim Hassan", 302: "Sara Khan"}[empNum]
		subID := mustUUID(1000 + empNum)
		payID := mustUUID(1100 + empNum)

		_, err := s.db.ExecContext(ctx, `
			INSERT INTO employee_subscriptions (id, employee_id, plan_id, status, starts_at, expires_at, created_at, updated_at)
			VALUES ($1, $2, $3, 'active', $4, $5, $6, $6)
			ON CONFLICT (id) DO UPDATE SET status = 'active', starts_at = EXCLUDED.starts_at, expires_at = EXCLUDED.expires_at, updated_at = EXCLUDED.updated_at`,
			subID, profileID, planStarter, starts, expires, s.now,
		)
		if err != nil {
			return fmt.Errorf("subscription %s: %w", name, err)
		}

		amount := int64(49900)
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO payments (id, employee_id, subscription_id, provider, provider_order_id, provider_payment_id, amount, currency, status, created_at, updated_at)
			VALUES ($1, $2, $3, 'razorpay', $4, $5, $6, 'INR', 'success', $7, $7)
			ON CONFLICT (provider, provider_order_id) DO NOTHING`,
			payID, profileID, subID, fmt.Sprintf("order_demo_%d", empNum), fmt.Sprintf("pay_demo_%d", empNum), amount, s.now,
		)
		if err != nil {
			return fmt.Errorf("payment %s: %w", name, err)
		}

		s.catalog.Subscriptions = append(s.catalog.Subscriptions, SubscriptionRow{
			EmployeeName: name, Plan: "Starter (₹499/mo, 3 services)", Status: "active",
			StartsAt: starts.Format("2006-01-02"), ExpiresAt: expires.Format("2006-01-02"),
		})
		s.catalog.Payments = append(s.catalog.Payments, PaymentRow{
			EmployeeName: name, Plan: "Starter", AmountINR: "499.00", Status: "success", Provider: "razorpay",
		})
		_ = i
	}
	return nil
}

func (s *seeder) seedServices(ctx context.Context) error {
	type svcDef struct {
		id       uuid.UUID
		empNum   int
		catNum   int
		title    string
		desc     string
		price    string
		duration int
		active   bool
	}
	services := []svcDef{
		{mustUUID(601), 301, 1, "Standard Home Cleaning", "2BHK deep clean including kitchen and bathrooms", "1499.00", 120, true},
		{mustUUID(602), 301, 1, "Move-Out Cleaning", "Full property clean for tenancy handover", "2499.00", 180, true},
		{mustUUID(603), 301, 8, "Pre-Clean Pest Spray", "Pest control prep before deep cleaning", "799.00", 60, true},
		{mustUUID(604), 302, 4, "Math Tutoring (Grade 10)", "CBSE math — algebra, geometry, trigonometry", "600.00", 60, true},
		{mustUUID(605), 302, 4, "Physics Crash Course", "Intensive 2-hour session for board exams", "900.00", 120, true},
		{mustUUID(606), 302, 4, "JEE Foundation Math", "Advanced problem solving for competitive exams", "1200.00", 90, true},
		{mustUUID(607), 303, 2, "Kitchen Pipe Repair", "Fix leaks and replace fittings", "899.00", 90, false},
		{mustUUID(608), 304, 5, "Bridal Makeup Trial", "Full bridal look trial at home", "3500.00", 120, false},
	}

	empNames := map[int]string{301: "Karim Hassan", 302: "Sara Khan", 303: "Dev Reddy", 304: "Meera Nair"}
	catNames := map[int]string{1: "Home Cleaning", 2: "Plumbing", 4: "Tutoring", 5: "Beauty & Wellness", 8: "Pest Control"}

	for _, svc := range services {
		profileID := mustUUID(svc.empNum + 100)
		_, err := s.db.ExecContext(ctx, `
			INSERT INTO employee_services (id, employee_id, category_id, title, description, price, duration_minutes, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
			ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, description = EXCLUDED.description,
				price = EXCLUDED.price, duration_minutes = EXCLUDED.duration_minutes,
				is_active = EXCLUDED.is_active, updated_at = EXCLUDED.updated_at`,
			svc.id, profileID, catUUID(svc.catNum), svc.title, svc.desc, svc.price, svc.duration, svc.active, s.now,
		)
		if err != nil {
			return fmt.Errorf("service %s: %w", svc.title, err)
		}
		s.catalog.Services = append(s.catalog.Services, ServiceRow{
			ID: svc.id, EmployeeName: empNames[svc.empNum], Category: catNames[svc.catNum],
			Title: svc.title, Description: svc.desc, PriceINR: svc.price,
			DurationMinutes: svc.duration, IsActive: svc.active,
		})
	}
	return nil
}

func (s *seeder) seedBookings(ctx context.Context) error {
	type bookingDef struct {
		id         uuid.UUID
		custNum    int
		empNum     int
		svcID      uuid.UUID
		daysOffset int
		start      string
		end        string
		status     string
		notes      string
		amount     string
	}
	bookings := []bookingDef{
		{mustUUID(701), 101, 301, mustUUID(601), -14, "10:00", "12:00", "completed", "Please focus on kitchen", "1499.00"},
		{mustUUID(702), 102, 301, mustUUID(602), -7, "09:00", "12:00", "completed", "Move-out clean for 2BHK", "2499.00"},
		{mustUUID(703), 103, 302, mustUUID(604), -5, "16:00", "17:00", "completed", "Grade 10 algebra revision", "600.00"},
		{mustUUID(704), 104, 302, mustUUID(605), -3, "14:00", "16:00", "completed", "Board exam prep", "900.00"},
		{mustUUID(705), 101, 302, mustUUID(604), 2, "17:00", "18:00", "accepted", "Weekly math session", "600.00"},
		{mustUUID(706), 102, 301, mustUUID(601), 3, "11:00", "13:00", "pending", "Regular fortnightly clean", "1499.00"},
		{mustUUID(707), 103, 301, mustUUID(603), 4, "09:00", "10:00", "pending", "Before pest treatment", "799.00"},
		{mustUUID(708), 104, 301, mustUUID(601), 1, "10:00", "12:00", "in_progress", "Deep clean today", "1499.00"},
		{mustUUID(709), 101, 301, mustUUID(602), 5, "09:00", "12:00", "accepted", "End of lease cleaning", "2499.00"},
		{mustUUID(710), 102, 302, mustUUID(606), -2, "15:00", "16:30", "cancelled", "Cancelled — schedule conflict", "1200.00"},
		{mustUUID(711), 103, 302, mustUUID(604), 6, "18:00", "19:00", "rejected", "Outside service area", "600.00"},
		{mustUUID(712), 104, 302, mustUUID(604), -10, "16:00", "17:00", "no_show", "Customer did not attend", "600.00"},
	}

	custNames := map[int]string{101: "Alice Sharma", 102: "Priya Patel", 103: "Rahul Mehta", 104: "Ananya Iyer"}
	empNames := map[int]string{301: "Karim Hassan", 302: "Sara Khan"}
	svcTitles := map[uuid.UUID]string{
		mustUUID(601): "Standard Home Cleaning", mustUUID(602): "Move-Out Cleaning",
		mustUUID(603): "Pre-Clean Pest Spray", mustUUID(604): "Math Tutoring (Grade 10)",
		mustUUID(605): "Physics Crash Course", mustUUID(606): "JEE Foundation Math",
	}

	adminID := mustUUID(1)

	for _, b := range bookings {
		custProfile := mustUUID(b.custNum + 100)
		empProfile := mustUUID(b.empNum + 100)
		date := s.now.AddDate(0, 0, b.daysOffset).Format("2006-01-02")

		_, err := s.db.ExecContext(ctx, `
			INSERT INTO bookings (id, customer_id, employee_id, service_id, booking_date, start_time, end_time, status, customer_notes, total_amount, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $11)
			ON CONFLICT (id) DO UPDATE SET status = EXCLUDED.status, updated_at = EXCLUDED.updated_at`,
			b.id, custProfile, empProfile, b.svcID, date, b.start, b.end, b.status, b.notes, b.amount, s.now,
		)
		if err != nil {
			return fmt.Errorf("booking %s: %w", b.id, err)
		}

		histID := mustUUID(1200 + int(b.id[15])%100)
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO booking_status_history (id, booking_id, old_status, new_status, changed_by_user_id, reason, created_at)
			VALUES ($1, $2, NULL, 'pending', $3, 'Booking created', $4)
			ON CONFLICT (id) DO NOTHING`,
			histID, b.id, mustUUID(b.custNum), s.now,
		)
		if err != nil {
			return fmt.Errorf("booking history %s: %w", b.id, err)
		}
		if b.status != "pending" {
			histID2 := mustUUID(1300 + int(b.id[15])%100)
			changedBy := mustUUID(b.empNum)
			if b.status == "cancelled" {
				changedBy = mustUUID(b.custNum)
			}
			_, err = s.db.ExecContext(ctx, `
				INSERT INTO booking_status_history (id, booking_id, old_status, new_status, changed_by_user_id, reason, created_at)
				VALUES ($1, $2, 'pending', $3, $4, $5, $6)
				ON CONFLICT (id) DO NOTHING`,
				histID2, b.id, b.status, changedBy, "Status updated", s.now,
			)
			if err != nil {
				return fmt.Errorf("booking history2 %s: %w", b.id, err)
			}
		}
		_ = adminID

		s.catalog.Bookings = append(s.catalog.Bookings, BookingRow{
			ID: b.id, CustomerName: custNames[b.custNum], EmployeeName: empNames[b.empNum],
			ServiceTitle: svcTitles[b.svcID], Date: date, StartTime: b.start, EndTime: b.end,
			Status: b.status, AmountINR: b.amount, CustomerNotes: b.notes,
		})
	}
	return nil
}

func (s *seeder) seedReviews(ctx context.Context) error {
	type reviewDef struct {
		id      uuid.UUID
		booking uuid.UUID
		custNum int
		empNum  int
		rating  int
		comment string
		status  string
		reply   string
	}
	reviews := []reviewDef{
		{mustUUID(801), mustUUID(701), 101, 301, 5, "Excellent deep clean! Kitchen sparkles.", "approved", "Thank you Alice, glad you liked it!"},
		{mustUUID(802), mustUUID(702), 102, 301, 4, "Good job overall, missed one balcony corner.", "approved", "Thanks for the feedback, we'll improve."},
		{mustUUID(803), mustUUID(703), 103, 302, 5, "Sara explains concepts very clearly.", "approved", "Happy to help Rahul with your exams!"},
		{mustUUID(804), mustUUID(704), 104, 302, 5, "Great physics session, highly recommend.", "pending", ""},
	}

	custNames := map[int]string{101: "Alice Sharma", 102: "Priya Patel", 103: "Rahul Mehta", 104: "Ananya Iyer"}
	empNames := map[int]string{301: "Karim Hassan", 302: "Sara Khan"}

	for _, r := range reviews {
		custProfile := mustUUID(r.custNum + 100)
		empProfile := mustUUID(r.empNum + 100)
		_, err := s.db.ExecContext(ctx, `
			INSERT INTO reviews (id, booking_id, customer_id, employee_id, rating, comment, review_images, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, '[]'::jsonb, $7, $8, $8)
			ON CONFLICT (booking_id) DO UPDATE SET rating = EXCLUDED.rating, comment = EXCLUDED.comment, status = EXCLUDED.status, updated_at = EXCLUDED.updated_at`,
			r.id, r.booking, custProfile, empProfile, r.rating, r.comment, r.status, s.now,
		)
		if err != nil {
			return fmt.Errorf("review %s: %w", r.id, err)
		}
		if r.reply != "" {
			replyID := mustUUID(1400 + int(r.id[15])%50)
			_, err = s.db.ExecContext(ctx, `
				INSERT INTO review_replies (id, review_id, employee_id, reply, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $5)
				ON CONFLICT (review_id) DO UPDATE SET reply = EXCLUDED.reply, updated_at = EXCLUDED.updated_at`,
				replyID, r.id, empProfile, r.reply, s.now,
			)
			if err != nil {
				return fmt.Errorf("review reply %s: %w", r.id, err)
			}
		}
		s.catalog.Reviews = append(s.catalog.Reviews, ReviewRow{
			BookingID: r.booking, CustomerName: custNames[r.custNum], EmployeeName: empNames[r.empNum],
			Rating: r.rating, Comment: r.comment, Status: r.status, HasReply: r.reply != "",
		})
	}

	// Update employee ratings for approved reviews
	for _, empNum := range []int{301, 302} {
		profileID := mustUUID(empNum + 100)
		_, err := s.db.ExecContext(ctx, `
			UPDATE employee_profiles ep SET
				average_rating = sub.avg_rating,
				total_reviews = sub.cnt,
				updated_at = $2
			FROM (
				SELECT employee_id, ROUND(AVG(rating)::numeric, 2) AS avg_rating, COUNT(*) AS cnt
				FROM reviews WHERE employee_id = $1 AND status = 'approved'
				GROUP BY employee_id
			) sub
			WHERE ep.id = sub.employee_id`,
			profileID, s.now,
		)
		if err != nil {
			return fmt.Errorf("update employee rating: %w", err)
		}
	}
	for i := range s.catalog.Employees {
		var avg sql.NullString
		var total int
		err := s.db.QueryRowContext(ctx, `
			SELECT average_rating, total_reviews FROM employee_profiles WHERE id = $1`,
			s.catalog.Employees[i].ProfileID,
		).Scan(&avg, &total)
		if err == nil {
			s.catalog.Employees[i].AverageRating = avg.String
			s.catalog.Employees[i].TotalReviews = total
		}
	}
	return nil
}

func (s *seeder) seedNotifications(ctx context.Context) error {
	type notifDef struct {
		userNum  int
		userName string
		typ      string
		title    string
		body     string
		read     bool
	}
	notifs := []notifDef{
		{101, "Alice Sharma", "booking_update", "Booking completed", "Your home cleaning with Karim Hassan is complete.", true},
		{102, "Priya Patel", "booking_update", "Booking accepted", "Karim Hassan accepted your cleaning request.", false},
		{301, "Karim Hassan", "booking_new", "New booking request", "Priya Patel requested Move-Out Cleaning.", false},
		{302, "Sara Khan", "review_new", "New review", "Rahul Mehta left you a 5-star review.", true},
		{303, "Dev Reddy", "kyc_update", "KYC under review", "Your documents are pending admin approval.", false},
		{104, "Ananya Iyer", "booking_update", "Booking in progress", "Your cleaning session with Karim Hassan has started.", false},
	}

	for i, n := range notifs {
		id := mustUUID(1500 + i)
		userID := mustUUID(n.userNum)
		readAt := sql.NullTime{}
		if n.read {
			readAt = sql.NullTime{Time: s.now.Add(-time.Hour), Valid: true}
		}
		data, _ := json.Marshal(map[string]string{"source": "seed"})
		_, err := s.db.ExecContext(ctx, `
			INSERT INTO notifications (id, user_id, type, title, body, data, read_at, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (id) DO NOTHING`,
			id, userID, n.typ, n.title, n.body, data, readAt, s.now,
		)
		if err != nil {
			return fmt.Errorf("notification %s: %w", n.title, err)
		}
		s.catalog.Notifications = append(s.catalog.Notifications, NotificationRow{
			UserName: n.userName, Type: n.typ, Title: n.title, Body: n.body, Read: n.read,
		})
	}
	return nil
}

func (s *seeder) seedChat(ctx context.Context) error {
	convID := mustUUID(1601)
	custProfile := mustUUID(101 + 100)
	empProfile := mustUUID(301 + 100)
	bookingID := mustUUID(709)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO chat_conversations (id, customer_id, employee_id, booking_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
		ON CONFLICT (id) DO NOTHING`,
		convID, custProfile, empProfile, bookingID, s.now,
	)
	if err != nil {
		return fmt.Errorf("chat conversation: %w", err)
	}

	messages := []struct {
		senderNum int
		text      string
	}{
		{101, "Hi Karim, can we start 15 minutes early?"},
		{301, "Sure Alice, I'll be there at 4:45 PM."},
		{101, "Perfect, thank you!"},
	}
	for i, m := range messages {
		msgID := mustUUID(1610 + i)
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO chat_messages (id, conversation_id, sender_id, message, created_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO NOTHING`,
			msgID, convID, mustUUID(m.senderNum), m.text, s.now,
		)
		if err != nil {
			return fmt.Errorf("chat message: %w", err)
		}
	}
	s.catalog.Chat = append(s.catalog.Chat, ChatRow{
		CustomerName: "Alice Sharma", EmployeeName: "Karim Hassan",
		BookingRef: "709 (End of lease cleaning — accepted)", MessageCount: 3,
		SampleMessage: messages[0].text,
	})
	return nil
}

func (s *seeder) seedSupport(ctx context.Context) error {
	ticketID := mustUUID(1701)
	custProfile := mustUUID(103 + 100)
	custUser := mustUUID(103)
	adminUser := mustUUID(1)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO support_tickets (id, customer_id, subject, status, priority, created_at, updated_at)
		VALUES ($1, $2, $3, 'in_progress', 'normal', $4, $4)
		ON CONFLICT (id) DO NOTHING`,
		ticketID, custProfile, "Refund request for cancelled tutoring session", s.now,
	)
	if err != nil {
		return fmt.Errorf("support ticket: %w", err)
	}

	msgs := []struct {
		id      uuid.UUID
		sender  uuid.UUID
		text    string
		isStaff bool
	}{
		{mustUUID(1711), custUser, "I cancelled my session 2 days ago but haven't received confirmation.", false},
		{mustUUID(1712), adminUser, "We're reviewing your refund request and will update within 24 hours.", true},
	}
	for _, m := range msgs {
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO support_messages (id, ticket_id, sender_id, message, is_staff, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id) DO NOTHING`,
			m.id, ticketID, m.sender, m.text, m.isStaff, s.now,
		)
		if err != nil {
			return fmt.Errorf("support message: %w", err)
		}
	}
	s.catalog.Support = append(s.catalog.Support, SupportRow{
		CustomerName: "Rahul Mehta", Subject: "Refund request for cancelled tutoring session",
		Status: "in_progress", Priority: "normal", Messages: 2,
	})
	return nil
}

func (s *seeder) seedReports(ctx context.Context) error {
	reportID := mustUUID(1801)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO reports (id, reporter_id, reported_user_id, booking_id, reason, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, 'open', $7, $7)
		ON CONFLICT (id) DO NOTHING`,
		reportID, mustUUID(104), mustUUID(302), mustUUID(712),
		"no_show", "Customer reported missed appointment", s.now,
	)
	if err != nil {
		return fmt.Errorf("report: %w", err)
	}
	s.catalog.Reports = append(s.catalog.Reports, ReportRow{
		ReporterName: "Ananya Iyer", ReportedName: "Sara Khan",
		Reason: "no_show", Status: "open",
	})
	return nil
}

func (s *seeder) seedBadges(ctx context.Context) error {
	for _, empNum := range []int{301, 302} {
		_, err := s.db.ExecContext(ctx, `
			INSERT INTO badges (id, employee_id, badge_type, created_at)
			VALUES ($1, $2, 'verified_booking_review', $3)
			ON CONFLICT (employee_id, badge_type) DO NOTHING`,
			mustUUID(1900+empNum), mustUUID(empNum+100), s.now,
		)
		if err != nil {
			return fmt.Errorf("badge: %w", err)
		}
	}
	return nil
}

func (s *seeder) addUser(id uuid.UUID, name, email, phone, role, portal, portalURL string) {
	s.catalog.Users = append(s.catalog.Users, UserRow{
		ID: id, Name: name, Email: email, Phone: phone, Role: role,
		Status: "active", Password: demoPassword, Portal: portal, PortalURL: portalURL,
	})
}

func joinComma(items []string) string {
	out := ""
	for i, v := range items {
		if i > 0 {
			out += ", "
		}
		out += v
	}
	return out
}
