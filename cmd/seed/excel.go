package main

import (
	"fmt"
	"strconv"

	"github.com/xuri/excelize/v2"
)

func exportExcel(catalog *Catalog, path string) error {
	f := excelize.NewFile()
	defer f.Close()

	writeSheet(f, "Overview", []string{"Section", "Detail"}, rowsFromOverview(catalog.Overview))
	writeSheet(f, "Login Credentials", []string{
		"Name", "Email", "Password", "Role", "Phone", "Portal", "Login URL",
	}, rowsFromUsers(catalog.Users))
	writeSheet(f, "Categories", []string{"ID", "Name", "Description", "Active"}, rowsFromCategories(catalog.Categories))
	writeSheet(f, "Employees", []string{
		"Name", "Email", "Display Name", "Phone", "Location", "Verification",
		"Experience (yrs)", "Skills", "Languages", "Avg Rating", "Reviews",
	}, rowsFromEmployees(catalog.Employees))
	writeSheet(f, "Services", []string{
		"ID", "Employee", "Category", "Title", "Description", "Price (INR)", "Duration (min)", "Active",
	}, rowsFromServices(catalog.Services))
	writeSheet(f, "Customers", []string{"Name", "Email", "Phone", "User ID", "Profile ID"}, rowsFromCustomers(catalog.Customers))
	writeSheet(f, "Addresses", []string{"Customer", "Label", "Address", "City", "State", "Pincode", "Default"}, rowsFromAddresses(catalog.Addresses))
	writeSheet(f, "Availability", []string{"Employee", "Day", "Start", "End"}, rowsFromAvailability(catalog.Availability))
	writeSheet(f, "KYC", []string{"Employee", "Status", "ID Proof", "Address Proof", "Notes"}, rowsFromKYC(catalog.KYC))
	writeSheet(f, "Subscriptions", []string{"Employee", "Plan", "Status", "Starts", "Expires"}, rowsFromSubscriptions(catalog.Subscriptions))
	writeSheet(f, "Payments", []string{"Employee", "Plan", "Amount (INR)", "Status", "Provider"}, rowsFromPayments(catalog.Payments))
	writeSheet(f, "Bookings", []string{
		"ID", "Customer", "Employee", "Service", "Date", "Start", "End", "Status", "Amount (INR)", "Notes",
	}, rowsFromBookings(catalog.Bookings))
	writeSheet(f, "Reviews", []string{
		"Booking ID", "Customer", "Employee", "Rating", "Comment", "Status", "Employee Reply",
	}, rowsFromReviews(catalog.Reviews))
	writeSheet(f, "Notifications", []string{"User", "Type", "Title", "Body", "Read"}, rowsFromNotifications(catalog.Notifications))
	writeSheet(f, "Chat", []string{"Customer", "Employee", "Booking", "Messages", "Sample"}, rowsFromChat(catalog.Chat))
	writeSheet(f, "Support Tickets", []string{"Customer", "Subject", "Status", "Priority", "Messages"}, rowsFromSupport(catalog.Support))
	writeSheet(f, "Reports", []string{"Reporter", "Reported User", "Reason", "Status"}, rowsFromReports(catalog.Reports))

	f.DeleteSheet("Sheet1")
	if err := f.SaveAs(path); err != nil {
		return fmt.Errorf("save excel: %w", err)
	}
	return nil
}

func writeSheet(f *excelize.File, name string, headers []string, rows [][]string) {
	idx, _ := f.NewSheet(name)
	f.SetActiveSheet(idx)
	for col, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		_ = f.SetCellValue(name, cell, h)
	}
	for r, row := range rows {
		for c, val := range row {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
			_ = f.SetCellValue(name, cell, val)
		}
	}
	_ = f.SetPanes(name, &excelize.Panes{Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})
}

func rowsFromOverview(rows []OverviewRow) [][]string {
	out := make([][]string, len(rows))
	for i, r := range rows {
		out[i] = []string{r.Section, r.Detail}
	}
	return out
}

func rowsFromUsers(rows []UserRow) [][]string {
	out := make([][]string, len(rows))
	for i, r := range rows {
		out[i] = []string{r.Name, r.Email, r.Password, r.Role, r.Phone, r.Portal, r.PortalURL}
	}
	return out
}

func rowsFromCategories(rows []CategoryRow) [][]string {
	out := make([][]string, len(rows))
	for i, r := range rows {
		out[i] = []string{r.ID.String(), r.Name, r.Description, strconv.FormatBool(r.IsActive)}
	}
	return out
}

func rowsFromEmployees(rows []EmployeeRow) [][]string {
	out := make([][]string, len(rows))
	for i, r := range rows {
		out[i] = []string{
			r.Name, r.Email, r.DisplayName, r.Phone, r.Location, r.VerificationStatus,
			strconv.Itoa(r.ExperienceYears), r.Skills, r.Languages, r.AverageRating, strconv.Itoa(r.TotalReviews),
		}
	}
	return out
}

func rowsFromServices(rows []ServiceRow) [][]string {
	out := make([][]string, len(rows))
	for i, r := range rows {
		out[i] = []string{
			r.ID.String(), r.EmployeeName, r.Category, r.Title, r.Description,
			r.PriceINR, strconv.Itoa(r.DurationMinutes), strconv.FormatBool(r.IsActive),
		}
	}
	return out
}

func rowsFromCustomers(rows []CustomerRow) [][]string {
	out := make([][]string, len(rows))
	for i, r := range rows {
		out[i] = []string{r.Name, r.Email, r.Phone, r.UserID.String(), r.ProfileID.String()}
	}
	return out
}

func rowsFromAddresses(rows []AddressRow) [][]string {
	out := make([][]string, len(rows))
	for i, r := range rows {
		out[i] = []string{r.CustomerName, r.Label, r.AddressLine, r.City, r.State, r.Pincode, strconv.FormatBool(r.IsDefault)}
	}
	return out
}

func rowsFromAvailability(rows []AvailabilityRow) [][]string {
	out := make([][]string, len(rows))
	for i, r := range rows {
		out[i] = []string{r.EmployeeName, r.DayOfWeek, r.StartTime, r.EndTime}
	}
	return out
}

func rowsFromKYC(rows []KYCRow) [][]string {
	out := make([][]string, len(rows))
	for i, r := range rows {
		out[i] = []string{r.EmployeeName, r.Status, r.IDProof, r.AddressProof, r.Notes}
	}
	return out
}

func rowsFromSubscriptions(rows []SubscriptionRow) [][]string {
	out := make([][]string, len(rows))
	for i, r := range rows {
		out[i] = []string{r.EmployeeName, r.Plan, r.Status, r.StartsAt, r.ExpiresAt}
	}
	return out
}

func rowsFromPayments(rows []PaymentRow) [][]string {
	out := make([][]string, len(rows))
	for i, r := range rows {
		out[i] = []string{r.EmployeeName, r.Plan, r.AmountINR, r.Status, r.Provider}
	}
	return out
}

func rowsFromBookings(rows []BookingRow) [][]string {
	out := make([][]string, len(rows))
	for i, r := range rows {
		out[i] = []string{
			r.ID.String(), r.CustomerName, r.EmployeeName, r.ServiceTitle,
			r.Date, r.StartTime, r.EndTime, r.Status, r.AmountINR, r.CustomerNotes,
		}
	}
	return out
}

func rowsFromReviews(rows []ReviewRow) [][]string {
	out := make([][]string, len(rows))
	for i, r := range rows {
		reply := "No"
		if r.HasReply {
			reply = "Yes"
		}
		out[i] = []string{
			r.BookingID.String(), r.CustomerName, r.EmployeeName,
			strconv.Itoa(r.Rating), r.Comment, r.Status, reply,
		}
	}
	return out
}

func rowsFromNotifications(rows []NotificationRow) [][]string {
	out := make([][]string, len(rows))
	for i, r := range rows {
		out[i] = []string{r.UserName, r.Type, r.Title, r.Body, strconv.FormatBool(r.Read)}
	}
	return out
}

func rowsFromChat(rows []ChatRow) [][]string {
	out := make([][]string, len(rows))
	for i, r := range rows {
		out[i] = []string{r.CustomerName, r.EmployeeName, r.BookingRef, strconv.Itoa(r.MessageCount), r.SampleMessage}
	}
	return out
}

func rowsFromSupport(rows []SupportRow) [][]string {
	out := make([][]string, len(rows))
	for i, r := range rows {
		out[i] = []string{r.CustomerName, r.Subject, r.Status, r.Priority, strconv.Itoa(r.Messages)}
	}
	return out
}

func rowsFromReports(rows []ReportRow) [][]string {
	out := make([][]string, len(rows))
	for i, r := range rows {
		out[i] = []string{r.ReporterName, r.ReportedName, r.Reason, r.Status}
	}
	return out
}
