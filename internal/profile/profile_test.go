package profile

import (
	"reflect"
	"testing"
)

func TestGetProfile(t *testing.T) {
	tests := []struct {
		name        string
		profileName string
		want        []string
		wantErr     bool
	}{
		{
			name:        "minimal profile",
			profileName: "minimal",
			want:        []string{"baseline", "user", "security", "swap", "updates"},
			wantErr:     false,
		},
		{
			name:        "dev profile",
			profileName: "dev",
			want:        []string{"baseline", "user", "security", "swap", "updates", "docker", "monitoring", "devtools"},
			wantErr:     false,
		},
		{
			name:        "web profile",
			profileName: "web",
			want:        []string{"baseline", "user", "security", "swap", "updates", "docker", "monitoring", "caddy"},
			wantErr:     false,
		},
		{
			name:        "database profile",
			profileName: "database",
			want:        []string{"baseline", "user", "security", "swap", "updates", "docker", "monitoring", "postgres", "redis"},
			wantErr:     false,
		},
		{
			name:        "coolify profile",
			profileName: "coolify",
			want:        []string{"baseline", "user", "security", "swap", "updates", "docker", "coolify"},
			wantErr:     false,
		},
		{
			name:        "non-existent profile",
			profileName: "nonexistent",
			want:        nil,
			wantErr:     true,
		},
		{
			name:        "empty profile name",
			profileName: "",
			want:        nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetProfile(tt.profileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetProfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetProfile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetProfileReturnsCopy(t *testing.T) {
	// Verify that GetProfile returns a copy, not the original slice
	modules1, err := GetProfile("minimal")
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}

	modules2, err := GetProfile("minimal")
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}

	// Modify the first result
	modules1[0] = "modified"

	// The second result should not be affected
	if modules2[0] == "modified" {
		t.Error("GetProfile() returned the same slice reference, expected a copy")
	}
}

func TestListProfiles(t *testing.T) {
	got := ListProfiles()

	// Verify all expected profiles are present
	expectedProfiles := map[string]bool{
		"minimal":  true,
		"dev":      true,
		"web":      true,
		"database": true,
		"coolify":  true,
	}

	if len(got) != len(expectedProfiles) {
		t.Errorf("ListProfiles() returned %d profiles, want %d", len(got), len(expectedProfiles))
	}

	for _, name := range got {
		if !expectedProfiles[name] {
			t.Errorf("ListProfiles() returned unexpected profile: %s", name)
		}
	}

	// Verify the list is sorted
	for i := 1; i < len(got); i++ {
		if got[i-1] > got[i] {
			t.Errorf("ListProfiles() returned unsorted list: %v", got)
			break
		}
	}
}

func TestProfileExists(t *testing.T) {
	tests := []struct {
		name        string
		profileName string
		want        bool
	}{
		{
			name:        "minimal exists",
			profileName: "minimal",
			want:        true,
		},
		{
			name:        "dev exists",
			profileName: "dev",
			want:        true,
		},
		{
			name:        "web exists",
			profileName: "web",
			want:        true,
		},
		{
			name:        "database exists",
			profileName: "database",
			want:        true,
		},
		{
			name:        "coolify exists",
			profileName: "coolify",
			want:        true,
		},
		{
			name:        "non-existent profile",
			profileName: "nonexistent",
			want:        false,
		},
		{
			name:        "empty profile name",
			profileName: "",
			want:        false,
		},
		{
			name:        "case sensitive check",
			profileName: "DEV",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ProfileExists(tt.profileName); got != tt.want {
				t.Errorf("ProfileExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAllProfilesAreValid(t *testing.T) {
	// Verify that all profiles in the map can be retrieved
	allProfiles := ListProfiles()
	for _, profileName := range allProfiles {
		modules, err := GetProfile(profileName)
		if err != nil {
			t.Errorf("Profile %s exists in map but GetProfile() returned error: %v", profileName, err)
		}
		if len(modules) == 0 {
			t.Errorf("Profile %s has no modules", profileName)
		}
	}
}
