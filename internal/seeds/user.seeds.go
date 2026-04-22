package seeds

import (
	"context"
	"log"

	"github.com/amirrajj-dev/taskio/internal/data"
	"github.com/amirrajj-dev/taskio/internal/repositories"
)

func UserSeeds() {
	log.Println("Starting users seeding...")
	for _, user := range data.Users {
		user.HashPassword(user.Password)
		repositories.UserRepo.CreateUser(context.Background(), user)
	}
	log.Println("users seeding completed ✅")
}
