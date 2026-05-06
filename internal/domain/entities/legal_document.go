// internal/domain/entities/legal_document.go
package entities

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ========================================
// LEGAL DOCUMENT TYPE CODES
// ========================================

const (
	LegalDocCodeTerms     = "terms"
	LegalDocCodePrivacy   = "privacy"
	LegalDocCodeMarketing = "marketing"
)

// ========================================
// LEGAL DOCUMENT STATUS (matches PG enum)
// ========================================

type LegalDocumentStatus string

const (
	LegalDocStatusDraft     LegalDocumentStatus = "draft"
	LegalDocStatusInReview  LegalDocumentStatus = "in_review"
	LegalDocStatusPublished LegalDocumentStatus = "published"
	LegalDocStatusArchived  LegalDocumentStatus = "archived"
	LegalDocStatusWithdrawn LegalDocumentStatus = "withdrawn"
)

// ========================================
// CHANGE SEVERITY
// ========================================

const (
	LegalDocSeverityMinor    = "minor"
	LegalDocSeverityMajor    = "major"
	LegalDocSeverityCritical = "critical"
)

// ========================================
// LegalDocumentType — catálogo de tipos
// ========================================

type LegalDocumentType struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Code         string    `gorm:"size:50;not null;uniqueIndex" json:"code"`
	Name         string    `gorm:"size:150;not null" json:"name"`
	Description  string    `json:"description,omitempty"`
	IsRequired   bool      `gorm:"not null;default:true" json:"is_required"`
	DisplayOrder int       `gorm:"not null;default:0" json:"display_order"`
	IsActive     bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (LegalDocumentType) TableName() string {
	return "auth.legal_document_types"
}

// ========================================
// LegalDocument — versiones de documentos legales
// ========================================

// LegalDocumentMetadata es el JSONB de metadata extensible.
type LegalDocumentMetadata map[string]interface{}

// Value implementa driver.Valuer (escritura a Postgres).
func (m LegalDocumentMetadata) Value() (interface{}, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}

// Scan implementa sql.Scanner (lectura desde Postgres).
func (m *LegalDocumentMetadata) Scan(src interface{}) error {
	if src == nil {
		*m = LegalDocumentMetadata{}
		return nil
	}
	var raw []byte
	switch v := src.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		*m = LegalDocumentMetadata{}
		return nil
	}
	if len(raw) == 0 {
		*m = LegalDocumentMetadata{}
		return nil
	}
	return json.Unmarshal(raw, m)
}

type LegalDocument struct {
	ID               uuid.UUID             `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	DocumentTypeID   uuid.UUID             `gorm:"type:uuid;not null;index" json:"document_type_id"`
	Version          string                `gorm:"size:20;not null" json:"version"`
	Locale           string                `gorm:"size:10;not null;default:'es-PE'" json:"locale"`
	Jurisdiction     string                `gorm:"size:10;not null;default:'PE'" json:"jurisdiction"`
	Title            string                `gorm:"size:255;not null" json:"title"`
	Summary          string                `json:"summary,omitempty"`
	ContentMarkdown  string                `gorm:"not null" json:"content_markdown"`
	ContentHTML      string                `gorm:"column:content_html" json:"content_html,omitempty"`
	ContentHash      string                `gorm:"size:64;not null" json:"content_hash"`
	Changelog        string                `json:"changelog,omitempty"`
	ChangeSeverity   string                `gorm:"size:20" json:"change_severity,omitempty"`
	Status           LegalDocumentStatus   `gorm:"type:auth.legal_document_status;not null;default:'draft'" json:"status"`
	EffectiveDate    *time.Time            `gorm:"type:date" json:"effective_date,omitempty"`
	PublishedAt      *time.Time            `json:"published_at,omitempty"`
	ArchivedAt       *time.Time            `json:"archived_at,omitempty"`
	CreatedBy        *uuid.UUID            `gorm:"type:uuid" json:"created_by,omitempty"`
	ReviewedBy       *uuid.UUID            `gorm:"type:uuid" json:"reviewed_by,omitempty"`
	ApprovedBy       *uuid.UUID            `gorm:"type:uuid" json:"approved_by,omitempty"`
	ApprovedAt       *time.Time            `json:"approved_at,omitempty"`
	Metadata         LegalDocumentMetadata `gorm:"type:jsonb;not null;default:'{}'" json:"metadata"`
	CreatedAt        time.Time             `json:"created_at"`
	UpdatedAt        time.Time             `json:"updated_at"`
	DeletedAt        *time.Time            `json:"deleted_at,omitempty"`

	// Relación opcional cargada via Preload
	DocumentType *LegalDocumentType `gorm:"foreignKey:DocumentTypeID" json:"document_type,omitempty"`
}

func (LegalDocument) TableName() string {
	return "auth.legal_documents"
}

// ComputeHash calcula el SHA-256 hex del contenido markdown.
// Útil al crear drafts y al validar integridad antes de publicar.
func ComputeContentHash(markdown string) string {
	sum := sha256.Sum256([]byte(markdown))
	return hex.EncodeToString(sum[:])
}

// IsPublished indica si el documento está vigente legalmente.
func (d *LegalDocument) IsPublished() bool {
	return d.Status == LegalDocStatusPublished && d.DeletedAt == nil
}

// CanTransitionTo valida transiciones de status permitidas (state machine).
func (d *LegalDocument) CanTransitionTo(target LegalDocumentStatus) bool {
	switch d.Status {
	case LegalDocStatusDraft:
		return target == LegalDocStatusInReview || target == LegalDocStatusPublished
	case LegalDocStatusInReview:
		return target == LegalDocStatusPublished || target == LegalDocStatusDraft
	case LegalDocStatusPublished:
		return target == LegalDocStatusArchived || target == LegalDocStatusWithdrawn
	case LegalDocStatusArchived, LegalDocStatusWithdrawn:
		return false
	}
	return false
}
