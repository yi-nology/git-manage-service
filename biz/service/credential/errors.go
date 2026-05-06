package credential

import "errors"

var (
	ErrCredentialNotFound = errors.New("credential not found")
	ErrCredentialInUse    = errors.New("credential is currently in use")
	ErrInvalidType        = errors.New("invalid credential type")
	ErrNameExists         = errors.New("credential name already exists")
)
