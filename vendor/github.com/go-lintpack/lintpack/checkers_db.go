package lintpack

import (
	"fmt"
	"sort"
	"strings"

	"github.com/go-toolsmith/astfmt"
)

type checkerProto struct {
	info        *CheckerInfo
	constructor func(*Context, parameters) *Checker
}

// prototypes is a set of registered checkers that are not yet instantiated.
// Registration should be done with AddChecker function.
// Initialized checkers can be obtained with NewChecker function.
var prototypes = make(map[string]checkerProto)

func getCheckersInfo() []*CheckerInfo {
	infoList := make([]*CheckerInfo, 0, len(prototypes))
	for _, proto := range prototypes {
		infoCopy := *proto.info
		infoList = append(infoList, &infoCopy)
	}
	sort.Slice(infoList, func(i, j int) bool {
		return infoList[i].Name < infoList[j].Name
	})
	return infoList
}

func addChecker(info *CheckerInfo, constructor func(*CheckerContext) FileWalker) {
	if _, ok := prototypes[info.Name]; ok {
		panic(fmt.Sprintf("checker with name %q already registered", info.Name))
	}

	trimDocumentation := func(d *CheckerInfo) {
		fields := []*string{
			&d.Summary,
			&d.Details,
			&d.Before,
			&d.After,
			&d.Note,
		}
		for _, f := range fields {
			*f = strings.TrimSpace(*f)
		}
	}
	validateDocumentation := func(d *CheckerInfo) {
		// TODO(Quasilyte): validate documentation.
	}

	trimDocumentation(info)
	validateDocumentation(info)

	proto := checkerProto{
		info: info,
		constructor: func(ctx *Context, params parameters) *Checker {
			var c Checker
			c.Info = info
			c.ctx = CheckerContext{
				Context: ctx,
				Params:  params,
				printer: astfmt.NewPrinter(ctx.FileSet),
			}
			c.fileWalker = constructor(&c.ctx)
			return &c
		},
	}

	prototypes[info.Name] = proto
}

func newChecker(ctx *Context, info *CheckerInfo, params map[string]interface{}) *Checker {
	proto, ok := prototypes[info.Name]
	if !ok {
		panic(fmt.Sprintf("checker with name %q not registered", info.Name))
	}
	return proto.constructor(ctx, params)
}
