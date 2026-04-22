package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/amirrajj-dev/taskio/internal/configs"
	"github.com/amirrajj-dev/taskio/internal/dtos"
	appErr "github.com/amirrajj-dev/taskio/internal/errors"
	"github.com/amirrajj-dev/taskio/internal/models"
	"github.com/amirrajj-dev/taskio/internal/repositories"
	"github.com/amirrajj-dev/taskio/internal/utils"
	"github.com/google/uuid"
)

type AuthService interface {
	Register(ctx context.Context, dto dtos.RegisterRequest) (*models.UserResponse, *string, error)
	Login(ctx context.Context, dto dtos.LoginRequest) (*models.UserResponse, *string, error)
	CreateRefreshToken(ctx context.Context, user *models.User) (*models.RefreshToken, error)
	RefreshToken(ctx context.Context, userID uuid.UUID) (*models.RefreshToken, error)
	FindRefreshTokenByUserID(ctx context.Context, userID uuid.UUID) (*models.RefreshToken, error)
	DeleteRefresh(ctx context.Context, refreshID uuid.UUID) error
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.UserResponse, error)
}

type authService struct {
	authRepo repositories.UserRepository
}

var UserService *authService

func NewAuthService() {
	UserService = &authService{
		authRepo: repositories.UserRepo,
	}
}

var boyAvatars = []string{"boy1.png", "boy2.png", "boy3.png"}
var girlAvatars = []string{"girl1.png", "girl2.png", "girl3.png"}

func (s *authService) Register(ctx context.Context, dto dtos.RegisterRequest) (*models.UserResponse, *string, error) {
	slog.Info("register attempt", "email", dto.Email)
	now := time.Now().UTC()
	_, findUserErr := s.authRepo.FindUserByEmail(ctx, dto.Email)
	if findUserErr != nil {
		if errors.Is(findUserErr, appErr.ErrUserNotFound) {
			return nil, nil, appErr.ErrUserNotFound
		}
		return nil, nil, findUserErr
	}

	var userAvatar string
	if dto.Gender == "male" {
		userAvatar = boyAvatars[rand.Intn(len(boyAvatars))]
	} else {
		userAvatar = girlAvatars[rand.Intn(len(girlAvatars))]
	}
	user := models.User{
		ID:        uuid.New(),
		Email:     dto.Email,
		Password:  dto.Password,
		FullName:  dto.FullName,
		ImageUrl:  userAvatar,
		Gender:    dto.Gender,
		LastLogin: now,
		CreatedAt: now,
		UpdatedAt: now,
	}
	user.HashPassword(user.Password)
	createdUser, err := s.authRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, nil, err
	}
	slog.Info("user registered successfully", "user_id", createdUser.ID, "email", dto.Email)
	registerToken, genRegTokerr := utils.GenerateToken(createdUser.ToResponse(), configs.Configs.JWT.JWT_EXPIRY_HOURS_REGISTER, configs.Configs.JWT.JWT_SECRET)
	if genRegTokerr != nil {
		return nil, nil, fmt.Errorf("failed to generate register token")
	}
	if _, createRefreshErr := s.CreateRefreshToken(ctx, createdUser); createRefreshErr != nil {
		return nil, nil, createRefreshErr
	}
	return createdUser.ToResponse(), &registerToken, nil
}

func (s *authService) Login(ctx context.Context, dto dtos.LoginRequest) (*models.UserResponse, *string, error) {
	slog.Info("login attempt", "email", dto.Email)
	exisitngUser, _ := s.authRepo.FindUserByEmail(ctx, dto.Email)
	if exisitngUser == nil {
		return nil, nil, appErr.ErrUserNotFound
	}
	if !exisitngUser.ComparePassword(dto.Password) {
		slog.Warn("failed login attempt", "email", dto.Email, "reason", "invalid credentials")
		return nil, nil, fmt.Errorf("invalid credentials")
	}
	user := models.User{
		ID:        exisitngUser.ID,
		Email:     exisitngUser.Email,
		Password:  exisitngUser.Password,
		FullName:  exisitngUser.FullName,
		ImageUrl:  exisitngUser.ImageUrl,
		Gender:    exisitngUser.Gender,
		LastLogin: time.Now().UTC(),
		CreatedAt: exisitngUser.CreatedAt,
		UpdatedAt: exisitngUser.UpdatedAt,
	}
	token, genTokerr := utils.GenerateToken(user.ToResponse(), configs.Configs.JWT.JWT_EXPIRY_HOURS_LOGIN, configs.Configs.JWT.JWT_SECRET)
	if genTokerr != nil {
		return nil, nil, fmt.Errorf("failed to generate token")
	}
	existingRefreshToken, existingRefErr := s.authRepo.FindRefreshByUserID(ctx, user.ID)
	if existingRefErr != nil {
		if errors.Is(existingRefErr, appErr.ErrRefreshNotFound) {
			_, createErr := s.CreateRefreshToken(ctx, &user)
			if createErr != nil {
				return nil, nil, fmt.Errorf("failed to create refresh token: %w", createErr)
			}
			existingRefreshToken, _ = s.authRepo.FindRefreshByUserID(ctx, user.ID)
		} else {
			return nil, nil, fmt.Errorf("failed to find refresh token : %w", existingRefErr)
		}
	}
	isExpired := existingRefreshToken.ExpiresAt.Before(time.Now())
	if isExpired {
		if deleteErr := s.authRepo.DeleteRefresh(ctx, existingRefreshToken.ID); deleteErr != nil {
			return nil, nil, fmt.Errorf("failed to delete refresh token : %w", deleteErr)
		}
		if _, createErr := s.CreateRefreshToken(ctx, &user); createErr != nil {
			return nil, nil, fmt.Errorf("failed to create refresh token : %w", createErr)
		}
	}
	slog.Info("user logged in successfully", "user_id", user.ID, "email", dto.Email)
	return user.ToResponse(), &token, nil
}

func (s *authService) CreateRefreshToken(ctx context.Context, user *models.User) (*models.RefreshToken, error) {
	slog.Info("creating refresh token", "user_id", user.ID)
	refToken, genRefTokerr := utils.GenerateToken(user.ToResponse(), configs.Configs.JWT.JWT_EXPIRY_HOURS_REFRESH, configs.Configs.JWT.JWT_SECRET)
	if genRefTokerr != nil {
		return nil, fmt.Errorf("failed to generate refresh token")
	}
	refreshToken := models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     refToken,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}
	token, createRefreshErr := s.authRepo.CreateRefresh(ctx, refreshToken)
	if createRefreshErr != nil {
		return nil, fmt.Errorf("failed to create refresh token : %w", createRefreshErr)
	}
	return token, nil
}

func (s *authService) RefreshToken(ctx context.Context, userID uuid.UUID) (*models.RefreshToken, error) {
	slog.Info("refreshing token", "user_id", userID)
	user, findUserErr := s.authRepo.FindUserByID(ctx, userID)
	if findUserErr != nil {
		if errors.Is(findUserErr, appErr.ErrUserNotFound) {
			return nil, appErr.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user : %w", findUserErr)
	}
	existingRefreshToken, existingRefErr := s.authRepo.FindRefreshByUserID(ctx, userID)
	if existingRefErr != nil {
		if errors.Is(existingRefErr, appErr.ErrRefreshNotFound) {
			_, createErr := s.CreateRefreshToken(ctx, user)
			if createErr != nil {
				return nil, fmt.Errorf("failed to create refresh token: %w", createErr)
			}
			existingRefreshToken, _ = s.authRepo.FindRefreshByUserID(ctx, user.ID)
		}
	}
	isExpired := existingRefreshToken.ExpiresAt.Before(time.Now())
	if isExpired {
		if deleteErr := s.authRepo.DeleteRefresh(ctx, existingRefreshToken.ID); deleteErr != nil {
			return nil, fmt.Errorf("failed to delete refresh token : %w", deleteErr)
		}
	}
	return existingRefreshToken, nil
}

func (s *authService) FindRefreshTokenByUserID(ctx context.Context, userID uuid.UUID) (*models.RefreshToken, error) {
	refreshToken, err := s.authRepo.FindRefreshByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return refreshToken, nil
}

func (s *authService) DeleteRefresh(ctx context.Context, refreshID uuid.UUID) error {
	slog.Info("deleting refresh token", "refresh_id", refreshID)
	if err := s.authRepo.DeleteRefresh(ctx, refreshID); err != nil {
		return err
	} else {
		return nil
	}
}

func (s *authService) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.UserResponse, error) {
	slog.Info("fetching user by ID", "user_id", userID)
	user, err := s.authRepo.FindUserByID(ctx, userID)
	if err != nil {
		slog.Error("failed to get user", "user_id", userID, "error", err)
		return nil, err
	}
	return user.ToResponse(), nil
}
