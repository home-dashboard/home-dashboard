package nodejs

import (
	"golang.org/x/mod/semver"
	"testing"
)

func TestInstaller_ResolvePath(t *testing.T) {
	type args struct {
		version  string
		platform string
		cpu      string
	}
	tests := []struct {
		name    string
		r       *Installer
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "windows",
			r:    &Installer{},
			args: args{
				version:  "12.18.3",
				platform: "windows",
				cpu:      "amd64",
			},
			want:    "node-v12.18.3-windows-x64.zip",
			wantErr: false,
		},
		{
			name: "linux",
			r:    &Installer{},
			args: args{
				version:  "12.18.3",
				platform: "linux",
				cpu:      "amd64",
			},
			want:    "node-v12.18.3-linux-x64.tar.gz",
			wantErr: false,
		},
		{
			name: "darwin",
			r:    &Installer{},
			args: args{
				version:  "12.18.3",
				platform: "darwin",
				cpu:      "amd64",
			},
			want:    "node-v12.18.3-darwin-x64.tar.gz",
			wantErr: false,
		},
		{
			name: "unsupported platform",
			r:    &Installer{},
			args: args{
				version:  "12.18.3",
				platform: "unsupported",
				cpu:      "amd64",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "unsupported cpu",
			r:    &Installer{},
			args: args{
				version:  "12.18.3",
				platform: "linux",
				cpu:      "unsupported",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "unsupported cpu",
			r:    &Installer{},
			args: args{
				version:  "12.18.3",
				platform: "linux",
				cpu:      "unsupported",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "unsupported cpu",
			r:    &Installer{},
			args: args{
				version: "12.18.3",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "unsupported cpu",
			r:    &Installer{},
			args: args{
				version: "12.18.3",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Installer{}
			got, err := r.ResolvePath(tt.args.version, tt.args.platform, tt.args.cpu)
			if (err != nil) != tt.wantErr {
				t.Errorf("Installer.ResolvePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf(
					"Installer.ResolvePath() = %v, want %v",
					got,
					tt.want,
				)
			}
		})
	}
}

func TestInstaller_Version(t *testing.T) {
	r := &Installer{
		MirrorURL:     "https://nodejs.org",
		WorkDirectory: "cron_service/nodejs",
	}

	if !semver.IsValid(r.Version()) {
		t.Errorf("not valid %s", r.Version())
	}
}
