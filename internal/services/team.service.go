package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"time"

	"github.com/amirrajj-dev/taskio/internal/constants"
	"github.com/amirrajj-dev/taskio/internal/consumers"
	"github.com/amirrajj-dev/taskio/internal/events"
	"github.com/amirrajj-dev/taskio/internal/infrastructure/database"
	"github.com/amirrajj-dev/taskio/internal/infrastructure/queue"
	"github.com/amirrajj-dev/taskio/internal/models"
	"github.com/amirrajj-dev/taskio/internal/repositories"
	"github.com/amirrajj-dev/taskio/internal/websocket"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

type TeamService interface {
	CreateTeam(ctx context.Context, orgID, userID uuid.UUID, name string) (*models.Team, error)
	ListOrganizationTeams(ctx context.Context, orgID uuid.UUID) ([]*models.TeamWithMemberCount, error)
	UpdateTeam(ctx context.Context, teamID uuid.UUID, name string) (*models.Team, error)
	AddMemberToTeam(ctx context.Context, teamID, userID, actorID uuid.UUID) (*models.TeamMember, error)
	DeleteMemberFromTeam(ctx context.Context, teamID, userID, targetUserID uuid.UUID) (bool, error)
	GetTeamMembers(ctx context.Context, teamID uuid.UUID) ([]*models.TeamMemberResult, error)
	ChangeTeamMemberRole(ctx context.Context, newRole string, teamID, actorID, targetID uuid.UUID) (bool, error)
}

type teamService struct {
	teamRepo         repositories.TeamRepository
	orgRepo          repositories.OrganiztionRepository
	userRepo         repositories.UserRepository
	activityProducer *queue.ActivityProducer
	queueConn        *amqp.Connection
}

var TeamServicee *teamService

func NewTeamService(producer *queue.ActivityProducer) {
	TeamServicee = &teamService{
		teamRepo:         repositories.TeamRepo,
		orgRepo:          repositories.OrgRepo,
		userRepo:         repositories.UserRepo,
		activityProducer: producer,
		queueConn:        queue.QueueConn,
	}
}

func (s *teamService) CreateTeam(ctx context.Context, orgID, userID uuid.UUID, name string) (*models.Team, error) {
	slog.Info("creating team", "name", name, "org_id", orgID, "user_id", userID)
	existingOrg, _ := s.orgRepo.FindOrganizationByID(ctx, orgID)
	if existingOrg == nil {
		return nil, fmt.Errorf("organization not found")
	}

	existingTeam, _ := s.teamRepo.FindTeamByName(ctx, name)

	if existingTeam != nil {
		return nil, fmt.Errorf("team with the given name already exists")
	}

	var createdTeam *models.Team
	now := time.Now().UTC()

	err := database.PG.Transaction(func(tx *gorm.DB) error {
		repo := s.teamRepo.WithTx(tx)

		team := models.Team{
			ID:        uuid.New(),
			OrgID:     orgID,
			Name:      name,
			CreatorID: userID,
			CreatedAt: now,
		}
		teamMember := models.TeamMember{
			ID:       uuid.New(),
			TeamID:   team.ID,
			UserID:   userID,
			Role:     "owner",
			JoinedAt: now,
		}

		t, err := repo.CreateTeam(ctx, team)
		if err != nil {
			return err
		}
		createdTeam = t

		_, err = repo.CreateTeamMember(ctx, teamMember)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	slog.Info("team created", "team_id", createdTeam.ID, "name", name)
	slog.Info("team member created" , "team_id" , createdTeam.ID , "user_id" , userID)
	// real time broadcast to org members
	if websocket.WsManager != nil {
		user, _ := s.userRepo.FindUserByID(ctx, userID)
		payload := websocket.TeamCreatedPayload{
			Name:      name,
			CreatedBy: user.FullName,
		}
		websocket.WsManager.BroadcastToRoom("org:"+orgID.String(), websocket.EventTeamCreated, payload)
	} else {
		slog.Warn("failed to send event via websocket for create team" , "team_name" , createdTeam.Name)
	}

	event := models.ActivityEvent{
		EventName: events.ActivityEvents.CreateTeam,
		UserID:    &userID,
		OrgID:     &orgID,
		Metadata: map[string]any{
			"created_team_name": name,
		},
	}

	if err := s.activityProducer.Publish(event); err != nil {
		slog.Error("failed to publish create team event" , "error" , err.Error())
	}

	return createdTeam, nil
}

func (s *teamService) ListOrganizationTeams(ctx context.Context, orgID uuid.UUID) ([]*models.TeamWithMemberCount, error) {
	slog.Info("listing organization teams", "org_id", orgID)
	existingOrg, err := s.orgRepo.FindOrganizationByID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check organization: %w", err)
	}
	if existingOrg == nil {
		return nil, fmt.Errorf("organization not found")
	}

	teams, err := s.teamRepo.FindTeamsWithMemberCounts(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list teams with member counts: %w", err)
	}
	slog.Info("teams listed", "org_id", orgID, "count", len(teams))
	return teams, nil
}

func (s *teamService) UpdateTeam(ctx context.Context, teamID uuid.UUID, name string) (*models.Team, error) {
	slog.Info("updating team", "team_id", teamID, "new_name", name)
	existingTeam, findTeamErr := s.teamRepo.FindTeamByID(ctx, teamID)
	if findTeamErr != nil {
		return nil, findTeamErr
	}
	if strings.TrimSpace(existingTeam.Name) == strings.TrimSpace(name) {
		return nil, fmt.Errorf("please provide a new name for the team")
	}
	if updated, updateTeamErr := s.teamRepo.UpdateTeamName(ctx, teamID, name); !updated {
		return nil, updateTeamErr
	}
	slog.Info("team updated", "team_id", teamID, "old_name", existingTeam.Name, "new_name", name)
	if websocket.WsManager != nil {
		payload := websocket.TeamUpdatedPayload{
			Name: name,
		}
		websocket.WsManager.BroadcastToRoom("team:"+teamID.String(), websocket.EventTeamUpdated, payload)
	} else {
		slog.Warn("failed to send event via websocket for update team" , "team_name" , name , "team_id" , teamID)
	}

	return &models.Team{
		ID:    existingTeam.ID,
		Name:  name,
		OrgID: existingTeam.OrgID,
	}, nil
}

func (s *teamService) AddMemberToTeam(ctx context.Context, teamID, userID, actorID uuid.UUID) (*models.TeamMember, error) {
	slog.Info("adding member to team", "team_id", teamID, "user_id", userID, "actor_id", actorID)
	existingTeam, findTeamErr := s.teamRepo.FindTeamByID(ctx, teamID)
	if findTeamErr != nil {
		return nil, findTeamErr
	}
	existingTeamMember, _ := s.teamRepo.FindTeamMember(ctx, teamID, userID)
	if existingTeamMember != nil {
		return nil, fmt.Errorf("already a member of the team")
	}
	_, findOrgUserErr := s.orgRepo.FindOrganizationUser(ctx, existingTeam.OrgID, userID)
	if findOrgUserErr != nil {
		return nil, findOrgUserErr
	}
	teamMember := models.TeamMember{
		ID:       uuid.New(),
		TeamID:   existingTeam.ID,
		UserID:   userID,
		Role:     "member",
		JoinedAt: time.Now().UTC(),
	}
	if added, err := s.teamRepo.AddMemberToTeam(ctx, teamMember); !added {
		return nil, err
	}

	slog.Info("member added to team", "team_id", teamID, "user_id", userID, "role", "member")

	if websocket.WsManager != nil {
		addedToTeamUser, _ := s.userRepo.FindUserByID(ctx, userID)
		payload := websocket.TeamMemberAddedPayload{
			TeamID:   teamID,
			TeamName: existingTeam.Name,
			UserID:   userID,
			UserName: addedToTeamUser.FullName,
			Role:     teamMember.Role,
			AddedBy:  actorID,
		}
		websocket.WsManager.BroadcastToRoom("team:"+teamID.String(), websocket.EventTeamMemberAdded, payload)
	} else {
		slog.Warn("failed to send event via websocket for add member to team")
	}

	event := models.ActivityEvent{
		EventName: events.ActivityEvents.AddMemberToTeam,
		UserID:    &userID,
		OrgID:     &existingTeam.OrgID,
		TeamID:    &teamID,
		Metadata: map[string]any{
			"new_team_member_user_id": teamMember.UserID,
		},
	}
	if err := s.activityProducer.Publish(event); err != nil {
		slog.Error("failed to publish add member to team event" , "error" , err.Error())
	}
	return &teamMember, nil
}

func (s *teamService) DeleteMemberFromTeam(
	ctx context.Context, teamID, userID, targetUserID uuid.UUID,
) (bool, error) {
	slog.Info("removing member from team", "team_id", teamID, "target_user_id", targetUserID, "actor_id", userID)
	if _, err := s.teamRepo.FindTeamByID(ctx, teamID); err != nil {
		return false, err
	}

	actor, err := s.teamRepo.FindTeamMember(ctx, teamID, userID)
	if err != nil {
		return false, err
	}
	target, err := s.teamRepo.FindTeamMember(ctx, teamID, targetUserID)
	if err != nil {
		return false, err
	}

	if actor.Role == "admin" && target.Role == "admin" {
		return false, fmt.Errorf("admins cannot remove other admins")
	}
	if target.Role == "owner" && actor.Role != "owner" {
		return false, fmt.Errorf("only owner can remove the owner")
	}

	deleted, err := s.teamRepo.DeleteMemberFromTeam(ctx, teamID, targetUserID)
	if err != nil {
		return false, err
	}
	slog.Info("member removed from team", "team_id", teamID, "target_user_id", targetUserID)

	if !deleted {
		return false, fmt.Errorf("unexpected: member not deleted")
	}

	team, _ := s.teamRepo.FindTeamByID(ctx, teamID)

	members, err := s.teamRepo.FindTeamMembers(ctx, teamID)
	if err != nil {
		return true, nil // don't block success
	}
	if len(members) == 0 {
		orgMembers, _ := s.orgRepo.FindOrganizationMembers(ctx, team.OrgID)
		if websocket.Publisher != nil && len(orgMembers) > 0 {
			userIDs := make([]uuid.UUID, len(orgMembers))
			for i, member := range orgMembers {
				userIDs[i] = member.UserID
			}
			payload := websocket.TeamDeletedPayload{
				Name: team.Name,
			}
			// non blocking batch publish
			go websocket.Publisher.PublishToMany(userIDs, websocket.EventTeamDeleted, payload)
		} else {
			slog.Warn("failed to send event via websocket for delete member from team" , "team_id" , teamID , "delete_member_id" , targetUserID)
		}

		s.teamRepo.DeleteTeam(ctx, teamID)
		if deleted, err := s.teamRepo.DeleteTeam(ctx, teamID); !deleted {
			log.Printf("failed to delete empty team %s: %v", teamID, err)
			return true, nil // member deleted, team cleanup failed but don't rollback member deletion
		}

		// delete related team members , projects , tasks , comments
		// Publish cleanup event for related data
		cleanupEvent := consumers.TeamCleanupEvent{
			TeamID:    teamID,
			OrgID:     team.OrgID,
			DeletedBy: userID,
			Timestamp: time.Now().UTC(),
		}

		if err := s.publishTeamCleanupEvent(cleanupEvent); err != nil {
			slog.Error("WARNING: failed to publish team cleanup event" , "error" , err.Error())
		}
	}

	if len(members) != 0 {
		if websocket.WsManager != nil {
			deletedFromTeamUser, _ := s.userRepo.FindUserByID(ctx, targetUserID)
			payload := websocket.TeamMemberDeletedPayload{
				TeamID:    teamID,
				TeamName:  team.Name,
				UserID:    targetUserID,
				UserName:  deletedFromTeamUser.FullName,
				Role:      target.Role,
				DeletedBy: actor.UserID,
			}
			websocket.WsManager.BroadcastToRoom("team:"+teamID.String(), websocket.EventTeamMemberDeleted, payload)
		}
	}

	event := models.ActivityEvent{
		EventName: events.ActivityEvents.DeleteMemberFromTeam,
		UserID:    &userID,
		TeamID:    &teamID,
		Metadata: map[string]any{
			"deleteD_member_id": targetUserID,
		},
	}
	if err := s.activityProducer.Publish(event); err != nil {
		slog.Error("failed to publish delete member from team event" , "error" , err.Error())
	}
	return true, nil
}

func (s *teamService) GetTeamMembers(ctx context.Context, teamID uuid.UUID) ([]*models.TeamMemberResult, error) {
	slog.Info("fetching team members", "team_id", teamID)
	_, findTeamErr := s.teamRepo.FindTeamByID(ctx, teamID)
	if findTeamErr != nil {
		return nil, findTeamErr
	}
	teamMembers, err := s.teamRepo.FindTeamMembers(ctx, teamID)
	if err != nil {
		return nil, err
	}
	slog.Info("team members fetched", "team_id", teamID, "count", len(teamMembers))
	return teamMembers, nil
}

func (s *teamService) ChangeTeamMemberRole(
	ctx context.Context,
	newRole string,
	teamID, actorID, targetID uuid.UUID,
) (bool, error) {
	slog.Info("changing team member role", "team_id", teamID, "target_id", targetID, "new_role", newRole, "actor_id", actorID)
	// Validate role input
	if !isValidRole(newRole) {
		return false, fmt.Errorf("invalid role: %s", newRole)
	}

	// Ensure team exists
	if _, err := s.teamRepo.FindTeamByID(ctx, teamID); err != nil {
		return false, err
	}

	// Fetch actor + target
	actor, err := s.teamRepo.FindTeamMember(ctx, teamID, actorID)
	if err != nil {
		return false, err
	}
	target, err := s.teamRepo.FindTeamMember(ctx, teamID, targetID)
	if err != nil {
		return false, err
	}

	// Prevent self‑change to owner by admin
	if actorID == targetID && newRole == "owner" && actor.Role != "owner" {
		return false, errors.New("only the current owner can assign ownership")
	}

	// Permission rules
	if actor.Role == "admin" && target.Role == "owner" {
		return false, errors.New("admins cannot modify the owner")
	}
	if actor.Role == "admin" && target.Role == "admin" && newRole == "owner" {
		return false, errors.New("admins cannot promote admins to owner")
	}

	// (ownership transfer)
	if actor.Role == "owner" && newRole == "owner" {
		return s.transferOwnership(ctx, teamID, actor, target)
	}

	// Normal role change
	if changed, changeErr := s.teamRepo.ChangeTeamMemberRole(ctx, newRole, teamID, targetID); !changed {
		return false, changeErr
	}

	slog.Info("team member role changed", "team_id", teamID, "target_id", targetID, "new_role", newRole)

	if websocket.WsManager != nil {
		paylaod := websocket.TeamMemberRoleChangedPayload{
			NewRole:  newRole,
			ActorID:  actor.UserID,
			TargetID: target.UserID,
		}
		websocket.WsManager.BroadcastToRoom("team:"+teamID.String(), websocket.EventTeamMemberRoleChanged, paylaod)
	}else {
		slog.Warn("failed to send event via websocket for change team member role" , "team_id" , teamID , "role" , newRole)
	}

	event := models.ActivityEvent{
		EventName: events.ActivityEvents.ChangeTeamMemberRole,
		UserID:    &actorID,
		TeamID:    &teamID,
		Metadata: map[string]any{
			"target_user_id": targetID,
			"new_role":       newRole,
		},
	}
	if err := s.activityProducer.Publish(event); err != nil {
		slog.Error("failed to publish change team member role event" , "error" , err.Error())
	}
	return true, nil
}

func (s *teamService) transferOwnership(
	ctx context.Context,
	teamID uuid.UUID,
	oldOwner, newOwner *models.TeamMember,
) (bool, error) {
	slog.Info("transferring team ownership", "team_id", teamID, "old_owner", oldOwner.UserID, "new_owner", newOwner.UserID)
	err := database.PG.Transaction(func(tx *gorm.DB) error {

		txRepo := s.teamRepo.WithTx(tx)

		// Promote new owner
		changed, err := txRepo.ChangeTeamMemberRole(ctx, "owner", teamID, newOwner.UserID)
		if err != nil {
			return err // Returning error triggers rollback
		}
		if !changed {
			return errors.New("failed to promote new owner")
		}

		// Demote old owner to admin
		changed, err = txRepo.ChangeTeamMemberRole(ctx, "admin", teamID, oldOwner.UserID)
		if err != nil {
			return err
		}
		if !changed {
			return errors.New("failed to demote old owner")
		}

		return nil // Returning nil triggers commit
	})

	if err != nil {
		return false, err
	}

	slog.Info("team ownership transferred", "team_id", teamID)

	if websocket.WsManager != nil {
		payload := websocket.TransferOwnerShipPayload{
			ActorID:  oldOwner.UserID,
			NewOwner: newOwner.UserID,
		}
		websocket.WsManager.BroadcastToRoom("team:"+teamID.String(), websocket.EventTransferOwnerShip, payload)
	}else {
		slog.Warn("failed to send event via websocket for transfer ownership")
	}

	event := models.ActivityEvent{
		EventName: events.ActivityEvents.ChangeTeamMemberRole,
		UserID:    &oldOwner.UserID,
		TeamID:    &teamID,
		Metadata: map[string]any{
			"new_owner_id": newOwner.UserID,
		},
	}

	if err := s.activityProducer.Publish(event); err != nil {
		slog.Error("failed to publish change team member role event" , "error" , err.Error())
	}

	return true, nil
}

func isValidRole(role string) bool {
	return role == "owner" || role == "admin" || role == "member" || role == "viewer"
}

func (s *teamService) publishTeamCleanupEvent(event consumers.TeamCleanupEvent) error {
	slog.Info("publishing team cleanup event", "team_id", event.TeamID)
	conn := queue.QueueConn
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(constants.TEAM_CLEANUP_QUEUE, true, false, false, false, nil)
	if err != nil {
		return err
	}

	body, err := json.Marshal(event)
	if err != nil {
		slog.Error("failed to marshal event" , "error" , err.Error() , "queue" , constants.TEAM_CLEANUP_QUEUE)
		return err
	}

	slog.Info("publishing team cleanup event ...")

	return ch.Publish(
		"",
		constants.TEAM_CLEANUP_QUEUE,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
}
