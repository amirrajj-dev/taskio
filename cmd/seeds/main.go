package main

import (
	"github.com/amirrajj-dev/taskio/internal/configs"
	"github.com/amirrajj-dev/taskio/internal/infrastructure/database"
	"github.com/amirrajj-dev/taskio/internal/repositories"
	"github.com/amirrajj-dev/taskio/internal/seeds"
)

func main() {
	configs.LoadConfig()
	database.ConnectToPostgresDb()
	defer database.DisconnectFromPostgresDb()
	repositories.NewOrganizationRepository()
	repositories.NewUserRepository()
	seeds.OrganizationSeeds()
	seeds.UserSeeds()
}
