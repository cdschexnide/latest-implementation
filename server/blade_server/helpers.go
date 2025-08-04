package blade_server

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "mime/multipart"
    "net/http"
    
    "blade-ingestion-service/database/models"
)

// CatalogUploader handles uploading items to the catalog
type CatalogUploader struct {
    catalogURL string
    authToken  string
    httpClient *http.Client
}

// NewCatalogUploader creates a new catalog uploader
func NewCatalogUploader(catalogURL, authToken string) *CatalogUploader {
    return &CatalogUploader{
        catalogURL: catalogURL,
        authToken:  authToken,
        httpClient: &http.Client{},
    }
}

// UploadItem uploads a BLADE item to the catalog
func (cu *CatalogUploader) UploadItem(item *models.BLADEItem) error {
    // Create multipart writer
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    
    // Add file part
    fileName := fmt.Sprintf("%s_%s.json", item.DataType, item.ItemID)
    part, err := writer.CreateFormFile("file", fileName)
    if err != nil {
        return fmt.Errorf("failed to create form file: %w", err)
    }
    
    // Write JSON data
    if _, err := part.Write(item.Data); err != nil {
        return fmt.Errorf("failed to write data: %w", err)
    }
    
    // Add metadata fields
    writer.WriteField("dataSource", fmt.Sprintf("BLADE:%s", item.DataType))
    writer.WriteField("classificationMarking", item.ClassificationMarking)
    writer.WriteField("itemId", item.ItemID)
    
    // Add metadata JSON
    if len(item.Metadata) > 0 {
        writer.WriteField("metadata", string(item.Metadata))
    }
    
    // Close writer
    if err := writer.Close(); err != nil {
        return fmt.Errorf("failed to close writer: %w", err)
    }
    
    // Create request
    req, err := http.NewRequest("POST", cu.catalogURL+"/catalog/item", body)
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }
    
    // Set headers
    req.Header.Set("Authorization", "Bearer "+cu.authToken)
    req.Header.Set("Content-Type", writer.FormDataContentType())
    
    // Send request
    resp, err := cu.httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("failed to send request: %w", err)
    }
    defer resp.Body.Close()
    
    // Check response
    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
        bodyBytes, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("catalog returned status %d: %s", resp.StatusCode, string(bodyBytes))
    }
    
    return nil
}

// CheckItemExists checks if an item already exists in the catalog
func (cu *CatalogUploader) CheckItemExists(dataType, itemID string) (bool, error) {
    url := fmt.Sprintf("%s/catalog/exists?source=%s&id=%s", cu.catalogURL, dataType, itemID)
    
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return false, err
    }
    
    req.Header.Set("Authorization", "Bearer "+cu.authToken)
    
    resp, err := cu.httpClient.Do(req)
    if err != nil {
        return false, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode == http.StatusOK {
        return true, nil
    } else if resp.StatusCode == http.StatusNotFound {
        return false, nil
    }
    
    return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
}

// BuildMetadata builds metadata for catalog upload
func BuildMetadata(item *models.BLADEItem, source string) map[string]interface{} {
    metadata := map[string]interface{}{
        "dataType":       item.DataType,
        "itemId":         item.ItemID,
        "classification": item.ClassificationMarking,
        "source":         source,
        "importTime":     item.CreatedAt.Format("2006-01-02T15:04:05Z"),
    }
    
    // Add any existing metadata
    if len(item.Metadata) > 0 {
        var existingMeta map[string]interface{}
        if err := json.Unmarshal(item.Metadata, &existingMeta); err == nil {
            for k, v := range existingMeta {
                metadata[k] = v
            }
        }
    }
    
    return metadata
}