package tools

import (
	"github.com/Cyclone1070/iav/internal/config"
	"github.com/Cyclone1070/iav/internal/tools/model"
	"github.com/Cyclone1070/iav/internal/tools/service"
)

// NewWorkspaceContext returns a default workspace context with system implementations.
// The workspaceRoot is canonicalised (absolute and symlink-resolved).
// Each context gets its own checksum cache instance and file size limit from config.
func NewWorkspaceContext(cfg *config.Config, workspaceRoot string) (*model.WorkspaceContext, error) {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	canonicalRoot, err := service.CanonicaliseRoot(workspaceRoot)
	if err != nil {
		return nil, err
	}

	fs := service.NewOSFileSystem()

	// Initialize gitignore service (handles missing .gitignore gracefully)
	gitignoreSvc, err := service.NewGitignoreService(canonicalRoot, fs)
	if err != nil {
		// MVP DEFERRAL: Intentionally silent fallback for now.
		// TODO(logging): Add slog.Warn("gitignore initialization failed", "error", err) when logging is set up.
		gitignoreSvc = &service.NoOpGitignoreService{}
	}

	return &model.WorkspaceContext{
		Config:          *cfg,
		FS:              fs,
		BinaryDetector:  &service.SystemBinaryDetector{SampleSize: cfg.Tools.BinaryDetectionSampleSize},
		ChecksumManager: service.NewChecksumManager(),

		WorkspaceRoot:    canonicalRoot,
		GitignoreService: gitignoreSvc,
		CommandExecutor:  &service.OSCommandExecutor{},

		TodoStore: NewInMemoryTodoStore(),
		DockerConfig: model.DockerConfig{
			CheckCommand: []string{"docker", "info"},
			// TODO(cross-platform): MacOS-specific Docker commands. Linux uses systemctl, Windows uses Start-Service.
			StartCommand: []string{"docker", "desktop", "start"},
		},
	}, nil
}
