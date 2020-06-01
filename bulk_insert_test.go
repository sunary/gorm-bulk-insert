package bulk

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/require"
)

type user struct {
	ID       int
	UserName string `gorm:"column:name"`
	Age      int
	Hobby    string `gorm:"-"`
}

func insertData() []user {
	return []user{
		{
			UserName: "sunary",
			Age:      22,
			Hobby:    "chess",
		},
		{
			UserName: "aku",
			Age:      68,
			Hobby:    "manga",
		},
	}
}

func bulkData() []interface{} {
	insertData := insertData()
	bulkData := make([]interface{}, len(insertData))
	for i := range insertData {
		bulkData[i] = insertData[i]
	}

	return bulkData
}

func TestInsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	defer db.Close()

	gdb, err := gorm.Open("mysql", db)
	require.NoError(t, err)

	insertData := insertData()
	tableName := "user"

	mock.ExpectBegin()
	mock.ExpectExec(
		fmt.Sprintf("INSERT INTO %s", tableName),
	).WithArgs(
		reflect.ValueOf(insertData[0].UserName).Interface(), reflect.ValueOf(insertData[0].Age).Interface(),
		reflect.ValueOf(insertData[1].UserName).Interface(), reflect.ValueOf(insertData[1].Age).Interface(),
	).WillReturnResult(sqlmock.NewResult(2, 2))
	mock.ExpectCommit()

	bulkData := bulkData()
	err = Insert(gdb, tableName, bulkData)
	require.NoError(t, err)
}

func Test_getTags(t *testing.T) {
	bulkData := bulkData()

	tests := []struct {
		input    []interface{}
		expected [2][]string
	}{
		{
			input:    bulkData,
			expected: [2][]string{{"", "name", "age", ""}, {"name", "age"}},
		},
	}

	for _, tt := range tests {
		if got1, got2 := getTags(tt.input); !reflect.DeepEqual(got1, tt.expected[0]) || !reflect.DeepEqual(got2, tt.expected[1]) {
			t.Errorf("getTags() = %v, %v, want %v, %v", got1, got2, tt.expected[0], tt.expected[1])
		}
	}
}

func Test_sliceValues(t *testing.T) {
	insertData := insertData()
	bulkData := bulkData()

	type input struct {
		objs          []interface{}
		tags          []string
		availableTags []string
	}
	tests := []struct {
		input    input
		expected []interface{}
	}{
		{
			input: input{
				objs:          bulkData,
				tags:          []string{"", "name", "age", ""},
				availableTags: []string{"name", "age"},
			},
			expected: []interface{}{
				reflect.ValueOf(insertData[0].UserName).Interface(), reflect.ValueOf(insertData[0].Age).Interface(),
				reflect.ValueOf(insertData[1].UserName).Interface(), reflect.ValueOf(insertData[1].Age).Interface(),
			},
		},
	}
	for _, tt := range tests {
		if got := sliceValues(tt.input.objs, tt.input.tags, tt.input.availableTags); !reflect.DeepEqual(got, tt.expected) {
			t.Errorf("sliceValues() = %v, want %v", got, tt.expected)
		}
	}
}

func Test_isZeroOfUnderlyingType(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected bool
	}{
		{
			input:    0,
			expected: true,
		},
		{
			input:    "",
			expected: true,
		},
		{
			input:    42,
			expected: false,
		},
		{
			input:    "foo",
			expected: false,
		},
		{
			input:    0.0,
			expected: true,
		},
		{
			input:    0.1,
			expected: false,
		},
		{
			input:    time.Now(),
			expected: false,
		},
	}

	for _, tt := range tests {
		if got := isZeroOfUnderlyingType(tt.input); !reflect.DeepEqual(got, tt.expected) {
			t.Errorf("isZeroOfUnderlyingType() = %v, want %v", got, tt.expected)
		}
	}
}
