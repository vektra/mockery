package template_funcs

import "testing"

func TestFirstIsLower(t *testing.T) {
	tests := []struct {
		arg  string
		want bool
	}{
		{
			arg:  "Exported",
			want: false,
		},
		{
			arg:  "unexported",
			want: true,
		},
		{
			arg:  "1234",
			want: false,
		},
		{
			arg:  "MockargGetter",
			want: false,
		},
		{
			arg:  "",
			want: false,
		},
		{
			arg:  "état",
			want: true,
		},
		{
			arg:  "État",
			want: false,
		},
		{
			arg:  "πValue",
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			if got := FirstIsLower(tt.arg); got != tt.want {
				t.Errorf("FirstIsLower() = %v, want %v", got, tt.want)
			}
		})
	}
}
