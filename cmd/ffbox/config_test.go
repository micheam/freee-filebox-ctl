package main

import (
	"os"
	"testing"
)

func TestSelectEditor(t *testing.T) {
	tests := []struct {
		name          string
		visual        string
		editor        string
		wantEditor    string
		setupEnv      func()
		cleanupEnv    func()
	}{
		{
			name:       "VISUAL takes precedence over EDITOR",
			visual:     "vim",
			editor:     "nano",
			wantEditor: "vim",
		},
		{
			name:       "EDITOR is used when VISUAL is not set",
			visual:     "",
			editor:     "nano",
			wantEditor: "nano",
		},
		{
			name:       "vi is used when neither VISUAL nor EDITOR is set",
			visual:     "",
			editor:     "",
			wantEditor: "vi",
		},
		{
			name:       "VISUAL with empty string falls back to EDITOR",
			visual:     "",
			editor:     "emacs",
			wantEditor: "emacs",
		},
		{
			name:       "Both set to same value returns that value",
			visual:     "code",
			editor:     "code",
			wantEditor: "code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment variables
			originalVisual := os.Getenv("VISUAL")
			originalEditor := os.Getenv("EDITOR")

			// Cleanup function to restore original environment
			defer func() {
				if originalVisual != "" {
					os.Setenv("VISUAL", originalVisual)
				} else {
					os.Unsetenv("VISUAL")
				}
				if originalEditor != "" {
					os.Setenv("EDITOR", originalEditor)
				} else {
					os.Unsetenv("EDITOR")
				}
			}()

			// Set up test environment
			if tt.visual != "" {
				os.Setenv("VISUAL", tt.visual)
			} else {
				os.Unsetenv("VISUAL")
			}

			if tt.editor != "" {
				os.Setenv("EDITOR", tt.editor)
			} else {
				os.Unsetenv("EDITOR")
			}

			// Run custom setup if provided
			if tt.setupEnv != nil {
				tt.setupEnv()
			}

			// Test the function
			got := selectEditor()
			if got != tt.wantEditor {
				t.Errorf("selectEditor() = %q, want %q", got, tt.wantEditor)
			}

			// Run custom cleanup if provided
			if tt.cleanupEnv != nil {
				tt.cleanupEnv()
			}
		})
	}
}
