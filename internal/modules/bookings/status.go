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
	ActionCustomerCancel   TransitionAction = "customer_cancel"
	ActionEmployeeAccept   TransitionAction = "employee_accept"
	ActionEmployeeReject   TransitionAction = "employee_reject"
	ActionEmployeeStart    TransitionAction = "employee_start"
	ActionEmployeeComplete TransitionAction = "employee_complete"
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
	}
	return nil
}

// ActiveStatuses block overlapping bookings for the same employee slot.
func ActiveStatuses() []string {
	return []string{StatusPending, StatusAccepted, StatusInProgress}
}
