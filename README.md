# q
[![Build Status](https://travis-ci.org/y0ssar1an/q.svg?branch=develop)](https://travis-ci.org/y0ssar1an/q)
[![GoDoc](https://godoc.org/github.com/y0ssar1an/q?status.svg)](https://godoc.org/github.com/y0ssar1an/q)
[![Go Report Card](https://goreportcard.com/badge/github.com/y0ssar1an/q)](https://goreportcard.com/report/github.com/y0ssar1an/q)

q is a better way to do print statement debugging.

Type `q.Q` instead of `fmt.Printf` and your variables will be printed like this:

![q output examples](https://i.imgur.com/OFmm7pb.png)

## Why is this better than `fmt.Printf`?

* Faster to type
* Pretty-printed vars and expressions
* Easier to see inside structs
* Doesn't go to noisy-ass stdout. It goes to `$TMPDIR/q`.
* Pretty colors!

## Basic Usage

```go
import "github.com/y0ssar1an/q"
...
q.Q(a, b, c)

// Alternatively, use the . import and you can omit the package name.
// q only exports the Q function.
import . "github.com/y0ssar1an/q"
...
Q(a, b, c)
```


For best results, dedicate a terminal to tailing `$TMPDIR/q` while you work.

## Install

```sh
go get -u github.com/y0ssar1an/q
```

Put these aliases in your shell config. Typing `qq` will then start tailing
`$TMPDIR/q`.
```sh
alias qq=". $GOPATH/src/github.com/y0ssar1an/q/q.sh"
alias rmqq="rm $TMPDIR/q"
```

## Editor Integration

#### Sublime Text
```
cp $GOPATH/src/github.com/y0ssar1an/q/qq.sublime-snippet Packages/User/qq.sublime-snippet
```

#### Atom
Navigate to your `snippets.cson` file by either opening `~/.atom/snippets.cson`
directly or by selecting the `Atom > Open Your Snippets` menu. You can then add
this code snippet to the bottom and save the file:
```
'.source.go':
  'q log':
    'prefix': 'qq'
    'body': 'q.Q($1)'
```

#### VS Code
In the VS Code menu go to `Preferences` and choose `User Snippets`. When the
language dropdown menu appears select `GO`. Add the following snippet to the
array of snippets.
```
"q.Q ": {
	"prefix": "qq",
	"body": [
		"q.Q($1)"
	],
	"description": "Quick and dirty debugging output for tired Go programmers"
}
```

#### vim/Emacs
TBD Send me a PR, please :)

## Haven't I seen this somewhere before?

Python programmers will recognize this as a Golang port of the
[`q` module by zestyping](https://github.com/zestyping/q).

Ping does a great job of explaining `q` in his awesome lightning talk from
PyCon 2013. Watch it! It's funny :)

[![ping's PyCon 2013 lightning talk](https://i.imgur.com/7KmWvtG.jpg)](https://youtu.be/OL3De8BAhME?t=25m14s)

## FAQ

### Why `q.Q`?
It's quick to type and unlikely to cause naming collisions.

### Is `q.Q()` safe for concurrent use?
Yes
