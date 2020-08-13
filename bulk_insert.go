package bulk

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/jinzhu/gorm"
)

const (
	MaximumPlaceholders = 65536
	QSLitePlaceholders  = 999
	gormTag             = "gorm"
	columnPrefix        = "column:"
	funcTableName       = "TableName"
)

// BulkInsert
func BulkInsert(db *gorm.DB, bulks []interface{}) error {
	return BulkUpsert(db, bulks, nil)
}

// BulkInsertWithTableName
func BulkInsertWithTableName(db *gorm.DB, tableName string, bulks []interface{}) error {
	return BulkUpsertWithTableName(db, tableName, bulks, nil)
}

// BulkUpsert
func BulkUpsert(db *gorm.DB, bulks []interface{}, uniqueKeys []string) error {
	return BulkUpsertWithTableName(db, getTableName(bulks[0]), bulks, uniqueKeys)
}

// BulkUpsertWithTableName
func BulkUpsertWithTableName(db *gorm.DB, tableName string, bulks []interface{}, uniqueKeys []string) error {
	isUpsert := false
	if len(uniqueKeys) > 0 {
		isUpsert = true
	}

	tags, aTags := getTags(bulks)
	fields := strings.Join(aTags, ", ")
	objPlaceholders := len(aTags)
	if isUpsert {
		objPlaceholders = len(aTags)*2 - len(uniqueKeys)
	}

	batchSize := MaximumPlaceholders / objPlaceholders
	if strings.HasPrefix(db.Dialect().GetName(), "sqlite") {
		batchSize = QSLitePlaceholders / objPlaceholders
	}

	tx := db.Begin()

	for i := 0; i < len(bulks)/batchSize+1; i++ {
		maxBatchIndex := (i + 1) * batchSize
		if maxBatchIndex > len(bulks) {
			maxBatchIndex = len(bulks)
		}

		valueArgs, onUpdateFields := sliceValues(bulks[i*batchSize:maxBatchIndex], tags, aTags, uniqueKeys)

		phStrs := make([]string, maxBatchIndex-i*batchSize)
		placeholderStrs := "(?" + strings.Repeat(", ?", len(aTags)-1) + ")"
		for j := range bulks[i*batchSize : maxBatchIndex] {
			phStrs[j] = placeholderStrs
		}

		var upsertPhStrs []string
		if isUpsert {
			upsertPhStrs = make([]string, len(onUpdateFields))
			for j := range onUpdateFields {
				upsertPhStrs[j] = fmt.Sprintf("%s = ?", onUpdateFields[j])
			}
		}

		if isUpsert {
			numArgs := len(aTags) + len(onUpdateFields)
			upsertPlaceholderStr := strings.Join(upsertPhStrs, ",")
			for j := 0; j < len(valueArgs); j += numArgs {
				smt := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s ON DUPLICATE KEY UPDATE %s", tableName, fields, placeholderStrs, upsertPlaceholderStr)
				err := tx.Exec(smt, valueArgs[j:min(j+numArgs, len(valueArgs)-1)]...).Error
				if err != nil {
					tx.Rollback()
					return err
				}
			}
		} else {
			smt := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", tableName, fields, strings.Join(phStrs, ",\n"))
			err := tx.Exec(smt, valueArgs...).Error
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit().Error
}

// Deprecated: Insert is deprecated, using BulkInsertWithTableName instead
func Insert(db *gorm.DB, tableName string, bulks []interface{}) error {
	return BulkInsertWithTableName(db, tableName, bulks)
}

func getTableName(t interface{}) string {
	st := reflect.TypeOf(t)
	if _, ok := st.MethodByName(funcTableName); ok {
		v := reflect.ValueOf(t).MethodByName(funcTableName).Call(nil)
		if len(v) > 0 {
			return v[0].String()
		}
	}

	name := ""
	if t := reflect.TypeOf(t); t.Kind() == reflect.Ptr {
		name = t.Elem().Name()
	} else {
		name = t.Name()
	}

	return toSnakeCase(name)
}

func getTags(objs []interface{}) ([]string, []string) {
	re := regexp.MustCompile(fmt.Sprintf("(?i)%s[a-z0-9_\\-]+", columnPrefix))
	tags := make([]string, reflect.TypeOf(objs[0]).NumField())

	for i := range objs {
		v := reflect.ValueOf(objs[i])
		t := reflect.TypeOf(objs[i])
		for j := 0; j < t.NumField(); j++ {
			if tags[j] != "" || isZeroOfUnderlyingType(v.Field(j).Interface()) {
				continue
			}

			field := t.Field(j)
			tag := field.Tag.Get(gormTag)
			if tag == "-" {
				continue
			}

			tag = re.FindString(tag)
			if strings.HasPrefix(tag, columnPrefix) {
				tag = strings.TrimPrefix(tag, columnPrefix)
			} else {
				tag = toSnakeCase(field.Name)
			}

			tags[j] = tag
		}
	}

	availableTags := []string{}
	for i := range tags {
		if tags[i] != "" {
			availableTags = append(availableTags, tags[i])
		}
	}

	return tags, availableTags
}

func sliceValues(objs []interface{}, tags, aTags, uniqueKeys []string) ([]interface{}, []string) {
	uniqueTag := map[string]struct{}{}
	var upsertTags []string
	isUpsert := false
	updateSize := len(aTags)
	if len(uniqueKeys) > 0 && len(uniqueKeys) < len(aTags) {
		updateSize += updateSize - len(uniqueKeys)
		isUpsert = true

		for i := range uniqueKeys {
			uniqueTag[uniqueKeys[i]] = struct{}{}
		}

		upsertTags = make([]string, len(aTags)-len(uniqueKeys))
		j := 0
		for i := range aTags {
			if _, ok := uniqueTag[aTags[i]]; !ok {
				upsertTags[j] = aTags[i]
				j += 1
			}
		}
	}

	availableValues := make([]interface{}, len(objs)*updateSize)

	c := 0
	for i := range objs {
		v := reflect.ValueOf(objs[i])

		var upsertValues []interface{}
		if isUpsert {
			upsertValues = make([]interface{}, len(upsertTags))
		}

		k := 0
		for j := 0; j < v.NumField(); j++ {
			if tags[j] != "" {
				availableValues[c] = v.Field(j).Interface()

				if _, ok := uniqueTag[tags[j]]; !ok && isUpsert {
					upsertValues[k] = availableValues[c]
					k += 1
				}

				c += 1
			}
		}

		for j := range upsertValues {
			availableValues[c] = upsertValues[j]
			c += 1
		}
	}

	return availableValues, upsertTags
}

func isZeroOfUnderlyingType(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}
