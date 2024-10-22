package migorm

import (
	"context"
	"fmt"
	"os"
)

func Run(ctx context.Context, migrater Migrater) {
	args := os.Args

	log := migrater.Logger(ctx)

	if len(args) == 0 {
		log.Errorf(migrater.UpMigrations(ctx).Error())
		return
	}

	switch args[1] {
	case "up":
		if len(args) != 3 {
			log.Errorf("Up command format must be: go run migrate up 00000000000_migation_name ")
			return
		}
		log.Errorf(migrater.UpConcreteMigration(ctx, args[2]).Error())
	case "down":
		if len(args) != 3 {
			log.Errorf("Down command format must be: go run migrate down 00000000000_migation_name ")
			return
		}
		log.Errorf(migrater.DownConcreteMigration(ctx, args[2]).Error())
	case "make":
		if len(args) != 3 {
			log.Errorf("Make command format must be: go run migrate.go make my_new_migration_name")
			return
		}
		log.Errorf(migrater.MakeFileMigration(ctx, args[2]).Error())
	default:
		log.Errorf(fmt.Sprintf("Unknown command parameters: %+v", args[1:]))
	}
}
