package bulkinsert

import (
	"reflect"
	"testing"
)

func Test_toSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"Name", "name",
		},
		{
			"UserName", "user_name",
		},
		{
			"UUID", "uuid",
		},
		{
			"HTMLFile", "html_file",
		},
	}

	for _, tt := range tests {
		if got := toSnakeCase(tt.input); !reflect.DeepEqual(got, tt.expected) {
			t.Errorf("toSnakeCase() = %v, want %v", got, tt.expected)
		}
	}
}
