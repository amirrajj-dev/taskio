package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/amirrajj-dev/taskio/internal/infrastructure/database"
	"github.com/amirrajj-dev/taskio/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrganizationInviteRepository interface {
	CreateOrganizationInvite(ctx context.Context, invite models.OrganizationInvite) (*models.OrganizationInvite, error)
	FindOrganizationInviteByToken(ctx context.Context, token string) (*models.OrganizationInvite, error)
	FindOrganizationInviteByEmailAndOrgID(ctx context.Context, email string, orgID uuid.UUID) (*models.OrganizationInvite, error)
	UpdateOrganizationInviteStatus(ctx context.Context, inviteID uuid.UUID, status string) error
	DeleteOrganizationInvite(ctx context.Context, inviteID uuid.UUID) (bool, error)
	FindPendingInvitesByEmail(ctx context.Context , email string) ([]*models.OrganizationInvite , error)
	
}


type organizationInviteRepository struct {
	OrganizationInvites *gorm.DB
}

var OrgInviteRepo *organizationInviteRepository


func NewOrganizationInviteRepository() {
	OrgInviteRepo = &organizationInviteRepository{
		OrganizationInvites: database.PG.Model(&models.OrganizationInvite{}),
	}
}

func (r *organizationInviteRepository) CreateOrganizationInvite(ctx context.Context, invite models.OrganizationInvite) (*models.OrganizationInvite, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result := r.OrganizationInvites.WithContext(opCtx).Create(&invite)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return nil, fmt.Errorf("invite creation failed due to duplicate key: %w", result.Error)
		}
		return nil, fmt.Errorf("failed to create organization invite: %w", result.Error)
	}
	return &invite, nil
}

func (r *organizationInviteRepository) FindOrganizationInviteByToken(ctx context.Context, token string) (*models.OrganizationInvite, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var invite models.OrganizationInvite
	result := r.OrganizationInvites.WithContext(opCtx).Where("token = ?", token).First(&invite)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("invitation token not found or expired")
		}
		return nil, fmt.Errorf("failed to find organization invite by token: %w", result.Error)
	}
	return &invite, nil
}

func (r *organizationInviteRepository) FindOrganizationInviteByEmailAndOrgID(ctx context.Context, email string, orgID uuid.UUID) (*models.OrganizationInvite, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var invite models.OrganizationInvite
	result := r.OrganizationInvites.WithContext(opCtx).
		Where("email = ? AND org_id = ? AND status = ?", email, orgID, "invited").
		First(&invite)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find organization invite by email and orgID: %w", result.Error)
	}
	return &invite, nil
}

func (r *organizationInviteRepository) UpdateOrganizationInviteStatus(ctx context.Context, inviteID uuid.UUID, status string) error {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	allowedStatuses := map[string]struct{}{"invited": {}, "accepted": {}, "revoked": {}}
	if _, ok := allowedStatuses[status]; !ok {
		return fmt.Errorf("invalid status value: %s", status)
	}

	result := r.OrganizationInvites.WithContext(opCtx).Model(&models.OrganizationInvite{}).
		Where("id = ?", inviteID).
		Updates(map[string]any{
			"status":     status,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update organization invite status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("organization invite with ID %s not found", inviteID)
	}
	return nil
}

func (r *organizationInviteRepository) DeleteOrganizationInvite(ctx context.Context, inviteID uuid.UUID) (bool, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res := r.OrganizationInvites.WithContext(opCtx).Where("id = ?", inviteID).Delete(&models.OrganizationInvite{})
	if res.Error != nil {
		return false, fmt.Errorf("failed to delete organization invite %s : %w", inviteID, res.Error)
	}
	if res.RowsAffected == 0 {
		return false, fmt.Errorf("organization invite not found")
	}
	return true, nil
}


func (r *organizationInviteRepository) FindPendingInvitesByEmail(ctx context.Context , email string) ([]*models.OrganizationInvite , error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var pendingInvites []*models.OrganizationInvite
	res := r.OrganizationInvites.WithContext(opCtx).Where("email = ? AND status = ? AND expires_at > ?", email, "invited", time.Now()).Find(&pendingInvites)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil , fmt.Errorf("no pending invites found")
		}
		return nil, fmt.Errorf("failed to find pending organization invite %w :", res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, fmt.Errorf("organization invite not found")
	}
	return pendingInvites , nil
}