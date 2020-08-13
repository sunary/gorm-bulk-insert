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

func (user) TableName() string {
	return "tb_user"
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

func TestBulkInsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	defer db.Close()

	gdb, err := gorm.Open("mysql", db)
	require.NoError(t, err)

	insertData := insertData()

	mock.ExpectBegin()
	mock.ExpectExec(
		fmt.Sprintf("INSERT INTO %s", escapeSqlName(user{}.TableName())),
	).WithArgs(
		reflect.ValueOf(insertData[0].UserName).Interface(), reflect.ValueOf(insertData[0].Age).Interface(),
		reflect.ValueOf(insertData[1].UserName).Interface(), reflect.ValueOf(insertData[1].Age).Interface(),
	).WillReturnResult(sqlmock.NewResult(2, 2))
	mock.ExpectCommit()

	bulkData := bulkData()
	err = BulkInsert(gdb, bulkData)
	require.NoError(t, err)
}

func TestBulkInsertWithTableName(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	defer db.Close()

	gdb, err := gorm.Open("mysql", db)
	require.NoError(t, err)

	insertData := insertData()
	tableName := user{}.TableName()

	mock.ExpectBegin()
	mock.ExpectExec(
		fmt.Sprintf("INSERT INTO %s", escapeSqlName(tableName)),
	).WithArgs(
		reflect.ValueOf(insertData[0].UserName).Interface(), reflect.ValueOf(insertData[0].Age).Interface(),
		reflect.ValueOf(insertData[1].UserName).Interface(), reflect.ValueOf(insertData[1].Age).Interface(),
	).WillReturnResult(sqlmock.NewResult(2, 2))
	mock.ExpectCommit()

	bulkData := bulkData()
	err = BulkInsertWithTableName(gdb, tableName, bulkData)
	require.NoError(t, err)
}

func TestBulkUpsert(t *testing.T) {
	t.SkipNow()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	defer db.Close()

	gdb, err := gorm.Open("mysql", db)
	require.NoError(t, err)

	insertData := insertData()

	mock.ExpectBegin()
	mock.ExpectExec(
		fmt.Sprintf("INSERT INTO %s", escapeSqlName(user{}.TableName())),
	).WithArgs(
		reflect.ValueOf(insertData[0].UserName).Interface(), reflect.ValueOf(insertData[0].Age).Interface(), reflect.ValueOf(insertData[0].Age).Interface(),
	).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(
		fmt.Sprintf("INSERT INTO %s", escapeSqlName(user{}.TableName())),
	).WithArgs(
		reflect.ValueOf(insertData[1].UserName).Interface(), reflect.ValueOf(insertData[1].Age).Interface(), reflect.ValueOf(insertData[1].Age).Interface(),
	).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	bulkData := bulkData()
	err = BulkUpsert(gdb, bulkData, []string{"name"})
	require.NoError(t, err)
}

func Test_getTableName(t *testing.T) {
	tests := []struct {
		input interface{}
		want  string
	}{
		{
			input: insertData()[0],
			want:  insertData()[0].TableName(),
		},
	}

	for _, tt := range tests {
		if got := getTableName(tt.input); got != tt.want {
			t.Errorf("getTableName(%v) = %v want %v", tt.input, got, tt.want)
		}
	}
}

func Test_getTags(t *testing.T) {
	bulkData := bulkData()

	tests := []struct {
		input []interface{}
		want  [2][]string
	}{
		{
			input: bulkData,
			want:  [2][]string{{"", "name", "age", ""}, {"name", "age"}},
		},
	}

	for _, tt := range tests {
		if got1, got2 := getTags(tt.input); !reflect.DeepEqual(got1, tt.want[0]) || !reflect.DeepEqual(got2, tt.want[1]) {
			t.Errorf("getTags(%v) = %v, %v, want %v, %v", tt.input, got1, got2, tt.want[0], tt.want[1])
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
		uniqueKeys    []string
	}
	tests := []struct {
		input          input
		wantValues     []interface{}
		wantUpsertTags []string
	}{
		{
			input: input{
				objs:          bulkData,
				tags:          []string{"", "name", "age", ""},
				availableTags: []string{"name", "age"},
				uniqueKeys:    nil,
			},
			wantValues: []interface{}{
				reflect.ValueOf(insertData[0].UserName).Interface(), reflect.ValueOf(insertData[0].Age).Interface(),
				reflect.ValueOf(insertData[1].UserName).Interface(), reflect.ValueOf(insertData[1].Age).Interface(),
			},
			wantUpsertTags: nil,
		},
		{
			input: input{
				objs:          bulkData,
				tags:          []string{"", "name", "age", ""},
				availableTags: []string{"name", "age"},
				uniqueKeys:    []string{"name"},
			},
			wantValues: []interface{}{
				reflect.ValueOf(insertData[0].UserName).Interface(), reflect.ValueOf(insertData[0].Age).Interface(), reflect.ValueOf(insertData[0].Age).Interface(),
				reflect.ValueOf(insertData[1].UserName).Interface(), reflect.ValueOf(insertData[1].Age).Interface(), reflect.ValueOf(insertData[1].Age).Interface(),
			},
			wantUpsertTags: []string{"age"},
		},
	}
	for _, tt := range tests {
		if gotValues, gotTags := sliceValues(tt.input.objs, tt.input.tags, tt.input.availableTags, tt.input.uniqueKeys); !reflect.DeepEqual(gotValues, tt.wantValues) || !reflect.DeepEqual(gotTags, tt.wantUpsertTags) {
			t.Errorf("sliceValues(%v, %v, %v, %v) = %v, %v, want %v, %v",
				tt.input.objs, tt.input.tags, tt.input.availableTags, tt.input.uniqueKeys, gotValues, gotTags, tt.wantValues, tt.wantUpsertTags)
		}
	}
}

func Test_isZeroOfUnderlyingType(t *testing.T) {
	tests := []struct {
		input interface{}
		want  bool
	}{
		{
			input: 0,
			want:  true,
		},
		{
			input: "",
			want:  true,
		},
		{
			input: 42,
			want:  false,
		},
		{
			input: "foo",
			want:  false,
		},
		{
			input: 0.0,
			want:  true,
		},
		{
			input: 0.1,
			want:  false,
		},
		{
			input: time.Now(),
			want:  false,
		},
	}

	for _, tt := range tests {
		if got := isZeroOfUnderlyingType(tt.input); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("isZeroOfUnderlyingType(%v) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
