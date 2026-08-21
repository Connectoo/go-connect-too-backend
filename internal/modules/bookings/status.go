package bookings

// Booking status values.
const (
	StatusPending    = "pending"
	StatusAccepted   = "accepted"
	StatusRejected   = "rejected"
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"
	StatusCancelled  = "cancelled"
	StatusNoShow     = "no_show"
)

// TransitionAction identifies who performs a status change.
type TransitionAction string

const (
	ActionCustomerCancel     TransitionAction = "customer_cancel"
	ActionCustomerReschedule TransitionAction = "customer_reschedule"
	ActionEmployeeAccept     TransitionAction = "employee_accept"
	ActionEmployeeReject     TransitionAction = "employee_reject"
	ActionEmployeeStart      TransitionAction = "employee_start"
	ActionEmployeeComplete   TransitionAction = "employee_complete"
	ActionEmployeeCancel     TransitionAction = "employee_cancel"
	ActionEmployeeReschedule TransitionAction = "employee_reschedule"
	ActionEmployeeNoShow     TransitionAction = "employee_no_show"
	ActionAdminUpdateStatus  TransitionAction = "admin_update_status"
)

// ValidateTransition reports whether a status change is allowed for the given action.
func ValidateTransition(from, to string, action TransitionAction) error {
	if from == to {
		return ErrInvalidStatusTransition
	}

	allowed := transitionTargets(from, action)
	for _, status := range allowed {
		if status == to {
			return nil
		}
	}
	return ErrInvalidStatusTransition
}

func transitionTargets(from string, action TransitionAction) []string {
	switch action {
	case ActionCustomerCancel:
		if from == StatusPending || from == StatusAccepted {
			return []string{StatusCancelled}
		}
	case ActionEmployeeAccept:
		if from == StatusPending {
			return []string{StatusAccepted}
		}
	case ActionEmployeeReject:
		if from == StatusPending {
			return []string{StatusRejected}
		}
	case ActionEmployeeStart:
		if from == StatusAccepted {
			return []string{StatusInProgress}
		}
	case ActionEmployeeComplete:
		if from == StatusInProgress {
			return []string{StatusCompleted}
		}
	case ActionEmployeeCancel:
		if from == StatusPending || from == StatusAccepted {
			return []string{StatusCancelled}
		}
	case ActionEmployeeNoShow:
		if from == StatusAccepted || from == StatusInProgress {
			return []string{StatusNoShow}
		}
	case ActionAdminUpdateStatus:
		return allStatuses()
	}
	return nil
}

func allStatuses() []string {
	return []string{
		StatusPending,
		StatusAccepted,
		StatusRejected,
		StatusInProgress,
		StatusCompleted,
		StatusCancelled,
		StatusNoShow,
	}
}

// CanReschedule reports whether a booking may change its schedule.
func CanReschedule(status string) bool {
	return status == StatusPending || status == StatusAccepted
}

// ActiveStatuses block overlapping bookings for the same employee slot.
func ActiveStatuses() []string {
	return []string{StatusPending, StatusAccepted, StatusInProgress}
}
