package orm

import (
	"context"
	"errors"
	"strings"
	"time"

	"coin-radar-gin/internal/modules/user"
	"coin-radar-gin/internal/platform/database/orm/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository is the GORM implementation of user.Repository.
type UserRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) *UserRepository { return &UserRepository{db: db} }

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	if u.ID == "" {
		u.ID = uuid.NewString()
	}
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now().UTC()
	}
	if u.UpdatedAt.IsZero() {
		u.UpdatedAt = u.CreatedAt
	}
	return r.db.WithContext(ctx).Create(toUserRow(u)).Error
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	result := r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ? AND deleted_at IS NULL", u.ID).
		Updates(map[string]any{
			"telegram_id":   nullableInt64(u.TelegramID),
			"email":         nullableString(u.Email),
			"username":      nullableString(u.Username),
			"password_hash": nullableString(u.PasswordHash),
			"updated_at":    u.UpdatedAt,
			"deleted_at":    u.DeletedAt,
			"created_by":    nullableString(u.CreatedBy),
			"updated_by":    nullableString(u.UpdatedBy),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return user.ErrNotFound
	}
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*user.User, error) {
	return r.find(ctx, r.db.Where("id = ?", id))
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	if email = strings.ToLower(strings.TrimSpace(email)); email == "" {
		return nil, user.ErrNotFound
	}
	return r.find(ctx, r.db.Where("email = ?", email))
}

func (r *UserRepository) FindByTelegramID(ctx context.Context, telegramID int64) (*user.User, error) {
	if telegramID == 0 {
		return nil, user.ErrNotFound
	}
	return r.find(ctx, r.db.Where("telegram_id = ?", telegramID))
}

func (r *UserRepository) find(ctx context.Context, query *gorm.DB) (*user.User, error) {
	var row models.User
	err := query.WithContext(ctx).Where("deleted_at IS NULL").First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, user.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return fromUserRow(row), nil
}

func toUserRow(u *user.User) *models.User {
	return &models.User{ID: u.ID, TelegramID: nullableInt64(u.TelegramID), Email: nullableString(u.Email),
		Username: nullableString(u.Username), PasswordHash: nullableString(u.PasswordHash), CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt, DeletedAt: u.DeletedAt, CreatedBy: nullableString(u.CreatedBy), UpdatedBy: nullableString(u.UpdatedBy)}
}

func fromUserRow(row models.User) *user.User {
	u := &user.User{ID: row.ID, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt, DeletedAt: row.DeletedAt}
	if row.TelegramID != nil {
		u.TelegramID = *row.TelegramID
	}
	if row.Email != nil {
		u.Email = *row.Email
	}
	if row.Username != nil {
		u.Username = *row.Username
	}
	if row.PasswordHash != nil {
		u.PasswordHash = *row.PasswordHash
	}
	if row.CreatedBy != nil {
		u.CreatedBy = *row.CreatedBy
	}
	if row.UpdatedBy != nil {
		u.UpdatedBy = *row.UpdatedBy
	}
	return u
}

func nullableString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func nullableInt64(value int64) *int64 {
	if value == 0 {
		return nil
	}
	return &value
}
