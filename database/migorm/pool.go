package migorm

import (
	"fmt"
	"sync"
)

type migrationsPool struct {
	migrations map[string]any
	sync.Mutex
}

func (p *migrationsPool) Register(migrationName string, migration any) error {
	p.Lock()
	defer p.Unlock()

	switch migration.(type) {
	case NewMigration:
		break
	default:
		return fmt.Errorf("unknown migration type: %T, %s", migration, migrationName)
	}

	if migration == nil {
		return fmt.Errorf("migration '%s' cannot be nil", migrationName)
	}

	if p.migrations == nil {
		p.migrations = make(map[string]any)
	}

	_, ok := p.migrations[migrationName]
	if ok {
		return fmt.Errorf("migration with name '%s' already exist", migrationName)
	}

	p.migrations[migrationName] = migration

	return nil
}
