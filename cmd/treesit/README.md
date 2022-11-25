This small tool is a wrapper around the tree-sitter-cli framework See:
https://github.com/tree-sitter/tree-sitter.  The tree-sitter library provides
parsing support for a large number languages, via subpackages which provide
grammars for these various languages.  Such grammars must each be installed
using the framework, and rely on Rust, Node.js, and C, C++, and bash tooling.

For the current tool, we make a call to `tree-sitter parse file.lang`, then
if language support is available for a given file, we massage the produced
parse tree to include features needed by Semantic Diff Tool, and into a 
format easier to process by SDT.  Specifically, tree-sitter might provide a
parse tree such as:

    % tree-sitter parse samples/hello0.c
    (translation_unit [0, 0] - [15, 0]
      (comment [0, 0] - [0, 44])
      (comment [2, 0] - [2, 41])
      (preproc_include [3, 0] - [4, 0]
        path: (system_lib_string [3, 9] - [3, 18]))
     [...]
       (function_definition [7, 0] - [14, 1]
         type: (primitive_type [7, 0] - [7, 3])
         declarator: (function_declarator [7, 4] - [7, 10]
           declarator: (identifier [7, 4] - [7, 8])
     [...]
             (string_literal [11, 11] - [11, 24]))))
       (return_statement [13, 4] - [13, 13]
         (number_literal [13, 11] - [13, 12])))))

For SDT we want several things to be different, since our interest is simply
in identifying which lines might contain semantically meaningful changes.
Tree-sitter gives us both too much and too little for this goal, but in a
manner where we can mechanically transform the tree to the desired form.
The transformed format is (generally) compatible with that produced by
`gotree`.

Differences in line number and column number are not semantically meaningful
(in most programming languages), however changes in names and literals ARE 
important.  Comments are discarded by `treesit` unless the env variable
TREESIT_COMMENTS is set to a non-blank value.

To recover lines, we move their numbers to the lefthand column.  We also
fill in those literals within the source file that are important to us based
on their line/col offset.  For example (note that SDT wants 1-based line
numbers whereas tree-sitter uses 0-based):

    % treesit samples/hello0.c
    SrcLn | Node
    00001 | (translation_unit
    00004 |   (preproc_include
    00004 |     path: (system_lib_string <stdio.h>))
    [...]
    00008 |    (function_definition
    00008 |      type: (primitive_type)
    00008 |      declarator: (function_declarator
    00008 |        declarator: (identifier main)
    [...]
    00012 |          (string_literal "Hello World"))))
    00014 |    (return_statement
    00014 |      (number_literal 0)))))


