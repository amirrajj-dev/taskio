package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/amirrajj-dev/taskio/internal/errors"
	"github.com/amirrajj-dev/taskio/internal/infrastructure/database"
	"github.com/amirrajj-dev/taskio/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user models.User) (*models.User, error)
	FindUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	FindUserByEmail(ctx context.Context, email string) (*models.User, error)
	CreateRefresh(ctx context.Context, refresh models.RefreshToken) (*models.RefreshToken, error)
	FindRefreshByUserID(ctx context.Context, userID uuid.UUID) (*models.RefreshToken, error)
	DeleteRefresh(ctx context.Context, refreshID uuid.UUID) error
}

type userRepository struct {
	Users         *gorm.DB
	RefreshTokens *gorm.DB
}

var UserRepo *userRepository

func NewUserRepository() {
	UserRepo = &userRepository{
		Users:         database.PG.Model(&models.User{}),
		RefreshTokens: database.PG.Model(&models.RefreshToken{}),
	}
}

func (r *userRepository) CreateUser(ctx context.Context, user models.User) (*models.User, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := r.Users.WithContext(opCtx).Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user : %w", err)
	}
	return &user, nil
}

func (r *userRepository) FindUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var user models.User
	if err := r.Users.WithContext(opCtx).First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var user models.User
	if err := r.Users.WithContext(opCtx).Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) CreateRefresh(ctx context.Context, refresh models.RefreshToken) (*models.RefreshToken, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := r.RefreshTokens.WithContext(opCtx).Create(&refresh).Error; err != nil {
		return nil, fmt.Errorf("failed to create refresh : %w", err)
	}
	return &refresh, nil
}

func (r *userRepository) FindRefreshByUserID(ctx context.Context, userID uuid.UUID) (*models.RefreshToken, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var refreshToken models.RefreshToken
	if err := r.RefreshTokens.WithContext(opCtx).Where("user_id = ?", userID).First(&refreshToken).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrRefreshNotFound
		}
		return nil, err
	}
	return &refreshToken, nil
}

func (r *userRepository) DeleteRefresh(ctx context.Context, refreshID uuid.UUID) error {
	result := r.RefreshTokens.WithContext(ctx).Delete(&models.RefreshToken{}, refreshID)
	if result.Error != nil {
		return fmt.Errorf("delete refresh: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.ErrRefreshNotFound
	}
	return nil
}
