package ruby

import (
	"reflect"
	"testing"
)

func Test_getGemlockFiles(t *testing.T) {
	tests := []struct {
		name     string
		repoPath string
		want     []string
		wantErr  bool
	}{
		{
			name:     "",
			repoPath: "/Users/lpusok/Develop/docker-for-bitrise-web",
			want:     nil,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getGemlockFiles(tt.repoPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("getGemDeps() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getGemDeps() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseLockfile(t *testing.T) {
	tests := []struct {
		name        string
		gemlockPath string
		want        []string
		wantErr     bool
	}{
		{
			name:        "",
			gemlockPath: "/Users/lpusok/Develop/docker-for-bitrise-web/src/Gemfile.lock",
			want:        nil,
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseLockfile(tt.gemlockPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLockfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseLockfile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getGemDeps(t *testing.T) {
	tests := []struct {
		name     string
		repoPath string
		want     map[string]string
		wantErr  bool
	}{
		{
			name:     "",
			repoPath: "/Users/lpusok/Develop/docker-for-bitrise-web",
			want:     nil,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got, err := GetGemDeps(tt.repoPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("getGemDeps() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getGemDeps() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getLicensesForGem(t *testing.T) {
	tests := []struct {
		name    string
		gem     string
		want    []string
		wantErr bool
	}{
		{
			name:    "",
			gem:     "actionmailer",
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getLicensesForGem(tt.gem)
			if (err != nil) != tt.wantErr {
				t.Errorf("getLicensesForGem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getLicensesForGem() = %v, want %v", got, tt.want)
			}
		})
	}
}
