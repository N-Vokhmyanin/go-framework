package migorm

import (
	"path/filepath"
	"runtime"
	"strings"
)

var pool migrationsPool

// RegisterMigration Each migration file call this method in its init method
func RegisterMigration(migration any) {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		panic("cant get migration filename: fail invoke caller")
	}

	migrationName := strings.Replace(filepath.Base(file), ".go", "", -1)

	err := pool.Register(migrationName, migration)
	if err != nil {
		panic(err)
	}
}
