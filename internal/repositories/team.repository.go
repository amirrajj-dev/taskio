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

type TeamRepository interface {
	WithTx(tx *gorm.DB) *teamRepository
	CreateTeam(ctx context.Context, team models.Team) (*models.Team, error)
	CreateTeamMember(ctx context.Context, teamMember models.TeamMember) (*models.TeamMember, error)
	ListOrganizationTeams(ctx context.Context, orgID uuid.UUID) ([]*models.Team, error)
	UpdateTeamName(ctx context.Context, teamID uuid.UUID, name string) (bool, error)
	AddMemberToTeam(ctx context.Context, teamMember models.TeamMember) (bool, error)
	DeleteMemberFromTeam(ctx context.Context, teamID, userID uuid.UUID) (bool, error)
	FindTeamByID(ctx context.Context, teamID uuid.UUID) (*models.Team, error)
	FindTeamByName(ctx context.Context, name string) (*models.Team, error)
	FindTeamMember(ctx context.Context, teamID, userID uuid.UUID) (*models.TeamMember, error)
	FindTeamMembers(ctx context.Context, teamID uuid.UUID) ([]*models.TeamMemberResult, error)
	FindTeamsWithMemberCounts(ctx context.Context, orgID uuid.UUID) ([]*models.TeamWithMemberCount, error)
	DeleteTeam(ctx context.Context, teamID uuid.UUID) (bool, error)
	CheckTeamMemberShip(ctx context.Context , teamID , userID uuid.UUID) (bool , error)
	ChangeTeamMemberRole(ctx context.Context , role string , teamID , userID uuid.UUID) (bool , error)
}

type teamRepository struct {
	Teams       *gorm.DB
	TeamMembers *gorm.DB
}

var TeamRepo *teamRepository

func NewTeamRepository() {
	TeamRepo = &teamRepository{
		Teams:       database.PG.Model(&models.Team{}),
		TeamMembers: database.PG.Model(&models.TeamMember{}),
	}
}

func (r *teamRepository) WithTx(tx *gorm.DB) *teamRepository {
	return &teamRepository{
		Teams:       tx.Model(&models.Team{}),
		TeamMembers: tx.Model(&models.TeamMember{}),
	}
}

func (r *teamRepository) CreateTeam(ctx context.Context, team models.Team) (*models.Team, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res := r.Teams.WithContext(opCtx).Create(&team)
	if res.Error != nil {
		return nil, fmt.Errorf("failed to create team : %w", res.Error)
	}
	return &team, nil
}

func (r *teamRepository) CreateTeamMember(ctx context.Context, teamMember models.TeamMember) (*models.TeamMember, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res := r.TeamMembers.WithContext(opCtx).Create(&teamMember)
	if res.Error != nil {
		return nil, fmt.Errorf("failed to create team member : %w", res.Error)
	}
	return &teamMember, nil
}

func (r *teamRepository) ListOrganizationTeams(ctx context.Context, orgID uuid.UUID) ([]*models.Team, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var teams []*models.Team
	res := r.Teams.WithContext(opCtx).Where("org_id = ?", orgID).Find(&teams)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("organization has no team")
		}
		return nil, fmt.Errorf("failed to find organization teams : %w", res.Error)
	}
	return teams, nil
}

func (r *teamRepository) UpdateTeamName(ctx context.Context, teamID uuid.UUID, name string) (bool, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res := r.Teams.WithContext(opCtx).Where("id = ?", teamID).Update("name", name)
	if res.Error != nil {
		return false, fmt.Errorf("failed to update team name")
	}
	if res.RowsAffected == 0 {
		return false, fmt.Errorf("update operation failed for team name")
	}
	return true, nil
}

func (r *teamRepository) AddMemberToTeam(ctx context.Context, teamMember models.TeamMember) (bool, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res := r.TeamMembers.WithContext(opCtx).Create(&teamMember)
	if res.Error != nil {
		return false, fmt.Errorf("failed to add user to team")
	}
	return true, nil
}

func (r *teamRepository) DeleteMemberFromTeam(ctx context.Context, teamID, userID uuid.UUID) (bool, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res := r.TeamMembers.WithContext(opCtx).Where(&models.TeamMember{TeamID: teamID, UserID: userID}).Delete(&models.TeamMember{})
	if res.Error != nil {
		return false, fmt.Errorf("failed to delete team member : %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return false, fmt.Errorf("team member not found")
	}
	return true, nil
}

func (r *teamRepository) FindTeamByID(ctx context.Context, teamID uuid.UUID) (*models.Team, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var team models.Team
	res := r.Teams.WithContext(opCtx).Where("id = ?", teamID).First(&team)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("team not found")
		}
		return nil, fmt.Errorf("failed to find team : %w", res.Error)
	}
	return &team, nil
}

func (r *teamRepository) FindTeamByName(ctx context.Context, name string) (*models.Team, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var team models.Team
	res := r.Teams.WithContext(opCtx).Where("name = ?", name).First(&team)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("team not found")
		}
		return nil, fmt.Errorf("failed to find team : %w", res.Error)
	}
	return &team, nil
}

func (r *teamRepository) FindTeamMember(ctx context.Context, teamID, userID uuid.UUID) (*models.TeamMember, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var teamMember *models.TeamMember
	res := r.TeamMembers.WithContext(opCtx).Where("team_id = ?", teamID).Where("user_id = ?", userID).First(&teamMember)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("team member not found")
		}
		return nil, res.Error
	}
	return teamMember, nil
}

func (r *teamRepository) FindTeamMembers(ctx context.Context, teamID uuid.UUID) ([]*models.TeamMemberResult, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var teamMembers []*models.TeamMemberResult
	query := `
	SELECT u.full_name as name , tm.user_id , tm.role , u.image_url 
	from team_members tm JOIN users u
	ON tm.user_id = u.id WHERE tm.team_id = ?;
	`
	res := r.TeamMembers.WithContext(opCtx).Raw(query , teamID).Scan(&teamMembers)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("team has no team member")
		}
		return nil, fmt.Errorf("failed to find team members : %w", res.Error)
	}
	return teamMembers, nil
}

func (r *teamRepository) FindTeamsWithMemberCounts(ctx context.Context, orgID uuid.UUID) ([]*models.TeamWithMemberCount, error) {
	var results []*models.TeamWithMemberCount

	query := `
        SELECT t.id,
               t.name,
               COUNT(tm.id) AS members_count
        FROM teams t
        LEFT JOIN team_members tm ON tm.team_id = t.id
        WHERE t.org_id = ?
        GROUP BY t.id, t.name
        ORDER BY t.name
    `

	if err := r.Teams.WithContext(ctx).Raw(query, orgID).Scan(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}

func (r *teamRepository) DeleteTeam(ctx context.Context, teamID uuid.UUID) (bool, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res := r.Teams.WithContext(opCtx).Where("id = ?", teamID).Delete(&models.Team{})
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return false, fmt.Errorf("team not found")
		}
		return false, fmt.Errorf("failed to delete team : %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return false, fmt.Errorf("delete operation for team failed")
	}
	return true, nil
}


func (r *teamRepository) CheckTeamMemberShip(ctx context.Context , teamID , userID uuid.UUID) (bool , error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res := r.TeamMembers.WithContext(opCtx).Where("team_id = ?" , teamID).Where("user_id = ?" , userID)
	if res.Error != nil {
		if errors.Is(res.Error , gorm.ErrRecordNotFound) {
			return false , fmt.Errorf("not a part of the team")
		}
		return false , res.Error
	}
	return true , nil
}

func (r *teamRepository) ChangeTeamMemberRole(ctx context.Context, role string, teamID, userID uuid.UUID) (bool, error) {
    result := r.TeamMembers.WithContext(ctx).
        Where("team_id = ? AND user_id = ?", teamID, userID).
        Update("role", role)
        
    if result.Error != nil {
        return false, result.Error
    }
    return result.RowsAffected > 0, nil
}

