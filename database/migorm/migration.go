package migorm

type NewMigration interface {
	Up(ctx Context) error
	Down(ctx Context) error
}
