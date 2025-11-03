package tools

import (
	"fmt"
	"log"
	"os"
	"strings"

	"api/src/config"
)

func GenerateModelsFromDatabase() error {
	// Get all table names

	var tables []string
	if err := config.DB.Raw("SELECT tablename FROM pg_tables WHERE schemaname = 'public'").Scan(&tables).Error; err != nil {
		return fmt.Errorf("[ERROR] Failed to get table names: %v", err)
	}

	modelContent := `package models
	
import (
	"time"
)

`

	for _, table := range tables {
		if strings.HasPrefix(table, "pg_") || table == "schema_migrations" {
			continue // Skip system tables
		}

		modelStruct, err := generateStructFromTable(table)
		if err != nil {
			log.Printf("[WARN] Could not generate struct for table %s: %v", table, err)
			continue
		}

		modelContent += modelStruct + "\n\n"
	}

	file, err := os.Create("src/models/models.go")
	if err != nil {
		return fmt.Errorf("[ERROR] Failed to create models.go :%v", err)
	}
	defer file.Close()

	_, err = file.WriteString(modelContent)
	if err != nil {
		return fmt.Errorf("[ERROR] Failed to write to models.go :%v", err)
	}

	log.Println("[NOTICE] Models generated successfully. Can be found in models/models.go")
	return nil
}

func generateStructFromTable(tableName string) (string, error) {
	// Get column infomation

	type ColumnInfo struct {
		ColumnName    string
		DataType      string
		IsNullable    string
		ColumnDefault *string
	}

	var columns []ColumnInfo
	query := `
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns
		WHERE table_name = ?
		ORDER BY ordinal_position
	`

	if err := config.DB.Raw(query, tableName).Scan(&columns).Error; err != nil {
		return "", fmt.Errorf("[ERROR] Failed to get column info: %v", err)
	}

	stuctName := toPascalCase(tableName)
	structContent := fmt.Sprintf("type %s struct {\n", stuctName)

	for _, col := range columns {
		fieldName := toPascalCase(col.ColumnName)
		goType := mapPostgresToGoType(col.DataType, col.IsNullable == "YES")
		jsonTag := fmt.Sprintf("`json:\"%s\"`", col.ColumnName)

		// Add GORM tags
		gormTags := generateGormTags(col.ColumnName, col.DataType, col.IsNullable == "YES", col.ColumnDefault)
		if gormTags != "" {
			jsonTag = fmt.Sprintf("`json:\"%s\" %s`", col.ColumnName, gormTags)
		}

		structContent += fmt.Sprintf("\t%s\t%s\t%s\n", fieldName, goType, jsonTag)
	}
	structContent += "}\n"

	// Add TableName method

	structContent += fmt.Sprintf("\nfunc (%s) TableName() string {\n", stuctName)
	structContent += fmt.Sprintf("\treturn \"%s\"\n", tableName)
	structContent += "}"

	return structContent, nil
}

func toPascalCase(s string) string {
	words := strings.Split(s, "_")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, "")
}

func mapPostgresToGoType(pgType string, nullable bool) string {
	var goType string

	switch pgType {
	case "integer", "int4":
		goType = "int"
	case "bigint", "int8":
		goType = "int64"
	case "smallint", "int2":
		goType = "int16"
	case "serial", "serial4":
		goType = "uint"
	case "bigserial", "serial8":
		goType = "uint64"
	case "boolean", "bool":
		goType = "bool"
	case "character varying", "varchar", "text":
		goType = "string"
	case "timestamp with time zone", "timestamp without time zone", "timestamp":
		goType = "time.Time"
	case "date":
		goType = "time.Time"
	case "numeric", "decimal":
		goType = "float64"
	case "real", "float4":
		goType = "float32"
	case "double precision", "float8":
		goType = "float64"
	case "uuid":
		goType = "string"
	case "json", "jsonb":
		goType = "string" // or you could use json.RawMessage
	default:
		goType = "interface{}" // fallback for unknown types
	}

	if nullable && goType != "string" {
		return "*" + goType
	}

	return goType
}

func generateGormTags(columnName, dataType string, nullable bool, defaultValue *string) string {
	var gormOptions []string

	// Primary key detection
	if columnName == "id" {
		if dataType == "uuid" {
			// For UUID primary keys with auto-generation
			return "gorm:\"primaryKey;type:uuid;default:gen_random_uuid()\""
		} else {
			return "gorm:\"primaryKey\""
		}
	}

	// Handle UUID fields (foreign keys or other UUID columns)
	if dataType == "uuid" {
		gormOptions = append(gormOptions, "type:uuid")
		// Check if it has a default value (auto-generation)
		if defaultValue != nil && strings.Contains(*defaultValue, "gen_random_uuid()") {
			gormOptions = append(gormOptions, "default:gen_random_uuid()")
		}
	}

	// Handle foreign key UUID columns (like user_id, group_id, etc.)
	if dataType == "uuid" && strings.HasSuffix(columnName, "_id") && columnName != "id" {
		// This is likely a foreign key UUID column
		if !contains(gormOptions, "type:uuid") {
			gormOptions = append(gormOptions, "type:uuid")
		}
	}

	// Common GORM tags
	if !nullable {
		gormOptions = append(gormOptions, "not null")
	}

	if len(gormOptions) > 0 {
		return fmt.Sprintf("gorm:\"%s\"", strings.Join(gormOptions, ";"))
	}

	return ""
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
