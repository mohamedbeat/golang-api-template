package auth

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID               uuid.UUID  `db:"id"`
	UserID           uuid.UUID  `db:"user_id"`
	FamilyID         uuid.UUID  `db:"family_id"`
	RefreshTokenHash string     `db:"refresh_token_hash"`
	UserAgent        *string    `db:"user_agent"`
	IPAddress        *string    `db:"ip_address"`
	IsActive         bool       `db:"is_active"`
	ReplacedBy       *uuid.UUID `db:"replaced_by"`
	RevokedAt        *time.Time `db:"revoked_at"`
	ExpiresAt        time.Time  `db:"expires_at"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"`
}

func (s *Session) IsExpired() bool {
	return time.Now().UTC().After(s.ExpiresAt.UTC())
}

type OrgSession struct {
	ID               uuid.UUID  `db:"id"`
	OrgID            uuid.UUID  `db:"organization_id"`
	FamilyID         uuid.UUID  `db:"family_id"`
	RefreshTokenHash string     `db:"refresh_token_hash"`
	UserAgent        *string    `db:"user_agent"`
	IPAddress        *string    `db:"ip_address"`
	IsActive         bool       `db:"is_active"`
	ReplacedBy       *uuid.UUID `db:"replaced_by"`
	RevokedAt        *time.Time `db:"revoked_at"`
	ExpiresAt        time.Time  `db:"expires_at"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"`
}

func (s *OrgSession) IsExpired() bool {
	return time.Now().UTC().After(s.ExpiresAt.UTC())
}
