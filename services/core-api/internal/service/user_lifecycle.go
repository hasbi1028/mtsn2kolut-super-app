package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type userLifecycleStore interface {
	GetUserByID(ctx context.Context, id pgtype.UUID) (db.GetUserByIDRow, error)
	UpdateUserStatus(ctx context.Context, arg db.UpdateUserStatusParams) error
	SoftDeleteUser(ctx context.Context, id pgtype.UUID) error
	UpdateUserPassword(ctx context.Context, arg db.UpdateUserPasswordParams) error
	IncrementUserAuthVersion(ctx context.Context, userID pgtype.UUID) (int32, error)
	UpdateUserProfileLink(ctx context.Context, arg db.UpdateUserProfileLinkParams) error
	RevokeAllAuthSessionsForUser(ctx context.Context, userID pgtype.UUID) (int64, error)
	UserHasRbacRole(ctx context.Context, arg db.UserHasRbacRoleParams) (bool, error)
	CountActiveAdminsByRbac(ctx context.Context) (int64, error)
	CreateAuditLog(ctx context.Context, arg db.CreateAuditLogParams) (db.AuditLog, error)
}

type userLifecycleTxStarter interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type UserLifecycle struct {
	q  userLifecycleStore
	tx userLifecycleTxStarter
}

type ProfileLink struct {
	EmployeeID pgtype.UUID
	StudentID  pgtype.UUID
	ParentID   pgtype.UUID
}

func NewUserLifecycle(q *db.Queries) *UserLifecycle {
	return &UserLifecycle{q: q}
}

func NewUserLifecycleWithPool(pool *pgxpool.Pool) *UserLifecycle {
	return &UserLifecycle{q: db.New(pool), tx: pool}
}

func (s *UserLifecycle) DeleteAsDeactivate(ctx context.Context, id pgtype.UUID, actorID pgtype.UUID) error {
	if err := s.ensureCanDeactivate(ctx, id); err != nil {
		return err
	}
	return s.withStore(ctx, func(store userLifecycleStore) error {
		if err := store.SoftDeleteUser(ctx, id); err != nil {
			return err
		}
		if _, err := store.RevokeAllAuthSessionsForUser(ctx, id); err != nil {
			return err
		}
		return auditUserLifecycle(ctx, store, actorID, "USER_DELETED", "user", uuidEntityID(id), map[string]any{"user_id": uuidEntityID(id), "deleted_at": "now"})
	})
}

func (s *UserLifecycle) UpdateStatus(ctx context.Context, id pgtype.UUID, isActive bool, actorID pgtype.UUID) error {
	return s.setActive(ctx, id, isActive, actorID)
}

func (s *UserLifecycle) ResetPassword(ctx context.Context, id pgtype.UUID, newPassword string, actorID pgtype.UUID) error {
	user, err := s.q.GetUserByID(ctx, id)
	if err != nil {
		return err
	}
	if err := ValidatePassword(user.Username, newPassword); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.withStore(ctx, func(store userLifecycleStore) error {
		if err := store.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{ID: id, PasswordHash: string(hash)}); err != nil {
			return err
		}
		if _, err := store.IncrementUserAuthVersion(ctx, id); err != nil {
			return err
		}
		if _, err := store.RevokeAllAuthSessionsForUser(ctx, id); err != nil {
			return err
		}
		return auditUserLifecycle(ctx, store, actorID, "USER_PASSWORD_RESET", "user", uuidEntityID(id), map[string]any{"user_id": uuidEntityID(id), "username": user.Username})
	})
}

func (s *UserLifecycle) UpdateProfileLink(ctx context.Context, id pgtype.UUID, link ProfileLink, actorID pgtype.UUID) error {
	if err := validateProfileLink(link); err != nil {
		return err
	}
	return s.withStore(ctx, func(store userLifecycleStore) error {
		if err := store.UpdateUserProfileLink(ctx, db.UpdateUserProfileLinkParams{ID: id, EmployeeID: link.EmployeeID, StudentID: link.StudentID, ParentID: link.ParentID, DisplayName: pgtype.Text{}}); err != nil {
			return err
		}
		return auditUserLifecycle(ctx, store, actorID, "USER_PROFILE_LINK_UPDATED", "user", uuidEntityID(id), map[string]any{
			"user_id":     uuidEntityID(id),
			"employee_id": uuidEntityID(link.EmployeeID),
			"student_id":  uuidEntityID(link.StudentID),
			"parent_id":   uuidEntityID(link.ParentID),
		})
	})
}

func (s *UserLifecycle) setActive(ctx context.Context, id pgtype.UUID, isActive bool, actorID pgtype.UUID) error {
	if !isActive {
		if err := s.ensureCanDeactivate(ctx, id); err != nil {
			return err
		}
	}
	return s.withStore(ctx, func(store userLifecycleStore) error {
		if err := store.UpdateUserStatus(ctx, db.UpdateUserStatusParams{
			ID:       id,
			IsActive: isActive,
		}); err != nil {
			return err
		}
		if !isActive {
			if _, err := store.RevokeAllAuthSessionsForUser(ctx, id); err != nil {
				return err
			}
		}
		return auditUserLifecycle(ctx, store, actorID, "USER_STATUS_UPDATED", "user", uuidEntityID(id), map[string]any{"user_id": uuidEntityID(id), "is_active": isActive})
	})
}

func (s *UserLifecycle) ensureCanDeactivate(ctx context.Context, id pgtype.UUID) error {
	isAdmin, err := s.q.UserHasRbacRole(ctx, db.UserHasRbacRoleParams{UserID: id, Code: "admin"})
	if err != nil {
		return err
	}
	if !isAdmin {
		return nil
	}
	admins, err := s.q.CountActiveAdminsByRbac(ctx)
	if err != nil {
		return err
	}
	if admins <= 1 {
		return fmt.Errorf("admin aktif terakhir tidak boleh dinonaktifkan")
	}
	return nil
}

func (s *UserLifecycle) withStore(ctx context.Context, fn func(userLifecycleStore) error) error {
	if s.tx == nil {
		return fn(s.q)
	}
	tx, err := s.tx.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := fn(db.New(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func validateProfileLink(link ProfileLink) error {
	linked := 0
	if link.EmployeeID.Valid {
		linked++
	}
	if link.StudentID.Valid {
		linked++
	}
	if link.ParentID.Valid {
		linked++
	}
	if linked > 1 {
		return fmt.Errorf("satu akun hanya boleh ditautkan ke satu jenis profil")
	}
	return nil
}

func auditUserLifecycle(ctx context.Context, store userLifecycleStore, actorID pgtype.UUID, action string, entityType string, entityID string, metadata map[string]any) error {
	payload, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	_, err = store.CreateAuditLog(ctx, db.CreateAuditLogParams{UserID: actorID, Action: action, EntityType: entityType, EntityID: entityID, Metadata: payload})
	return err
}
