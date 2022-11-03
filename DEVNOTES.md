# Overview

As described in the general README for this project, much of the work in
finding semantic differences between file revisions relies on external
tools.  Often these tools are the actual interpreters or compilers for the
respective programming or declarative languages.

There are two basic classes of tools that will be used: parsers and
canonicalizers.  What we get from each is somewhat different.  In the case
of a canonicalizer, it's pretty simple: Create the fully normalized version
of the *A* version and the *B* version, and do a diff on them.  If nothing
semantically important has changed, hopefully these canonical versions are
identical.  If not, we can present a highlighted diff of where they differ.

In the case of a parse tree, we instead transform the surface of the code
into its underlying AST (abstract syntax tree).  Some tools might settle for
a CST (concrete syntax tree) is that is more available.  At that point, we
basically need to compare the ASTs with position information (line/column)
stripped out, then if they look different at that level, put back in the
position information and find those portions of the diff of the actual
source files that corresponds to those parts of the AST (usually by looking
at the non-stripped AST).

# Adding a new supported language

Suppose you'd like `sdt` to support the language FizzBuzz.  These are the
steps you'd take.

* What identifies a FizzBuzz source file.  For now, we only look at file
  extensions, so suppose FizzBuzz fils use the extension `*.fzb`.

  In `$ROOT/main.go` is a `switch ext` within `compareFileType()`.  Within a
  `case` of that switch, you'll see a call similar to:

    ruby.Diff(filename, options, config)

  For FizzBuzz support, simply use `fizzbuzz.Diff(filename, options,
  config)` instead.

* Create a package at `$ROOT/pkg/fizzbuzz/fizzbuzz.go`.  Be sure to import
  this package within `main.go` as well, of course.  The call to
  `fizzbuzz.Diff()` needs to return a string that simply _is_ the reported
  analysis.

  In particular though, you want to support one or both of the report types
  for `options.Parsetree` and `options.Semantic`.  The latter is more
  important, but the former will often more-or-less "fall out" for free.
  Let's suppose FizzBuzz uses the AST approach.  Within the `Diff` function
  (which is likely to be the only exported function in the package), you
  will make an external call utilizing
  `config.Commands["fizzbuzz"].Executable` and
  `config.Commands["fizzbuzz"].Switches`.  Follow the pattern in one of the
  existing languages for this.

* The point of the executable/switches pattern is that the executable and
  switches may optionally be configured in `$HOME/.sdt.toml` by each
  individual user.  However, you must also configure default options in
  `main.go` for those users who wish to simply use their system path
  versions of tools.

  In the languages supported so far, using the language executables
  themselves (and a few switches) has worked. However, for FizzBuzz, you may
  need to include a small FizzBuzz program to produce an AST. Ideally, it
  should be possible to include this program as an executable in the
  repository (either compiled for various platforms or simply a runnable
  script for interpreted languages).

* Most likely, for the `options.Parsetree` path, you'll call
  `utils.ColorDiff()`, with `types.FizzBuzz` passed as an argument.
  Similarly, for the `options.Semantic` path, you'll call
  `utils.SemanticChanges()` using the same type constant.

  By implication, of course, these functions in the `utils` package also
  need to switch on the parse type.  And you'll need to add the actual
  constant into the `types` package, which is a single line in an
  enumeration.

  Within `utils.ColorDiff()` the switch should be particularly easy.  All
  that existing cases have done is run a series of regexp transformations to
  "cleanup" the respective trees for better presentation.  We'd like a tree
  to look tree-like, i.e. line oriented, which is what most tools product.

  In `utils.SemanticChanges()` the switch includes slightly more, but even
  there most functionality is common between languages.  Basically, all that
  a particular case like `types.FizzBuzz` needs to do is add to a set
  `diffLines` the contains all the lines of the source (not destination) the
  correspond to differences in the AST.  In general, an annotated AST should
  contain exactly this information already.
