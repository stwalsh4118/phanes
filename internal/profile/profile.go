package profile

import (
	"fmt"
	"sort"
)

// profiles defines the available server profiles as maps of module name lists.
// Each profile represents a common server configuration that combines multiple
// modules to achieve a specific server setup.
var profiles = map[string][]string{
	"minimal": {
		"baseline",
		"user",
		"security",
		"swap",
		"updates",
	},
	"dev": {
		"baseline",
		"user",
		"security",
		"swap",
		"updates",
		"docker",
		"monitoring",
		"devtools",
	},
	"web": {
		"baseline",
		"user",
		"security",
		"swap",
		"updates",
		"docker",
		"monitoring",
		"caddy",
	},
	"database": {
		"baseline",
		"user",
		"security",
		"swap",
		"updates",
		"docker",
		"monitoring",
		"postgres",
		"redis",
	},
	"coolify": {
		"baseline",
		"user",
		"security",
		"swap",
		"updates",
		"docker",
		"coolify",
	},
}

// GetProfile returns the list of module names for the specified profile.
// Returns an error if the profile does not exist.
//
// Example:
//
//	modules, err := profile.GetProfile("dev")
//	if err != nil {
//	    log.Error("Profile not found: %v", err)
//	    return
//	}
//	// modules = ["baseline", "user", "security", "swap", "updates", "docker", "monitoring", "devtools"]
func GetProfile(name string) ([]string, error) {
	modules, exists := profiles[name]
	if !exists {
		return nil, fmt.Errorf("profile %s not found", name)
	}
	// Return a copy to prevent external modification
	result := make([]string, len(modules))
	copy(result, modules)
	return result, nil
}

// ListProfiles returns a sorted list of all available profile names.
//
// Example:
//
//	profiles := profile.ListProfiles()
//	// profiles = ["coolify", "database", "dev", "minimal", "web"]
func ListProfiles() []string {
	names := make([]string, 0, len(profiles))
	for name := range profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ProfileExists checks if a profile with the given name exists.
// Returns true if the profile exists, false otherwise.
//
// Example:
//
//	if profile.ProfileExists("dev") {
//	    log.Info("Dev profile is available")
//	}
func ProfileExists(name string) bool {
	_, exists := profiles[name]
	return exists
}

