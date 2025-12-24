package search

import (
	"testing"

	"github.com/Cyclone1070/iav/internal/config"
)

func TestSearchContentRequest_Validation(t *testing.T) {
	cfg := config.DefaultConfig()

	tests := []struct {
		name         string
		req          SearchContentRequest
		wantErr      error
		verifyValues func(t *testing.T, req SearchContentRequest)
	}{
		{"Valid", SearchContentRequest{Query: "foo"}, nil, nil},
		{"EmptyQuery", SearchContentRequest{Query: ""}, ErrQueryRequired, nil},
		{"NegativeOffset_Clamps", SearchContentRequest{Query: "foo", Offset: -1}, nil, func(t *testing.T, req SearchContentRequest) {
			if req.Offset != 0 {
				t.Errorf("expected offset 0, got %d", req.Offset)
			}
		}},
		{"NegativeLimit_Defaults", SearchContentRequest{Query: "foo", Limit: -1}, nil, func(t *testing.T, req SearchContentRequest) {
			if req.Limit != cfg.Tools.DefaultSearchContentLimit {
				t.Errorf("expected default limit %d, got %d", cfg.Tools.DefaultSearchContentLimit, req.Limit)
			}
		}},
		{"LimitExceedsMax_Caps", SearchContentRequest{Query: "foo", Limit: cfg.Tools.MaxSearchContentLimit + 1}, nil, func(t *testing.T, req SearchContentRequest) {
			if req.Limit != cfg.Tools.MaxSearchContentLimit {
				t.Errorf("expected max limit %d, got %d", cfg.Tools.MaxSearchContentLimit, req.Limit)
			}
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate(cfg)
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("Validate() error = %v, want nil", err)
				}
			} else {
				if err == nil {
					t.Errorf("Validate() error = nil, want %v", tt.wantErr)
				}
			}
			if err == nil && tt.verifyValues != nil {
				tt.verifyValues(t, tt.req)
			}
		})
	}
}
