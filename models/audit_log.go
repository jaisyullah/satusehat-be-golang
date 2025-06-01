package models

type AuditLog struct {
	User       string                 `bson:"user" json:"user"`
	Action     string                 `bson:"action" json:"action"`
	Resource   string                 `bson:"resource" json:"resource"`
	ResourceID string                 `bson:"resource_id" json:"resource_id"`
	StatusCode int                    `bson:"status_code" json:"status_code"`
	Details    map[string]interface{} `bson:"details" json:"details"`
	Timestamp  int64                  `bson:"timestamp"` // optional, add if you want
}
