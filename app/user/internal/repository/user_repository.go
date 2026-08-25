package repository

import (
	"context"

	"flamingo/app/user/internal/model"
	"flamingo/app/user/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// userRepositoryImpl user repository implementation
type userRepositoryImpl struct {
	db  *gorm.DB
	log *log.Helper
}

// NewUserRepository creates user repository
func NewUserRepository(db *gorm.DB, logger log.Logger) service.UserRepository {
	return &userRepositoryImpl{
		db:  db,
		log: log.NewHelper(logger),
	}
}

// Create creates user
func (r *userRepositoryImpl) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID gets user by ID
func (r *userRepositoryImpl) GetByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByPhone gets user by phone number
func (r *userRepositoryImpl) GetByPhone(ctx context.Context, phone string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail gets user by email
func (r *userRepositoryImpl) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByAccount gets user by account (phone or email)
func (r *userRepositoryImpl) GetByAccount(ctx context.Context, account string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Where("phone = ? OR email = ?", account, account).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates user
func (r *userRepositoryImpl) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// UpdatePhone updates phone number
func (r *userRepositoryImpl) UpdatePhone(ctx context.Context, userID string, phone *string) error {
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", userID).
		Update("phone", phone).Error
}

// UpdateEmail updates email
func (r *userRepositoryImpl) UpdateEmail(ctx context.Context, userID string, email *string) error {
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", userID).
		Update("email", email).Error
}

// UpdatePassword updates password
func (r *userRepositoryImpl) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", userID).
		Update("password_hash", passwordHash).Error
}

// UpdateStatus updates status
func (r *userRepositoryImpl) UpdateStatus(ctx context.Context, userID string, status int) error {
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", userID).
		Update("status", status).Error
}

// WithTx returns a repository instance that uses the given transaction
func (r *userRepositoryImpl) WithTx(tx *gorm.DB) service.UserRepository {
	return &userRepositoryImpl{db: tx, log: r.log}
}
