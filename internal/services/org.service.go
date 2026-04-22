package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/amirrajj-dev/taskio/internal/configs"
	"github.com/amirrajj-dev/taskio/internal/constants"
	"github.com/amirrajj-dev/taskio/internal/consumers"
	"github.com/amirrajj-dev/taskio/internal/dtos"
	"github.com/amirrajj-dev/taskio/internal/events"
	"github.com/amirrajj-dev/taskio/internal/infrastructure/queue"
	"github.com/amirrajj-dev/taskio/internal/models"
	"github.com/amirrajj-dev/taskio/internal/repositories"
	"github.com/amirrajj-dev/taskio/internal/websocket"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type OrganizationService interface {
	CreateOrganization(ctx context.Context, org dtos.CreateOrganizationRequest, userID uuid.UUID) (*models.Organization, error)
	GetOrganizations(ctx context.Context, userID uuid.UUID) ([]*models.CurrentUserOrganizationsResponse, error)
	GetOrganization(ctx context.Context, orgID uuid.UUID) (*models.OrganizationWithMembersCount, error)
	CheckMemberShip(ctx context.Context, orgID uuid.UUID, userID uuid.UUID) (bool, error)
	GetOrganizationMember(ctx context.Context, orgID, userID uuid.UUID) (*models.OrganizationUser, error)
	GetOrganizationMembers(ctx context.Context, orgID uuid.UUID) ([]*models.OrganizationUserResult, error)
	UpdateOrganizationName(ctx context.Context, orgID, userID uuid.UUID, name string) (bool, error)
	DeleteOrganization(ctx context.Context, orgID, userID uuid.UUID) (bool, error)
	UpdateOrganizationUserRole(
		ctx context.Context,
		orgID uuid.UUID,
		currentUserID uuid.UUID,
		targetOrgUserID uuid.UUID,
		newRole string,
	) (*models.OrganizationUser, error)
	InviteToOrganization(ctx context.Context, orgID uuid.UUID, email string, invitedBy uuid.UUID) (bool, error)
	AcceptInvitation(ctx context.Context, token string, userID uuid.UUID) (*models.OrganizationUser, error)
	RejectInvitation(ctx context.Context, token string, userID uuid.UUID) error
	GetPendingInvitations(ctx context.Context, userID uuid.UUID) ([]*models.OrganizationInvite, error)
}

type organizationService struct {
	orgRepo          repositories.OrganiztionRepository
	orgInviteRepo    repositories.OrganizationInviteRepository
	authRepo         repositories.UserRepository
	queueConn        *amqp.Connection
	activityProducer *queue.ActivityProducer
}

type EmailMessage struct {
	Subject string `json:"subject"`
	To      string `json:"to"`
	Body    string `json:"body"`
	IsHtml  bool   `json:"isHtml"`
	LinkUrl string `json:"linkUrl"`
}

var OrgService *organizationService

func NewOrganizationService(producer *queue.ActivityProducer) {
	OrgService = &organizationService{
		orgRepo:          repositories.OrgRepo,
		orgInviteRepo:    repositories.OrgInviteRepo,
		authRepo:         repositories.UserRepo,
		queueConn:        queue.QueueConn,
		activityProducer: producer,
	}
}

func (s *organizationService) CreateOrganization(ctx context.Context, org dtos.CreateOrganizationRequest, userID uuid.UUID) (*models.Organization, error) {
	slog.Info("creating organization", "name", org.Name, "owner_id", userID)
	_, err := s.orgRepo.FindOrganizationByName(ctx, org.Name)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	organization := models.Organization{
		ID:        uuid.New(),
		Name:      org.Name,
		OwnerID:   userID,
		CreatedAt: now,
	}
	createdOrg, createOrgErr := s.orgRepo.CreateOrganization(ctx, organization)
	if createOrgErr != nil {
		return nil, err
	}
	organizationUser := models.OrganizationUser{
		ID:       uuid.New(),
		OrgID:    createdOrg.ID,
		UserID:   userID,
		Role:     "owner",
		JoinedAt: now,
	}
	_, createOrgUserErr := s.orgRepo.CreateOrganizationUser(ctx, organizationUser)
	if createOrgUserErr != nil {
		return nil, createOrgUserErr
	}
	slog.Info("organization created", "org_id", createdOrg.ID, "name", org.Name)
	return createdOrg, nil
}

func (s *organizationService) GetOrganizations(ctx context.Context, userID uuid.UUID) ([]*models.CurrentUserOrganizationsResponse, error) {
	slog.Info("fetching organizations user involved", "user_id", userID)
	organizations, err := s.orgRepo.FindOrganizationsUserInvolved(ctx, userID)
	if err != nil {
		return nil, err
	}
	slog.Info("user involved organizations fetched successfully")
	return organizations, nil
}

func (s *organizationService) GetOrganization(ctx context.Context, orgID uuid.UUID) (*models.OrganizationWithMembersCount, error) {
	slog.Info("fetching organization", "org_id", orgID)
	organization, err := s.orgRepo.FindOrganizationByID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	slog.Info("organization fetched successfully", "org_id", orgID)
	return organization, nil
}

func (s *organizationService) CheckMemberShip(ctx context.Context, orgID uuid.UUID, userID uuid.UUID) (bool, error) {
	slog.Info("checking membership", "org_id", orgID, "user_id", userID)
	organization, err := s.orgRepo.FindOrganizationByID(ctx, orgID)
	if err != nil {
		return false, err
	}
	organizationMembers, findOrgMembersErr := s.orgRepo.FindOrganizationMembers(ctx, organization.ID)
	if findOrgMembersErr != nil {
		return false, nil
	}
	for _, member := range organizationMembers {
		if member.UserID == userID {
			slog.Info("part of the organization", "org_id", orgID, "user_id", userID)
			return true, nil
		}
	}
	slog.Info("not a part of the organization", "org_id", orgID, "user_id", userID)
	return false, fmt.Errorf("your not a member of this organization")
}

func (s *organizationService) GetOrganizationMember(ctx context.Context, orgID, userID uuid.UUID) (*models.OrganizationUser, error) {
	slog.Info("fetching organization member" , "org_id" ,orgID , "user_id" , userID)
	organization, getOrgErr := s.GetOrganization(ctx, orgID)
	if organization == nil {
		return nil, getOrgErr
	}
	orgMembers, orgMembersErr := s.orgRepo.FindOrganizationMembers(ctx, organization.ID)
	if orgMembersErr != nil {
		return nil, orgMembersErr
	}
	var orgMember models.OrganizationUser
	for _, member := range orgMembers {
		if member.UserID == userID {
			orgMember = *member
		}
	}
	slog.Info("organization member fetched successfully" , "org_id" , orgID , "user_id" , userID)
	return &orgMember, nil
}

func (s *organizationService) GetOrganizationMembers(ctx context.Context, orgID uuid.UUID) ([]*models.OrganizationUserResult, error) {
	slog.Info("fetching organization members" , "org_id" , orgID)
	organization, getOrgErr := s.GetOrganization(ctx, orgID)
	if organization == nil {
		return nil, getOrgErr
	}
	orgMembers, findOrgMembersErr := s.orgRepo.FindOrganizationMembersResult(ctx, orgID)
	if findOrgMembersErr != nil {
		return nil, findOrgMembersErr
	}
	slog.Info("organization members fetched successfully" , "org_id" , orgID)
	return orgMembers, nil
}

func (s *organizationService) UpdateOrganizationName(ctx context.Context, orgID, userID uuid.UUID, name string) (bool, error) {
	slog.Info("updating organization name", "org_id", orgID, "new_name", name, "updated_by", userID)
	organization, getOrgErr := s.GetOrganization(ctx, orgID)
	if organization == nil {
		return false, getOrgErr
	}
	if strings.Trim(organization.Name, "") == strings.TrimSpace(name) {
		return false, fmt.Errorf("please provide a new name for the organization")
	}
	updated, err := s.orgRepo.UpdateOrganizationName(ctx, orgID, name)
	if !updated {
		return false, err
	}
	slog.Info("organization name updated successfuly" , "org_id" , orgID , "user_id" , userID)

	if websocket.WsManager != nil {
		payload := websocket.OrganizationUpdatedPayload{
			Name: name,
		}
		websocket.WsManager.BroadcastToRoom("org:"+orgID.String(), websocket.EventOrganizationUpdated, payload)
	} else {
		slog.Warn("failed to send event via websocket for update organization name" , "org_name" , name)
	}

	event := models.ActivityEvent{
		EventName: events.ActivityEvents.UpdateOrg,
		UserID:    &userID,
		OrgID:     &orgID,
		Metadata: map[string]any{
			"new_org_name": name,
		},
	}
	if err := s.activityProducer.Publish(event); err != nil {
		slog.Error("failed to publish update org event" , "err" , err.Error())
	}
	return true, nil
}

func (s *organizationService) DeleteOrganization(ctx context.Context, orgID, userID uuid.UUID) (bool, error) {
	existingOrg, findOrgErr := s.orgRepo.FindOrganizationByID(ctx, orgID)
	
	if findOrgErr != nil {
		return false, findOrgErr
	}
	
	if existingOrg == nil {
		return false, fmt.Errorf("organization not found")
	}
	slog.Info("deleting organization", "org_id", orgID, "name", existingOrg.Name, "deleted_by", userID)

	orgMembers, _ := s.orgRepo.FindOrganizationMembers(ctx, orgID)

	if websocket.Publisher != nil && len(orgMembers) > 0 {
		userIDs := make([]uuid.UUID, len(orgMembers))
		for i, member := range orgMembers {
			userIDs[i] = member.UserID
		}

		payload := websocket.OrganizationDeletedPayload{
			Name: existingOrg.Name,
		}

		// Non-blocking batch publish
		go websocket.Publisher.PublishToMany(userIDs, websocket.EventOrganizationDeleted, payload)
	} else {
		slog.Warn("failed to send event via websocket for delete organization" , "org_id" , orgID)
	}

	deletedOrganization, deleteOrganizationErr := s.orgRepo.DeleteOrganization(ctx, orgID)
	if !deletedOrganization {
		slog.Error("failed to delete organization", "org_id", orgID, "error", deleteOrganizationErr)
		return false, fmt.Errorf("failed to delete organization %s : %w", orgID, deleteOrganizationErr)
	}

	slog.Info("organization deleted successfully" , "org_id" , orgID , "user_id" , userID)

	// delete related organization members , teams , team members , projects , tasks , comments
	// Publish cleanup event to separate queue
	cleanupEvent := consumers.OrgCleanupEvent{
		OrgID:     orgID,
		DeletedBy: userID,
		Timestamp: time.Now().UTC(),
	}

	if err := s.publishOrgCleanupEvent(s.queueConn, cleanupEvent); err != nil {
		slog.Warn("WARNING: failed to publish org cleanup event" , "error" , err.Error())
	}

	// Publish activity event for logging
	event := models.ActivityEvent{
		EventName: events.ActivityEvents.DeleteOrg,
		UserID:    &userID,
		OrgID:     &orgID,
		Metadata: map[string]any{
			"name": existingOrg.Name,
		},
	}
	if err := s.activityProducer.Publish(event); err != nil {
		slog.Warn("failed to publish update org event" , "org_id" , orgID)
	}
	return true, nil
}

func (s *organizationService) UpdateOrganizationUserRole(ctx context.Context, orgID uuid.UUID, currentUserID uuid.UUID, targetOrgUserID uuid.UUID, newRole string) (*models.OrganizationUser, error) {
	slog.Info("updating user role", "org_id", orgID, "target_user", targetOrgUserID, "new_role", newRole, "updated_by", currentUserID)
	const allowedRoles = "owner,admin"
	validRoles := []string{"owner", "admin", "member"}

	org, err := s.orgRepo.FindOrganizationByID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch org: %w", err)
	}
	if org == nil {
		return nil, fmt.Errorf("organization not found")
	}

	currentOrgUser, err := s.orgRepo.FindOrganizationUser(ctx, orgID, currentUserID)
	if err != nil || currentOrgUser == nil {
		return nil, fmt.Errorf("unauthorized to modify roles")
	}

	targetOrgUser, err := s.orgRepo.FindOrganizationUserByID(ctx, orgID, targetOrgUserID)
	if err != nil || targetOrgUser == nil {
		return nil, fmt.Errorf("target user not found in organization")
	}

	if currentOrgUser.Role == "owner" && targetOrgUser.Role == "owner" {
		return nil, fmt.Errorf("forbidden: can not do this in this route")
	}

	if currentOrgUser.Role != "owner" && currentOrgUser.Role != "admin" {
		return nil, fmt.Errorf("forbidden: insufficient privileges")
	}

	if !slices.Contains(validRoles, newRole) {
		return nil, fmt.Errorf("invalid role specified")
	}

	if currentOrgUser.Role == "admin" && targetOrgUser.Role == "owner" {
		return nil, fmt.Errorf("admin cannot modify owner roles")
	}

	if newRole == targetOrgUser.Role {
		return nil, fmt.Errorf("already has role: %s", newRole)
	}

	if err := s.orgRepo.UpdateOrganizationUserRole(ctx, targetOrgUser.ID, orgID, newRole); err != nil {
		return nil, fmt.Errorf("update operation failed: %w", err)
	}

	slog.Info("organizaion user role updated" , "org_id" , orgID , "target_user" , targetOrgUser.UserID , "new_role" , newRole)

	if websocket.WsManager != nil {
		user, _ := s.authRepo.FindUserByID(ctx, targetOrgUser.UserID)
		payload := websocket.OrganizationUserRoleChanged{
			NewRole: newRole,
			Name:    user.FullName,
		}
		websocket.WsManager.BroadcastToRoom("org:"+orgID.String(), websocket.EventOrganizationUserRoleChanged, payload)
	} else {
		slog.Warn("failed to send event via websocket for update org user role" , "org_id" , orgID , "role" , newRole)
	}

	event := models.ActivityEvent{
		EventName: events.ActivityEvents.UpdateOrgUserRole,
		UserID:    &currentUserID,
		OrgID:     &orgID,
		Metadata: map[string]any{
			"role":           newRole,
			"old_role":       targetOrgUser.Role,
			"target_user_id": targetOrgUserID,
		},
	}

	if err := s.activityProducer.Publish(event); err != nil {
		slog.Warn("failed to publish event" , "error", err.Error())
	}

	return &models.OrganizationUser{
		ID:       targetOrgUser.ID,
		OrgID:    orgID,
		UserID:   targetOrgUser.UserID,
		Role:     newRole,
		JoinedAt: targetOrgUser.JoinedAt,
	}, nil
}

func (s *organizationService) InviteToOrganization(ctx context.Context, orgID uuid.UUID, email string, invitedBy uuid.UUID) (bool, error) {
	slog.Info("inviting user to organization", "org_id", orgID, "email", email, "invited_by", invitedBy)
	existingOrg, _ := s.orgRepo.FindOrganizationByID(ctx, orgID)
	if existingOrg == nil {
		log.Printf("Error finding organization %s", orgID)
		return false, fmt.Errorf("failed to find organization")
	}
	existingOrgUser, _ := s.orgRepo.FindOrganizationUserByOrgAndEmail(ctx, orgID, email)
	if existingOrgUser != nil {
		return false, fmt.Errorf("user with email %s is already a member of %s organization", email, existingOrg.Name)
	}
	existingInvite, _ := s.orgInviteRepo.FindOrganizationInviteByEmailAndOrgID(ctx, email, existingOrg.ID)
	if existingInvite != nil {
		if existingInvite.IsExpired() {
			if deleted, err := s.orgInviteRepo.DeleteOrganizationInvite(ctx, existingInvite.ID); !deleted {
				return false, fmt.Errorf("failed to delete expired organization invite : %w", err)
			}
			log.Println("expired invite token deleted succesfully")
		} else {
			return false, fmt.Errorf("user with this already has a valid invitation with status %s for organization %s", existingInvite.Status, existingOrg.Name)
		}
	}
	token := s.GenerateSecureToken()
	expiresAt := time.Now().Add(48 * time.Hour)
	newInvite := models.OrganizationInvite{
		ID:        uuid.New(),
		OrgID:     orgID,
		Email:     email,
		InvitedBy: invitedBy,
		Token:     token,
		ExpiresAt: expiresAt,
		Status:    "invited",
	}
	_, createOrgInviteErr := s.orgInviteRepo.CreateOrganizationInvite(ctx, newInvite)
	if createOrgInviteErr != nil {
		log.Printf("Error creating organization invite: %v", createOrgInviteErr)
		return false, fmt.Errorf("failed to create invitation")
	}
	inviteLink := fmt.Sprintf("%s/invite/accept?token=%s", configs.Configs.FRONTEND_URL, token)
	subject := fmt.Sprintf("You’re invited to join %s Organization", existingOrg.Name)
	body := fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
    <head>
        <title>Invitation</title>
    </head>
    <body>
        <p>Hello! You’ve been invited to join the <strong>%s</strong> organization on TaskFlow.</p>
        <p>Click the link below to accept the invitation:</p>
        <p><a href="%s">Accept Invitation</a></p>
        <p>If you don’t have an account yet, you’ll be asked to create one first.</p>
        <p>This link expires in 48 hours.</p>
    </body>
    </html>
`, existingOrg.Name, inviteLink)

	go s.SendEmail(subject, email, inviteLink, body, true)
	slog.Info("invitation sent", "org_id", orgID, "email", email, "token", token)
	return true, nil
}

func (s *organizationService) GenerateSecureToken() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

// PRODUCER
func (s *organizationService) SendEmail(subject, to, linkUrl, emailBody string, isHtml bool) {
	slog.Info("sending email", "to", to, "subject", subject)
	ch, err := s.queueConn.Channel()
	if err != nil {
		slog.Error("failed to open a channel" , "error" , err.Error())
		return
	}
	defer ch.Close()
	queue, err := ch.QueueDeclare(constants.EMAIL_QUEUE, true, false, false, false, nil)
	if err != nil {
		slog.Error("failed to declare queue" , "queue" , constants.EMAIL_QUEUE , "error" , err.Error())
		return
	}
	emailPayload := EmailMessage{
		Subject: subject,
		To:      to,
		IsHtml:  isHtml,
		LinkUrl: linkUrl,
		Body:    emailBody,
	}

	body, err := json.Marshal(emailPayload)
	if err != nil {
		slog.Error("failed to marshall email payload" , "error" , err.Error())
		return
	}
	err = ch.Publish("", queue.Name, false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent,
	})
	if err != nil {
		slog.Error("failed to publish messgae" , "error" , err.Error())
		return
	}
	slog.Info("Email Task enqueued succesfully")
}

func (s *organizationService) publishOrgCleanupEvent(conn *amqp.Connection, event consumers.OrgCleanupEvent) error {
	slog.Info("publishing org cleanup event", "org_id", event.OrgID)
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(constants.ORG_CLEANUP_QUEUE, true, false, false, false, nil)
	if err != nil {
		slog.Error("failed to declare queue" , "queue" , constants.ORG_CLEANUP_QUEUE , "error" , err.Error())
		return err
	}

	body, err := json.Marshal(event)
	if err != nil {
		slog.Error("failed to marshal body" , "error" , err.Error())
		return err
	}
	slog.Info("event is publishing ...")
	return ch.Publish(
		"",
		constants.ORG_CLEANUP_QUEUE,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
}

func (s *organizationService) AcceptInvitation(ctx context.Context, token string, userID uuid.UUID) (*models.OrganizationUser, error) {
	slog.Info("accepting invitation", "token", token, "user_id", userID)
	// Find the invite by token
	invite, err := s.orgInviteRepo.FindOrganizationInviteByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired invitation")
	}

	// Check if invite is expired
	if invite.IsExpired() {
		slog.Info("invite is expired" , "email" , invite.Email , "token" , invite.Token , "exipred_at"  ,invite.ExpiresAt)
		s.orgInviteRepo.DeleteOrganizationInvite(ctx, invite.ID)
		return nil, fmt.Errorf("invitation has expired")
	}

	// Check if invite is already accepted or revoked
	if invite.Status != "invited" {
		slog.Info("invalid invite status" , "status" , invite.Status , "email" , invite.Email , "token" , invite.Token)
		return nil, fmt.Errorf("invitation is already %s", invite.Status)
	}

	// Get the user by email
	user, err := s.authRepo.FindUserByEmail(ctx, invite.Email)
	if err != nil {
		slog.Info("user not found" , "email" , invite.Email)
		return nil, fmt.Errorf("user not found with this email. Please register first")
	}

	// Check if user is already a member
	existingMember, _ := s.orgRepo.FindOrganizationUser(ctx, invite.OrgID, user.ID)
	if existingMember != nil {
		slog.Info("user is already a member")
		// User is already a member, just update invite status
		s.orgInviteRepo.UpdateOrganizationInviteStatus(ctx, invite.ID, "accepted")
		return existingMember, fmt.Errorf("user is already a member of this organization")
	}

	// Create organization user
	orgUser := models.OrganizationUser{
		ID:       uuid.New(),
		OrgID:    invite.OrgID,
		UserID:   user.ID,
		Role:     "member", // Default role for invited users
		JoinedAt: time.Now().UTC(),
	}

	createdOrgUser, err := s.orgRepo.CreateOrganizationUser(ctx, orgUser)
	if err != nil {
		return nil, fmt.Errorf("failed to add user to organization: %w", err)
	}

	// Update invite status
	if err := s.orgInviteRepo.UpdateOrganizationInviteStatus(ctx, invite.ID, "accepted"); err != nil {
		slog.Warn("Warning: failed to update invite status to accepted")
	}

	// Send WebSocket notification to organization members
	if websocket.WsManager != nil {
		org, _ := s.orgRepo.FindOrganizationByID(ctx, invite.OrgID)
		payload := websocket.TeamMemberAddedPayload{
			TeamID:   invite.OrgID,
			TeamName: org.Name,
			UserID:   user.ID,
			UserName: user.FullName,
			Role:     "member",
			AddedBy:  invite.InvitedBy,
		}
		// Notify all org members
		members, _ := s.orgRepo.FindOrganizationMembers(ctx, invite.OrgID)
		for _, member := range members {
			websocket.WsManager.BroadcastToUser(member.UserID, websocket.EventTeamMemberAdded, payload)
		}
		slog.Info("org members notified successfully")
	}

	// Publish activity event
	event := models.ActivityEvent{
		EventName: events.ActivityEvents.AddMemberToTeam,
		UserID:    &user.ID,
		OrgID:     &invite.OrgID,
		Metadata: map[string]any{
			"invited_by": invite.InvitedBy,
			"role":       "member",
		},
	}
	s.activityProducer.Publish(event)

	return createdOrgUser, nil
}

func (s *organizationService) RejectInvitation(ctx context.Context, token string, userID uuid.UUID) error {
	slog.Info("rejecting invitation", "token", token, "user_id", userID)
	// Find the invite by token
	invite, err := s.orgInviteRepo.FindOrganizationInviteByToken(ctx, token)
	if err != nil {
		slog.Info("invalid invitation")
		return fmt.Errorf("invalid invitation")
	}

	slog.Info("updating invite status to revoked")
	// Update status to revoked
	return s.orgInviteRepo.UpdateOrganizationInviteStatus(ctx, invite.ID, "revoked")
}

func (s *organizationService) GetPendingInvitations(ctx context.Context, userID uuid.UUID) ([]*models.OrganizationInvite, error) {
	slog.Info("fetching pending invitations")
	// Get user by ID to get email
	user, err := s.authRepo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.orgInviteRepo.FindPendingInvitesByEmail(ctx, user.Email)
}
