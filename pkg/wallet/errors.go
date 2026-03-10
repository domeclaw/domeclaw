package wallet

import (
	"errors"
	"fmt"
)

var (
	// General errors
	ErrWalletDisabled     = errors.New("wallet functionality is disabled")
	ErrInvalidAddress     = errors.New("invalid Ethereum address")
	ErrInvalidAmount      = errors.New("invalid amount")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrAccountNotFound    = errors.New("account not found")
	ErrChainNotConfigured = errors.New("chain not configured")
	
	// Network errors
	ErrNetworkFailure     = errors.New("network failure")
	ErrRPCConnection      = errors.New("RPC connection failed")
	ErrTransactionFailed  = errors.New("transaction failed")
	ErrInsufficientBalance = errors.New("insufficient balance")
	
	// Configuration errors
	ErrInvalidConfiguration = errors.New("invalid configuration")
	ErrChainNotFound        = errors.New("chain not found")
	ErrInvalidChainID       = errors.New("invalid chain ID")
	
	// Security errors
	ErrAccessDenied     = errors.New("access denied")
	ErrInvalidUserID    = errors.New("invalid user ID")
	ErrUnauthorized     = errors.New("unauthorized access to wallet commands")
)

// WalletError represents a wallet-specific error with additional context
type WalletError struct {
	Type    ErrorType
	Message string
	Cause   error
	Context map[string]interface{}
}

// ErrorType represents the type of wallet error
type ErrorType int

const (
	ErrorTypeGeneral ErrorType = iota
	ErrorTypeNetwork
	ErrorTypeConfiguration
	ErrorTypeSecurity
	ErrorTypeValidation
)

func (e *WalletError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *WalletError) Unwrap() error {
	return e.Cause
}

// NewWalletError creates a new wallet error
func NewWalletError(errType ErrorType, message string, cause error, context map[string]interface{}) *WalletError {
	return &WalletError{
		Type:    errType,
		Message: message,
		Cause:   cause,
		Context: context,
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string, cause error) *WalletError {
	return NewWalletError(ErrorTypeValidation, message, cause, nil)
}

// NewNetworkError creates a network error
func NewNetworkError(message string, cause error) *WalletError {
	return NewWalletError(ErrorTypeNetwork, message, cause, nil)
}

// NewConfigurationError creates a configuration error
func NewConfigurationError(message string, cause error) *WalletError {
	return NewWalletError(ErrorTypeConfiguration, message, cause, nil)
}

// NewSecurityError creates a security error
func NewSecurityError(message string, cause error) *WalletError {
	return NewWalletError(ErrorTypeSecurity, message, cause, nil)
}

// IsRetryable determines if an error is retryable
func IsRetryable(err error) bool {
	var walletErr *WalletError
	if errors.As(err, &walletErr) {
		switch walletErr.Type {
		case ErrorTypeNetwork:
			return true
		case ErrorTypeConfiguration:
			return false // Configuration errors should not be retried
		case ErrorTypeSecurity:
			return false // Security errors should not be retried
		case ErrorTypeValidation:
			return false // Validation errors should not be retried
		}
	}
	
	// Check for specific network errors that are retryable
	if errors.Is(err, ErrNetworkFailure) || 
	   errors.Is(err, ErrRPCConnection) ||
	   errors.Is(err, ErrTransactionFailed) {
		return true
	}
	
	return false
}

// IsUserError determines if an error is caused by user input
func IsUserError(err error) bool {
	var walletErr *WalletError
	if errors.As(err, &walletErr) {
		return walletErr.Type == ErrorTypeValidation || walletErr.Type == ErrorTypeSecurity
	}
	
	return errors.Is(err, ErrInvalidAddress) ||
	       errors.Is(err, ErrInvalidAmount) ||
	       errors.Is(err, ErrInvalidPassword) ||
	       errors.Is(err, ErrAccessDenied) ||
	       errors.Is(err, ErrUnauthorized)
}

// GetUserFriendlyMessage returns a user-friendly error message
func GetUserFriendlyMessage(err error) string {
	var walletErr *WalletError
	if errors.As(err, &walletErr) {
		return walletErr.Message
	}
	
	// Map common errors to user-friendly messages
	switch {
	case errors.Is(err, ErrWalletDisabled):
		return "Wallet functionality is currently disabled. Please contact an administrator."
	case errors.Is(err, ErrInvalidAddress):
		return "Invalid Ethereum address. Please check the address format."
	case errors.Is(err, ErrInvalidAmount):
		return "Invalid amount. Please enter a valid number."
	case errors.Is(err, ErrInvalidPassword):
		return "Invalid password. Please check your password and try again."
	case errors.Is(err, ErrAccountNotFound):
		return "Account not found. Please ensure the wallet exists."
	case errors.Is(err, ErrInsufficientBalance):
		return "Insufficient balance. Please ensure you have enough funds."
	case errors.Is(err, ErrAccessDenied):
		return "Access denied. You don't have permission to use wallet commands."
	case errors.Is(err, ErrUnauthorized):
		return "Wallet commands not available for this user."
	case errors.Is(err, ErrChainNotConfigured):
		return "Chain not configured. Please check your configuration."
	case errors.Is(err, ErrNetworkFailure):
		return "Network connection failed. Please try again later."
	case errors.Is(err, ErrTransactionFailed):
		return "Transaction failed. Please check the error details and try again."
	default:
		return "An unexpected error occurred. Please try again or contact support."
	}
}