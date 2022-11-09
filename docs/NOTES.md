## Early thoughts

* I installed a tool called `astdiff` that is Python-only, but does pretty much
exactly what we've discussed for that language
(https://github.com/auntbertha/ASTdiff).

* `astdiff` seems cool, but basically a once-person project that hasn't been
touched since 2018.

* I switched to a repository that I am somewhat familiar with, but is not my
code (other than some very small contributions
(git@github.com:Chad-Mowbray/iamb-classifier.git)

* I ran `black` on the whole project, which reformatted 36 files.  Black is
very high quality, widely-used, and I would be shocked if it created any
functional changes.

* `astdiff` is integrated with `git`, so I can run:

```
% astdiff
Running: git diff --name-only 9da3bbfa2cac4518dae5b01de7c62a100a785a82
...
```

* 35 of the 36 changed files show "ok".  1 file shows (abridged):

```
Checking ipclassifier/iambic_line_processors/iambic_line.py ... failed
different nodes:
line 7 in first commit:
    Takes a tokenized line of IP
    [... bunch more lines of docstring of function ...]

line 6 in second commit:
    Takes a tokenized line of IP
    [... bunch more lines of docstring of function ...]
```

That is, the indented stuff was all the content of a function docstring.  That
is unchanged except for removal of trailing whitespace on one of the 15 or so
lines in the docstring.

* So I was curious what was *actually* different:

```
-class IambicLine():
+class IambicLine:
```

The docstring whitespace thing is a red herring, it's the different style of
class declaration that is causing the failure.

## Early thoughts, part 2

* `astexplorer` (https://github.com/fkling/astexplorer) is MIT licensed and a
good starting place for my Atlantis project.

* We can rip out all the web interface parts (which is most of it) and just use
the parsers

* The parsers are not part of `astexplorer` itself, but rather all other
libraries the front-end utilizes

* Even though many languages are parsed, the parsing itself is always done in
Javascript.  Apparently a fairly rich ecosystem of "parse-language-X-using-JS"
exists.

* Ruby is not one of the supported languages (but I'm not sure if maybe some
non-included JS parser for Ruby source exists; it might)

* *Important* There are multiple parsers for many of the languages (and you can
choose among them in [astexplorer.net](http://astexplorer.net)).  However, each
of the many parsers produces a *different* parse tree.  I.e. it's not the AST
actually used by the given programming language, but a representation that some
library happens to use (mostly libraries written to support linters and similar
tools).

* Therefore, for programming languages with multiple parsers available, they
will likely not all agree on whether a given change is actually *semantic*.  In
any case, the AST *representation* of on tool is always going to be
incompatible with the one created by a different tool.

* I feel like this basically eliminates any sort of authoritative "semantic
hash" of a file or project, but it doesn't eliminate the utility of a richer
diff tool.

* **ABANDONED**: After closer examination, the parsers used by `astexplorer`
are not generally of sufficient quality, and this is not a usable approach.

## Early thoughts, part 3

The adopted approach is to write the tool in Golang, but utilize various
external tools to create parse trees or canonicalized representations.  See
other documentation for discussion of this successful approach.
