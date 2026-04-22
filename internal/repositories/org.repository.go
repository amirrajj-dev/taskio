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

type OrganiztionRepository interface {
	CreateOrganization(ctx context.Context, org models.Organization) (*models.Organization, error)
	FindOrganizationsUserInvolved(ctx context.Context, userID uuid.UUID) ([]*models.CurrentUserOrganizationsResponse, error)
	CreateOrganizationUser(ctx context.Context, orgUser models.OrganizationUser) (*models.OrganizationUser, error)
	FindOrganizationByName(ctx context.Context, name string) (*models.Organization, error)
	FindOrganizationByID(ctx context.Context, orgID uuid.UUID) (*models.OrganizationWithMembersCount, error)
	FindOrganizationMembers(ctx context.Context, orgId uuid.UUID) ([]*models.OrganizationUser, error)
	// used by /api/orgs/:orgId/members
	FindOrganizationMembersResult(ctx context.Context, orgId uuid.UUID) ([]*models.OrganizationUserResult, error)
	UpdateOrganizationName(ctx context.Context, orgID uuid.UUID, name string) (bool, error)
	DeleteOrganization(ctx context.Context, orgID uuid.UUID) (bool, error)
	DeleteOrganizationUserByOrgID(ctx context.Context, orgID uuid.UUID) (bool, error)
	FindOrganizationUser(ctx context.Context, orgID, userID uuid.UUID) (*models.OrganizationUser, error)
	FindOrganizationUserByID(ctx context.Context, orgID, orgUserID uuid.UUID) (*models.OrganizationUser, error)
	FindOrganizationUserByOrgAndEmail(ctx context.Context, orgID uuid.UUID, email string) (*models.OrganizationUser, error)
	UpdateOrganizationUserRole(ctx context.Context, orgUserID uuid.UUID, orgID uuid.UUID, role string) error
}

type organizationRepository struct {
	Organization        *gorm.DB
	OrganizationUsers   *gorm.DB
	OrganizationInvites *gorm.DB
	Users               *gorm.DB
}

var OrgRepo *organizationRepository

func NewOrganizationRepository() {
	OrgRepo = &organizationRepository{
		Organization:        database.PG.Model(&models.Organization{}),
		OrganizationUsers:   database.PG.Model(&models.OrganizationUser{}),
		OrganizationInvites: database.PG.Model(&models.OrganizationInvite{}),
		Users:               database.PG.Model(&models.User{}),
	}
}

func (r *organizationRepository) CreateOrganization(ctx context.Context, org models.Organization) (*models.Organization, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := r.Organization.WithContext(opCtx).Create(&org).Error; err != nil {
		return nil, fmt.Errorf("failed to create organization : %w", err)
	}
	return &org, nil
}

func (r *organizationRepository) FindOrganizationsUserInvolved(ctx context.Context, userID uuid.UUID) ([]*models.CurrentUserOrganizationsResponse, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var userInvolvedOrganizations []models.OrganizationUser
	if err := r.OrganizationUsers.WithContext(opCtx).Where("user_id", userID).Find(&userInvolvedOrganizations).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return []*models.CurrentUserOrganizationsResponse{}, nil
		}
		return nil, fmt.Errorf("failed to find current user joined or member organizations : %w", err)
	}
	var data []*models.CurrentUserOrganizationsResponse
	for _, user := range userInvolvedOrganizations {
		currentOrg, err := r.FindOrganizationByID(opCtx, user.OrgID)
		if err != nil {
			return []*models.CurrentUserOrganizationsResponse{}, fmt.Errorf("failed to fetch org with id %s : %w", user.OrgID, err)
		}
		data = append(data, &models.CurrentUserOrganizationsResponse{
			ID:   user.OrgID,
			Name: currentOrg.Name,
			Role: user.Role,
		})
	}
	return data, nil
}

func (r *organizationRepository) CreateOrganizationUser(ctx context.Context, orgUser models.OrganizationUser) (*models.OrganizationUser, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := r.OrganizationUsers.WithContext(opCtx).Create(&orgUser).Error; err != nil {
		return nil, err
	}
	return &orgUser, nil
}

func (r *organizationRepository) FindOrganizationByName(ctx context.Context, name string) (*models.Organization, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var organization models.Organization
	if err := r.Organization.WithContext(opCtx).Find(&organization).Where("name = ?", name).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, err
	}
	return &organization, nil
}

func (r *organizationRepository) FindOrganizationByID(ctx context.Context, orgID uuid.UUID) (*models.OrganizationWithMembersCount, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var organization models.Organization
	if err := r.Organization.WithContext(opCtx).First(&organization, orgID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, err
	}

	organizationMembers, err := r.FindOrganizationMembers(opCtx, organization.ID)

	if err != nil {
		return nil, err
	}

	organizationsWithMembersCount := models.OrganizationWithMembersCount{
		ID:           organization.ID,
		Name:         organization.Name,
		OwnerID:      organization.OwnerID,
		MembersCount: len(organizationMembers),
		CreatedAt:    organization.CreatedAt,
	}
	return &organizationsWithMembersCount, nil
}

func (r *organizationRepository) FindOrganizationMembers(ctx context.Context, orgId uuid.UUID) ([]*models.OrganizationUser, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var organizationMembers []*models.OrganizationUser
	if err := r.OrganizationUsers.WithContext(opCtx).Where("org_id = ?", orgId).Find(&organizationMembers).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return []*models.OrganizationUser{}, fmt.Errorf("organization has no member")
		}
		return nil, err
	}
	return organizationMembers, nil
}

func (r *organizationRepository) FindOrganizationMembersResult(ctx context.Context, orgId uuid.UUID) ([]*models.OrganizationUserResult, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	var results []*models.OrganizationUserResult
	defer cancel()
	query := `
	SELECT u.full_name as name , ou.role 
	, u.id as user_id , ou.joined_at , u.image_url
	FROM organization_users ou 
	LEFT JOIN users u on ou.user_id = u.id 
	WHERE org_id = ?;
	`
	if err := r.OrganizationUsers.WithContext(opCtx).Raw(query , orgId).Scan(&results).Error;err != nil {
		return nil , err
	}
	return  results , nil
}

func (r *organizationRepository) UpdateOrganizationName(ctx context.Context, orgID uuid.UUID, name string) (bool, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := r.Organization.WithContext(opCtx).Where("id = ?", orgID).Update("name", name).Error; err != nil {
		return false, fmt.Errorf("failed to update organization name => %w", err)
	}
	return true, nil
}

func (r *organizationRepository) DeleteOrganization(ctx context.Context, orgID uuid.UUID) (bool, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res := r.Organization.WithContext(opCtx).
		Where("id = ?", orgID).
		Delete(&models.Organization{})

	if res.Error != nil {
		return false, fmt.Errorf("failed to delete organization : %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return false, fmt.Errorf("organization not found")
	}
	return true, nil
}

func (r *organizationRepository) DeleteOrganizationUserByOrgID(ctx context.Context, orgID uuid.UUID) (bool, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := r.OrganizationUsers.WithContext(opCtx).Where("org_id = ?", orgID).Delete(&models.OrganizationUser{}).Error; err != nil {
		return false, fmt.Errorf("failed to delete organization user")
	}
	return true, nil
}

func (r *organizationRepository) FindOrganizationUser(ctx context.Context, orgID uuid.UUID, userID uuid.UUID) (*models.OrganizationUser, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var orgUser models.OrganizationUser

	err := r.OrganizationUsers.WithContext(opCtx).
		Where("org_id = ?", orgID).Where("user_id = ?" , userID).
		First(&orgUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("organization user not found")
		}
		return nil, fmt.Errorf("database error fetching organization user: %w", err)
	}
	return &orgUser, nil
}

func (r *organizationRepository) FindOrganizationUserByID(ctx context.Context, orgID uuid.UUID, orgUserID uuid.UUID) (*models.OrganizationUser, error) {

	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var orgUser models.OrganizationUser

	err := r.OrganizationUsers.WithContext(opCtx).
		Where("id = ? AND org_id = ?", orgUserID, orgID).
		First(&orgUser).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("database error fetching org user by id: %w", err)
	}

	return &orgUser, nil
}

func (r *organizationRepository) FindOrganizationUserByOrgAndEmail(ctx context.Context, orgID uuid.UUID, email string) (*models.OrganizationUser, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var user models.User
	resUser := r.Users.WithContext(opCtx).Where("email = ?", email).First(&user)
	if resUser.Error != nil {
		if resUser.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("user not found : %w", resUser.Error)
	}
	return r.FindOrganizationUser(ctx, orgID, user.ID)
}

func (r *organizationRepository) UpdateOrganizationUserRole(ctx context.Context, orgUserID uuid.UUID, orgID uuid.UUID, role string) error {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res := r.OrganizationUsers.WithContext(opCtx).
		Where("id = ? AND org_id = ?", orgUserID, orgID).
		Update("role", role)

	if res.Error != nil {
		return fmt.Errorf("database error updating role: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("organization user not found")
	}
	return nil
}
