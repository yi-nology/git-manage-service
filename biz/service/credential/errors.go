package credential

import "errors"

var (
	ErrCredentialNotFound = errors.New("credential not found")
	ErrCredentialInUse    = errors.New("credential is currently in use")
	ErrInvalidType        = errors.New("invalid credential type")
	ErrNameExists         = errors.New("credential name already exists")
	ErrNameRequired       = errors.New("name is required")
	ErrSSHKeyRequired     = errors.New("ssh_key_id or ssh_key_path is required for ssh_key type")
)

// validCredentialTypes lists the credential types accepted by the API.
var validCredentialTypes = map[string]bool{
	"ssh_key":        true,
	"http_basic":     true,
	"http_token":     true,
	"platform_token": true,
}

// validateCredential validates the common fields of a create/update request.
func validateCredential(name, credType string, sshKeyID uint64, sshKeyPath string) error {
	if name == "" {
		return ErrNameRequired
	}
	if !validCredentialTypes[credType] {
		return ErrInvalidType
	}
	if credType == "ssh_key" && sshKeyID == 0 && sshKeyPath == "" {
		return ErrSSHKeyRequired
	}
	return nil
}
