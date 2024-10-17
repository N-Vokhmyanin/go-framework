package migorm

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gorm.io/gorm"
	"os"
	"path"
	"runtime"
	"sort"
	"strings"
	"text/template"
	"time"
)

type Migrater interface {
	Logger(ctx context.Context) Logger
	UpMigrations(ctx context.Context) error
	UpConcreteMigration(ctx context.Context, name string) error
	DownConcreteMigration(ctx context.Context, name string) error
	MakeFileMigration(ctx context.Context, name string) error
}

func NewMigrater(db *gorm.DB, cfg *Config) Migrater {
	if cfg == nil {
		cfg = &Config{}
	}
	if cfg.Log == nil {
		cfg.Log = NewLogger()
	}
	if cfg.MigrationsDir == "" {
		cfg.MigrationsDir = "migrations"
	}
	if cfg.TableName == "" {
		cfg.TableName = DefaultMigrationsTableName
	}
	if db != nil {
		db = db.Scopes(WithCustomMigrationsTableName(cfg.TableName))
	}
	return &migrater{
		db:     db,
		Config: cfg,
	}
}

type migrater struct {
	db *gorm.DB
	*Config
}

func (m *migrater) Logger(_ context.Context) Logger {
	return m.Log
}

func (m *migrater) newContext(ctx context.Context) Context {
	return &migrationContext{
		Context: ctx,
		db: m.db.Session(&gorm.Session{
			Context: ctx,
			NewDB:   true,
		}).Unscoped().Begin(),
		log: m.Log,
	}
}

func (m *migrater) UpMigrations(ctx context.Context) error {

	m.Log.Infof("Start migrations")

	m.checkMigrationTable()

	newMigrations, err := m.getNewMigrations(ctx)
	if err != nil {
		return err
	}

	successCnt := 0
	for _, migration := range newMigrations {
		if migration.Id == 0 {
			err = m.doUp(ctx, migration)
			if err != nil {
				return err
			}
			successCnt += 1
		}
	}

	if successCnt > 0 {
		m.Log.Infof("All migrations are done success!")
	} else {
		m.Log.Infof("Nothing to migrate.")
	}

	return nil
}

func (m *migrater) doUp(ctx context.Context, model *migrationModel) (err error) {
	migCtx := m.newContext(ctx)
	defer func() {
		if err != nil {
			migCtx.DB().Rollback()
		} else {
			migCtx.DB().Commit()
		}
	}()

	migration := pool.migrations[model.Name]
	if migration == nil {
		return fmt.Errorf("migration '%s' not found", model.Name)
	}

	switch xMigration := migration.(type) {
	case NewMigration:
		err = xMigration.Up(migCtx)
	default:
		err = fmt.Errorf("migration '%s' failed: unknown type %T", model.Name, migration)
	}

	if err != nil {
		return fmt.Errorf("up migration '%s' error: %+v", model.Name, err)
	}

	if model.Id == 0 {
		err = m.db.Create(model).Error
		if err != nil {
			return fmt.Errorf("save migration '%s' error: %+v", model.Name, err)
		}
	}

	return nil
}

func (m *migrater) doDown(ctx context.Context, model *migrationModel) (err error) {
	migCtx := m.newContext(ctx)
	defer func() {
		if err != nil {
			migCtx.DB().Rollback()
		} else {
			migCtx.DB().Commit()
		}
	}()

	migration := pool.migrations[model.Name]
	if migration == nil {
		return fmt.Errorf("migration '%s' not found", model.Name)
	}

	switch xMigration := migration.(type) {
	case NewMigration:
		err = xMigration.Down(migCtx)
	default:
		err = fmt.Errorf("migration '%s' failed: unknown type %T", model.Name, migration)
	}

	if err != nil {
		return fmt.Errorf("down migration '%s' error, err: %+v", model.Name, err)
	}

	if model.Id != 0 {
		err = m.db.Delete(model).Error
		if err != nil {
			return fmt.Errorf("delete migration '%s' error: %+v", model.Name, err)
		}
	}

	return nil
}

func (m *migrater) UpConcreteMigration(ctx context.Context, name string) error {

	migration := pool.migrations[name]
	if migration == nil {
		return fmt.Errorf("migration '%s' not found", name)
	}

	model, err := m.findOrNewMigrationByName(ctx, name)
	if err != nil {
		return err
	}

	err = m.doUp(ctx, model)
	if err != nil {
		return err
	}

	return nil
}

func (m *migrater) DownConcreteMigration(ctx context.Context, name string) error {

	migration := pool.migrations[name]
	if migration == nil {
		return fmt.Errorf("migration '%s' not found", name)
	}

	model, err := m.findOrNewMigrationByName(ctx, name)
	if err != nil {
		return err
	}

	err = m.doDown(ctx, model)
	if err != nil {
		return err
	}

	return nil
}

func (m *migrater) MakeFileMigration(_ context.Context, name string) error {
	migrationsPath := m.Config.MigrationsDir

	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		m.Log.Infof("Create new directory : %v", migrationsPath)
		if err := os.MkdirAll(migrationsPath, os.ModePerm); err != nil {
			return err
		}
	}

	err := checkFileExists(migrationsPath, name+".go")
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	realName := fmt.Sprintf("%d_%s.go", now, name)

	migrationPath := migrationsPath + "/" + realName

	f, err := os.Create(migrationPath)
	if err != nil {
		return fmt.Errorf("create migration file: %v", err)
	}

	partsName := strings.Split(name, "_")
	structName := "migration"
	caser := cases.Title(language.Und, cases.NoLower)
	for _, p := range partsName {
		structName += caser.String(p)
	}

	partsDir := strings.Split(m.Config.MigrationsDir, "/")
	packageName := partsDir[len(partsDir)-1]

	tmpl, err := getTemplate()
	if err != nil {
		return err
	}
	err = tmpl.Execute(f, map[string]interface{}{"struct_name": structName, "package": packageName})

	if err != nil {
		return err
	}

	m.Log.Infof("migration file created: %v", realName)

	return nil
}

func (m *migrater) findOrNewMigrationByName(ctx context.Context, name string) (*migrationModel, error) {
	var model *migrationModel
	err := m.
		db.
		WithContext(ctx).
		Where("name = ?", name).
		First(&model).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	if err != nil {
		return nil, err
	}
	if model == nil {
		model = &migrationModel{
			Name: name,
		}
	}
	return model, nil
}

// Finds not yet completed migration files
func (m *migrater) getNewMigrations(ctx context.Context) ([]*migrationModel, error) {

	var names []string
	for k := range pool.migrations {
		names = append(names, k)
	}

	sort.Strings(names)

	result := make([]*migrationModel, 0)
	existMigrations := make([]*migrationModel, 0)

	err := m.
		db.
		WithContext(ctx).
		Model(migrationModel{}).
		Scan(&existMigrations).
		Error
	if err != nil {
		return nil, err
	}

	for _, name := range names {
		var exists bool
		for _, existing := range existMigrations {
			if existing.Name == name {
				exists = true
				break
			}
		}
		if !exists {
			model := &migrationModel{}
			model.Name = name
			result = append(result, model)
		}
	}

	return result, nil
}

// ***  helpers ***

// check or create table to register successful migrations
func (m *migrater) checkMigrationTable() {
	if !m.db.Migrator().HasTable(m.Config.TableName) {
		m.Log.Infof("Init table: %v", m.Config.TableName)
		err := m.db.AutoMigrate(&migrationModel{})
		if err != nil {
			panic(err)
		}
	}
}

// сheck the existence of a file in the directory with migrations
func checkFileExists(dir string, name string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, f := range files {
		split := strings.Split(f.Name(), "_")

		if name == strings.Join(split[1:], "_") {
			return fmt.Errorf("File %v already exists in dir: %v", name, dir)
		}
	}

	return nil
}

func getTemplate() (*template.Template, error) {

	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return nil, fmt.Errorf("Template caller")
	}

	tmpl, err := template.ParseFiles(path.Dir(filename) + "/../assets/template")
	if err != nil {
		return nil, fmt.Errorf("parse template : %v", err)
	}

	return tmpl, nil
}
