package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type JsonMap map[string]interface{}

func (j JsonMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

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

type RegistroAuditoria struct {
	ID              int       `json:"id" gorm:"primaryKey;autoIncrement"`
	IdUsuario       int       `json:"id_usuario"`
	NombreUsuario   string    `json:"nombre_usuario"`
	NombreModulo    string    `json:"nombre_modulo"`
	Accion          string    `json:"accion"`        // CREATE, READ, UPDATE, DELETE
	PermisoUsado    string    `json:"permiso_usado"` // R, W, X, D
	TipoEntidad     string    `json:"tipo_entidad"`  // users, roles, modules, etc.
	IdEntidad       int       `json:"id_entidad"`
	ValorAnterior   JsonMap   `json:"valor_anterior" gorm:"type:jsonb"`
	ValorNuevo      JsonMap   `json:"valor_nuevo" gorm:"type:jsonb"`
	DireccionIP     string    `json:"direccion_ip"`
	AgenteUsuario   string    `json:"agente_usuario"`
	CodigoEstado    int       `json:"codigo_estado"`
	RutaSolicitud   string    `json:"ruta_solicitud"`
	MetodoSolicitud string    `json:"metodo_solicitud"`
	FechaCreacion   time.Time `json:"fecha_creacion" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}

func (RegistroAuditoria) TableName() string {
	return "registros_auditoria"
}

type RegistroAuditoriaResponse struct {
	ID              int       `json:"id"`
	IdUsuario       int       `json:"id_usuario"`
	NombreUsuario   string    `json:"nombre_usuario"`
	NombreModulo    string    `json:"nombre_modulo"`
	Accion          string    `json:"accion"`
	PermisoUsado    string    `json:"permiso_usado"`
	TipoEntidad     string    `json:"tipo_entidad"`
	IdEntidad       int       `json:"id_entidad"`
	ValorAnterior   JsonMap   `json:"valor_anterior,omitempty"`
	ValorNuevo      JsonMap   `json:"valor_nuevo,omitempty"`
	DireccionIP     string    `json:"direccion_ip"`
	CodigoEstado    int       `json:"codigo_estado"`
	RutaSolicitud   string    `json:"ruta_solicitud"`
	MetodoSolicitud string    `json:"metodo_solicitud"`
	FechaCreacion   time.Time `json:"fecha_creacion"`
}

func (r *RegistroAuditoria) ToResponse() RegistroAuditoriaResponse {
	return RegistroAuditoriaResponse{
		ID:              r.ID,
		IdUsuario:       r.IdUsuario,
		NombreUsuario:   r.NombreUsuario,
		NombreModulo:    r.NombreModulo,
		Accion:          r.Accion,
		PermisoUsado:    r.PermisoUsado,
		TipoEntidad:     r.TipoEntidad,
		IdEntidad:       r.IdEntidad,
		ValorAnterior:   r.ValorAnterior,
		ValorNuevo:      r.ValorNuevo,
		DireccionIP:     r.DireccionIP,
		CodigoEstado:    r.CodigoEstado,
		RutaSolicitud:   r.RutaSolicitud,
		MetodoSolicitud: r.MetodoSolicitud,
		FechaCreacion:   r.FechaCreacion,
	}
}
