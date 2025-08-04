package datasource

import (
    "time"
    "gorm.io/gorm"
    "gorm.io/datatypes"
)

// DataSource represents a configured BLADE data source
type DataSource struct {
    gorm.Model
    TypeName         string         `gorm:"uniqueIndex;not null" json:"type_name"`
    DisplayName      string         `json:"display_name"`
    DataType         string         `gorm:"index" json:"data_type"` // maintenance, sortie, etc.
    Enabled          bool           `json:"enabled"`
    Parameters       datatypes.JSON `json:"parameters"`
    
    // Sync configuration
    SyncEnabled      bool           `json:"sync_enabled"`
    SyncSchedule     string         `json:"sync_schedule,omitempty"` // Cron expression
    LastSyncTime     *time.Time     `json:"last_sync_time,omitempty"`
    LastSyncStatus   string         `json:"last_sync_status,omitempty"`
    
    // Statistics
    ItemCount        int            `json:"item_count"`
    LastErrorMessage string         `json:"last_error_message,omitempty"`
    
    // Databricks specific
    WarehouseID      string         `json:"warehouse_id"`
    CatalogName      string         `json:"catalog_name"`
    SchemaName       string         `json:"schema_name"`
    TableName        string         `json:"table_name"`
}

// GetParameters unmarshals parameters JSON
func (ds *DataSource) GetParameters() (map[string]interface{}, error) {
    var params map[string]interface{}
    if err := ds.Parameters.Scan(&params); err != nil {
        return nil, err
    }
    return params, nil
}

// SetParameters marshals parameters to JSON
func (ds *DataSource) SetParameters(params map[string]interface{}) error {
    data, err := datatypes.NewJSONType(params).MarshalJSON()
    if err != nil {
        return err
    }
    ds.Parameters = data
    return nil
}

// GetFullTableName returns the fully qualified table name
func (ds *DataSource) GetFullTableName() string {
    if ds.CatalogName != "" && ds.SchemaName != "" {
        return ds.CatalogName + "." + ds.SchemaName + "." + ds.TableName
    }
    return ds.TableName
}