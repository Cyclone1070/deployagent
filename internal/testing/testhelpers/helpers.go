// Package testhelpers provides shared utilities for integration testing
package testhelpers

import (
	"testing"

	"github.com/Cyclone1070/iav/internal/testing/mocks"
)

// Re-export types for backward compatibility
type MockProvider = mocks.MockProvider
type MockUI = mocks.MockUI

var NewMockProvider = mocks.NewMockProvider
var NewMockUI = mocks.NewMockUI

// CreateTestWorkspace creates a temporary workspace for integration tests
func CreateTestWorkspace(t *testing.T) string {
	return t.TempDir()
}
