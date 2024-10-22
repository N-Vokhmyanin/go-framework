package database

import (
	"errors"
	"github.com/N-Vokhmyanin/go-framework/contracts"
	"github.com/N-Vokhmyanin/go-framework/database/migorm"
	"github.com/urfave/cli/v2"
	"regexp"
)

var migrationNamePattern = regexp.MustCompile(`^[a-z0-9_]+$`)

//goland:noinspection SpellCheckingInspection
type migrateCommand struct {
	migrater migorm.Migrater
}

func NewMigrateCommands(app contracts.Application, mig migorm.Migrater) []*cli.Command {
	cmd := &migrateCommand{
		migrater: mig,
	}

	initService := func(ctx *cli.Context) error {
		app.InitService()
		return nil
	}

	return []*cli.Command{
		{
			Category: "migrate",
			Name:     "migrate:all",
			Usage:    "Run all migrations",
			Before:   initService,
			Action:   cmd.exitOnError(cmd.migrateAll),
		},
		{
			Category: "migrate",
			Name:     "migrate:up",
			Usage:    "Run concrete migration",
			Before:   initService,
			Action:   cmd.exitOnError(cmd.migrateUp),
		},
		{
			Category: "migrate",
			Name:     "migrate:down",
			Usage:    "Down concrete migration",
			Before:   initService,
			Action:   cmd.exitOnError(cmd.migrateDown),
		},
		{
			Category: "migrate",
			Name:     "migrate:make",
			Usage:    "Make new migration",
			Action:   cmd.exitOnError(cmd.migrateMake),
		},
	}
}

func (c *migrateCommand) exitOnError(fn func(*cli.Context) error) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		err := fn(ctx)
		if err != nil {
			return cli.Exit(err, 1)
		}

		return nil
	}
}

func (c *migrateCommand) migrateAll(ctx *cli.Context) error {
	return c.migrater.UpMigrations(ctx.Context)
}

func (c *migrateCommand) migrateUp(ctx *cli.Context) error {
	name, err := c.getMigrationName(ctx)
	if err != nil {
		return err
	}

	return c.migrater.UpConcreteMigration(ctx.Context, name)
}

func (c *migrateCommand) migrateDown(ctx *cli.Context) error {
	name, err := c.getMigrationName(ctx)
	if err != nil {
		return err
	}

	return c.migrater.DownConcreteMigration(ctx.Context, name)
}

func (c *migrateCommand) migrateMake(ctx *cli.Context) error {
	name, err := c.getMigrationName(ctx)
	if err != nil {
		return err
	}

	return c.migrater.MakeFileMigration(ctx.Context, name)
}

func (c *migrateCommand) getMigrationName(ctx *cli.Context) (name string, err error) {
	if name = ctx.Args().First(); name == "" {
		return name, errors.New("migration name required")
	}
	if !migrationNamePattern.MatchString(name) {
		return name, errors.New("migration name must match " + migrationNamePattern.String())
	}

	return name, nil
}
