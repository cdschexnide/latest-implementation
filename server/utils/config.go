package utils

import (
    "fmt"
    "os"
    "strconv"
    "time"
)

// Config holds all configuration for the BLADE Ingestion Service
type Config struct {
    // Server Configuration
    Host     string
    GRPCPort string
    RESTPort string
    
    // Database Configuration
    DBHost     string
    DBPort     string
    DBUser     string
    DBPassword string
    DBName     string
    DBSchema   string
    
    // Mock Databricks Configuration
    MockDatabricksURL    string
    MockDatabricksToken  string
    MockWarehouseID      string
    MockRequestTimeout   time.Duration
    
    // Catalog Configuration
    CatalogURL           string
    CatalogAuthToken     string
    CatalogTimeout       time.Duration
    CatalogBatchSize     int
    CatalogRetryAttempts int
    
    // Data Processing Configuration
    DefaultClassification string
    MaxRecordsPerQuery   int
    EnableDataValidation bool
    
    // BLADE-specific Configuration
    BLADEDataTypes []string
    DataTypeMapping map[string]string
    
    // Performance Configuration
    ConcurrentUploads  int
    RateLimitPerSecond int
    ProcessingTimeout  time.Duration
    
    // Logging
    LogLevel  string
    LogFormat string
    
    // Feature Flags
    UseSSL           bool
    EnableSwaggerUI  bool
    EnableMetrics    bool
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
    config := &Config{
        // Server defaults
        Host:     getEnvOrDefault("HOST", "0.0.0.0"),
        GRPCPort: getEnvOrDefault("GRPC_PORT", "9090"),
        RESTPort: getEnvOrDefault("REST_PORT", "9091"),
        
        // Database
        DBHost:     getEnvOrDefault("PGHOST", "localhost"),
        DBPort:     getEnvOrDefault("PGPORT", "5432"),
        DBUser:     getEnvOrDefault("PG_USER", "blade_user"),
        DBPassword: os.Getenv("APP_DB_ADMIN_PASSWORD"),
        DBName:     getEnvOrDefault("PG_DATABASE", "blade_ingestion"),
        DBSchema:   getEnvOrDefault("DB_SCHEMA_NAME", "public"),
        
        // Mock Databricks
        MockDatabricksURL:   getEnvOrDefault("MOCK_DATABRICKS_URL", "http://localhost:8080"),
        MockDatabricksToken: os.Getenv("MOCK_DATABRICKS_TOKEN"),
        MockWarehouseID:     getEnvOrDefault("MOCK_DATABRICKS_WAREHOUSE_ID", "test-warehouse"),
        MockRequestTimeout:  getDurationOrDefault("MOCK_REQUEST_TIMEOUT", 30*time.Second),
        
        // Catalog
        CatalogURL:           getEnvOrDefault("CATALOG_URL", "http://localhost:8092"),
        CatalogAuthToken:     os.Getenv("CATALOG_AUTH_TOKEN"),
        CatalogTimeout:       getDurationOrDefault("CATALOG_TIMEOUT", 60*time.Second),
        CatalogBatchSize:     getIntOrDefault("CATALOG_BATCH_SIZE", 10),
        CatalogRetryAttempts: getIntOrDefault("CATALOG_RETRY_ATTEMPTS", 3),
        
        // Data processing
        DefaultClassification: getEnvOrDefault("DEFAULT_CLASSIFICATION", "UNCLASSIFIED"),
        MaxRecordsPerQuery:   getIntOrDefault("MAX_RECORDS_PER_QUERY", 1000),
        EnableDataValidation: getBoolOrDefault("ENABLE_DATA_VALIDATION", true),
        
        // BLADE data types
        BLADEDataTypes: []string{"maintenance", "sortie", "deployment", "logistics"},
        DataTypeMapping: map[string]string{
            "maintenance": "blade_maintenance_data",
            "sortie":      "blade_sortie_data",
            "deployment":  "blade_deployment_data",
            "logistics":   "blade_logistics_data",
        },
        
        // Performance
        ConcurrentUploads:  getIntOrDefault("CONCURRENT_UPLOADS", 5),
        RateLimitPerSecond: getIntOrDefault("RATE_LIMIT_PER_SECOND", 10),
        ProcessingTimeout:  getDurationOrDefault("PROCESSING_TIMEOUT", 5*time.Minute),
        
        // Logging
        LogLevel:  getEnvOrDefault("LOG_LEVEL", "debug"),
        LogFormat: getEnvOrDefault("LOG_FORMAT", "json"),
        
        // Features
        UseSSL:          getBoolOrDefault("USE_SSL", false),
        EnableSwaggerUI: getBoolOrDefault("ENABLE_SWAGGER_UI", true),
        EnableMetrics:   getBoolOrDefault("ENABLE_METRICS", false),
    }
    
    // Validate required configuration
    if err := validateConfig(config); err != nil {
        return nil, fmt.Errorf("configuration validation failed: %w", err)
    }
    
    return config, nil
}

// validateConfig ensures all required configuration is present
func validateConfig(config *Config) error {
    if config.DBPassword == "" {
        return fmt.Errorf("database password (APP_DB_ADMIN_PASSWORD) is required")
    }
    
    if config.MockDatabricksURL == "" {
        return fmt.Errorf("mock Databricks URL is required")
    }
    
    if config.CatalogURL == "" {
        return fmt.Errorf("catalog URL is required")
    }
    
    if config.MaxRecordsPerQuery <= 0 {
        return fmt.Errorf("max records per query must be positive")
    }
    
    if config.CatalogBatchSize <= 0 {
        return fmt.Errorf("catalog batch size must be positive")
    }
    
    return nil
}

// GetDatabricksQuery builds a SQL query for a BLADE data type
func (c *Config) GetDatabricksQuery(dataType string, filter string, limit int) string {
    tableName, exists := c.DataTypeMapping[dataType]
    if !exists {
        tableName = fmt.Sprintf("blade_%s_data", dataType)
    }
    
    query := fmt.Sprintf("SELECT * FROM %s.%s", c.DBSchema, tableName)
    
    if filter != "" {
        query += " WHERE " + filter
    }
    
    if limit > 0 {
        query += fmt.Sprintf(" LIMIT %d", limit)
    } else if c.MaxRecordsPerQuery > 0 {
        query += fmt.Sprintf(" LIMIT %d", c.MaxRecordsPerQuery)
    }
    
    return query
}

// GetCatalogDataSource returns the catalog-compatible data source name
func (c *Config) GetCatalogDataSource(dataType string) string {
    return fmt.Sprintf("BLADE Databricks: %s", dataType)
}

// IsBLADEDataType checks if the given type is valid
func (c *Config) IsBLADEDataType(dataType string) bool {
    for _, validType := range c.BLADEDataTypes {
        if validType == dataType {
            return true
        }
    }
    return false
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getIntOrDefault(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if parsed, err := strconv.Atoi(value); err == nil {
            return parsed
        }
    }
    return defaultValue
}

func getBoolOrDefault(key string, defaultValue bool) bool {
    if value := os.Getenv(key); value != "" {
        if parsed, err := strconv.ParseBool(value); err == nil {
            return parsed
        }
    }
    return defaultValue
}

func getDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
    if value := os.Getenv(key); value != "" {
        if parsed, err := time.ParseDuration(value); err == nil {
            return parsed
        }
    }
    return defaultValue
}

// QueryConfig represents configuration for a specific query operation
type QueryConfig struct {
	DataType           string
	SQLQuery           string
	MaxResults         int
	IncludeMetadata    bool
	FilterCriteria     map[string]interface{}
	ClassificationFilter string
}

// NewQueryConfig creates a query configuration
func (c *Config) NewQueryConfig(dataType, sqlQuery string) *QueryConfig {
	return &QueryConfig{
			DataType:        dataType,
			SQLQuery:        sqlQuery,
			MaxResults:      c.MaxRecordsPerQuery,
			IncludeMetadata: true,
			FilterCriteria:  make(map[string]interface{}),
			ClassificationFilter: c.DefaultClassification,
	}
}

// CatalogUploadConfig represents configuration for catalog upload operations
type CatalogUploadConfig struct {
	BatchSize          int
	MaxRetries         int
	RetryDelay         time.Duration
	ValidateBeforeUpload bool
	SkipDuplicates     bool
	MetadataEnrichment bool
}

// NewCatalogUploadConfig creates upload configuration
func (c *Config) NewCatalogUploadConfig() *CatalogUploadConfig {
	return &CatalogUploadConfig{
			BatchSize:            c.CatalogBatchSize,
			MaxRetries:           c.CatalogRetryAttempts,
			RetryDelay:           2 * time.Second,
			ValidateBeforeUpload: c.EnableDataValidation,
			SkipDuplicates:       true,
			MetadataEnrichment:   true,
	}
}