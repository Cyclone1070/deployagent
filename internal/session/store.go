package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/Cyclone1070/iav/internal/config"
	"github.com/Cyclone1070/iav/internal/provider"
	"github.com/google/uuid"
)

// Store manages session creation, loading, and listing.
type Store struct {
	storageDir string
}

// NewStore creates a new session store.
func NewStore(cfg *config.Config) *Store {
	return &Store{storageDir: cfg.Session.StorageDir}
}

// NewSession creates a new session with a unique ID.
func (st *Store) NewSession() (*Session, error) {
	if err := os.MkdirAll(st.storageDir, 0755); err != nil {
		return nil, fmt.Errorf("create storage dir: %w", err)
	}
	s := &Session{
		id:         uuid.New().String(),
		messages:   []provider.Message{},
		storageDir: st.storageDir,
	}
	if err := s.Save(); err != nil {
		return nil, err
	}
	return s, nil
}

// LoadSession loads a session from disk by ID.
func (st *Store) LoadSession(id string) (*Session, error) {
	path := filepath.Join(st.storageDir, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read session file: %w", err)
	}

	var dto sessionDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}
	return &Session{
		id:         dto.ID,
		messages:   dto.Messages,
		storageDir: st.storageDir,
	}, nil
}

// ListSessions returns all session IDs sorted by modification time (newest first).
func (st *Store) ListSessions() ([]string, error) {
	entries, err := os.ReadDir(st.storageDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("read storage dir: %w", err)
	}

	type sessionInfo struct {
		id    string
		mtime time.Time
	}
	var infos []sessionInfo

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		infos = append(infos, sessionInfo{
			id:    entry.Name()[:len(entry.Name())-5], // strip .json
			mtime: info.ModTime(),
		})
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].mtime.After(infos[j].mtime)
	})

	ids := make([]string, len(infos))
	for i, info := range infos {
		ids[i] = info.id
	}
	return ids, nil
}
