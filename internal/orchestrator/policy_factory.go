package orchestrator

import (
	"github.com/Cyclone1070/iav/internal/config"
	"github.com/Cyclone1070/iav/internal/orchestrator/models"
)

// NewPolicy creates a Policy with values from config.
// It initializes all maps to ensure thread safety.
func NewPolicy(cfg *config.Config) *models.Policy {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	return &models.Policy{
		Shell: models.ShellPolicy{
			Allow:        cfg.Policy.ShellAllow,
			Deny:         cfg.Policy.ShellDeny,
			SessionAllow: make(map[string]bool),
		},
		Tools: models.ToolPolicy{
			Allow:        cfg.Policy.ToolsAllow,
			Deny:         cfg.Policy.ToolsDeny,
			SessionAllow: make(map[string]bool),
		},
	}
}
