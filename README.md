# Semantic Diff Tool (sdt)

The command-line tool `sdt` compares source files to identify which changes 
create semantic differences in the program operation, and specifically to 
exclude many changes which cannot be *functionally important* to the operation
of a program or library.

Use of `sdt` will allow code reviewers or submitters to assure that 
modifications made to improve stylistic formatting of the code—whether
made by hand or using code-formatting tools—does not modify the underlying
*meaning* of the code.

As designed, the tool is much more likely to produce false positives for
the presence of semantic changes than false negatives.  That is to say, `sdt`
might indicate that a certain segment of the diff between versions is *likely*
to contain a semantic difference in code, but upon human examination, a developer
might decide that no actual behavior will change (or, of course, she might 
decide that the change in code behavior is a desired change).

It is unlikely that `sdt` will identify an overall file change, or any change 
to a particular segment of a diff as semantically irrelevant where that change
actually does change behavior.  However, this tool provides NO WARRANTY, and
it remains up to your human developers and your CI/CD process to make final
decisions on whether to accept a given change.

## Future plans

It would be nice to integrate `sdt` as a git subcommand, which should be
straightforward.

It would also be nice to allow `sdt` to be used as an integration or extension 
to GitHub or other collaborative development services (Bitbucket, GitLab, etc) 
such that views of pull requests could be accompanied by the analysis `sdt` 
provides.

## Related tools

### Difftastic

[Difftastic](https://github.com/Wilfred/difftastic) (`difft`) serves a largely
overlapping purpose to `sdt`.  Difftastic builds on the substantial and
longstanding work in [tree-sitter](https://github.com/tree-sitter/tree-sitter),
which is a parser generator tool for which many grammars and interfaces have
been created.

Semantic Diff Tool is much newer and less developed; but `sdt` also takes a
much more heuristic approach.  That is, `difft` will answer the question
"*exactly* what has changed between these two versions of a file at an AST
level?"  In contrast, `sdt` aims to answer the somewhat narrower question
"which segments of the diff are likely to be semantically important?"
Specifically, `sdt` is meant to aid code reviewers in excluding merely
stylistic changes and focus on (likely) functional changes.

This difference is reflected in the different approaches to comparing files the
two tools take.  Instead of relying on fixed and compiled-in versions of
parsers as `difft` does, `sdt` dynamically calls external tools (which can be
whichever specific version of those is used by a project).  

As a consequence, using `difft` with a project that, e.g., utilizes Python 3.8
is difficult to impossible.  At the least, it would require recompiling the
Rust tool with dependencies on whichever older version of
[tree-sitter-python](https://github.com/tree-sitter/tree-sitter-python) has
specifically 3.8-level syntax.  In contrast, `sdt` can be easily and
dynamically configured to utilize an appropriate Python executable, such as
`/usr/local/bin/python3.8`.

As an illustration, in the following screenshots, `git diff` shows three
segments where surface changes were made to a small test file.  The first and
third—but not the second—segment is semantically meaningful to what the program
does.  Both `difft` and `sdt` correctly identify this fact; but `sdt` presents
a display closely modeled on `git diff` itself while `difft` highlights
*exactly* those characters that represent a change (using a somewhat
idiosyncratic, but clear, format).  These examples are included as screenshots
rather than as marked text to preserve the use of color highlighting by all
three tool.

```
% git diff
```

<img src="assets/img/diff-python-git.png" alt="git diff" width="50%"/>

```
% GIT_EXTERNAL_DIFF='difft --display=inline' git diff
```

<img src="assets/img/difft-python-git.png" alt="difft as GIT_EXTERNAL_DIFF" width="50%"/>

```
% sdt semantic
```

<img src="assets/img/sdt-python-git.png" alt="sdt semantic" width="50%/">

### AST Explorer

[AST Explorer](https://github.com/fkling/astexplorer) is somewhat similar in
concept to Difftastic.  It is written in JavaScript, and uses language-native
parsers to support numerous programming languages.  It does not attempt to
provide diff'ing, but adding that would be relatively straightforward.  That
was, in fact, the initial but discarded approach taken by the creator of `sdt`.

However, after some initial development, I realized that the quality of the
third-party JS parsers that AST Explorer utilizes are often poor, and fail to
recognize many commonplace constructs of the languages they respectively
process.  AST Explorer itself exposes a very nice [web-based front
end](https://astexplorer.net/), but beyond the sample files of many source
languages it provides tool often fails to parse valid source code.

# Supported languages

Much of the work that Semantic Diff Tool accomplishes is done by means of utilizing
other tools.  You will need to install those other tools in your development 
environment separately.  However, this requirement is generally fairly trivial,
since the tools used are often the underlying runtime engines or compilers for the
very same programming languages of those files whose changes are analyzed (in 
other words, the programming languages your project uses).

The configuration file `$HOME/.sdt.toml` allows you to choose specific versions of
tools and specific switches to use.  This is useful especially if a particular 
project utilizes a different version of a programming language than the one 
installed to the default path of a development environment.  Absent an overriding
configuration, each tool is assumed to reside on your $PATH, and a default 
collection of flags and switches are used.

For example, for Ruby files, the default command `ruby --dump=parsetree` is used
to create an AST of the file being analyzed.  Similarly, for Python files, 
`python -m ast -a` is used for the same purpose.  Other tools produce canonical 
representations rather than ASTs, depending on what best serves the needs of
a particular language (and depending on what tools are available and their 
quality).  While overriding the configuration between different version of a 
programming language or tool will *probably* not break the code that performs the
semantic comparison, not all languages have been tested in all versions (especially
for versions that will be created in the future and do not yet exist).

## Ruby

Initial support created.  Supports both --semantic and --parsetree flags.

Only a `ruby` interpreter is required (by default)

## Python

Initial support created.  Supports both --semantic and --parsetree flags.

Only a `python` interpreter is required (by default)

## SQL

Initial support created.  Uses canonicalization rather than parsing, so
only the --semantic flag is supported.

Requires the tool `sqlformat` (by default).  See:

* https://github.com/andialbrecht/sqlparse
* https://manpages.ubuntu.com/manpages/jammy/man1/sqlformat.1.html

## JavaScript

TODO.  Supports both --semantic and --parsetree flags.

Requires the `node` interpreter and the library `acorn` (by default). See:

* https://github.com/acornjs/acorn/tree/master/acorn/

Use of the error-tolerant variant `acorn-loose` was contemplated and
rejected (as a default).  We believe this tool would be more likely to
produce spurious difference in parse trees. See:

* https://github.com/acornjs/acorn/tree/master/acorn-loose/

Note that Node 18.10 was used during development. However, any node version
supported by Acorn should work identically.  However, if you wish to treat
the files being analyzed as a specific ECMAScript version, see the option
`ecmaVersion` that can be configured in `.sdt.toml` and is discussed in the
sample version of that file.

## Golang

TODO

## Others

What do you want most?

