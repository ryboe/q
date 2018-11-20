package config

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/golangci/golangci-lint/pkg/logutils"
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
	{
		Pattern: "Use of unsafe calls should be audited",
		Linter:  "gosec",
		Why:     "Too many false-positives on 'unsafe' usage",
	},
	{
		Pattern: "Subprocess launch(ed with variable|ing should be audited)",
		Linter:  "gosec",
		Why:     "Too many false-positives for parametrized shell calls",
	},
	{
		Pattern: "G104",
		Linter:  "gosec",
		Why:     "Duplicated errcheck checks",
	},
	{
		Pattern: "(Expect directory permissions to be 0750 or less|Expect file permissions to be 0600 or less)",
		Linter:  "gosec",
		Why:     "Too many issues in popular repos",
	},
	{
		Pattern: "Potential file inclusion via variable",
		Linter:  "gosec",
		Why:     "False positive is triggered by 'src, err := ioutil.ReadFile(filename)'",
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
	Govet struct {
		CheckShadowing bool `mapstructure:"check-shadowing"`
	}
	Golint struct {
		MinConfidence float64 `mapstructure:"min-confidence"`
	}
	Gofmt struct {
		Simplify bool
	}
	Goimports struct {
		LocalPrefixes string `mapstructure:"local-prefixes"`
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
	Errcheck ErrcheckSettings
	Gocritic GocriticSettings
}

type ErrcheckSettings struct {
	CheckTypeAssertions bool       `mapstructure:"check-type-assertions"`
	CheckAssignToBlank  bool       `mapstructure:"check-blank"`
	Ignore              IgnoreFlag `mapstructure:"ignore"`
	Exclude             string     `mapstructure:"exclude"`
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

type GocriticCheckSettings map[string]interface{}

type GocriticSettings struct {
	EnabledChecks    []string                         `mapstructure:"enabled-checks"`
	DisabledChecks   []string                         `mapstructure:"disabled-checks"`
	SettingsPerCheck map[string]GocriticCheckSettings `mapstructure:"settings"`

	inferredEnabledChecks map[string]bool
}

func (s *GocriticSettings) InferEnabledChecks(log logutils.Log) {
	enabledChecks := s.EnabledChecks
	if len(enabledChecks) == 0 {
		if len(s.DisabledChecks) != 0 {
			for _, defaultCheck := range defaultGocriticEnabledChecks {
				if !s.isCheckDisabled(defaultCheck) {
					enabledChecks = append(enabledChecks, defaultCheck)
				}
			}
		} else {
			enabledChecks = defaultGocriticEnabledChecks
		}
	}

	s.inferredEnabledChecks = map[string]bool{}
	for _, check := range enabledChecks {
		s.inferredEnabledChecks[strings.ToLower(check)] = true
	}
	log.Infof("Gocritic enabled checks: %s", enabledChecks)
}

func (s GocriticSettings) isCheckDisabled(name string) bool {
	for _, disabledCheck := range s.DisabledChecks {
		if disabledCheck == name {
			return true
		}
	}

	return false
}

func (s GocriticSettings) Validate(log logutils.Log) error {
	if len(s.EnabledChecks) != 0 && len(s.DisabledChecks) != 0 {
		return errors.New("both enabled and disabled check aren't allowed for gocritic")
	}

	for checkName := range s.SettingsPerCheck {
		if !s.IsCheckEnabled(checkName) {
			log.Warnf("Gocritic settings were provided for not enabled check %q", checkName)
		}
	}

	return nil
}

func (s GocriticSettings) IsCheckEnabled(name string) bool {
	return s.inferredEnabledChecks[strings.ToLower(name)]
}

var defaultGocriticEnabledChecks = []string{
	"appendAssign",
	"assignOp",
	"caseOrder",
	"dupArg",
	"dupBranchBody",
	"dupCase",
	"flagDeref",
	"ifElseChain",
	"regexpMust",
	"singleCaseSwitch",
	"sloppyLen",
	"switchTrue",
	"typeSwitchVar",
	"underef",
	"unlambda",
	"unslice",
	"defaultCaseOrder",
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
	Errcheck: ErrcheckSettings{
		Ignore: IgnoreFlag{},
	},
	Gocritic: GocriticSettings{
		SettingsPerCheck: map[string]GocriticCheckSettings{},
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

// IgnoreFlags was taken from errcheck in order to keep the API identical.
// https://github.com/kisielk/errcheck/blob/1787c4bee836470bf45018cfbc783650db3c6501/main.go#L25-L60
type IgnoreFlag map[string]*regexp.Regexp

func (f IgnoreFlag) String() string {
	pairs := make([]string, 0, len(f))
	for pkg, re := range f {
		prefix := ""
		if pkg != "" {
			prefix = pkg + ":"
		}
		pairs = append(pairs, prefix+re.String())
	}
	return fmt.Sprintf("%q", strings.Join(pairs, ","))
}

func (f IgnoreFlag) Set(s string) error {
	if s == "" {
		return nil
	}
	for _, pair := range strings.Split(s, ",") {
		colonIndex := strings.Index(pair, ":")
		var pkg, re string
		if colonIndex == -1 {
			pkg = ""
			re = pair
		} else {
			pkg = pair[:colonIndex]
			re = pair[colonIndex+1:]
		}
		regex, err := regexp.Compile(re)
		if err != nil {
			return err
		}
		f[pkg] = regex
	}
	return nil
}

// Type returns the type of the flag follow the pflag format.
func (IgnoreFlag) Type() string {
	return "stringToRegexp"
}
