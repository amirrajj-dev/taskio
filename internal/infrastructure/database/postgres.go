package database

import (
	"fmt"
	"log"

	"github.com/amirrajj-dev/taskio/internal/configs"
	"github.com/amirrajj-dev/taskio/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var PG *gorm.DB

func ConnectToPostgresDb() {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", configs.Configs.POSTGRES.POSTGRES_HOST, configs.Configs.POSTGRES.POSTGRES_PORT, configs.Configs.POSTGRES.POSTGRES_USER, configs.Configs.POSTGRES.POSTGRES_PASSWORD, configs.Configs.POSTGRES.POSTGRES_DB)
	db, connErr := gorm.Open(postgres.Open(dsn))
	if connErr != nil {
		log.Fatalf("failed to connect to postgres db : %v", connErr)
		return
	}
	sqlDB, sqlDbErr := db.DB()
	if sqlDbErr != nil {
		log.Fatalf("failed to get sqlDB for pinging : %v", sqlDbErr)
		return
	}
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("failed to ping postgres db : %v", err)
		return
	}
	log.Println("connected to Postgres DB succesfully")
	if err := db.AutoMigrate(&models.ActivityLog{},
		&models.Comment{}, &models.Organization{}, &models.OrganizationUser{}, &models.OrganizationInvite{} , &models.Project{}, &models.RefreshToken{}, &models.Task{}, &models.Team{}, &models.TeamMember{}, &models.User{}); err != nil {
		log.Fatalf("failed to migrate models : %v", err)
	}
	PG = db
}

func DisconnectFromPostgresDb() {
	if PG == nil {
		log.Println("already disconnected from postgres db")
		return
	}
	sqlDB, sqlDbErr := PG.DB()
	if sqlDbErr != nil {
		log.Fatalf("failed to get sqlDB for closing : %v", sqlDbErr)
		return
	}
	if err := sqlDB.Close(); err != nil {
		log.Println("failed to disconnect from postgres db")
	} else {
		log.Println("disconnected from postgres db successfully")
	}
}
