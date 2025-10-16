package seeder

import (
	"context"
	"fmt"
	"time"

	"errors"

	"github.com/samaasi/uptime-application/services/api-services/internal/api/models"
	"github.com/samaasi/uptime-application/services/api-services/pkg/logger"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Seeder defines the interface for a seeder
type Seeder interface {
	Name() string
	Dependencies() []string
	Seed(context.Context) error
}

// SeedError represents a custom error type for seeding operations
type SeedError struct {
	EntityType string
	EntityName string
	Operation  string
	Err        error
}

func (e *SeedError) Error() string {
	return fmt.Sprintf("seeding %s '%s' failed during %s: %v",
		e.EntityType, e.EntityName, e.Operation, e.Err)
}

// Unwrap returns the wrapped error
func (e *SeedError) Unwrap() error {
	return e.Err
}

// SeederOption defines a functional option for configuring a GenericSeeder.
type SeederOption[T any] func(*GenericSeeder[T])

// WithBatchSize sets the batch size for seeding operations.
func WithBatchSize[T any](size int) SeederOption[T] {
	return func(gs *GenericSeeder[T]) {
		if size > 0 {
			gs.batchSize = size
		}
	}
}

// WithConflictColumns sets the conflict columns for upsert operations.
func WithConflictColumns[T any](columns ...string) SeederOption[T] {
	return func(gs *GenericSeeder[T]) {
		if len(columns) > 0 {
			gs.conflictColumns = columns
		}
	}
}

// WithValidator sets the validation function for entities.
func WithValidator[T any](validator func(T) error) SeederOption[T] {
	return func(gs *GenericSeeder[T]) {
		gs.validator = validator
	}
}

// GenericSeeder provides a generic seeding implementation
type GenericSeeder[T any] struct {
	db              *gorm.DB
	entityName      string
	batchSize       int
	validator       func(T) error
	conflictColumns []string
}

// NewGenericSeeder creates a new generic seeder
func NewGenericSeeder[T any](db *gorm.DB, entityName string, options ...SeederOption[T]) *GenericSeeder[T] {
	gs := &GenericSeeder[T]{
		db:              db,
		entityName:      entityName,
		batchSize:       100,
		conflictColumns: []string{"name"},
	}

	for _, option := range options {
		option(gs)
	}

	return gs
}

// SeedEntities performs generic seeding with batch operations
func (gs *GenericSeeder[T]) SeedEntities(ctx context.Context, entities []T) error {
	if gs.validator != nil {
		for i, entity := range entities {
			if err := gs.validator(entity); err != nil {
				return &SeedError{
					EntityType: gs.entityName,
					EntityName: fmt.Sprintf("index_%d", i),
					Operation:  "validation",
					Err:        err,
				}
			}
		}
	}

	if len(entities) == 0 {
		logger.Info(fmt.Sprintf("No %s to seed", gs.entityName))
		return nil
	}

	logger.Info(fmt.Sprintf("Starting %s seeding process", gs.entityName), logger.Int("count", len(entities)))

	return gs.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		conflictColumns := make([]clause.Column, len(gs.conflictColumns))
		for i, col := range gs.conflictColumns {
			conflictColumns[i] = clause.Column{Name: col}
		}

		result := tx.Clauses(clause.OnConflict{
			Columns:   conflictColumns,
			DoNothing: true,
		}).CreateInBatches(entities, gs.batchSize)

		if result.Error != nil {
			return &SeedError{
				EntityType: gs.entityName,
				EntityName: "batch",
				Operation:  "upsert",
				Err:        result.Error,
			}
		}

		logger.Info(fmt.Sprintf("Successfully processed %s", gs.entityName),
			logger.Int64("affected_rows", result.RowsAffected),
			logger.Int("total_entities", len(entities)))

		return nil
	})
}

// PermissionSeeder handles seeding of permissions into the database
type PermissionSeeder struct {
	*GenericSeeder[models.Permission]
	config *SeedConfig
}

// NewPermissionSeeder creates a new instance of PermissionSeeder
func NewPermissionSeeder(db *gorm.DB, config *SeedConfig) Seeder {
	return &PermissionSeeder{
		GenericSeeder: NewGenericSeeder(db, "permissions",
			WithConflictColumns[models.Permission]("name"),
			WithValidator[models.Permission](validatePermission)),
		config: config,
	}
}

// Name returns the name of the seeder.
func (ps *PermissionSeeder) Name() string {
	return "permissions"
}

// Dependencies returns the dependencies of the seeder.
func (ps *PermissionSeeder) Dependencies() []string {
	return []string{}
}

// getPermissions returns the permissions from configuration
func (ps *PermissionSeeder) getPermissions() []models.Permission {
	permissions := make([]models.Permission, len(ps.config.Permissions))
	for i, permConfig := range ps.config.Permissions {
		permissions[i] = permConfig.ToModel()
	}
	return permissions
}

// validatePermission validates a Permission entity
func validatePermission(p models.Permission) error {
	if p.Name == "" {
		return errors.New("permission name cannot be empty")
	}
	if len(p.Name) > 50 {
		return errors.New("permission name cannot exceed 50 characters")
	}
	return nil
}

// Seed seeds the predefined permissions into the database
func (ps *PermissionSeeder) Seed(ctx context.Context) error {
	permissions := ps.getPermissions()
	return ps.GenericSeeder.SeedEntities(ctx, permissions)
}

// OrganizationTypeSeeder handles seeding of organization types into the database
type OrganizationTypeSeeder struct {
	*GenericSeeder[models.OrganizationType]
	config *SeedConfig
}

// NewOrganizationTypeSeeder creates a new instance of OrganizationTypeSeeder
func NewOrganizationTypeSeeder(db *gorm.DB, config *SeedConfig) Seeder {
	return &OrganizationTypeSeeder{
		GenericSeeder: NewGenericSeeder(db, "organization_types",
			WithConflictColumns[models.OrganizationType]("name"),
			WithValidator[models.OrganizationType](validateOrganizationType)),
		config: config,
	}
}

// Name returns the name of the seeder.
func (ots *OrganizationTypeSeeder) Name() string {
	return "organization_types"
}

// Dependencies returns the dependencies of the seeder.
func (ots *OrganizationTypeSeeder) Dependencies() []string {
	return []string{}
}

// getOrganizationTypes returns the organization types from configuration
func (ots *OrganizationTypeSeeder) getOrganizationTypes() []models.OrganizationType {
	organizationTypes := make([]models.OrganizationType, len(ots.config.OrganizationTypes))
	for i, orgTypeConfig := range ots.config.OrganizationTypes {
		organizationTypes[i] = orgTypeConfig.ToModel()
	}
	return organizationTypes
}

// validateOrganizationType validates an OrganizationType entity
func validateOrganizationType(ot models.OrganizationType) error {
	if ot.Name == "" {
		return errors.New("organization type name cannot be empty")
	}
	if len(ot.Name) > 100 {
		return errors.New("organization type name cannot exceed 100 characters")
	}
	return nil
}

// Seed seeds the predefined organization types into the database
func (ots *OrganizationTypeSeeder) Seed(ctx context.Context) error {
	organizationTypes := ots.getOrganizationTypes()
	return ots.GenericSeeder.SeedEntities(ctx, organizationTypes)
}

// ApplicationTypeSeeder handles seeding of application types into the database
type ApplicationTypeSeeder struct {
	*GenericSeeder[models.ApplicationType]
	config *SeedConfig
}

// NewApplicationTypeSeeder creates a new instance of ApplicationTypeSeeder
func NewApplicationTypeSeeder(db *gorm.DB, config *SeedConfig) Seeder {
	return &ApplicationTypeSeeder{
		GenericSeeder: NewGenericSeeder(db, "application_types",
			WithConflictColumns[models.ApplicationType]("name"),
			WithValidator[models.ApplicationType](validateApplicationType)),
		config: config,
	}
}

// Name returns the name of the seeder.
func (ats *ApplicationTypeSeeder) Name() string {
	return "application_types"
}

// Dependencies returns the dependencies of the seeder.
func (ats *ApplicationTypeSeeder) Dependencies() []string {
	return []string{}
}

// getApplicationTypes returns the application types from configuration
func (ats *ApplicationTypeSeeder) getApplicationTypes() []models.ApplicationType {
	applicationTypes := make([]models.ApplicationType, len(ats.config.ApplicationTypes))
	for i, appTypeConfig := range ats.config.ApplicationTypes {
		applicationTypes[i] = appTypeConfig.ToModel()
	}
	return applicationTypes
}

// validateApplicationType validates an ApplicationType entity
func validateApplicationType(at models.ApplicationType) error {
	if at.Name == "" {
		return errors.New("application type name cannot be empty")
	}
	if len(at.Name) > 100 {
		return errors.New("application type name cannot exceed 100 characters")
	}
	return nil
}

// Seed seeds the predefined application types into the database
func (ats *ApplicationTypeSeeder) Seed(ctx context.Context) error {
	applicationTypes := ats.getApplicationTypes()
	return ats.GenericSeeder.SeedEntities(ctx, applicationTypes)
}

// SeedManager manages seeding operations with dependency resolution
type SeedManager struct {
	db      *gorm.DB
	config  *SeedConfig
	seeders map[string]Seeder
}

// NewSeedManager creates a new seed manager with configuration
func NewSeedManager(db *gorm.DB, config *SeedConfig) *SeedManager {
	sm := &SeedManager{
		db:      db,
		config:  config,
		seeders: make(map[string]Seeder),
	}

	sm.Register(NewPermissionSeeder(db, config))
	sm.Register(NewOrganizationTypeSeeder(db, config))
	sm.Register(NewApplicationTypeSeeder(db, config))

	return sm
}

// Register registers a seeder with the manager
func (sm *SeedManager) Register(seeder Seeder) {
	sm.seeders[seeder.Name()] = seeder
}

// resolveDependencies performs topological sort to resolve seeding order
func (sm *SeedManager) resolveDependencies() ([]Seeder, error) {
	visited := make(map[string]bool)
	visiting := make(map[string]bool)
	result := make([]Seeder, 0, len(sm.seeders))

	var visit func(string) error
	visit = func(name string) error {
		if visiting[name] {
			return fmt.Errorf("circular dependency detected involving %s", name)
		}
		if visited[name] {
			return nil
		}

		seeder, exists := sm.seeders[name]
		if !exists {
			return fmt.Errorf("seeder dependency '%s' not found/registered", name)
		}

		visiting[name] = true

		for _, depName := range seeder.Dependencies() {
			if err := visit(depName); err != nil {
				return err
			}
		}

		visiting[name] = false
		visited[name] = true
		result = append(result, seeder)
		return nil
	}

	for name := range sm.seeders {
		if err := visit(name); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// SeedWithDependencies executes all seeders in dependency order
func (sm *SeedManager) SeedWithDependencies(ctx context.Context) error {
	orderedSeeders, err := sm.resolveDependencies()
	if err != nil {
		return fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	logger.Info("Starting dependency-ordered seeding", logger.Int("total_seeders", len(orderedSeeders)))

	for _, seeder := range orderedSeeders {
		logger.Info(fmt.Sprintf("Executing seeder: %s", seeder.Name()))
		if err := seeder.Seed(ctx); err != nil {
			return &SeedError{
				EntityType: seeder.Name(),
				EntityName: "seeder",
				Operation:  "execution",
				Err:        err,
			}
		}
	}

	logger.Info("All seeders completed successfully")
	return nil
}

// withRetry executes a function with exponential backoff retry
func withRetry(ctx context.Context, maxRetries int, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 100ms, 200ms, 400ms, etc.
			backoff := time.Duration(100*attempt*attempt) * time.Millisecond
			logger.Info("Retrying operation", logger.Int("attempt", attempt), logger.Duration("backoff", backoff))

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		if err := fn(); err != nil {
			lastErr = err
			logger.Warn("Operation failed", logger.Int("attempt", attempt), logger.ErrorField(err))
			continue
		}

		return nil
	}

	return fmt.Errorf("operation failed after %d attempts: %w", maxRetries, lastErr)
}

// SeedDefaultData seeds all default data with dependency management and retry logic
func SeedDefaultData(ctx context.Context, db *gorm.DB) error {
	return SeedDefaultDataWithConfig(ctx, db, "")
}

// SeedDefaultDataWithConfig seeds all default data with custom configuration path
func SeedDefaultDataWithConfig(ctx context.Context, db *gorm.DB, configPath string) error {
	logger.Info("Starting default data seeding with enhanced error handling")

	// Use default config path if not provided
	if configPath == "" {
		configPath = "configs/seed_config.yaml"
	}

	config, err := LoadSeedConfig(configPath)
	if err != nil {
		logger.Warn("Failed to load seed config, using defaults", logger.ErrorField(err))
		config, err = getDefaultSeedConfig()
		if err != nil {
			return fmt.Errorf("failed to load default seed config: %w", err)
		}
	}

	seedManager := NewSeedManager(db, config)
	seedManager.Register(NewPermissionSeeder(db, config))
	seedManager.Register(NewOrganizationTypeSeeder(db, config))
	seedManager.Register(NewApplicationTypeSeeder(db, config))

	return withRetry(ctx, 3, func() error {
		return seedManager.SeedWithDependencies(ctx)
	})
}
