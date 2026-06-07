package main

import (
	"context"
	"database/sql"
	"fmt"
)

func cleanupDemoData(ctx context.Context, db *sql.DB) error {
	// Remove dependent rows first, then demo users (email domain marker).
	steps := []string{
		`DELETE FROM admin_audit_logs WHERE admin_user_id IN (SELECT id FROM users WHERE email LIKE '%@` + demoEmailDomain + `')`,
		`DELETE FROM support_messages WHERE ticket_id IN (
			SELECT st.id FROM support_tickets st
			JOIN customer_profiles cp ON cp.id = st.customer_id
			JOIN users u ON u.id = cp.user_id
			WHERE u.email LIKE '%@` + demoEmailDomain + `'
		)`,
		`DELETE FROM support_tickets WHERE customer_id IN (
			SELECT cp.id FROM customer_profiles cp
			JOIN users u ON u.id = cp.user_id
			WHERE u.email LIKE '%@` + demoEmailDomain + `'
		)`,
		`DELETE FROM chat_messages WHERE conversation_id IN (
			SELECT cc.id FROM chat_conversations cc
			JOIN customer_profiles cp ON cp.id = cc.customer_id
			JOIN users u ON u.id = cp.user_id
			WHERE u.email LIKE '%@` + demoEmailDomain + `'
		)`,
		`DELETE FROM chat_conversations WHERE customer_id IN (
			SELECT cp.id FROM customer_profiles cp
			JOIN users u ON u.id = cp.user_id
			WHERE u.email LIKE '%@` + demoEmailDomain + `'
		)`,
		`DELETE FROM notifications WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@` + demoEmailDomain + `')`,
		`DELETE FROM review_replies WHERE review_id IN (
			SELECT r.id FROM reviews r
			JOIN customer_profiles cp ON cp.id = r.customer_id
			JOIN users u ON u.id = cp.user_id
			WHERE u.email LIKE '%@` + demoEmailDomain + `'
		)`,
		`DELETE FROM reviews WHERE customer_id IN (
			SELECT cp.id FROM customer_profiles cp
			JOIN users u ON u.id = cp.user_id
			WHERE u.email LIKE '%@` + demoEmailDomain + `'
		)`,
		`DELETE FROM booking_status_history WHERE booking_id IN (
			SELECT b.id FROM bookings b
			JOIN customer_profiles cp ON cp.id = b.customer_id
			JOIN users u ON u.id = cp.user_id
			WHERE u.email LIKE '%@` + demoEmailDomain + `'
		)`,
		`DELETE FROM bookings WHERE customer_id IN (
			SELECT cp.id FROM customer_profiles cp
			JOIN users u ON u.id = cp.user_id
			WHERE u.email LIKE '%@` + demoEmailDomain + `'
		) OR employee_id IN (
			SELECT ep.id FROM employee_profiles ep
			JOIN users u ON u.id = ep.user_id
			WHERE u.email LIKE '%@` + demoEmailDomain + `'
		)`,
		`DELETE FROM refunds WHERE payment_id IN (
			SELECT p.id FROM payments p
			JOIN employee_profiles ep ON ep.id = p.employee_id
			JOIN users u ON u.id = ep.user_id
			WHERE u.email LIKE '%@` + demoEmailDomain + `'
		)`,
		`DELETE FROM payments WHERE employee_id IN (
			SELECT ep.id FROM employee_profiles ep
			JOIN users u ON u.id = ep.user_id
			WHERE u.email LIKE '%@` + demoEmailDomain + `'
		)`,
		`DELETE FROM subscription_changes WHERE subscription_id IN (
			SELECT es.id FROM employee_subscriptions es
			JOIN employee_profiles ep ON ep.id = es.employee_id
			JOIN users u ON u.id = ep.user_id
			WHERE u.email LIKE '%@` + demoEmailDomain + `'
		)`,
		`DELETE FROM employee_subscriptions WHERE employee_id IN (
			SELECT ep.id FROM employee_profiles ep
			JOIN users u ON u.id = ep.user_id
			WHERE u.email LIKE '%@` + demoEmailDomain + `'
		)`,
		`DELETE FROM employee_availability WHERE employee_id IN (
			SELECT ep.id FROM employee_profiles ep
			JOIN users u ON u.id = ep.user_id
			WHERE u.email LIKE '%@` + demoEmailDomain + `'
		)`,
		`DELETE FROM employee_services WHERE employee_id IN (
			SELECT ep.id FROM employee_profiles ep
			JOIN users u ON u.id = ep.user_id
			WHERE u.email LIKE '%@` + demoEmailDomain + `'
		)`,
		`DELETE FROM employee_kyc WHERE employee_id IN (
			SELECT ep.id FROM employee_profiles ep
			JOIN users u ON u.id = ep.user_id
			WHERE u.email LIKE '%@` + demoEmailDomain + `'
		)`,
		`DELETE FROM badges WHERE employee_id IN (
			SELECT ep.id FROM employee_profiles ep
			JOIN users u ON u.id = ep.user_id
			WHERE u.email LIKE '%@` + demoEmailDomain + `'
		)`,
		`DELETE FROM employee_profile_views WHERE employee_id IN (
			SELECT ep.id FROM employee_profiles ep
			JOIN users u ON u.id = ep.user_id
			WHERE u.email LIKE '%@` + demoEmailDomain + `'
		)`,
		`DELETE FROM reports WHERE reporter_id IN (SELECT id FROM users WHERE email LIKE '%@` + demoEmailDomain + `')`,
		`DELETE FROM user_addresses WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@` + demoEmailDomain + `')`,
		`DELETE FROM users WHERE email LIKE '%@` + demoEmailDomain + `'`,
		`DELETE FROM categories WHERE id >= '00000000-0000-4000-b000-000000000001'::uuid AND id <= '00000000-0000-4000-b000-000000000099'::uuid`,
	}

	for _, q := range steps {
		if _, err := db.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("cleanup: %w", err)
		}
	}
	return nil
}
