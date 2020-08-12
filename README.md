# q

[![CircleCI](https://circleci.com/gh/ryboe/q/tree/master.svg?style=svg)](https://circleci.com/gh/ryboe/q/tree/master)
[![Go Report Card](https://goreportcard.com/badge/github.com/ryboe/q)](https://goreportcard.com/report/github.com/ryboe/q)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/ryboe/q)](https://pkg.go.dev/github.com/ryboe/q)

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
import "q"
...
q.Q(a, b, c)
```

For best results, dedicate a terminal to tailing `$TMPDIR/q` while you work.

## Install

```sh
git clone https://github.com/ryboe/q $GOPATH/src/q
```

Put these functions in your shell config. Typing `qq` or `rmqq` will then start
tailing `$TMPDIR/q`.

```sh
qq() {
    clear

    logpath="$TMPDIR/q"
    if [[ -z "$TMPDIR" ]]; then
        logpath="/tmp/q"
    fi

    if [[ ! -f "$logpath" ]]; then
        echo 'Q LOG' > "$logpath"
    fi

    tail -100f -- "$logpath"
}

rmqq() {
    logpath="$TMPDIR/q"
    if [[ -z "$TMPDIR" ]]; then
        logpath="/tmp/q"
    fi
    if [[ -f "$logpath" ]]; then
        rm "$logpath"
    fi
    qq
}
```

You also can simply `tail -f $TMPDIR/q`, but it's highly recommended to use the above commands.

## Editor Integration

### VS Code

`Preferences > User Snippets > Go`

```json
"qq": {
    "prefix": "qq",
    "body": "q.Q($1) // DEBUG",
    "description": "Pretty-print to $TMPDIR/q"
}
```

### Sublime Text

`Tools > Developer > New Snippet`

```xml
<snippet>
    <content><![CDATA[
q.Q($1) // DEBUG
]]></content>
    <tabTrigger>qq</tabTrigger>
    <scope>source.go</scope>
</snippet>
```

### Atom

`Atom > Open Your Snippets`

```coffee
'.source.go':
    'qq':
        'prefix': 'qq'
        'body': 'q.Q($1) // DEBUG'
```

### JetBrains GoLand

`Settings > Editor > Live Templates`

In `Go`, add a new template with:

* Abbreviation: `qq`
* Description: `Pretty-print to $TMPDIR/q`
* Template text: `q.Q($END$) // DEBUG`
* Applicable in: select the `Go` scope

### Emacs

Add a new snippet file to the go-mode snippets directory
(`$HOME/.emacs.d/snippets/go-mode/qq`). This should
contain:

```emacs
# -*- mode: snippet -*-
# name: qq
# key: qq
# --
q.Q(${1:...}) // DEBUG
```

### vim

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

Yes.
