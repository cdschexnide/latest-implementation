package blade_server

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
    
    "blade-ingestion-service/database/models"
    pb "blade-ingestion-service/generated/proto"
    
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

// DatabricksClient handles communication with mock Databricks
type DatabricksClient struct {
    baseURL     string
    token       string
    warehouseID string
    httpClient  *http.Client
}

// NewDatabricksClient creates a new Databricks client
func NewDatabricksClient(baseURL, token, warehouseID string) *DatabricksClient {
    return &DatabricksClient{
        baseURL:     baseURL,
        token:       token,
        warehouseID: warehouseID,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

// ExecuteQuery executes a SQL query against Databricks
func (dc *DatabricksClient) ExecuteQuery(ctx context.Context, query string) ([]map[string]interface{}, error) {
    // Prepare request
    reqBody := map[string]interface{}{
        "warehouse_id": dc.warehouseID,
        "statement":    query,
        "wait_timeout": "30s",
    }
    
    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %w", err)
    }
    
    // Create HTTP request
    req, err := http.NewRequestWithContext(ctx, "POST", 
        fmt.Sprintf("%s/api/2.0/sql/statements", dc.baseURL), 
        bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    // Set headers
    req.Header.Set("Authorization", "Bearer "+dc.token)
    req.Header.Set("Content-Type", "application/json")
    
    // Execute request
    resp, err := dc.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to execute query: %w", err)
    }
    defer resp.Body.Close()
    
    // Read response
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("databricks returned status %d: %s", resp.StatusCode, string(body))
    }
    
    // Parse response
    var result struct {
        Result struct {
            Data [][]interface{} `json:"data"`
            Schema struct {
                Columns []struct {
                    Name string `json:"name"`
                } `json:"columns"`
            } `json:"schema"`
        } `json:"result"`
    }
    
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }
    
    // Convert to map format
    var rows []map[string]interface{}
    columns := result.Result.Schema.Columns
    
    for _, row := range result.Result.Data {
        rowMap := make(map[string]interface{})
        for i, col := range columns {
            if i < len(row) {
                rowMap[col.Name] = row[i]
            }
        }
        rows = append(rows, rowMap)
    }
    
    return rows, nil
}

// FetchBLADEItem fetches a specific BLADE item
func (dc *DatabricksClient) FetchBLADEItem(ctx context.Context, dataType, itemID string, tableName string) (map[string]interface{}, error) {
    query := fmt.Sprintf("SELECT * FROM %s WHERE item_id = '%s' LIMIT 1", tableName, itemID)
    
    rows, err := dc.ExecuteQuery(ctx, query)
    if err != nil {
        return nil, err
    }
    
    if len(rows) == 0 {
        return nil, fmt.Errorf("item not found")
    }
    
    return rows[0], nil
}

// TransformToBLADEItem converts raw data to BLADE item format
func TransformToBLADEItem(dataType string, rawData map[string]interface{}) (*models.BLADEItem, error) {
    // Generate item ID if not present
    itemID, ok := rawData["item_id"].(string)
    if !ok || itemID == "" {
        itemID = fmt.Sprintf("%s-%d", dataType, time.Now().Unix())
    }
    
    // Marshal data to JSON
    dataJSON, err := json.Marshal(rawData)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal data: %w", err)
    }
    
    // Determine classification
    classification := models.GetDefaultClassification(models.BLADEItemType(dataType))
    if class, ok := rawData["classification"].(string); ok {
        classification = class
    }
    
    // Create BLADE item
    item := &models.BLADEItem{
        ItemID:                itemID,
        DataType:              dataType,
        Data:                  dataJSON,
        ClassificationMarking: classification,
        LastModified:          time.Now(),
    }
    
    // Add metadata
    metadata := map[string]interface{}{
        "source":      "databricks",
        "import_time": time.Now().Format(time.RFC3339),
    }
    
    metadataJSON, _ := json.Marshal(metadata)
    item.Metadata = metadataJSON
    
    return item, nil
}