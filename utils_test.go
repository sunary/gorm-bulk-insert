package bulk

import (
	"reflect"
	"testing"
)

func Test_toSnakeCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
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
		if got := toSnakeCase(tt.input); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("toSnakeCase(%v) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
