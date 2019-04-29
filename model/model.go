package model

import (
	"net/url"

	//"github.com/mongodb/mongo-go-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	PublicKey  string `json:"publicKey" bson:"publicKey" mapstructure:"publicKey" structs:"publicKey"`
	PrivateKey string `json:"privateKey" bson:"privateKey" mapstructure:"privateKey" structs:"privateKey"`
	// PrivateKey *rsa.PrivateKey `json:"privateKey" bson:"privateKey" mapstructure:"privateKey" structs:"privateKey"`
}

// API Struct
type API struct {
	ID            primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty" mapstructure:"_id"`
	Name          string             `json:"name" bson:"name" mapstructure:"name" structs:"name"`
	DeviceAccount string             `json:"deviceAccount" bson:"deviceAccount" mapstructure:"deviceAccount" structs:"deviceAccount"`
	Path          string             `json:"path" bson:"path" mapstructure:"path" structs:"path"`
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
	ID           primitive.ObjectID     `json:"id,omitempty" bson:"_id,omitempty" mapstructure:"_id" structs:"_id"`
	Name         string                 `json:"name" bson:"name" mapstructure:"name" structs:"name"`
	BaseURL      string                 `json:"baseURL" bson:"baseURL" mapstructure:"baseURL" structs:"baseURL"`
	AuthType     string                 `json:"authType" bson:"authType" mapstructure:"authType" structs:"authType"`
	RequireProxy bool                   `json:"requireProxy" bson:"requireProxy" mapstructure:"requireProxy" structs:"requireProxy"`
	AuthObj      map[string]interface{} `json:"authObj" bson:"authObj" mapstructure:"authObj" structs:"authObj"`
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
	Category    string             `json:"category" bson:"category" mapstructure:"category" structs:"category"`
	Type        string             `json:"type" bson:"type" mapstructure:"type" structs:"type"`
	Steps       []Step             `json:"steps" bson:"steps" mapstructure:"steps" structs:"steps"`
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
	ID            primitive.ObjectID  `json:"id,omitempty" bson:"_id,omitempty" mapstructure:"_id" structs:"_id"`
	Title         string              `json:"title" bson:"title" mapstructure:"title" structs:"title"`
	Description   string              `json:"description" bson:"description" mapstructure:"description" structs:"description"`
	APICall       string              `json:"apiCall" bson:"apiCall" mapstructure:"apiCall" structs:"apiCall"`
	Verb          string              `json:"verb" bson:"verb" mapstructure:"verb" structs:"verb"`
	DeviceAccount string              `json:"deviceAccount" bson:"deviceAccount" mapstructure:"deviceAccount" structs:"deviceAccount"`
	Headers       []map[string]string `json:"headers" bson:"headers" mapstructure:"headers" structs:"headers"`
	Variables     []map[string]string `json:"variables" bson:"variables" mapstructure:"variables" structs:"variables"`
	Body          []map[string]string `json:"body" bson:"body" mapstructure:"body" structs:"body"`
	Query         []map[string]string `json:"query" bson:"query" mapstructure:"query" structs:"query"`
	VarMap        []map[string]string `json:"varMap" bson:"varMap" mapstructure:"varMap" structs:"varMap"`
}

// StepResult Struct
type StepResult struct {
	API        map[string]interface{} `json:"api" bson:"api" mapstructure:"api" structs:"api"`
	Account    string                 `json:"account" bson:"account" mapstructure:"account" structs:"account"`
	ReqHeaders map[string]string      `json:"reqHeaders" bson:"reqHeaders" mapstructure:"reqHeaders" structs:"reqHeaders"`
	ReqQuery   url.Values             `json:"reqQuery" bson:"reqQuery" mapstructure:"reqQuery" structs:"reqQuery"`
	ReqBody    map[string]interface{} `json:"reqBody" bson:"reqBody" mapstructure:"reqBody" structs:"reqBody"`
	ResBody    string                 `json:"resBody" bson:"resBody" mapstructure:"resBody" structs:"resBody"`
	ResStatus  int                    `json:"resStatus" bson:"resStatus" mapstructure:"resStatus" structs:"resStatus"`
	Error      string                 `json:"error" bson:"error" mapstructure:"error" structs:"error"`
	Status     int                    `json:"status" bson:"status" mapstructure:"status" structs:"status"`
}
