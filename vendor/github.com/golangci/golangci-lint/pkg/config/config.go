package config

import (
	"time"
)

const (
	OutFormatJSON              = "json"
	OutFormatLineNumber        = "line-number"
	OutFormatColoredLineNumber = "colored-line-number"
	OutFormatTab               = "tab"
	OutFormatCheckstyle        = "checkstyle"
)

var OutFormats = []string{
	OutFormatColoredLineNumber,
	OutFormatLineNumber,
	OutFormatJSON,
	OutFormatTab,
	OutFormatCheckstyle,
}

type ExcludePattern struct {
	Pattern string
	Linter  string
	Why     string
}

var DefaultExcludePatterns = []ExcludePattern{
	{
		Pattern: "Error return value of .((os\\.)?std(out|err)\\..*|.*Close" +
			"|.*Flush|os\\.Remove(All)?|.*printf?|os\\.(Un)?Setenv). is not checked",
		Linter: "errcheck",
		Why:    "Almost all programs ignore errors on these functions and in most cases it's ok",
	},
	{
		Pattern: "(comment on exported (method|function|type|const)|" +
			"should have( a package)? comment|comment should be of the form)",
		Linter: "golint",
		Why:    "Annoying issue about not having a comment. The rare codebase has such comments",
	},
	{
		Pattern: "func name will be used as test\\.Test.* by other packages, and that stutters; consider calling this",
		Linter:  "golint",
		Why:     "False positive when tests are defined in package 'test'",
	},
	{
		Pattern: "(possible misuse of unsafe.Pointer|should have signature)",
		Linter:  "govet",
		Why:     "Common false positives",
	},
	{
		Pattern: "ineffective break statement. Did you mean to break out of the outer loop",
		Linter:  "megacheck",
		Why:     "Developers tend to write in C-style with an explicit 'break' in a 'switch', so it's ok to ignore",
	},
}

func GetDefaultExcludePatternsStrings() []string {
	var ret []string
	for _, p := range DefaultExcludePatterns {
		ret = append(ret, p.Pattern)
	}

	return ret
}

type Run struct {
	IsVerbose           bool `mapstructure:"verbose"`
	Silent              bool
	CPUProfilePath      string
	MemProfilePath      string
	Concurrency         int
	PrintResourcesUsage bool `mapstructure:"print-resources-usage"`

	Config   string
	NoConfig bool

	Args []string

	BuildTags []string `mapstructure:"build-tags"`

	ExitCodeIfIssuesFound int  `mapstructure:"issues-exit-code"`
	AnalyzeTests          bool `mapstructure:"tests"`
	Deadline              time.Duration
	PrintVersion          bool

	SkipFiles []string `mapstructure:"skip-files"`
	SkipDirs  []string `mapstructure:"skip-dirs"`
}

type LintersSettings struct {
	Errcheck struct {
		CheckTypeAssertions bool `mapstructure:"check-type-assertions"`
		CheckAssignToBlank  bool `mapstructure:"check-blank"`
	}
	Govet struct {
		CheckShadowing       bool `mapstructure:"check-shadowing"`
		UseInstalledPackages bool `mapstructure:"use-installed-packages"`
	}
	Golint struct {
		MinConfidence float64 `mapstructure:"min-confidence"`
	}
	Gofmt struct {
		Simplify bool
	}
	Gocyclo struct {
		MinComplexity int `mapstructure:"min-complexity"`
	}
	Varcheck struct {
		CheckExportedFields bool `mapstructure:"exported-fields"`
	}
	Structcheck struct {
		CheckExportedFields bool `mapstructure:"exported-fields"`
	}
	Maligned struct {
		SuggestNewOrder bool `mapstructure:"suggest-new"`
	}
	Dupl struct {
		Threshold int
	}
	Goconst struct {
		MinStringLen        int `mapstructure:"min-len"`
		MinOccurrencesCount int `mapstructure:"min-occurrences"`
	}
	Depguard struct {
		ListType      string `mapstructure:"list-type"`
		Packages      []string
		IncludeGoRoot bool `mapstructure:"include-go-root"`
	}
	Misspell struct {
		Locale string
	}
	Unused struct {
		CheckExported bool `mapstructure:"check-exported"`
	}

	Lll      LllSettings
	Unparam  UnparamSettings
	Nakedret NakedretSettings
	Prealloc PreallocSettings
}

type LllSettings struct {
	LineLength int `mapstructure:"line-length"`
	TabWidth   int `mapstructure:"tab-width"`
}

type UnparamSettings struct {
	CheckExported bool `mapstructure:"check-exported"`
	Algo          string
}

type NakedretSettings struct {
	MaxFuncLines int `mapstructure:"max-func-lines"`
}

type PreallocSettings struct {
	Simple     bool
	RangeLoops bool `mapstructure:"range-loops"`
	ForLoops   bool `mapstructure:"for-loops"`
}

var defaultLintersSettings = LintersSettings{
	Lll: LllSettings{
		LineLength: 120,
		TabWidth:   1,
	},
	Unparam: UnparamSettings{
		Algo: "cha",
	},
	Nakedret: NakedretSettings{
		MaxFuncLines: 30,
	},
	Prealloc: PreallocSettings{
		Simple:     true,
		RangeLoops: true,
		ForLoops:   false,
	},
}

type Linters struct {
	Enable     []string
	Disable    []string
	EnableAll  bool `mapstructure:"enable-all"`
	DisableAll bool `mapstructure:"disable-all"`
	Fast       bool

	Presets []string
}

type Issues struct {
	ExcludePatterns    []string `mapstructure:"exclude"`
	UseDefaultExcludes bool     `mapstructure:"exclude-use-default"`

	MaxIssuesPerLinter int `mapstructure:"max-issues-per-linter"`
	MaxSameIssues      int `mapstructure:"max-same-issues"`

	DiffFromRevision  string `mapstructure:"new-from-rev"`
	DiffPatchFilePath string `mapstructure:"new-from-patch"`
	Diff              bool   `mapstructure:"new"`
}

type Config struct { //nolint:maligned
	Run Run

	Output struct {
		Format              string
		PrintIssuedLine     bool `mapstructure:"print-issued-lines"`
		PrintLinterName     bool `mapstructure:"print-linter-name"`
		PrintWelcomeMessage bool `mapstructure:"print-welcome"`
	}

	LintersSettings LintersSettings `mapstructure:"linters-settings"`
	Linters         Linters
	Issues          Issues

	InternalTest bool // Option is used only for testing golangci-lint code, don't use it
}

func NewDefault() *Config {
	return &Config{
		LintersSettings: defaultLintersSettings,
	}
}
