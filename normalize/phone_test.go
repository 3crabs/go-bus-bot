package normalize

import "testing"

func TestPhone(t *testing.T) {
	type args struct {
		phone string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "normal phone",
			args:    args{phone: "+7-000-000-00-00"},
			want:    "+7-000-000-00-00",
			wantErr: false,
		},
		{
			name:    "empty phone",
			args:    args{phone: ""},
			want:    "",
			wantErr: true,
		},
		{
			name:    "phone without +-",
			args:    args{phone: "71234567890"},
			want:    "+7-123-456-78-90",
			wantErr: false,
		},
		{
			name:    "phone without +- and start with 8",
			args:    args{phone: "81234567890"},
			want:    "+7-123-456-78-90",
			wantErr: false,
		},
		{
			name:    "phone without +- with len 10",
			args:    args{phone: "1234567890"},
			want:    "+7-123-456-78-90",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Phone(tt.args.phone)
			if (err != nil) != tt.wantErr {
				t.Errorf("Phone() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Phone() got = %v, want %v", got, tt.want)
			}
		})
	}
}
