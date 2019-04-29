package account

import "cane-project/model"

// Device Alias
type Device model.DeviceAccount

// User Alias
type User model.UserAccount

// BasicAuth Alias
type BasicAuth model.BasicAuth

// APIKeyAuth Alias
type APIKeyAuth model.APIKeyAuth

// SessionAuth Alias
type SessionAuth model.SessionAuth

// RFC3447Auth Alias
type RFC3447Auth model.Rfc3447Auth

// JSONBody Alias
type JSONBody map[string]interface{}

// AuthTypes Slice
var AuthTypes = []string{
	"none",
	"basic",
	"apikey",
	"session",
	"rfc3447",
}

// PrivilegeLevels Map
var PrivilegeLevels = map[string]int{
	"admin":    1,
	"user":     2,
	"readonly": 3,
}
