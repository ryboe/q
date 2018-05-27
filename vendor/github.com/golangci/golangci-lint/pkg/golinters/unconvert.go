package golinters

import (
	"context"

	"github.com/golangci/golangci-lint/pkg/result"
	unconvertAPI "github.com/golangci/unconvert"
)

type Unconvert struct{}

func (Unconvert) Name() string {
	return "unconvert"
}

func (Unconvert) Desc() string {
	return "Remove unnecessary type conversions"
}

func (lint Unconvert) Run(ctx context.Context, lintCtx *Context) ([]result.Issue, error) {
	positions := unconvertAPI.Run(lintCtx.Program)
	var res []result.Issue
	for _, pos := range positions {
		res = append(res, result.Issue{
			Pos:        pos,
			Text:       "unnecessary conversion",
			FromLinter: lint.Name(),
		})
	}

	return res, nil
}
