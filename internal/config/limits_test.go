package config

import (
	"testing"
)

func TestDefaultConfig_SearchContentLimits(t *testing.T) {
	cfg := DefaultConfig()

	// Verify new search limits have sensible defaults
	if cfg.Tools.DefaultSearchContentLimit == 0 {
		t.Error("DefaultSearchContentLimit should not be zero")
	}
	if cfg.Tools.MaxSearchContentLimit == 0 {
		t.Error("MaxSearchContentLimit should not be zero")
	}
	if cfg.Tools.DefaultSearchContentLimit > cfg.Tools.MaxSearchContentLimit {
		t.Error("DefaultSearchContentLimit should not exceed MaxSearchContentLimit")
	}
}

func TestDefaultConfig_FindFileLimits(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Tools.DefaultFindFileLimit == 0 {
		t.Error("DefaultFindFileLimit should not be zero")
	}
	if cfg.Tools.MaxFindFileLimit == 0 {
		t.Error("MaxFindFileLimit should not be zero")
	}
	if cfg.Tools.DefaultFindFileLimit > cfg.Tools.MaxFindFileLimit {
		t.Error("DefaultFindFileLimit should not exceed MaxFindFileLimit")
	}
}

func TestMergeConfig_SearchContentLimits(t *testing.T) {
	t.Run("override DefaultSearchContentLimit", func(t *testing.T) {
		dst := DefaultConfig()
		src := &Config{Tools: ToolsConfig{DefaultSearchContentLimit: 50}}
		// internal/config/loader.go: mergeConfig is not exported but logic is inside loader
		// However, we can't test unexported function easily if we are in 'config_test' package.
		// Since we are in 'config' package (package config), we can access mergeConfig if it is in same package.
		// Let's assume mergeConfig is unexported but accessible since we are in package config.
		mergeConfig(dst, src)
		if dst.Tools.DefaultSearchContentLimit != 50 {
			t.Errorf("expected 50, got %d", dst.Tools.DefaultSearchContentLimit)
		}
	})

	t.Run("override MaxSearchContentLimit", func(t *testing.T) {
		dst := DefaultConfig()
		src := &Config{Tools: ToolsConfig{MaxSearchContentLimit: 500}}
		mergeConfig(dst, src)
		if dst.Tools.MaxSearchContentLimit != 500 {
			t.Errorf("expected 500, got %d", dst.Tools.MaxSearchContentLimit)
		}
	})

	t.Run("zero values do not override", func(t *testing.T) {
		dst := DefaultConfig()
		originalDefault := dst.Tools.DefaultSearchContentLimit
		src := &Config{Tools: ToolsConfig{DefaultSearchContentLimit: 0}}
		mergeConfig(dst, src)
		if dst.Tools.DefaultSearchContentLimit != originalDefault {
			t.Errorf("zero should not override, expected %d, got %d", originalDefault, dst.Tools.DefaultSearchContentLimit)
		}
	})
}

func TestMergeConfig_FindFileLimits(t *testing.T) {
	t.Run("override DefaultFindFileLimit", func(t *testing.T) {
		dst := DefaultConfig()
		src := &Config{Tools: ToolsConfig{DefaultFindFileLimit: 50}}
		mergeConfig(dst, src)
		if dst.Tools.DefaultFindFileLimit != 50 {
			t.Errorf("expected 50, got %d", dst.Tools.DefaultFindFileLimit)
		}
	})

	t.Run("override MaxFindFileLimit", func(t *testing.T) {
		dst := DefaultConfig()
		src := &Config{Tools: ToolsConfig{MaxFindFileLimit: 500}}
		mergeConfig(dst, src)
		if dst.Tools.MaxFindFileLimit != 500 {
			t.Errorf("expected 500, got %d", dst.Tools.MaxFindFileLimit)
		}
	})

	t.Run("zero values do not override", func(t *testing.T) {
		dst := DefaultConfig()
		originalDefault := dst.Tools.DefaultFindFileLimit
		src := &Config{Tools: ToolsConfig{DefaultFindFileLimit: 0}}
		mergeConfig(dst, src)
		if dst.Tools.DefaultFindFileLimit != originalDefault {
			t.Errorf("zero should not override")
		}
	})
}
