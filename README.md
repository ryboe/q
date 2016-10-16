# q

This is a Golang port of Python's [`q` module by zestyping](https://github.com/zestyping/q).
It's a better way to do print statement debugging.

## tl;dr

It prints your variables like this:

![q output examples](https://i.imgur.com/4M125tLl.png)

## Why `q.Q()` is Better than `fmt.Printf()` and `log.Printf()`

* Faster to type
* Pretty-printed vars and expressions
* Easy to see what's in
* Highly readable, colorized output
* Doesn't go to stdout. It goes to the $TMPDIR/q file.
* Basically, it's a better way to do print statement debugging

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
q.Q(query)
```
Or, if you use the `.` import, you can just just type `Q(query)`
```go
import . "github.com/y0ssar1an/q"
...
Q(query)
```

Then you'll see this in `$TMPDIR/q`:

![imgur link](https://i.imgur.com/hUgIKyA.png)

If you're still not sure why you should care, Ping does a better job of
explaining this in his awesome lightning talk from PyCon 2013. Most of what he
says applies to Go.

[![ping's PyCon 2013 lightning talk](https://i.imgur.com/7KmWvtG.jpg)](https://youtu.be/OL3De8BAhME?t=25m14s)

## Install

### Sublime Text
1) Install Package Control
```

```
go get -u github.com/y0ssar1an/q
```

Put this alias in your shell config. Typing `q` will then start tailing
`$TMPDIR/q`.
```
alias q=". $GOPATH/src/github.com/y0ssar1an/q/q.sh"
```

It's common to dedicate a terminal to just tailing `$TMPDIR/q`.

## Basic Usage

99% of the time, you'll be using this one function.

```go
import "github.com/y0ssar1an/q"
...
q.Q(a, b, c)
```

## FAQ

### Why `q.Q`?
It's quick to type and unlikely to cause naming collisions with other variables,
functions, or packages.

### Is `q.Q()` concurrency safe?
Yes
