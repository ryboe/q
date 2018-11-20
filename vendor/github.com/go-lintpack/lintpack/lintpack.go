package lintpack

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/go-toolsmith/astfmt"
)

// CheckerInfo holds checker metadata and structured documentation.
type CheckerInfo struct {
	// Name is a checker name.
	Name string

	// Tags is a list of labels that can be used to enable or disable checker.
	// Common tags are "experimental" and "performance".
	Tags []string

	// Summary is a short one sentence description.
	// Should not end with a period.
	Summary string

	// Details extends summary with additional info. Optional.
	Details string

	// Before is a code snippet of code that will violate rule.
	Before string

	// After is a code snippet of fixed code that complies to the rule.
	After string

	// Note is an optional caution message or advice.
	Note string
}

// GetCheckersInfo returns a checkers info list for all registered checkers.
// The slice is sorted by a checker name.
//
// Info objects can be used to instantiate checkers with NewChecker function.
func GetCheckersInfo() []*CheckerInfo {
	return getCheckersInfo()
}

// HasTag reports whether checker described by the info has specified tag.
func (info *CheckerInfo) HasTag(tag string) bool {
	for i := range info.Tags {
		if info.Tags[i] == tag {
			return true
		}
	}
	return false
}

// Checker is an implementation of a check that is described by the associated info.
type Checker struct {
	// Info is an info object that was used to instantiate this checker.
	Info *CheckerInfo

	ctx CheckerContext

	fileWalker FileWalker
}

// Check runs rule checker over file f.
func (c *Checker) Check(f *ast.File) []Warning {
	c.ctx.warnings = c.ctx.warnings[:0]
	c.fileWalker.WalkFile(f)
	return c.ctx.warnings
}

// Warning represents issue that is found by checker.
type Warning struct {
	// Node is an AST node that caused warning to trigger.
	// Can be used to obtain proper error location.
	Node ast.Node

	// Text is warning message without source location info.
	Text string
}

// AddChecker registers a new checker into a checkers pool.
// Constructor is used to create a new checker instance.
// Checker name (defined in CheckerInfo.Name) must be unique.
//
// If checker is never needed, for example if it is disabled,
// constructor will not be called.
func AddChecker(info *CheckerInfo, constructor func(*CheckerContext) FileWalker) {
	addChecker(info, constructor)
}

// NewChecker returns initialized checker identified by an info.
// info must be non-nil.
// Panics if info describes a checker that was not properly registered.
//
// params argument specifies per-checker options.NewChecker. Can be nil.
func NewChecker(ctx *Context, info *CheckerInfo, params map[string]interface{}) *Checker {
	return newChecker(ctx, info, params)
}

// Context is a readonly state shared among every checker.
type Context struct {
	// TypesInfo carries parsed packages types information.
	TypesInfo *types.Info

	// SizesInfo carries alignment and type size information.
	// Arch-dependent.
	SizesInfo types.Sizes

	// FileSet is a file set that was used during the program loading.
	FileSet *token.FileSet

	// Pkg describes package that is being checked.
	Pkg *types.Package

	// Filename is a currently checked file name.
	Filename string

	// Require records what optional resources are required
	// by the checkers set that use this context.
	//
	// Every require fields makes associated context field
	// to be properly initialized.
	// For example, Context.require.PkgObjects => Context.PkgObjects.
	Require struct {
		PkgObjects bool
		PkgRenames bool
	}

	// PkgObjects stores all imported packages and their local names.
	PkgObjects map[*types.PkgName]string

	// PkgRenames maps package path to its local renaming.
	// Contains no entries for packages that were imported without
	// explicit local names.
	PkgRenames map[string]string
}

// NewContext returns new shared context to be used by every checker.
//
// All data carried by the context is readonly for checkers,
// but can be modified by the integrating application.
func NewContext(fset *token.FileSet, sizes types.Sizes) *Context {
	return &Context{
		FileSet:   fset,
		SizesInfo: sizes,
		TypesInfo: &types.Info{},
	}
}

// SetPackageInfo sets package-related metadata.
//
// Must be called for every package being checked.
func (c *Context) SetPackageInfo(info *types.Info, pkg *types.Package) {
	if info != nil {
		// We do this kind of assignment to avoid
		// changing c.typesInfo field address after
		// every re-assignment.
		*c.TypesInfo = *info
	}
	c.Pkg = pkg
}

// SetFileInfo sets file-related metadata.
//
// Must be called for every source code file being checked.
func (c *Context) SetFileInfo(name string, f *ast.File) {
	c.Filename = name
	if c.Require.PkgObjects {
		resolvePkgObjects(c, f)
	}
	if c.Require.PkgRenames {
		resolvePkgRenames(c, f)
	}
}

// CheckerContext is checker-local context copy.
// Fields that are not from Context itself are writeable.
type CheckerContext struct {
	*Context

	// Params hold checker-specific set of options.
	Params parameters

	// printer used to format warning text.
	printer *astfmt.Printer

	warnings []Warning
}

// Warn adds a Warning to checker output.
func (ctx *CheckerContext) Warn(node ast.Node, format string, args ...interface{}) {
	ctx.warnings = append(ctx.warnings, Warning{
		Text: ctx.printer.Sprintf(format, args...),
		Node: node,
	})
}

// FileWalker is an interface every checker should implement.
//
// The WalkFile method is executed for every Go file inside the
// package that is being checked.
type FileWalker interface {
	WalkFile(*ast.File)
}
