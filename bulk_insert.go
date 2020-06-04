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
)

func Insert(db *gorm.DB, tableName string, bulks []interface{}) error {
	tags, aTags := getTags(bulks)
	objPlaceholders := len(aTags)
	fields := strings.Join(aTags, ", ")

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

		valueArgs := sliceValues(bulks[i*batchSize:maxBatchIndex], tags, aTags)
		phStrs := make([]string, maxBatchIndex-i*batchSize)
		placeholderStrs := "(?" + strings.Repeat(", ?", objPlaceholders-1) + ")"
		for j := range bulks[i*batchSize : maxBatchIndex] {
			phStrs[j] = placeholderStrs
		}

		smt := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", tableName, fields, strings.Join(phStrs, ",\n"))
		err := tx.Exec(smt, valueArgs...).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
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

func sliceValues(objs []interface{}, tags, aTags []string) []interface{} {
	availableValues := make([]interface{}, len(objs)*len(aTags))

	c := 0
	for i := range objs {
		v := reflect.ValueOf(objs[i])
		for j := 0; j < v.NumField(); j++ {
			if tags[j] != "" {
				availableValues[c] = v.Field(j).Interface()
				c += 1
			}
		}
	}

	return availableValues
}

func isZeroOfUnderlyingType(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}
