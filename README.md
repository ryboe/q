# qq

Better print-statement debugging for Go.

This is a port of Python's `q` module by [zestyping](https://github.com/zestyping/q).
I changed the name to `qq` to avoid naming collisions with `q` variables.

## tl;dr

It prints your variables like this:
![qq output examples]()

## Why `qq` is better than `fmt.Print*()` and `log.Print*()` for debugging

You've probably written something like this a thousand times:

```
fmt.Println("\n\n\n\nDEBUG!!!!") // gee, i hope i see this when it flies by.
fmt.Println("query:", query)     // add "query:" so i know which var this is
```

That's a lot of effort to see one variable, and it still fails because you
forgot to import `fmt`, or you forgot that stdout gets redirected, or there's
so much noise on stdout/stderr that you can't find it, or some other dumb
reason.

Try this instead:

```
qq.Log(query)
```

Then you'll see this in `qq.log`:

![imgur link]()

For a better demonstration, check out [Ping's Lightning Talk from PyCon 2013](https://youtu.be/OL3De8BAhME?t=25m14s).
Everything he says applies to Go.

[![still from ping's lightning talk]()]()


## Install

```
go get github.com/y0ssar1an/qq
```

## Basic Usage

```golang
import "github.com/y0ssar1an/qq"
...
qq.Log(a, b, c)
```

Then `tail -f` the `qq.log` file in your $TMPDIR. That's it.

## Snippets

You _could_ type `qq.Log(a, b, c)`, but who's got time to type _all those
characters_? A better way is to add one of the provided snippets to your editor.
Then you'll just type `qq<TAB>` and it will expand to `qq.Log()`.

### Sublime Text
```
cd $GOPATH/src/github.com/y0ssar1an/qq

# OS X
cp qq.sublime-snippet ~/Library/Application\ Support/Sublime\ Text\ 3/Packages/User

# Linux
cp qq.sublime-snippet ~/.config/sublime-text-3/Packages/User

# Windows
???

```

### Atom
```

```

### Vim
TBD Somebody send me a PR, please.


### Easy log tailing

Put this alias in your shell config right meow!
```
alias qq=". $GOPATH/src/github.com/y0ssar1an/qq/qq.sh"
```

## Advanced Usage

### Customize the Header Line

`qq` uses the same flags as the `log` package, with the addition of `Lfuncname`

```golang
qq.SetFlags(LUTC | Llongfile | Lmicroseconds | Lfuncname)`
```

### Use Multiple Log Files

Create a separate `Logger` associated with the new file. Don't worry about
opening and closing the log file. `qq` will take care of that.

```golang
myqq := qq.New("/tmp/myqq.log", LstdFlags)
myqq.Log("herpa derp")
```

## FAQ

### Is `qq.Log()` concurrency safe?
Yes

### Why does `New()` take a file path instead of an `io.Writer`?
You would have to open and close the de

## Troubleshooting

### I can't find the qq.log file

`$TMPDIR/qq.log`. If `$TMPDIR` isn't set, they're going to
`/tmp/qq.log`. If you're on Android, they're going to `/data/local/tmp/qq.log`.



## Why this is better than `fmt.Print*()` and `log.Print*()`

`qq` logs are...
	- optimized for human readability. your ordinary program logs should be
	optimized for machine readability (see structured logging). the most
	important features are colorized. long lines are broken up. redundant info
	is minimized.
	- separated from your

`qq` logs never...
	- get lost in the noise of stdout or stderr
	- get redirected




## Install

```
go get github.com/y0ssar1an/qq
```

## How to Use

`qq.Log(a, b, c)`. Then `tail -f` the `qq.log` file in your $TMPDIR. That's it.

Now you _could_ type `qq.Log(a, b, c)`, but who's got time to type _all those
characters_? A better way is to add one of the provided snippets to your editor.
Then you'll just type `qq<TAB>` and it will expand to `qq.Log()`.

Also, put this alias in your shell config right meow!
```
alias qq=""
```



Like the `log` package, `qq` loggers are safe for concurrent use.



### Set a Prefix

###


# Where Are My Logs Going?

They're going to `$TMPDIR/qq.log`. If `$TMPDIR` isn't set, they're going to
`/tmp/qq.log`. If you're on Android, they're going to `/data/local/tmp/qq.log`.


When you're debugging in Go, and you just want to insert a quick print statement
so you can see the value of a variable

When you're debugging in Go, and you just want to insert a quick print statement
to see the value of a variable, would you rather look at a wall of text or this?



1) It prints the variable name with the value
2) ...in pretty colors
3) ...in a dedicated log file, away from all the noise of stdout and stderr

## Why it's better than fmt.Println or log.Println

Stdout

