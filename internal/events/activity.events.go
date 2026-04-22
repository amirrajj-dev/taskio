package events

type ActivityEvent struct {
	LogOut                   string
	CreateComment            string
	UpdateOrg                string
	DeleteOrg                string
	OrgDeletedForCleanup     string
	UpdateOrgUserRole        string
	CreateProject            string
	UpdateProject            string
	DeleteProject            string
	ProjectDeletedForCleanup string
	CreateTask               string
	CreateSubTask            string
	UpdateTask               string
	DeleteTask               string
	CreateTeam               string
	TaskDeletedForCleanUp    string
	AddMemberToTeam          string
	DeleteMemberFromTeam     string
	ChangeTeamMemberRole     string
	TeamDeletedForCleanup    string
}

var ActivityEvents = ActivityEvent{
	LogOut:                   "user.logout",
	CreateComment:            "comment.created",
	UpdateOrg:                "org.updated",
	DeleteOrg:                "org.deleted",
	OrgDeletedForCleanup:     "org.deleted.for.cleanup",
	UpdateOrgUserRole:        "org.user.role.updated",
	CreateProject:            "project.created",
	UpdateProject:            "project.updated",
	DeleteProject:            "project.deleted",
	ProjectDeletedForCleanup: "project.deleted.for.cleanup",
	CreateTask:               "task.created",
	CreateSubTask:            "task.subtask.created",
	UpdateTask:               "task.updated",
	DeleteTask:               "task.deleted",
	TaskDeletedForCleanUp:    "task.deleted.for.cleanup",
	CreateTeam:               "team.created",
	AddMemberToTeam:          "team.member.added",
	DeleteMemberFromTeam:     "team.member.deleted",
	ChangeTeamMemberRole:     "team.member.role.changed",
	TeamDeletedForCleanup:    "team.deleted.for.cleanup",
}
