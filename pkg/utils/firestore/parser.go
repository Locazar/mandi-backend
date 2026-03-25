package firestore

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// ParseFields safely extracts all Firestore typed values from fields map.
// Firestore returns values in a typed format:
// - stringValue: "text"
// - integerValue: "123"
// - doubleValue: 123.45
// - booleanValue: true/false
// - nullValue: "NULL_VALUE"
// - arrayValue: {values: [...]}
// - mapValue: {fields: {...}}
// - timestampValue: "2024-01-01T00:00:00Z"
// - geoPointValue: {latitude: ..., longitude: ...}
// - bytesValue: "base64string"
func ParseFields(fieldsMap map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for fieldName, fieldValue := range fieldsMap {
		if fieldValue == nil {
			result[fieldName] = nil
			continue
		}

		// Type assert to map for Firestore structured value
		fieldMap, ok := fieldValue.(map[string]interface{})
		if !ok {
			// Raw value (shouldn't happen with Firestore events, but handle gracefully)
			result[fieldName] = fieldValue
			continue
		}

		// Extract actual value based on Firestore type
		value := ExtractFirestoreValue(fieldMap)
		result[fieldName] = value
	}

	return result
}

// ExtractFirestoreValue safely extracts typed Firestore value from a field map.
// Returns the actual Go value or nil if unable to extract.
func ExtractFirestoreValue(fieldData map[string]interface{}) interface{} {
	if fieldData == nil {
		return nil
	}

	// stringValue
	if v, ok := fieldData["stringValue"].(string); ok {
		return v
	}

	// integerValue (comes as string in JSON to preserve precision)
	if v, ok := fieldData["integerValue"].(string); ok {
		return v // keep as string to preserve large integers
	}

	// doubleValue
	if v, ok := fieldData["doubleValue"].(float64); ok {
		return v
	}

	// booleanValue
	if v, ok := fieldData["booleanValue"].(bool); ok {
		return v
	}

	// timestampValue (RFC 3339 formatted string)
	if v, ok := fieldData["timestampValue"].(string); ok {
		return v
	}

	// nullValue
	if v, ok := fieldData["nullValue"].(string); ok {
		if v == "NULL_VALUE" {
			return nil
		}
	}

	// arrayValue
	if arrData, ok := fieldData["arrayValue"].(map[string]interface{}); ok {
		if values, ok := arrData["values"].([]interface{}); ok {
			extracted := make([]interface{}, 0, len(values))
			for _, item := range values {
				if itemMap, ok := item.(map[string]interface{}); ok {
					extracted = append(extracted, ExtractFirestoreValue(itemMap))
				} else {
					extracted = append(extracted, item)
				}
			}
			return extracted
		}
	}

	// mapValue (nested object)
	if mapData, ok := fieldData["mapValue"].(map[string]interface{}); ok {
		if fields, ok := mapData["fields"].(map[string]interface{}); ok {
			return ParseFields(fields)
		}
	}

	// geoPointValue
	if geoPoint, ok := fieldData["geoPointValue"].(map[string]interface{}); ok {
		return geoPoint // return as-is with latitude and longitude
	}

	// bytesValue
	if v, ok := fieldData["bytesValue"].(string); ok {
		return v // base64 encoded bytes
	}

	// referenceValue (Firestore document reference)
	if v, ok := fieldData["referenceValue"].(string); ok {
		return v // full path to referenced document
	}

	// No recognized type - return nil
	return nil
}

// ValuesEqual safely compares two values for equality
func ValuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Try direct comparison first
	if a == b {
		return true
	}

	// Convert to JSON for complex comparisons to handle type mismatches
	// (e.g., int64 string vs float64)
	aJSON, errA := json.Marshal(a)
	bJSON, errB := json.Marshal(b)

	if errA != nil || errB != nil {
		return false
	}

	return string(aJSON) == string(bJSON)
}

// ExtractDocumentPath extracts the document ID and collection path from full resource name
// Input: "projects/{project}/databases/(default)/documents/enquiries/doc-123"
// Returns: documentPath="enquiries/doc-123", documentID="doc-123"
func ExtractDocumentPath(fullName string) (documentPath string, documentID string) {
	// Format: projects/{project}/databases/{database}/documents/{path}
	parts := strings.Split(fullName, "/documents/")
	if len(parts) != 2 {
		log.Printf("WARN: Could not parse document path from: %s", fullName)
		return "", ""
	}

	documentPath = parts[1]
	pathParts := strings.Split(documentPath, "/")
	if len(pathParts) > 0 {
		documentID = pathParts[len(pathParts)-1]
	}

	return documentPath, documentID
}

// ExtractFieldValue safely extracts a field value from parsed fields with fallback alternatives
func ExtractFieldValue(fields map[string]interface{}, fieldNames ...string) interface{} {
	for _, fieldName := range fieldNames {
		if val, ok := fields[fieldName]; ok && val != nil {
			return val
		}
	}
	return nil
}

// GetFieldAsString safely converts a field value to string
func GetFieldAsString(fields map[string]interface{}, fieldName string) string {
	val := ExtractFieldValue(fields, fieldName)
	if val == nil {
		return ""
	}
	return fmt.Sprintf("%v", val)
}

// GetFieldAsInt safely converts a field value to int64
func GetFieldAsInt(fields map[string]interface{}, fieldName string) int64 {
	val := ExtractFieldValue(fields, fieldName)
	if val == nil {
		return 0
	}

	switch v := val.(type) {
	case float64:
		return int64(v)
	case int:
		return int64(v)
	case int64:
		return v
	case string:
		var i int64
		fmt.Sscanf(v, "%d", &i)
		return i
	default:
		return 0
	}
}

// GetFieldAsBool safely converts a field value to bool
func GetFieldAsBool(fields map[string]interface{}, fieldName string) bool {
	val := ExtractFieldValue(fields, fieldName)
	if val == nil {
		return false
	}

	switch v := val.(type) {
	case bool:
		return v
	case string:
		return v == "true" || v == "1" || v == "yes"
	default:
		return false
	}
}
