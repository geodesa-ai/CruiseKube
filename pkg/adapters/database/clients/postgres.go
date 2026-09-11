package clients

import (
	"fmt"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PostgreSQLClientFactory creates PostgreSQL database clients
type PostgreSQLClientFactory struct {
	config FactoryConfig
}

// NewPostgreSQLClientFactory creates a new PostgreSQL client factory
func NewPostgreSQLClientFactory(config FactoryConfig) *PostgreSQLClientFactory {
	return &PostgreSQLClientFactory{config: config}
}

// buildDSN validates and constructs the libpq connection string for the
// given config. Exported at the package level (unexported name, but a
// free function rather than a method) so its validation and precedence
// rules are directly testable without opening a real connection.
func buildDSN(config FactoryConfig) (string, error) {
	// Set defaults
	if config.Host == "" {
		config.Host = "localhost"
	}
	if config.Port == 0 {
		config.Port = 5432
	}
	if config.SSLMode == "" {
		config.SSLMode = "disable"
	}

	usingClientCert := config.SSLCert != "" || config.SSLKey != ""
	if usingClientCert && (config.SSLCert == "" || config.SSLKey == "") {
		return "", fmt.Errorf("client-certificate authentication requires both SSLCert and SSLKey to be set")
	}
	if !usingClientCert && config.Password == "" {
		return "", fmt.Errorf("PostgreSQL connection requires either Password or SSLCert/SSLKey")
	}

	dsnParts := []string{
		fmt.Sprintf("host=%s", config.Host),
		fmt.Sprintf("user=%s", config.Username),
		fmt.Sprintf("dbname=%s", config.Database),
		fmt.Sprintf("port=%d", config.Port),
		fmt.Sprintf("sslmode=%s", config.SSLMode),
	}
	if config.Password != "" {
		dsnParts = append(dsnParts, fmt.Sprintf("password=%s", config.Password))
	}
	if usingClientCert {
		// Client-certificate authentication: the certificate's subject maps
		// to the Postgres role via pg_hba.conf's "cert" auth method, so no
		// password is required (or used, if one is also present above).
		dsnParts = append(dsnParts,
			fmt.Sprintf("sslcert=%s", config.SSLCert),
			fmt.Sprintf("sslkey=%s", config.SSLKey),
		)
		if config.SSLRootCert != "" {
			dsnParts = append(dsnParts, fmt.Sprintf("sslrootcert=%s", config.SSLRootCert))
		}
	}
	return strings.Join(dsnParts, " "), nil
}

func (f *PostgreSQLClientFactory) CreateClient() (*gorm.DB, error) {
	dsn, err := buildDSN(f.config)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL database: %w", err)
	}

	// Configure connection pool for PostgreSQL
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	return db, nil
}
