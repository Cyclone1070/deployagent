package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Cyclone1070/iav/internal/tools/model"
	"github.com/Cyclone1070/iav/internal/tools/service"
)

// SearchContent searches for content matching a regex pattern using ripgrep.
// It validates the search path is within workspace boundaries, respects gitignore rules
// (unless includeIgnored is true), and returns matches with pagination support.
func SearchContent(ctx context.Context, wCtx *model.WorkspaceContext, req model.SearchContentRequest) (*model.SearchContentResponse, error) {
	// Resolve search path
	absSearchPath, _, err := service.Resolve(wCtx, req.SearchPath)
	if err != nil {
		return nil, err
	}

	// Check if search path exists
	info, err := wCtx.FS.Stat(absSearchPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, model.ErrFileMissing
		}
		return nil, fmt.Errorf("failed to stat search path: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("search path is not a directory")
	}

	// Use configured limits - Validate() already checked bounds
	limit := wCtx.Config.Tools.DefaultSearchContentLimit
	if req.Limit != 0 {
		limit = req.Limit
	}
	offset := req.Offset

	maxResults := wCtx.Config.Tools.MaxSearchContentResults

	// Hard limit on line length to avoid memory issues
	maxLineLength := wCtx.Config.Tools.MaxLineLength

	// 10MB default for crazy long lines (minified JS etc)
	maxScanTokenSize := wCtx.Config.Tools.MaxScanTokenSize

	// Configure scanner buffer
	initialBufSize := wCtx.Config.Tools.InitialScannerBufferSize

	// Build ripgrep command
	// rg --json "query" searchPath [--no-ignore]
	cmd := []string{"rg", "--json"}
	if !req.CaseSensitive {
		cmd = append(cmd, "-i")
	}
	if req.IncludeIgnored {
		cmd = append(cmd, "--no-ignore")
	}
	cmd = append(cmd, req.Query, absSearchPath)

	// Execute command with streaming
	proc, stdout, _, err := wCtx.CommandExecutor.Start(ctx, cmd, model.ProcessOptions{Dir: absSearchPath})
	if err != nil {
		return nil, fmt.Errorf("failed to start rg command: %w", err)
	}
	// process will be waited on explicitly later

	// Stream and process JSON output line by line
	var matches []model.SearchContentMatch
	scanner := bufio.NewScanner(stdout)
	// Increase buffer size to handle very long lines (e.g. minified JS)
	buf := make([]byte, 0, initialBufSize)
	scanner.Buffer(buf, maxScanTokenSize)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Parse JSON line
		var rgMatch struct {
			Type string `json:"type"`
			Data struct {
				Path struct {
					Text string `json:"text"`
				} `json:"path"`
				Lines struct {
					Text string `json:"text"`
				} `json:"lines"`
				LineNumber int `json:"line_number"`
			} `json:"data"`
		}

		if err := json.Unmarshal([]byte(line), &rgMatch); err != nil {
			// Skip malformed lines (though rg output should be reliable)
			continue
		}

		if rgMatch.Type == "match" {
			// Convert absolute path to workspace-relative
			relPath, err := filepath.Rel(wCtx.WorkspaceRoot, rgMatch.Data.Path.Text)
			if err != nil {
				// Should work if using absolute paths, but fallback to absolute if fails
				relPath = rgMatch.Data.Path.Text
			}

			lineContent := strings.TrimSpace(rgMatch.Data.Lines.Text)
			// Check if line is too long, truncate if necessary
			// This prevents returning massive lines that could crash the response
			if len(lineContent) > maxLineLength {
				lineContent = lineContent[:maxLineLength] + "...[truncated]"
			}

			matches = append(matches, model.SearchContentMatch{
				File:        filepath.ToSlash(relPath),
				LineNumber:  rgMatch.Data.LineNumber,
				LineContent: lineContent,
			})

			// Safety check for memory
			if len(matches) >= maxResults {
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading rg output: %w", err)
	}

	// Wait for command to complete
	if err := proc.Wait(); err != nil {
		exitCode := service.GetExitCode(err)
		if exitCode == 1 {
			// rg returns 1 for no matches (valid case)
			// We just continue with empty matches
		} else {
			// Exit code 2+ = real error
			return nil, fmt.Errorf("rg command failed: %w", err)
		}
	}

	// Sort results for consistency (by file, then line number)
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].File != matches[j].File {
			return matches[i].File < matches[j].File
		}
		return matches[i].LineNumber < matches[j].LineNumber
	})

	// Apply pagination
	paginatedMatches, pagination := service.ApplyPagination(matches, offset, limit)

	return &model.SearchContentResponse{
		Matches:    paginatedMatches,
		Offset:     offset,
		Limit:      limit,
		TotalCount: pagination.TotalCount,
		Truncated:  pagination.Truncated,
	}, nil
}
