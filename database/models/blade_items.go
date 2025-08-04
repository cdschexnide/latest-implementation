package models

import (
    "time"
    "gorm.io/gorm"
    "gorm.io/datatypes"
)

// BLADEItemType represents the type of BLADE data
type BLADEItemType string

const (
    MaintenanceData BLADEItemType = "maintenance"
    SortieData      BLADEItemType = "sortie"
    DeploymentData  BLADEItemType = "deployment"
    LogisticsData   BLADEItemType = "logistics"
)

// BLADEItem represents a generic BLADE data item in the database
type BLADEItem struct {
    gorm.Model
    ItemID                string         `gorm:"uniqueIndex;not null" json:"item_id"`
    DataType              string         `gorm:"index;not null" json:"data_type"`
    Data                  datatypes.JSON `json:"data"`
    ClassificationMarking string         `json:"classification_marking"`
    LastModified          time.Time      `json:"last_modified"`
    Metadata              datatypes.JSON `json:"metadata,omitempty"`
    
    // Tracking fields
    DataSourceID   uint      `gorm:"index" json:"data_source_id"`
    IngestionJobID string    `gorm:"index" json:"ingestion_job_id,omitempty"`
    CatalogID      string    `json:"catalog_id,omitempty"`
    UploadedAt     *time.Time `json:"uploaded_at,omitempty"`
}

// TableName specifies the table name for BLADE items
func (BLADEItem) TableName() string {
    return "blade_items"
}

// BLADEMaintenanceData represents aircraft maintenance data
type BLADEMaintenanceData struct {
    ItemID                 string    `json:"item_id"`
    AircraftTail          string    `json:"aircraft_tail"`
    AircraftType          string    `json:"aircraft_type"`
    MaintenanceType       string    `json:"maintenance_type"`
    MaintenanceCode       string    `json:"maintenance_code"`
    Description           string    `json:"description"`
    Priority              string    `json:"priority"`
    EstimatedCompletion   *time.Time `json:"estimated_completion,omitempty"`
    ActualCompletion      *time.Time `json:"actual_completion,omitempty"`
    TechnicianAssigned    string    `json:"technician_assigned"`
    BaseLocation          string    `json:"base_location"`
    WorkOrder             string    `json:"work_order"`
    NextScheduledDate     *time.Time `json:"next_scheduled_date,omitempty"`
}

// BLADESortieData represents flight mission data
type BLADESortieData struct {
    ItemID              string    `json:"item_id"`
    MissionID           string    `json:"mission_id"`
    AircraftTail        string    `json:"aircraft_tail"`
    AircraftType        string    `json:"aircraft_type"`
    PilotCallsign       string    `json:"pilot_callsign"`
    MissionType         string    `json:"mission_type"`
    DepartureBase       string    `json:"departure_base"`
    DestinationBase     string    `json:"destination_base"`
    ScheduledDeparture  time.Time `json:"scheduled_departure"`
    ActualDeparture     *time.Time `json:"actual_departure,omitempty"`
    ScheduledArrival    time.Time `json:"scheduled_arrival"`
    ActualArrival       *time.Time `json:"actual_arrival,omitempty"`
    FlightHours         *float64  `json:"flight_hours,omitempty"`
    MissionStatus       string    `json:"mission_status"`
}

// BLADEDeploymentData represents military deployment information
type BLADEDeploymentData struct {
    ItemID               string    `json:"item_id"`
    DeploymentID         string    `json:"deployment_id"`
    UnitDesignation      string    `json:"unit_designation"`
    UnitType             string    `json:"unit_type"`
    PersonnelCount       int       `json:"personnel_count"`
    CommandingOfficer    string    `json:"commanding_officer"`
    DeploymentLocation   string    `json:"deployment_location"`
    OriginBase           string    `json:"origin_base"`
    DeploymentStartDate  time.Time `json:"deployment_start_date"`
    DeploymentEndDate    *time.Time `json:"deployment_end_date,omitempty"`
    MissionObjective     string    `json:"mission_objective"`
    OperationalStatus    string    `json:"operational_status"`
}

// BLADELogisticsData represents supply chain and logistics data
type BLADELogisticsData struct {
    ItemID              string    `json:"item_id"`
    ShipmentID          string    `json:"shipment_id"`
    SupplyType          string    `json:"supply_type"`
    Description         string    `json:"description"`
    Quantity            int       `json:"quantity"`
    UnitOfMeasure       string    `json:"unit_of_measure"`
    Vendor              string    `json:"vendor"`
    OriginLocation      string    `json:"origin_location"`
    DestinationLocation string    `json:"destination_location"`
    ShippedDate         *time.Time `json:"shipped_date,omitempty"`
    EstimatedArrival    *time.Time `json:"estimated_arrival,omitempty"`
    Priority            string    `json:"priority"`
}

// GetBLADEItemType determines the BLADE item type from a string
func GetBLADEItemType(itemType string) BLADEItemType {
    switch itemType {
    case "maintenance", "engine_maintenance", "avionics_check":
        return MaintenanceData
    case "sortie", "training_mission", "combat_mission":
        return SortieData
    case "deployment", "unit_deployment":
        return DeploymentData
    case "logistics", "supply_shipment", "parts_delivery":
        return LogisticsData
    default:
        return MaintenanceData // Default
    }
}

// GetDefaultClassification returns default classification for item type
func GetDefaultClassification(itemType BLADEItemType) string {
    switch itemType {
    case MaintenanceData:
        return "UNCLASSIFIED"
    case SortieData:
        return "CONFIDENTIAL"
    case DeploymentData:
        return "SECRET"
    case LogisticsData:
        return "UNCLASSIFIED"
    default:
        return "UNCLASSIFIED"
    }
}

// ValidateClassificationMarking validates classification marking
func ValidateClassificationMarking(marking string) bool {
    validMarkings := map[string]bool{
        "U":             true,
        "UNCLASSIFIED":  true,
        "CUI":           true,
        "C":             true,
        "CONFIDENTIAL":  true,
        "S":             true,
        "SECRET":        true,
        "TS":            true,
        "TOP SECRET":    true,
    }
    return validMarkings[marking]
}