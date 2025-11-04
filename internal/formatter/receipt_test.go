package formatter

import (
	"slices"
	"testing"
)

// TestReceiptFieldDefinitionsConsistency ensures that allReceiptFieldNames,
// defaultReceiptFieldNames, and allFields are consistent with each other.
func TestReceiptFieldDefinitionsConsistency(t *testing.T) {
	t.Run("allReceiptFieldNames contains all fields in allFields", func(t *testing.T) {
		for fieldName := range allFields {
			if !slices.Contains(allReceiptFieldNames, fieldName) {
				t.Errorf("field %q exists in allFields but not in allReceiptFieldNames", fieldName)
			}
		}
	})

	t.Run("allFields contains all fields in allReceiptFieldNames", func(t *testing.T) {
		for _, fieldName := range allReceiptFieldNames {
			if _, ok := allFields[fieldName]; !ok {
				t.Errorf("field %q exists in allReceiptFieldNames but not in allFields", fieldName)
			}
		}
	})

	t.Run("defaultReceiptFieldNames is a subset of allReceiptFieldNames", func(t *testing.T) {
		for _, fieldName := range defaultReceiptFieldNames {
			if !slices.Contains(allReceiptFieldNames, fieldName) {
				t.Errorf("field %q exists in defaultReceiptFieldNames but not in allReceiptFieldNames", fieldName)
			}
		}
	})

	t.Run("allFields contains all fields in defaultReceiptFieldNames", func(t *testing.T) {
		for _, fieldName := range defaultReceiptFieldNames {
			if _, ok := allFields[fieldName]; !ok {
				t.Errorf("field %q exists in defaultReceiptFieldNames but not in allFields", fieldName)
			}
		}
	})

	t.Run("all fieldDef entries have required fields set", func(t *testing.T) {
		for fieldName, fd := range allFields {
			if fd.Header == "" {
				t.Errorf("field %q has empty Header", fieldName)
			}
			if fd.Extractor == nil {
				t.Errorf("field %q has nil Extractor", fieldName)
			}
		}
	})
}

// TestReceiptFieldDefinitionsNoDuplicates ensures there are no duplicate entries.
func TestReceiptFieldDefinitionsNoDuplicates(t *testing.T) {
	t.Run("allReceiptFieldNames has no duplicates", func(t *testing.T) {
		seen := make(map[string]bool)
		for _, fieldName := range allReceiptFieldNames {
			if seen[fieldName] {
				t.Errorf("duplicate field %q found in allReceiptFieldNames", fieldName)
			}
			seen[fieldName] = true
		}
	})

	t.Run("defaultReceiptFieldNames has no duplicates", func(t *testing.T) {
		seen := make(map[string]bool)
		for _, fieldName := range defaultReceiptFieldNames {
			if seen[fieldName] {
				t.Errorf("duplicate field %q found in defaultReceiptFieldNames", fieldName)
			}
			seen[fieldName] = true
		}
	})
}
