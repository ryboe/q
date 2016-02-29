# qq

Better print-statement debugging for Go.

This is a port of Python's [`q` module by zestyping](https://github.com/zestyping/q).
I changed the name to `qq` to avoid naming collisions with single-letter `q`
variables (common in Go).

## tl;dr

It prints your variables like this:

![qq output examples](https://i.imgur.com/4M125tLl.png)

## Why `qq` is Better than `fmt.Print*()` and `log.Print*()` for Debugging

You've probably written this a thousand times:

```go
fmt.Println("\n\n\n\nDEBUG!!!!") // gee, i hope i see this when it flies by.
fmt.Println("query:", query)     // add "query:" so i know which var this is
```

That's a lot of typing, and it still fails because stdout is getting redirected
somewhere, or there's so much noise on stdout/stderr that you can't find it, or
some other dumb reason.

Try this instead:

```go
qq.Log(query)
```

Then you'll see this in `qq.log`:

![imgur link](https://i.imgur.com/hUgIKyA.png)

If you're still not sure why you should care, Ping does a way better job of
explaining this in his awesome lightning talk from PyCon 2013. Most of what he
says applies to Go.

[![still from ping's lightning talk](https://i.imgur.com/7KmWvtG.jpg)](https://youtu.be/OL3De8BAhME?t=25m14s)

## Install

```
go get github.com/y0ssar1an/qq
```

## Basic Usage

99% of the time, you'll be using this one function.

```go
import "github.com/y0ssar1an/qq"
...
qq.Log(a, b, c)
```

Then `tail -f` the `qq.log` file in your `$TMPDIR`. That's it. Note that typing
`tail -f $TMPDIR/qq.log` might not work because `$TMPDIR` may have a trailing
slash. You think you're typing `/tmp/qq.log`, but you're actually typing
`/tmp//qq.log`. To avoid these headaches, it's recommended that you use the
alias [below](https://github.com/y0ssar1an/qq/blob/master/README.md#easy-log-tailing).

## Snippets

You _could_ type `qq.Log(a, b, c)`, but who's got time to type _all those
characters_? A better way is to add one of the provided snippets to your editor.
Then you'll just type `qq<TAB>` and it will expand to `qq.Log()`.

#### Sublime Text
```
# OS X
cp $GOPATH/src/github.com/y0ssar1an/qq/qq.sublime-snippet ~/Library/Application\ Support/Sublime\ Text\ 3/Packages/User

# Linux
cp $GOPATH/src/github.com/y0ssar1an/qq/qq.sublime-snippet ~/.config/sublime-text-3/Packages/User

# Windows
???

```

#### Atom
TBD

#### Vim
TBD Somebody send me a PR, please.


### Easy Log Tailing

Put this alias in your shell config right meow!
```
alias qq=". $GOPATH/src/github.com/y0ssar1an/qq/qq.sh"
```

It's common to dedicate a terminal to just tailing `qq.log`.

## Advanced Usage

Everything works just like the [`log` package](https://golang.org/pkg/log/).

### The Full Docs

[https://godoc.org/github.com/y0ssar1an/qq](https://godoc.org/github.com/y0ssar1an/qq)

### Customize the Header Line

`qq` uses the same flags as the `log` package, with the addition of `Lfuncname`

```go
qq.SetFlags(LUTC | Llongfile | Lmicroseconds | Lfuncname)`
```

### Use Multiple Log Files

Create a separate `Logger` associated with the new file. Don't worry about
opening and closing the log file. `qq` will take care of that.

```go
myqq := qq.New("/tmp/myqq.log", LstdFlags)
myqq.Log("herpa derp")
```

### Set a Prefix
```go
qq.SetPrefix("main goroutine")
```

## FAQ

### Is `qq.Log()` concurrency safe?
Yes

### Why does `New()` take a file path instead of an `io.Writer`?
You would have to open and close the output destination every time you wanted
to write. By giving the `qq` Logger a file path, it can take care of opening
and closing the log file for you. The goal of this library is to be "quick and
dirty debugging output for tired programmers". We're willing to trade some
flexibility to minimize typing.

## Troubleshooting

### I can't find the `qq.log` file

It's in `$TMPDIR`. If `$TMPDIR` isn't set, they're going to `/tmp/qq.log`. If
you're on Android, they're going to `/data/local/tmp/qq.log`.
