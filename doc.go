// Package q provides quick and dirty debugging output for tired programmers.
// Just type Q(foo)
// It's the fastest way to print a variable. Typing Q(foo) is easier than
// fmt.Printf("%#v whatever"). The output is easy on the eyes, with
// colorization, pretty-printing, and nice formatting. The output goes to
// the $TMPDIR/q log file, away from the noisy output of your program. Try it
// and give up using fmt.Printf() for good!
//
// For best results, import it like this:
// import . "github.com/y0ssar1an/q"

package q
