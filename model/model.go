package model

import (
	"crypto/rsa"

	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

// BasicAuth Type
type BasicAuth struct {
	UserName string `json:"username" bson:"username" mapstructure:"username" structs:"username"`
	Password string `json:"password" bson:"password" mapstructure:"password" structs:"password"`
}

// SessionAuth Type
type SessionAuth struct {
	UserName       string              `json:"username" bson:"username" mapstructure:"username" structs:"username"`
	Password       string              `json:"password" bson:"password" mapstructure:"password" structs:"password"`
	AuthBody       string              `json:"authBody" bson:"authBody" mapstructure:"authBody" structs:"authBody"`
	AuthBodyMap    []map[string]string `json:"authBodyMap" bson:"authBodyMap" mapstructure:"authBodyMap" structs:"authBodyMap"`
	CookieLifetime int32               `json:"cookieLifetime" bson:"cookieLifetime" mapstructure:"cookieLifetime" structs:"cookieLifetime"`
}

// APIKeyAuth Type
type APIKeyAuth struct {
	Header string `json:"header" bson:"header" mapstructure:"header" structs:"header"`
	Key    string `json:"key" bson:"key" mapstructure:"key" structs:"key"`
}

// Rfc3447Auth Type
type Rfc3447Auth struct {
	PublicKey  string          `json:"publicKey" bson:"publicKey" mapstructure:"publicKey" structs:"publicKey"`
	PrivateKey *rsa.PrivateKey `json:"privateKey" bson:"privateKey" mapstructure:"privateKey" structs:"privateKey"`
}

// API Struct
type API struct {
	ID            primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty" mapstructure:"_id"`
	Name          string             `json:"name" bson:"name" mapstructure:"name" structs:"name"`
	DeviceAccount string             `json:"deviceAccount" bson:"deviceAccount" mapstructure:"deviceAccount" structs:"deviceAccount"`
	Method        string             `json:"method" bson:"method" mapstructure:"method" structs:"method"`
	URL           string             `json:"url" bson:"url" mapstructure:"url" structs:"url"`
	Body          string             `json:"body" bson:"body" mapstructure:"body" structs:"body"`
	Type          string             `json:"type" bson:"type" mapstructure:"type" structs:"type"`
}

// UserAccount Struct
type UserAccount struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty" mapstructure:"_id" structs:"_id"`
	FirstName string             `json:"fname" bson:"fname" mapstructure:"fname" structs:"fname"`
	LastName  string             `json:"lname" bson:"lname" mapstructure:"lname" structs:"lname"`
	UserName  string             `json:"username" bson:"username" mapstructure:"username" structs:"username"`
	Password  string             `json:"password" bson:"password" mapstructure:"password" structs:"password"`
	Privilege int                `json:"privilege" bson:"privilege" mapstructure:"privilege" structs:"privilege"`
	Enable    bool               `json:"enable" bson:"enable" mapstructure:"enable" structs:"enable"`
	Token     string             `json:"token" bson:"token" mapstructure:"token" structs:"token"`
}

// DeviceAccount Struct
type DeviceAccount struct {
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty" mapstructure:"_id" structs:"_id"`
	Name         string             `json:"name" bson:"name" mapstructure:"name" structs:"name"`
	URL          string             `json:"url" bson:"url" mapstructure:"url" structs:"url"`
	AuthType     string             `json:"authType" bson:"authType" mapstructure:"authType" structs:"authType"`
	AuthObj      primitive.ObjectID `json:"authObj" bson:"authObj" mapstructure:"authObj" structs:"authObj"`
	RequireProxy bool               `json:"requireProxy" bson:"requireProxy" mapstructure:"requireProxy" structs:"requireProxy"`
}

// RouteValue Struct
type RouteValue struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty" mapstructure:"_id" structs:"_id"`
	Enable   bool               `json:"enable" bson:"enable" mapstructure:"enable" structs:"enable"`
	Verb     string             `json:"verb" bson:"verb" mapstructure:"verb" structs:"verb"`
	Version  int                `json:"version" bson:"version" mapstructure:"version" structs:"version"`
	Category string             `json:"category" bson:"category" mapstructure:"category" structs:"category"`
	Route    string             `json:"route" bson:"route" mapstructure:"route" structs:"route"`
	Message  map[string]string  `json:"message" bson:"message" mapstructure:"message" structs:"message"`
}

// Workflow Struct
type Workflow struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty" mapstructure:"_id" structs:"_id"`
	Name        string             `json:"name" bson:"name" mapstructure:"name" structs:"name"`
	Description string             `json:"description" bson:"description" mapstructure:"description" structs:"description"`
	Type        string             `json:"type" bson:"type" mapstructure:"type" structs:"type"`
	Steps       []Step             `json:"steps" bson:"steps" mapstructure:"steps" structs:"steps"`
	// ClaimCode   int                `json:"claimCode" bson:"claimCode" mapstructure:"claimCode" structs:"claimCode"`
	// Note, add OutputMap []map[string]string
}

// WorkflowClaim Struct
type WorkflowClaim struct {
	ID              primitive.ObjectID    `json:"id,omitempty" bson:"_id,omitempty" mapstructure:"_id" structs:"_id"`
	Timestamp       string                `json:"timestamp" bson:"timestamp" mapstructure:"timestamp" structs:"timestamp"`
	WorkflowResults map[string]StepResult `json:"workflowResults" bson:"workflowResults" mapstructure:"workflowResults" structs:"workflowResults"`
	ClaimCode       string                `json:"claimCode" bson:"claimCode" mapstructure:"claimCode" structs:"claimCode"`
	CurrentStatus   int                   `json:"currentStatus" bson:"currentStatus" mapstructure:"currentStatus" structs:"currentStatus"`
}

// Step Struct
type Step struct {
	ID primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty" mapstructure:"_id" structs:"_id"`
	// StepNum       int                 `json:"stepNum" bson:"stepNum" mapstructure:"stepNum" structs:"stepNum"`
	APICall       string              `json:"apiCall" bson:"apiCall" mapstructure:"apiCall" structs:"apiCall"`
	DeviceAccount string              `json:"deviceAccount" bson:"deviceAccount" mapstructure:"deviceAccount" structs:"deviceAccount"`
	VarMap        []map[string]string `json:"varMap" bson:"varMap" mapstructure:"varMap" structs:"varMap"`
	// Status        int                 `json:"status" bson:"status" mapstructure:"status" structs:"status"`
}

// StepResult Struct
type StepResult struct {
	APICall    string `json:"apiCall" bson:"apiCall" mapstructure:"apiCall" structs:"apiCall"`
	APIAccount string `json:"apiAccount" bson:"apiAccount" mapstructure:"apiAccount" structs:"apiAccount"`
	ReqBody    string `json:"reqBody" bson:"reqBody" mapstructure:"reqBody" structs:"reqBody"`
	ResBody    string `json:"resBody" bson:"resBody" mapstructure:"resBody" structs:"resBody"`
	Error      string `json:"error" bson:"error" mapstructure:"error" structs:"error"`
	Status     int    `json:"status" bson:"status" mapstructure:"status" structs:"status"`
}
