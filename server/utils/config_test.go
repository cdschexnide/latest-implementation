package utils

import (
    "os"
    "testing"
    
    "github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
    // Set test environment variables
    os.Setenv("APP_DB_ADMIN_PASSWORD", "test-password")
    os.Setenv("MOCK_DATABRICKS_URL", "http://test-databricks:8080")
    os.Setenv("CATALOG_URL", "http://test-catalog:8092")
    
    // Load config
    config, err := LoadConfig()
    
    // Assert no error
    assert.NoError(t, err)
    assert.NotNil(t, config)
    
    // Test defaults
    assert.Equal(t, "9090", config.GRPCPort)
    assert.Equal(t, "9091", config.RESTPort)
    assert.Equal(t, "blade_ingestion", config.DBName)
    
    // Test environment overrides
    assert.Equal(t, "test-password", config.DBPassword)
    assert.Equal(t, "http://test-databricks:8080", config.MockDatabricksURL)
    
    // Test BLADE data types
    assert.Contains(t, config.BLADEDataTypes, "maintenance")
    assert.Contains(t, config.BLADEDataTypes, "sortie")
    assert.True(t, config.IsBLADEDataType("maintenance"))
    assert.False(t, config.IsBLADEDataType("invalid"))
}

func TestGetDatabricksQuery(t *testing.T) {
    config := &Config{
        DBSchema: "public",
        DataTypeMapping: map[string]string{
            "maintenance": "blade_maintenance_data",
        },
        MaxRecordsPerQuery: 1000,
    }
    
    // Test basic query
    query := config.GetDatabricksQuery("maintenance", "", 0)
    expected := "SELECT * FROM public.blade_maintenance_data LIMIT 1000"
    assert.Equal(t, expected, query)
    
    // Test with filter
    query = config.GetDatabricksQuery("maintenance", "priority = 'HIGH'", 10)
    expected = "SELECT * FROM public.blade_maintenance_data WHERE priority = 'HIGH' LIMIT 10"
    assert.Equal(t, expected, query)
}

func TestValidateConfig(t *testing.T) {
    // Test missing password
    config := &Config{
        MockDatabricksURL: "http://test",
        CatalogURL: "http://test",
        MaxRecordsPerQuery: 100,
        CatalogBatchSize: 10,
    }
    err := validateConfig(config)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "password")
    
    // Test valid config
    config.DBPassword = "password"
    err = validateConfig(config)
    assert.NoError(t, err)
}