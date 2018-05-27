package golinters

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/golang/lint"
	"github.com/golangci/golangci-lint/pkg/result"
	"github.com/sirupsen/logrus"
)

type Golint struct{}

func (Golint) Name() string {
	return "golint"
}

func (Golint) Desc() string {
	return "Golint differs from gofmt. Gofmt reformats Go source code, whereas golint prints out style mistakes"
}

func (g Golint) Run(ctx context.Context, lintCtx *Context) ([]result.Issue, error) {
	var issues []result.Issue
	var lintErr error
	for _, pkgFiles := range lintCtx.Paths.FilesGrouppedByDirs() {
		i, err := g.lintFiles(lintCtx.Settings().Golint.MinConfidence, pkgFiles...)
		if err != nil {
			lintErr = err
			continue
		}
		issues = append(issues, i...)
	}
	if lintErr != nil {
		logrus.Infof("golint: %s", lintErr)
	}

	return issues, nil
}

func (g Golint) lintFiles(minConfidence float64, filenames ...string) ([]result.Issue, error) {
	files := make(map[string][]byte)
	for _, filename := range filenames {
		src, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("can't read file %s: %s", filename, err)
		}
		files[filename] = src
	}

	l := new(lint.Linter)
	ps, err := l.LintFiles(files)
	if err != nil {
		return nil, fmt.Errorf("can't lint files %s: %s", filenames, err)
	}

	var issues []result.Issue
	for _, p := range ps {
		if p.Confidence >= minConfidence {
			issues = append(issues, result.Issue{
				Pos:        p.Position,
				Text:       p.Text,
				FromLinter: g.Name(),
			})
			// TODO: use p.Link and p.Category
		}
	}

	return issues, nil
}
