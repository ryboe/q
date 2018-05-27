package golinters

import (
	"context"
	"fmt"
	"go/token"
	"io/ioutil"
	"log"
	"strconv"

	"github.com/GoASTScanner/gas"
	"github.com/GoASTScanner/gas/rules"
	"github.com/golangci/golangci-lint/pkg/result"
	"github.com/sirupsen/logrus"
)

type Gas struct{}

func (Gas) Name() string {
	return "gas"
}

func (Gas) Desc() string {
	return "Inspects source code for security problems"
}

func (lint Gas) Run(ctx context.Context, lintCtx *Context) ([]result.Issue, error) {
	gasConfig := gas.NewConfig()
	enabledRules := rules.Generate()
	logger := log.New(ioutil.Discard, "", 0)
	analyzer := gas.NewAnalyzer(gasConfig, logger)
	analyzer.LoadRules(enabledRules.Builders())

	analyzer.ProcessProgram(lintCtx.Program)
	issues, _ := analyzer.Report()

	var res []result.Issue
	for _, i := range issues {
		text := fmt.Sprintf("%s: %s", i.RuleID, i.What) // TODO: use severity and confidence
		var r result.Range
		line, err := strconv.Atoi(i.Line)
		if err != nil {
			if n, rerr := fmt.Sscanf(i.Line, "%d-%d", &r.From, &r.To); rerr != nil || n != 2 {
				logrus.Infof("Can't convert gas line number %q of %v to int: %s", i.Line, i, err)
				continue
			}
			line = r.From
		}

		res = append(res, result.Issue{
			Pos: token.Position{
				Filename: i.File,
				Line:     line,
			},
			Text:       text,
			LineRange:  r,
			FromLinter: lint.Name(),
		})
	}

	return res, nil
}
