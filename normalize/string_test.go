package normalize

import "testing"

func TestString(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "normal string",
			args:    args{s: "word"},
			want:    "word",
			wantErr: false,
		},
		{
			name:    "string with white suffix and prefix",
			args:    args{s: " word "},
			want:    "word",
			wantErr: false,
		},
		{
			name:    "empty string",
			args:    args{s: ""},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := String(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("String() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("String() got = %v, want %v", got, tt.want)
			}
		})
	}
}
