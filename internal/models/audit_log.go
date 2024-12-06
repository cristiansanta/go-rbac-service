package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// JsonMap tipo personalizado para almacenar JSON
type JsonMap map[string]interface{}

// Value implementa la interfaz driver.Valuer
func (j JsonMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implementa la interfaz sql.Scanner
func (j *JsonMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("tipo de valor inválido para JsonMap")
	}
	return json.Unmarshal(bytes, &j)
}

// AuditLog modelo principal para logs de auditoría
type AuditLog struct {
	ID             int       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID         int       `json:"user_id"`
	Username       string    `json:"username"`
	ModuleName     string    `json:"module_name"`
	Action         string    `json:"action"`          // CREATE, READ, UPDATE, DELETE
	PermissionUsed string    `json:"permission_used"` // R, W, X, D
	EntityType     string    `json:"entity_type"`     // users, roles, modules, etc.
	EntityID       int       `json:"entity_id"`
	OldValue       JsonMap   `json:"old_value" gorm:"type:jsonb"`
	NewValue       JsonMap   `json:"new_value" gorm:"type:jsonb"`
	IPAddress      string    `json:"ip_address"`
	UserAgent      string    `json:"user_agent"`
	StatusCode     int       `json:"status_code"`
	RequestPath    string    `json:"request_path"`
	RequestMethod  string    `json:"request_method"`
	FechaCreacion  time.Time `json:"fecha_creacion" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

// AuditLogResponse estructura para respuestas API
type AuditLogResponse struct {
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	Username       string    `json:"username"`
	ModuleName     string    `json:"module_name"`
	Action         string    `json:"action"`
	PermissionUsed string    `json:"permission_used"`
	EntityType     string    `json:"entity_type"`
	EntityID       int       `json:"entity_id"`
	OldValue       JsonMap   `json:"old_value,omitempty"`
	NewValue       JsonMap   `json:"new_value,omitempty"`
	IPAddress      string    `json:"ip_address"`
	StatusCode     int       `json:"status_code"`
	RequestPath    string    `json:"request_path"`
	RequestMethod  string    `json:"request_method"`
	FechaCreacion  time.Time `json:"fecha_creacion"`
}

// ToResponse convierte AuditLog a AuditLogResponse
func (a *AuditLog) ToResponse() AuditLogResponse {
	return AuditLogResponse{
		ID:             a.ID,
		UserID:         a.UserID,
		Username:       a.Username,
		ModuleName:     a.ModuleName,
		Action:         a.Action,
		PermissionUsed: a.PermissionUsed,
		EntityType:     a.EntityType,
		EntityID:       a.EntityID,
		OldValue:       a.OldValue,
		NewValue:       a.NewValue,
		IPAddress:      a.IPAddress,
		StatusCode:     a.StatusCode,
		RequestPath:    a.RequestPath,
		RequestMethod:  a.RequestMethod,
		FechaCreacion:  a.FechaCreacion,
	}
}
