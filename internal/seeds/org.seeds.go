package seeds

import (
	"context"
	"log"

	"github.com/amirrajj-dev/taskio/internal/data"
	"github.com/amirrajj-dev/taskio/internal/repositories"
)

func OrganizationSeeds() {
	log.Println("Starting organization seeding...")	
	for _ , org := range data.Organizations {
		repositories.OrgRepo.CreateOrganization(context.Background() , org)
	}
	log.Println("organizations seeding completed ✅")
	for _ , orgUser := range data.OrganizationUsers {
		repositories.OrgRepo.CreateOrganizationUser(context.Background() , orgUser)
	}
	log.Println("organization users seeding completed ✅")
}
