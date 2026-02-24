package wallet

import "errors"

var (
	// ErrWalletNotCreated is returned when wallet doesn't exist
	ErrWalletNotCreated = errors.New("wallet not created yet")

	// ErrWalletAlreadyExists is returned when trying to create duplicate wallet
	ErrWalletAlreadyExists = errors.New("wallet already exists")

	// ErrInvalidPIN is returned when PIN is incorrect
	ErrInvalidPIN = errors.New("invalid PIN")

	// ErrPINRequired is returned when PIN is required but not provided
	ErrPINRequired = errors.New("PIN required")

	// ErrInvalidPINFormat is returned when PIN format is invalid
	ErrInvalidPINFormat = errors.New("PIN must be 4 digits")

	// ErrKeystoreFailed is returned when keystore operation fails
	ErrKeystoreFailed = errors.New("keystore operation failed")
)
