package result

import "go/token"

type Range struct {
	From, To int
}

type Issue struct {
	FromLinter string
	Text       string

	Pos       token.Position
	LineRange Range
	HunkPos   int
}

func (i Issue) FilePath() string {
	return i.Pos.Filename
}

func (i Issue) Line() int {
	return i.Pos.Line
}

func (i Issue) GetLineRange() Range {
	if i.LineRange.From == 0 {
		return Range{
			From: i.Line(),
			To:   i.Line(),
		}
	}

	return i.LineRange
}
