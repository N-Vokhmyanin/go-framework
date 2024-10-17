package migorm

import (
	"context"
	"fmt"
	"os"
)

func Run(ctx context.Context, migrater Migrater) {

	args := os.Args

	log := migrater.Logger(ctx)

	var err error
	if len(args) > 1 {

		switch args[1] {
		case "up":
			if len(args) != 3 {
				log.Errorf("Up command format must be: go run migrate up 00000000000_migation_name ")
				return
			}
			err = migrater.UpConcreteMigration(ctx, args[2])
		case "down":
			if len(args) != 3 {
				log.Errorf("Down command format must be: go run migrate down 00000000000_migation_name ")
				return
			}
			err = migrater.DownConcreteMigration(ctx, args[2])
		case "make":
			if len(args) != 3 {
				log.Errorf("Make command format must be: go run migrate.go make my_new_migration_name")
				return
			}
			err = migrater.MakeFileMigration(ctx, args[2])
		default:
			err = fmt.Errorf("Unknown command parameters: %+v", args[1:])
		}
	} else {
		err = migrater.UpMigrations(ctx)
	}

	if err != nil {
		log.Errorf(err.Error())
		return
	}
}
