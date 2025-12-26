package config

import (
	"os"
	"testing"
)

func TestAuthDisabled(t *testing.T) {
	tests := []struct {
		name              string
		envVars           map[string]string
		expectError       bool
		expectAuthEnabled bool
		expectDisabled    bool
	}{
		{
			name: "Default behavior - auth enabled with password",
			envVars: map[string]string{
				"SNIPO_MASTER_PASSWORD": "test123",
			},
			expectError:       false,
			expectAuthEnabled: true,
			expectDisabled:    false,
		},
		{
			name: "Auth disabled - no password required",
			envVars: map[string]string{
				"SNIPO_DISABLE_AUTH": "true",
			},
			expectError:       false,
			expectAuthEnabled: false,
			expectDisabled:    true,
		},
		{
			name: "Auth disabled - password ignored",
			envVars: map[string]string{
				"SNIPO_DISABLE_AUTH":    "true",
				"SNIPO_MASTER_PASSWORD": "test123",
			},
			expectError:       false,
			expectAuthEnabled: false,
			expectDisabled:    true,
		},
		{
			name: "Auth enabled without password - should error",
			envVars: map[string]string{
				"SNIPO_DISABLE_AUTH": "false",
			},
			expectError:       true,
			expectAuthEnabled: true,
			expectDisabled:    false,
		},
		{
			name: "Auth disabled with hash - hash ignored",
			envVars: map[string]string{
				"SNIPO_DISABLE_AUTH":          "true",
				"SNIPO_MASTER_PASSWORD_HASH": "$argon2id$salt$hash",
			},
			expectError:       false,
			expectAuthEnabled: false,
			expectDisabled:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all relevant env vars
			_ = os.Unsetenv("SNIPO_MASTER_PASSWORD")
			_ = os.Unsetenv("SNIPO_MASTER_PASSWORD_HASH")
			_ = os.Unsetenv("SNIPO_DISABLE_AUTH")
			_ = os.Unsetenv("SNIPO_SESSION_SECRET")

			// Set test env vars
			for k, v := range tt.envVars {
				_ = os.Setenv(k, v)
			}
			// Always set session secret to avoid auto-generation
			_ = os.Setenv("SNIPO_SESSION_SECRET", "test-session-secret-32chars!!")

			cfg, err := Load()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if cfg.Auth.Disabled != tt.expectDisabled {
				t.Errorf("Expected Auth.Disabled=%v, got %v", tt.expectDisabled, cfg.Auth.Disabled)
			}

			if tt.expectDisabled {
				// When auth is disabled, password fields should be cleared
				if cfg.Auth.MasterPassword != "" || cfg.Auth.MasterPasswordHash != "" {
					t.Errorf("Expected passwords to be cleared when auth is disabled")
				}
			}
		})
	}

	// Clean up
	_ = os.Unsetenv("SNIPO_MASTER_PASSWORD")
	_ = os.Unsetenv("SNIPO_MASTER_PASSWORD_HASH")
	_ = os.Unsetenv("SNIPO_DISABLE_AUTH")
	_ = os.Unsetenv("SNIPO_SESSION_SECRET")
}

func TestAuthDisabledBackwardCompatibility(t *testing.T) {
	// Test that without SNIPO_DISABLE_AUTH set, behavior is the same as before

	// Clear env
	_ = os.Unsetenv("SNIPO_MASTER_PASSWORD")
	_ = os.Unsetenv("SNIPO_MASTER_PASSWORD_HASH")
	_ = os.Unsetenv("SNIPO_DISABLE_AUTH")
	_ = os.Unsetenv("SNIPO_SESSION_SECRET")

	// Set minimal config (old way)
	_ = os.Setenv("SNIPO_MASTER_PASSWORD", "test123")
	_ = os.Setenv("SNIPO_SESSION_SECRET", "test-secret-key-123456789012")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Backward compatibility broken: %v", err)
	}

	if cfg.Auth.Disabled {
		t.Error("Auth should be enabled by default (backward compatibility)")
	}

	if cfg.Auth.MasterPassword == "" {
		t.Error("Master password should be set")
	}

	// Clean up
	_ = os.Unsetenv("SNIPO_MASTER_PASSWORD")
	_ = os.Unsetenv("SNIPO_SESSION_SECRET")
}
