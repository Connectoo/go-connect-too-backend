package auth

import "github.com/google/uuid"

// RegisterCustomerRequest registers a customer account.
type RegisterCustomerRequest struct {
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Phone    *string `json:"phone,omitempty"`
	Password string  `json:"password"`
}

// RegisterEmployeeRequest registers an employee account.
type RegisterEmployeeRequest struct {
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Phone    *string `json:"phone,omitempty"`
	Password string  `json:"password"`
}

// LoginRequest authenticates a user.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshRequest exchanges a refresh token for new tokens.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// LogoutRequest revokes a refresh token.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// TokenPair is returned after login, register, or refresh.
type TokenPair struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	AccessExpiresAt  string `json:"access_expires_at"`
	RefreshExpiresAt string `json:"refresh_expires_at"`
}

// UserResponse is the public user profile.
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     *string   `json:"phone,omitempty"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt string    `json:"created_at"`
}

// AuthResponse wraps tokens and optional user on register/login.
type AuthResponse struct {
	User   *UserResponse `json:"user,omitempty"`
	Tokens TokenPair     `json:"tokens"`
}
