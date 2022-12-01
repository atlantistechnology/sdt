# Semantic Diff Tool (sdt)

The command-line tool `sdt` compares source files to identify which changes
create semantic differences in the program operation, and specifically to
exclude many changes which cannot be *functionally important* to the
operation of a program or library.

Use of `sdt` will allow code reviewers or submitters to assure that
modifications made to improve stylistic formatting of the code—whether made
by hand or using code-formatting tools—does not modify the underlying
*meaning* of the code.

The operation of the utility, pre-compiled binaries, bundled support tools,
installation details, language support, integration with editors and version
control management, and other details are discussed at the [Semantic Diff
Tool Assets](https://www.sdt.dev) site.

A quick example of usage shows an overview of capabilities (the switches
added simply minimize the output context and remove reliance on colorized
output).  The comparison shown compares what is currently on disk to the
HEAD of the working git branch (many other combinations are enabled with
command flags).

```
% sdt semantic --dumbterm --minimal 2>/dev/null
Changes to be committed:
    modified:   .github/workflows/test-treesit.yaml
| No available semantic analyzer for this format
    new file:   pkg/types/types_test.go
Changes not staged for commit:
    modified:   samples/filter.rb
| Segments with likely semantic changes
| @@ -3,4 +3,4 @@ def mod5?(items)
| -puts mod5? 1..100
| +puts mod5? 1..50
    modified:   samples/funcs.py
| No semantic differences detected
    modified:   samples/running-total.sql
|        count(DISTINCT co.order_id) AS {{-num_}}order{{-s}}{{+_count}},
```

## Installation

For users wishing to aid in developing SDT, or who simply wish to install
from source, you may install `sdt` by cloning this repository, and
installing the tool(s) using `go install`.

For example:

```bash
% git clone https://github.com/atlantistechnology/sdt.git
% cd sdt
% go install ./...
```

For installation of binaries, see the [asset site](https://www.sdt.dev).

# Supported languages

Much of the work that Semantic Diff Tool accomplishes is done by means of
utilizing other tools.  You will need to install those other tools in your
development environment separately.  However, this requirement is generally
fairly trivial, since the tools used are often the underlying runtime
engines or compilers for the very same programming languages of those files
whose changes are analyzed (in other words, the programming languages your
project uses).

An additional "fallback" means of supporting programming languages for
analysis is using the `tree-sitter-cli` parser/generator with any grammars
that happen to be installed.  This loses the ability to tailor analysis to a
specific language version, but adds many additional languges.  See the main
documentation for details.

In one manner or another (or via multiple, configurable options), `sdt` can
support:

| Mechanism     | Languages 
| ------------- | ------------------------------------------------------- 
| Bundled tools | Ruby, Python, SQL, JavaScript, JSON, Golang
| Tree-sitter   | Agda; Bash; C; C#; C++; Common Lisp; CSS; CUDA;
|               | Dockerfile; DOT; Elixir; Elm; Emacs Lisp; Eno; ERB/EJS; 
|               | Erlang; Fennel; GLSL (OpenGL Shading Language); Hack;
|               | Haskell; HCL; HTML; Java; Julia; Kotlin; Lua; Make;
|               | Markdown; Nix; Objective-C; OCaml; Org; Perl; PHP;
|               | Protocol Buffers; R; Racket; Rust; Scala; S-expressions;
|               | Sourcepawn; SPARQL; Svelte; Swift; SystemRDL; TOML;
|               | Turtle; Twig; TypeScript; Verilog; VHDL; Vue; WASM;
|               | WGSLi; WebGPU Shading Language; YAML.

