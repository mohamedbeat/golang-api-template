package auth

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type SessionRepository interface {
	// user
	Create(ctx context.Context, s *Session) error
	FindByToken(hash string) (*Session, error)
	WithTx(ctx context.Context, fn func(*sqlx.Tx) error) error
	GetForUpdate(ctx context.Context, tx *sqlx.Tx, hash string) (*Session, error)
	RotateSession(ctx context.Context, tx *sqlx.Tx, oldID uuid.UUID, newSession *Session) error
	RevokeFamily(ctx context.Context, tx *sqlx.Tx, familyID uuid.UUID) error
	RevokeByID(ctx context.Context, id uuid.UUID) error

	// organization
	CreateOrg(ctx context.Context, s *OrgSession) error
	FindForOrgByToken(hash string) (*OrgSession, error)
	GetForUpdateOrg(ctx context.Context, tx *sqlx.Tx, hash string) (*OrgSession, error)
	RotateOrgSession(ctx context.Context, tx *sqlx.Tx, oldID uuid.UUID, newSession *OrgSession) error
	RevokeOrgFamily(ctx context.Context, tx *sqlx.Tx, familyID uuid.UUID) error
	RevokeOrgByID(ctx context.Context, id uuid.UUID) error
}

type sessionRepo struct {
	DB *sqlx.DB
}

func NewSessionRepo(db *sqlx.DB) SessionRepository {
	return &sessionRepo{
		DB: db,
	}
}

func (r *sessionRepo) Create(ctx context.Context, s *Session) error {
	query := `
    INSERT INTO user_sessions (user_id, refresh_token_hash, expires_at)
    VALUES ($1, $2, $3)`

	_, err := r.DB.ExecContext(ctx, query, s.UserID, s.RefreshTokenHash, s.ExpiresAt)
	return err
}

func (r *sessionRepo) FindByToken(hash string) (*Session, error) {
	var s Session
	err := r.DB.Get(&s,
		"SELECT * FROM user_sessions WHERE refresh_token_hash=$1 AND expires_at > NOW()",
		hash,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

// WithTx runs fn inside a transaction, committing on success and rolling
// back on any error fn returns.
func (r *sessionRepo) WithTx(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := r.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

// GetForUpdate locks the session row for the duration of the transaction so
// concurrent refresh requests using the same token serialize on it instead
// of racing each other.
func (r *sessionRepo) GetForUpdate(ctx context.Context, tx *sqlx.Tx, hash string) (*Session, error) {
	var s Session
	err := tx.GetContext(ctx, &s,
		"SELECT * FROM user_sessions WHERE refresh_token_hash=$1 FOR UPDATE",
		hash,
	)
	if err != nil {
		return nil, err // caller checks errors.Is(err, sql.ErrNoRows)
	}
	return &s, nil
}

// RotateSession inserts the successor session in the same family and points
// the old row at it via replaced_by. The old row is never deleted, which is
// what lets reuse of an old token be detected later.
func (r *sessionRepo) RotateSession(ctx context.Context, tx *sqlx.Tx, oldID uuid.UUID, newSession *Session) error {
	var newID uuid.UUID
	err := tx.QueryRowContext(ctx, `
		INSERT INTO user_sessions (user_id, family_id, refresh_token_hash, user_agent, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, newSession.UserID, newSession.FamilyID, newSession.RefreshTokenHash, newSession.UserAgent, newSession.IPAddress, newSession.ExpiresAt).Scan(&newID)
	if err != nil {
		return err
	}
	newSession.ID = newID

	_, err = tx.ExecContext(ctx, `
		UPDATE user_sessions SET replaced_by = $1, updated_at = NOW() WHERE id = $2
	`, newID, oldID)
	return err
}

// RevokeFamily kills every session descended from the same login. Used when
// an already-rotated token is presented again, which signals theft.
func (r *sessionRepo) RevokeFamily(ctx context.Context, tx *sqlx.Tx, familyID uuid.UUID) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE user_sessions SET revoked_at = NOW(), updated_at = NOW()
		WHERE family_id = $1 AND revoked_at IS NULL
	`, familyID)
	return err
}

// RevokeByID kills a single session, e.g. on logout.
func (r *sessionRepo) RevokeByID(ctx context.Context, id uuid.UUID) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE user_sessions SET revoked_at = NOW(), updated_at = NOW() WHERE id = $1
	`, id)
	return err
}

/////////////////////////
//
// ogganization related
//
////////////////////////

func (r *sessionRepo) CreateOrg(ctx context.Context, s *OrgSession) error {
	query := `
    INSERT INTO organization_sessions (organization_id, refresh_token_hash, expires_at)
    VALUES ($1, $2, $3)`

	_, err := r.DB.ExecContext(ctx, query, s.OrgID, s.RefreshTokenHash, s.ExpiresAt)
	return err
}

func (r *sessionRepo) FindForOrgByToken(hash string) (*OrgSession, error) {
	var s OrgSession
	err := r.DB.Get(&s,
		"SELECT * FROM organization_sessions WHERE refresh_token_hash=$1 AND expires_at > NOW()",
		hash,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

// GetForUpdateOrg locks the org session row for the duration of the
// transaction so concurrent refresh requests using the same token
// serialize on it instead of racing each other.
func (r *sessionRepo) GetForUpdateOrg(ctx context.Context, tx *sqlx.Tx, hash string) (*OrgSession, error) {
	var s OrgSession
	err := tx.GetContext(ctx, &s,
		"SELECT * FROM organization_sessions WHERE refresh_token_hash=$1 FOR UPDATE",
		hash,
	)
	if err != nil {
		return nil, err // caller checks errors.Is(err, sql.ErrNoRows)
	}
	return &s, nil
}

// RotateOrgSession inserts the successor session in the same family and
// points the old row at it via replaced_by. The old row is never deleted.
func (r *sessionRepo) RotateOrgSession(ctx context.Context, tx *sqlx.Tx, oldID uuid.UUID, newSession *OrgSession) error {
	var newID uuid.UUID
	err := tx.QueryRowContext(ctx, `
		INSERT INTO organization_sessions (organization_id, family_id, refresh_token_hash, user_agent, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, newSession.OrgID, newSession.FamilyID, newSession.RefreshTokenHash, newSession.UserAgent, newSession.IPAddress, newSession.ExpiresAt).Scan(&newID)
	if err != nil {
		return err
	}
	newSession.ID = newID

	_, err = tx.ExecContext(ctx, `
		UPDATE organization_sessions SET replaced_by = $1, updated_at = NOW() WHERE id = $2
	`, newID, oldID)
	return err
}

// RevokeOrgFamily kills every session descended from the same org login.
// Used when an already-rotated token is presented again, which signals theft.
func (r *sessionRepo) RevokeOrgFamily(ctx context.Context, tx *sqlx.Tx, familyID uuid.UUID) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE organization_sessions SET revoked_at = NOW(), updated_at = NOW()
		WHERE family_id = $1 AND revoked_at IS NULL
	`, familyID)
	return err
}

// RevokeOrgByID kills a single org session, e.g. on logout.
func (r *sessionRepo) RevokeOrgByID(ctx context.Context, id uuid.UUID) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE organization_sessions SET revoked_at = NOW(), updated_at = NOW() WHERE id = $1
	`, id)
	return err
}
