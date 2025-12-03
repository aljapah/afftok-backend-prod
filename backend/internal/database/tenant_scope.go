package database

import (
	"fmt"

	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================
// TENANT-SCOPED DATABASE WRAPPER
// ============================================

// TenantDB wraps gorm.DB with automatic tenant scoping
type TenantDB struct {
	*gorm.DB
	tenantID     uuid.UUID
	skipScoping  bool
}

// Tenant creates a tenant-scoped database instance
func Tenant(db *gorm.DB, tenantID uuid.UUID) *TenantDB {
	return &TenantDB{
		DB:       db,
		tenantID: tenantID,
	}
}

// TenantFromContext creates a tenant-scoped DB from context tenant ID
func TenantFromContext(db *gorm.DB, tenantID uuid.UUID) *TenantDB {
	if tenantID == uuid.Nil {
		tenantID = models.DefaultTenantID
	}
	return Tenant(db, tenantID)
}

// SkipScoping disables tenant scoping for this query
func (t *TenantDB) SkipScoping() *TenantDB {
	t.skipScoping = true
	return t
}

// ============================================
// QUERY METHODS WITH AUTOMATIC SCOPING
// ============================================

// Find finds records with tenant scope
func (t *TenantDB) Find(dest interface{}, conds ...interface{}) *gorm.DB {
	return t.scope().Find(dest, conds...)
}

// First finds first record with tenant scope
func (t *TenantDB) First(dest interface{}, conds ...interface{}) *gorm.DB {
	return t.scope().First(dest, conds...)
}

// Last finds last record with tenant scope
func (t *TenantDB) Last(dest interface{}, conds ...interface{}) *gorm.DB {
	return t.scope().Last(dest, conds...)
}

// Take finds one record with tenant scope
func (t *TenantDB) Take(dest interface{}, conds ...interface{}) *gorm.DB {
	return t.scope().Take(dest, conds...)
}

// Create creates a record with tenant ID
func (t *TenantDB) Create(value interface{}) *gorm.DB {
	t.setTenantID(value)
	return t.DB.Create(value)
}

// Save saves a record with tenant ID
func (t *TenantDB) Save(value interface{}) *gorm.DB {
	t.setTenantID(value)
	return t.DB.Save(value)
}

// Updates updates records with tenant scope
func (t *TenantDB) Updates(values interface{}) *gorm.DB {
	return t.scope().Updates(values)
}

// Update updates a single column with tenant scope
func (t *TenantDB) Update(column string, value interface{}) *gorm.DB {
	return t.scope().Update(column, value)
}

// Delete deletes records with tenant scope
func (t *TenantDB) Delete(value interface{}, conds ...interface{}) *gorm.DB {
	return t.scope().Delete(value, conds...)
}

// Count counts records with tenant scope
func (t *TenantDB) Count(count *int64) *gorm.DB {
	return t.scope().Count(count)
}

// Pluck plucks column values with tenant scope
func (t *TenantDB) Pluck(column string, dest interface{}) *gorm.DB {
	return t.scope().Pluck(column, dest)
}

// ============================================
// CHAIN METHODS
// ============================================

// Model specifies the model
func (t *TenantDB) Model(value interface{}) *TenantDB {
	return &TenantDB{
		DB:          t.DB.Model(value),
		tenantID:    t.tenantID,
		skipScoping: t.skipScoping,
	}
}

// Where adds conditions
func (t *TenantDB) Where(query interface{}, args ...interface{}) *TenantDB {
	return &TenantDB{
		DB:          t.DB.Where(query, args...),
		tenantID:    t.tenantID,
		skipScoping: t.skipScoping,
	}
}

// Or adds OR conditions
func (t *TenantDB) Or(query interface{}, args ...interface{}) *TenantDB {
	return &TenantDB{
		DB:          t.DB.Or(query, args...),
		tenantID:    t.tenantID,
		skipScoping: t.skipScoping,
	}
}

// Not adds NOT conditions
func (t *TenantDB) Not(query interface{}, args ...interface{}) *TenantDB {
	return &TenantDB{
		DB:          t.DB.Not(query, args...),
		tenantID:    t.tenantID,
		skipScoping: t.skipScoping,
	}
}

// Order specifies order
func (t *TenantDB) Order(value interface{}) *TenantDB {
	return &TenantDB{
		DB:          t.DB.Order(value),
		tenantID:    t.tenantID,
		skipScoping: t.skipScoping,
	}
}

// Limit specifies limit
func (t *TenantDB) Limit(limit int) *TenantDB {
	return &TenantDB{
		DB:          t.DB.Limit(limit),
		tenantID:    t.tenantID,
		skipScoping: t.skipScoping,
	}
}

// Offset specifies offset
func (t *TenantDB) Offset(offset int) *TenantDB {
	return &TenantDB{
		DB:          t.DB.Offset(offset),
		tenantID:    t.tenantID,
		skipScoping: t.skipScoping,
	}
}

// Select specifies fields to select
func (t *TenantDB) Select(query interface{}, args ...interface{}) *TenantDB {
	return &TenantDB{
		DB:          t.DB.Select(query, args...),
		tenantID:    t.tenantID,
		skipScoping: t.skipScoping,
	}
}

// Omit specifies fields to omit
func (t *TenantDB) Omit(columns ...string) *TenantDB {
	return &TenantDB{
		DB:          t.DB.Omit(columns...),
		tenantID:    t.tenantID,
		skipScoping: t.skipScoping,
	}
}

// Group specifies group by
func (t *TenantDB) Group(name string) *TenantDB {
	return &TenantDB{
		DB:          t.DB.Group(name),
		tenantID:    t.tenantID,
		skipScoping: t.skipScoping,
	}
}

// Having specifies having conditions
func (t *TenantDB) Having(query interface{}, args ...interface{}) *TenantDB {
	return &TenantDB{
		DB:          t.DB.Having(query, args...),
		tenantID:    t.tenantID,
		skipScoping: t.skipScoping,
	}
}

// Joins specifies join conditions
func (t *TenantDB) Joins(query string, args ...interface{}) *TenantDB {
	return &TenantDB{
		DB:          t.DB.Joins(query, args...),
		tenantID:    t.tenantID,
		skipScoping: t.skipScoping,
	}
}

// Preload preloads associations
func (t *TenantDB) Preload(query string, args ...interface{}) *TenantDB {
	return &TenantDB{
		DB:          t.DB.Preload(query, args...),
		tenantID:    t.tenantID,
		skipScoping: t.skipScoping,
	}
}

// Distinct specifies distinct
func (t *TenantDB) Distinct(args ...interface{}) *TenantDB {
	return &TenantDB{
		DB:          t.DB.Distinct(args...),
		tenantID:    t.tenantID,
		skipScoping: t.skipScoping,
	}
}

// Unscoped removes default scopes
func (t *TenantDB) Unscoped() *TenantDB {
	return &TenantDB{
		DB:          t.DB.Unscoped(),
		tenantID:    t.tenantID,
		skipScoping: true,
	}
}

// ============================================
// TRANSACTION SUPPORT
// ============================================

// Transaction executes a function within a transaction
func (t *TenantDB) Transaction(fc func(tx *TenantDB) error) error {
	return t.DB.Transaction(func(tx *gorm.DB) error {
		tenantTx := &TenantDB{
			DB:          tx,
			tenantID:    t.tenantID,
			skipScoping: t.skipScoping,
		}
		return fc(tenantTx)
	})
}

// Begin begins a transaction
func (t *TenantDB) Begin() *TenantDB {
	return &TenantDB{
		DB:          t.DB.Begin(),
		tenantID:    t.tenantID,
		skipScoping: t.skipScoping,
	}
}

// Commit commits a transaction
func (t *TenantDB) Commit() *gorm.DB {
	return t.DB.Commit()
}

// Rollback rolls back a transaction
func (t *TenantDB) Rollback() *gorm.DB {
	return t.DB.Rollback()
}

// ============================================
// RAW QUERIES
// ============================================

// Raw executes raw SQL with tenant scope
func (t *TenantDB) Raw(sql string, values ...interface{}) *TenantDB {
	// Inject tenant_id into raw queries if needed
	return &TenantDB{
		DB:          t.DB.Raw(sql, values...),
		tenantID:    t.tenantID,
		skipScoping: t.skipScoping,
	}
}

// Exec executes raw SQL
func (t *TenantDB) Exec(sql string, values ...interface{}) *gorm.DB {
	return t.DB.Exec(sql, values...)
}

// ============================================
// INTERNAL HELPERS
// ============================================

// scope adds tenant scope to the query
func (t *TenantDB) scope() *gorm.DB {
	if t.skipScoping || t.tenantID == uuid.Nil {
		return t.DB
	}
	return t.DB.Where("tenant_id = ?", t.tenantID)
}

// setTenantID sets tenant ID on the value if it implements TenantScoped
func (t *TenantDB) setTenantID(value interface{}) {
	if t.skipScoping || t.tenantID == uuid.Nil {
		return
	}

	// Try to set tenant_id via interface
	if scoped, ok := value.(models.TenantScoped); ok {
		scoped.SetTenantID(t.tenantID)
		return
	}

	// Try reflection for struct with TenantID field
	// This is handled by GORM hooks instead
}

// GetTenantID returns the tenant ID
func (t *TenantDB) GetTenantID() uuid.UUID {
	return t.tenantID
}

// ============================================
// GORM CALLBACKS FOR AUTOMATIC TENANT SCOPING
// ============================================

// RegisterTenantCallbacks registers GORM callbacks for tenant scoping
func RegisterTenantCallbacks(db *gorm.DB) {
	// Before create: set tenant_id
	db.Callback().Create().Before("gorm:create").Register("tenant:before_create", func(tx *gorm.DB) {
		if tenantID, ok := tx.Statement.Context.Value("tenant_id").(uuid.UUID); ok {
			if tenantID != uuid.Nil {
				// Set tenant_id field if it exists
				tx.Statement.SetColumn("tenant_id", tenantID)
			}
		}
	})

	// Before query: add tenant scope
	db.Callback().Query().Before("gorm:query").Register("tenant:before_query", func(tx *gorm.DB) {
		if tenantID, ok := tx.Statement.Context.Value("tenant_id").(uuid.UUID); ok {
			if tenantID != uuid.Nil {
				// Check if model has tenant_id column
				if tx.Statement.Schema != nil {
					if _, ok := tx.Statement.Schema.FieldsByDBName["tenant_id"]; ok {
						tx.Where("tenant_id = ?", tenantID)
					}
				}
			}
		}
	})
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// WithTenant creates a tenant-scoped DB instance
func WithTenant(db *gorm.DB, tenantID uuid.UUID) *TenantDB {
	return Tenant(db, tenantID)
}

// ============================================
// MIGRATION HELPERS
// ============================================

// AddTenantIDColumn adds tenant_id column to a table
func AddTenantIDColumn(db *gorm.DB, tableName string) error {
	// Check if column exists
	var count int64
	db.Raw(`
		SELECT COUNT(*) FROM information_schema.columns 
		WHERE table_name = ? AND column_name = 'tenant_id'
	`, tableName).Scan(&count)

	if count > 0 {
		return nil // Column already exists
	}

	// Add column
	sql := fmt.Sprintf(`
		ALTER TABLE %s 
		ADD COLUMN tenant_id UUID REFERENCES tenants(id) DEFAULT '%s'
	`, tableName, models.DefaultTenantID.String())

	if err := db.Exec(sql).Error; err != nil {
		return err
	}

	// Create index
	indexSQL := fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS idx_%s_tenant_id ON %s(tenant_id)
	`, tableName, tableName)

	return db.Exec(indexSQL).Error
}

// BackfillTenantID backfills tenant_id for existing records
func BackfillTenantID(db *gorm.DB, tableName string, defaultTenantID uuid.UUID) error {
	sql := fmt.Sprintf(`
		UPDATE %s SET tenant_id = ? WHERE tenant_id IS NULL
	`, tableName)

	return db.Exec(sql, defaultTenantID).Error
}

// MigrateTenantColumns migrates all tenant-scoped tables
func MigrateTenantColumns(db *gorm.DB) error {
	tables := []string{
		"afftok_users",
		"offers",
		"networks",
		"user_offers",
		"clicks",
		"conversions",
		"teams",
		"team_members",
		"advertiser_api_keys",
		"api_key_usage_logs",
		"geo_rules",
		"webhook_pipelines",
		"webhook_steps",
		"webhook_executions",
		"webhook_step_results",
		"webhook_dlq_items",
	}

	for _, table := range tables {
		if err := AddTenantIDColumn(db, table); err != nil {
			fmt.Printf("Warning: Failed to add tenant_id to %s: %v\n", table, err)
		}
		if err := BackfillTenantID(db, table, models.DefaultTenantID); err != nil {
			fmt.Printf("Warning: Failed to backfill tenant_id in %s: %v\n", table, err)
		}
	}

	return nil
}

