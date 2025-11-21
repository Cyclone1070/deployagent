package services

import (
	"testing"

	"github.com/Cyclone1070/deployforme/internal/tools/models"
)

func TestEvaluatePolicy(t *testing.T) {
	tests := []struct {
		name    string
		policy  models.CommandPolicy
		command []string
		wantErr error
	}{
		{
			name:    "Allowed command",
			policy:  models.CommandPolicy{Allow: []string{"echo"}},
			command: []string{"echo", "hello"},
			wantErr: nil,
		},
		{
			name:    "Allowed via SessionAllow",
			policy:  models.CommandPolicy{SessionAllow: map[string]bool{"rm": true}},
			command: []string{"rm", "-rf", "/"},
			wantErr: nil,
		},
		{
			name:    "Ask list - Approval Required",
			policy:  models.CommandPolicy{Ask: []string{"deploy"}},
			command: []string{"deploy", "prod"},
			wantErr: models.ErrShellApprovalRequired,
		},
		{
			name:    "Ask list - Session Allowed",
			policy:  models.CommandPolicy{Ask: []string{"deploy"}, SessionAllow: map[string]bool{"deploy": true}},
			command: []string{"deploy", "prod"},
			wantErr: nil,
		},
		{
			name:    "Default Deny",
			policy:  models.CommandPolicy{},
			command: []string{"unknown"},
			wantErr: models.ErrShellRejected,
		},
		{
			name:    "Empty Command",
			policy:  models.CommandPolicy{},
			command: []string{},
			wantErr: models.ErrShellRejected, // Or specific error if parser fails first
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := EvaluatePolicy(tt.policy, tt.command)
			if err != tt.wantErr {
				t.Errorf("EvaluatePolicy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
