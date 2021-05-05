package healthcheck

import (
	"context"

	"github.com/pkg/errors"
)

var GitBranch string

const GitBranchKey = "git_branch"

// -X github.com/gtforge/go-healthcheck.GitBranch=`git rev-parse --abbrev-ref HEAD`
// Use it to report git branch on alive. If GitBranch was not set, return error
func MakeBranchPingerWithError() (Pinger, error) {
	if GitBranch == "" {
		return nil, errors.New("GitBranch not initialized")
	}
	return func(_ context.Context) (map[string]interface{}, error) {
		return map[string]interface{}{
			GitBranchKey: GitBranch,
		}, nil
	}, nil
}

func MakeBranchPinger() Pinger {
	return func(_ context.Context) (map[string]interface{}, error) {
		return map[string]interface{}{
			GitBranchKey: GitBranch,
		}, nil
	}
}
