package domain

type AuthInfo struct {
	Type   string `json:"type"`   // ssh, http, none
	Key    string `json:"key"`    // SSH Key Path or Username
	Secret string `json:"secret"` // Passphrase or Password (Encrypted in DB)
}
