package migorm

type Config struct {
	Log           Logger
	MigrationsDir string
	TableName     string
}
