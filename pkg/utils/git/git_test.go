package git_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/forsaken628/mapset" // Set of strings

	"github.com/atlantistechnology/sdt/pkg/types"
	"github.com/atlantistechnology/sdt/pkg/utils/git"
)

var options = types.Options{
	Status:    false,
	Semantic:  true,
	Parsetree: false,
	Glob:      "*",
	Minimal:   true,
	Dumbterm:  true,
	Verbose:   false,
	// When we don't do the semantic diff, the wrong type doesn't matter
	// However, a plain `diff -u` happens so we need *some* files
	Source:      "A.rb",
	Destination: "B.rb",
}

var config = types.Config{
	Description: "Configuration for git_test",
	Glob:        "*",
	Commands:    types.Commands,
}

func ExampleOutput() {
	fmt.Fprintf(os.Stdout, "foo")
	// Output: foo
}

// Update `known` as new "native" non-tree-sitter support is added
var known = mapset.New("Go", "JSON", "JavaScript", "SQL", "Python", "Ruby")
var manyExts = []string{
	".1", ".1in", ".1m", ".1x", ".2", ".2da", ".3", ".3in", ".3m", ".3p", ".3pm",
	".3qt", ".3x", ".4", ".4dform", ".4dm", ".4dproject", ".4gl", ".4th", ".5",
	".6", ".6pl", ".6pm", ".7", ".8", ".8xk", ".8xk.txt", ".8xp", ".8xp.txt",
	".9", ".a51", ".abap", ".abnf", ".ada", ".adb", ".adml", ".admx", ".ado",
	".adoc", ".adp", ".ads", ".afm", ".agc", ".agda", ".ahk", ".ahkl", ".aidl",
	".aj", ".al", ".als", ".ampl", ".angelscript", ".anim", ".ant",
	".antlers.html", ".antlers.php", ".antlers.xml", ".apacheconf", ".apib",
	".apl", ".applescript", ".app.src", ".arc", ".arpa", ".as", ".asax",
	".asc", ".asciidoc", ".ascx", ".asd", ".asddls", ".ash", ".ashx", ".asl",
	".asm", ".asmx", ".asn", ".asn1", ".asp", ".aspx", ".asset", ".astro",
	".asy", ".au3", ".aug", ".auk", ".aux", ".avdl", ".avsc", ".aw", ".awk",
	".axaml", ".axd", ".axi", ".axi.erb", ".axml", ".axs", ".axs.erb", ".b",
	".bal", ".bas", ".bash", ".bat", ".bats", ".bb", ".bbx", ".bdf", ".bdy",
	".be", ".befunge", ".bf", ".bi", ".bib", ".bibtex", ".bicep", ".bison",
	".blade", ".blade.php", ".bmx", ".bones", ".boo", ".boot", ".bpl", ".brd",
	".bro", ".brs", ".bs", ".bsl", ".bsv", ".builder", ".builds", ".bzl", ".c",
	".c++", ".cabal", ".cairo", ".cake", ".capnp", ".cats", ".cbl", ".cbx",
	".cc", ".ccp", ".ccproj", ".ccxml", ".cdc", ".cdf", ".cds", ".ceylon",
	".cfc", ".cfg", ".cfm", ".cfml", ".cgi", ".cginc", ".ch", ".chem", ".chpl",
	".chs", ".cil", ".cirru", ".cjs", ".cjsx", ".ck", ".cl", ".cl2", ".clar",
	".click", ".clixml", ".clj", ".cljc", ".cljs", ".cljscm", ".cljs.hl",
	".cljx", ".clp", ".cls", ".clw", ".cmake", ".cmake.in", ".cmd", ".cmp",
	".cnc", ".cob", ".c++-objdump", ".c++objdump", ".c-objdump", ".cobol",
	".cocci", ".code-snippets", "._coffee", ".coffee", ".coffee.md", ".com",
	".command", ".conll", ".conllu", ".coq", ".cp", ".cpp", ".cpp-objdump",
	".cppobjdump", ".cproject", ".cps", ".cpy", ".cql", ".cr", ".crc32",
	".creole", ".cs", ".csc", ".cscfg", ".csd", ".csdef", ".csh", ".cshtml",
	".csl", ".cson", ".csproj", ".css", ".csv", ".csx", ".ct", ".ctl", ".ctp",
	".cts", ".cu", ".cue", ".cuh", ".curry", ".cw", ".cwl", ".cxx",
	".cxx-objdump", ".cy", ".cyp", ".cypher", ".d", ".dae", ".darcspatch",
	".dart", ".dats", ".db2", ".dcl", ".ddl", ".decls", ".depproj", ".desktop",
	".desktop.in", ".dfm", ".dfy", ".dhall", ".di", ".diff", ".dircolors",
	".dita", ".ditamap", ".ditaval", ".djs", ".dll.config", ".dlm", ".dm",
	".do", ".d-objdump", ".dockerfile", ".dof", ".doh", ".dot", ".dotsettings",
	".dpatch", ".dpr", ".druby", ".dsc", ".dsl", ".dsp", ".dsr", ".dtx",
	".duby", ".dwl", ".dyalog", ".dyl", ".dylan", ".e", ".eam.fs", ".eb",
	".ebnf", ".ebuild", ".ec", ".ecl", ".eclass", ".eclxml", ".ecr", ".ect",
	".edc", ".editorconfig", ".edn", ".eex", ".eh", ".ejs", ".ejs.t", ".el",
	".eliom", ".eliomi", ".elm", ".elv", ".em", ".emacs", ".emacs.desktop",
	".emberscript", ".eml", ".env", ".epj", ".eps", ".epsi", ".eq", ".erb",
	".erb.deface", ".erl", ".es", ".es6", ".escript", ".ex", ".exs", ".eye",
	".f", ".f03", ".f08", ".f77", ".f90", ".f95", ".factor", ".fan",
	".fancypack", ".fcgi", ".fea", ".feature", ".filters", ".fish", ".flex",
	".flf", ".flux", ".fnc", ".fnl", ".for", ".forth", ".fp", ".fpp", ".fr",
	".frag", ".frg", ".frm", ".frt", ".fs", ".fsh", ".fshader", ".fsi",
	".fsproj", ".fst", ".fsti", ".fsx", ".fth", ".ftl", ".fun", ".fut", ".fx",
	".fxh", ".fxml", ".fy", ".g", ".g4", ".gaml", ".gap", ".gawk", ".gbl",
	".gbo", ".gbp", ".gbr", ".gbs", ".gco", ".gcode", ".gd", ".gdb",
	".gdbinit", ".ged", ".gemspec", ".geo", ".geojson", ".geom", ".gf", ".gi",
	".gitconfig", ".gitignore", ".gko", ".glade", ".gleam", ".glf", ".glsl",
	".glslf", ".glslv", ".gltf", ".glyphs", ".gmi", ".gml", ".gms", ".gmx",
	".gn", ".gni", ".gnu", ".gnuplot", ".go", ".god", ".golo", ".gp", ".gpb",
	".gpt", ".gql", ".grace", ".gradle", ".graphql", ".graphqls", ".groovy",
	".grt", ".grxml", ".gs", ".gsc", ".gsh", ".gshader", ".gsp", ".gst",
	".gsx", ".gtl", ".gto", ".gtp", ".gtpl", ".gts", ".gv", ".gvy", ".gyp",
	".gypi", ".h", ".h++", ".hack", ".haml", ".haml.deface", ".handlebars",
	".har", ".hats", ".hb", ".hbs", ".hc", ".hcl", ".hh", ".hhi", ".hic",
	".hlean", ".hlsl", ".hlsli", ".hocon", ".hoon", ".hpp", ".hqf", ".hql",
	".hrl", ".hs", ".hs-boot", ".hsc", ".hta", ".htm", ".html", ".html.heex",
	".html.hl", ".html.leex", ".http", ".hx", ".hxml", ".hxsl", ".hxx", ".hy",
	".hzp", ".i", ".i3", ".i7x", ".ice", ".iced", ".icl", ".idc", ".idr",
	".ig", ".ihlp", ".ijm", ".ijs", ".ik", ".ily", ".imba", ".iml", ".inc",
	".ini", ".ink", ".inl", ".ino", ".ins", ".intr", ".io", ".iol", ".ipf",
	".ipp", ".ipynb", ".irclog", ".isl", ".iss", ".iuml", ".ivy", ".ixx", ".j",
	".j2", ".jade", ".jake", ".janet", ".jav", ".java", ".javascript",
	".jbuilder", ".jelly", ".jflex", ".jinja", ".jinja2", ".jison",
	".jisonlex", ".jl", ".jq", "._js", ".js", ".jsb", ".jscad", ".js.erb",
	".jsfl", ".jsh", ".jslib", ".jsm", ".json", ".json5", ".jsonc", ".jsonl",
	".jsonld", ".jsonnet", ".json-tmlanguage", ".jsp", ".jspre", ".jsproj",
	".jss", ".jst", ".jsx", ".kak", ".kicad_mod", ".kicad_pcb", ".kicad_sch",
	".kicad_wks", ".kid", ".kit", ".kml", ".kojo", ".kql", ".krl", ".ksh",
	".ksy", ".kt", ".ktm", ".kts", ".kv", ".l", ".lagda", ".lark", ".las",
	".lasso", ".lasso8", ".lasso9", ".latte", ".launch", ".lbx", ".ld", ".lds",
	".lean", ".lektorproject", ".less", ".lex", ".lfe", ".lgt", ".lhs",
	".libsonnet", ".lid", ".lidr", ".ligo", ".linq", ".liquid", ".lisp",
	".litcoffee", ".livemd", ".ll", ".lmi", ".logtalk", ".lol", ".lookml",
	".lpr", "._ls", ".ls", ".lsl", ".lslp", ".lsp", ".ltx", ".lua", ".lvclass",
	".lvlib", ".lvproj", ".ly", ".m", ".m2", ".m3", ".m4", ".ma", ".mak",
	".make", ".makefile", ".mako", ".man", ".mao", ".markdown", ".marko",
	".mask", ".mat", ".mata", ".matah", ".mathematica", ".matlab", ".mawk",
	".maxhelp", ".maxpat", ".maxproj", ".mbox", ".mc", ".mcfunction",
	".mcmeta", ".mcr", ".md", ".md2", ".md4", ".md5", ".mdoc", ".mdown",
	".mdpolicy", ".mdwn", ".mdx", ".me", ".mediawiki", ".mermaid", ".meta",
	".metal", ".mg", ".minid", ".mint", ".mir", ".mirah", ".mjml", ".mjs",
	".mk", ".mkd", ".mkdn", ".mkdown", ".mkfile", ".mkii", ".mkiv", ".mkvi",
	".ml", ".ml4", ".mli", ".mligo", ".mlir", ".mll", ".mly", ".mm", ".mmd",
	".mmk", ".mms", ".mo", ".mod", ".model.lkml", ".monkey", ".monkey2",
	".moo", ".moon", ".move", ".mpl", ".mps", ".mq4", ".mq5", ".mqh", ".mrc",
	".ms", ".msd", ".mspec", ".mss", ".mt", ".mtl", ".mtml", ".mts", ".mu",
	".mud", ".muf", ".mumps", ".muse", ".mustache", ".mxml", ".mxt", ".mysql",
	".myt", ".n", ".nanorc", ".nas", ".nasl", ".nasm", ".natvis", ".nawk",
	".nb", ".nbp", ".nc", ".ncl", ".ndproj", ".ne", ".nearley", ".neon", ".nf",
	".nginx", ".nginxconf", ".ni", ".nim", ".nimble", ".nim.cfg", ".nimrod",
	".nims", ".ninja", ".nit", ".nix", ".njk", ".njs", ".nl", ".nlogo", ".no",
	".nomad", ".nproj", ".nqp", ".nr", ".nse", ".nsh", ".nsi", ".nss", ".nu",
	".numpy", ".numpyw", ".numsc", ".nuspec", ".nut", ".ny", ".obj",
	".objdump", ".odd", ".odin", ".ol", ".omgrofl", ".ooc", ".opa", ".opal",
	".opencl", ".orc", ".org", ".os", ".osm", ".outjob", ".owl", ".ox", ".oxh",
	".oxo", ".oxygene", ".oz", ".p", ".p4", ".p6", ".p6l", ".p6m", ".p8",
	".pac", ".pan", ".parrot", ".pas", ".pascal", ".pasm", ".pat", ".patch",
	".pb", ".pbi", ".pbt", ".pbtxt", ".pcbdoc", ".pck", ".pcss", ".pd",
	".pddl", ".pde", ".pd_lua", ".pegjs", ".pep", ".per", ".perl", ".pfa",
	".pgsql", ".ph", ".php", ".php3", ".php4", ".php5", ".phps", ".phpt",
	".phtml", ".pic", ".pig", ".pike", ".pir", ".pkb", ".pkgproj", ".pkl",
	".pks", ".pl", ".pl6", ".plantuml", ".plb", ".plist", ".plot", ".pls",
	".plsql", ".plt", ".pluginspec", ".plx", ".pm", ".pm6", ".pml", ".pmod",
	".po", ".pod", ".pod6", ".podsl", ".podspec", ".pogo", ".polar", ".pony",
	".por", ".postcss", ".pot", ".pov", ".pp", ".pprx", ".prawn", ".prc",
	".prefab", ".prefs", ".prg", ".pri", ".prisma", ".prjpcb", ".pro", ".proj",
	".prolog", ".properties", ".props", ".proto", ".prw", ".ps", ".ps1",
	".ps1xml", ".psc", ".psc1", ".psd1", ".psgi", ".psm1", ".pt", ".pub",
	".pug", ".puml", ".purs", ".pwn", ".pxd", ".pxi", ".py", ".py3", ".pyde",
	".pyi", ".pyp", ".pyt", ".pytb", ".pyw", ".pyx", ".q", ".qasm", ".qbs",
	".qhelp", ".ql", ".qll", ".qmd", ".qml", ".qs", ".r", ".r2", ".r3",
	".rabl", ".rake", ".raku", ".rakumod", ".raml", ".raw", ".razor", ".rb",
	".rbbas", ".rbfrm", ".rbi", ".rbmnu", ".rbres", ".rbtbar", ".rbuild",
	".rbuistate", ".rbw", ".rbx", ".rbxs", ".rchit", ".rd", ".rdf", ".rdoc",
	".re", ".reb", ".rebol", ".red", ".reds", ".reek", ".reg", ".regex",
	".regexp", ".rego", ".rei", ".religo", ".res", ".rest", ".rest.txt",
	".resx", ".rex", ".rexx", ".rg", ".rhtml", ".ring", ".riot", ".rkt",
	".rktd", ".rktl", ".rl", ".rmd", ".rmiss", ".rnh", ".rno", ".robot",
	".rockspec", ".roff", ".ronn", ".rpgle", ".rpy", ".rq", ".rs", ".rsc",
	".rsh", ".rs.in", ".rss", ".rst", ".rst.txt", ".rsx", ".rtf", ".ru",
	".ruby", ".rviz", ".s", ".sage", ".sagews", ".sas", ".sass", ".sats",
	".sbt", ".sc", ".scad", ".scala", ".scaml", ".scd", ".sce", ".scenic",
	".sch", ".schdoc", ".sci", ".scm", ".sco", ".scpt", ".scrbl", ".scss",
	".scxml", ".sdc", ".sed", ".self", ".service", ".sexp", ".sfd", ".sfproj",
	".sfv", ".sh", ".sha1", ".sha2", ".sha224", ".sha256", ".sha256sum",
	".sha3", ".sha384", ".sha512", ".shader", ".shen", ".sh.in", ".shproj",
	".sh-session", ".sieve", ".sig", ".sj", ".sjs", ".sl", ".sld", ".slim",
	".sln", ".sls", ".sma", ".smali", ".smithy", ".smk", ".sml", ".smt",
	".smt2", ".snap", ".snip", ".snippet", ".snippets", ".sol", ".soy", ".sp",
	".sparql", ".spc", ".spec", ".spin", ".sps", ".sqf", ".sql", ".sqlrpgle",
	".sra", ".srdf", ".srt", ".sru", ".srw", ".ss", ".ssjs", ".sss", ".st",
	".stan", ".star", ".sthlp", ".stl", ".ston", ".story", ".storyboard",
	".sttheme", ".sty", ".styl", ".sublime-build", ".sublime-commands",
	".sublime-completions", ".sublime-keymap", ".sublime-macro",
	".sublime-menu", ".sublime_metrics", ".sublime-mousemap",
	".sublime-project", ".sublime_session", ".sublime-settings",
	".sublime-snippet", ".sublime-syntax", ".sublime-theme",
	".sublime-workspace", ".sv", ".svelte", ".svg", ".svh", ".swift",
	".syntax", ".t", ".tab", ".tac", ".tag", ".talon", ".targets", ".tcc",
	".tcl", ".tcl.in", ".tcsh", ".te", ".tea", ".tesc", ".tese", ".tex",
	".texi", ".texinfo", ".textile", ".textproto", ".tf", ".tfstate",
	".tfstate.backup", ".tfvars", ".thor", ".thrift", ".thy", ".tl", ".tla",
	".tm", ".tmac", ".tmcommand", ".tml", ".tmlanguage", ".tmpreferences",
	".tmsnippet", ".tmtheme", ".tmux", ".toc", ".toml", ".tool", ".topojson",
	".tpb", ".tpl", ".tpp", ".tps", ".trg", ".ts", ".tst", ".tsv", ".tsx",
	".ttl", ".tu", ".twig", ".txi", ".txl", ".txt", ".uc", ".udf", ".udo",
	".ui", ".unity", ".uno", ".upc", ".ur", ".urdf", ".url", ".urs", ".ux",
	".v", ".vala", ".vapi", ".vark", ".vb", ".vba", ".vbhtml", ".vbproj",
	".vbs", ".vcl", ".vcxproj", ".vdf", ".veo", ".vert", ".vh", ".vhd",
	".vhdl", ".vhf", ".vhi", ".vho", ".vhost", ".vhs", ".vht", ".vhw",
	".view.lkml", ".vim", ".vimrc", ".viw", ".vmb", ".volt", ".vrx", ".vsh",
	".vshader", ".vsixmanifest", ".vssettings", ".vstemplate", ".vtl", ".vtt",
	".vue", ".vw", ".vxml", ".vy", ".w", ".wast", ".wat", ".watchr", ".wdl",
	".webapp", ".webidl", ".webmanifest", ".weechatlog", ".whiley", ".wiki",
	".wikitext", ".wisp", ".wixproj", ".wl", ".wlk", ".wlt", ".wlua",
	".workbook", ".workflow", ".wren", ".ws", ".wsdl", ".wsf", ".wsgi", ".wxi",
	".wxl", ".wxs", ".x", ".x10", ".x3d", ".x68", ".xacro", ".xaml", ".xbm",
	".xc", ".xdc", ".xht", ".xhtml", ".xi", ".xib", ".xlf", ".xliff", ".xm",
	".xmi", ".xml", ".xml.dist", ".xmp", ".xojo_code", ".xojo_menu",
	".xojo_report", ".xojo_script", ".xojo_toolbar", ".xojo_window", ".xpl",
	".xpm", ".xproc", ".xproj", ".xpy", ".xq", ".xql", ".xqm", ".xquery",
	".xqy", ".xrl", ".xs", ".xsd", ".xsh", ".xsjs", ".xsjslib", ".xsl",
	".xslt", ".xsp-config", ".xspec", ".xsp.metadata", ".xtend", ".xul",
	".xzap", ".y", ".yacc", ".yaml", ".yaml.sed", ".yaml-tmlanguage", ".yang",
	".yap", ".yar", ".yara", ".yasnippet", ".yml", ".yml.mysql", ".yrl",
	".yul", ".yy", ".yyp", ".zap", ".zcml", ".zeek", ".zep", ".zig", ".zil",
	".zimpl", ".zmpl", ".zone", ".zpl", ".zs", ".zsh", ".zsh-theme",
}

func TestRubyExt(t *testing.T) {
	exts := []string{
		".rb", ".rake", ".gemspec", ".god", ".irbrc", ".mspec",
		".pluginspec", ".podspec", ".rabl", ".rbuild", ".rbw",
		".rbx", ".ru", ".ruby", ".thor", ".watchr"}
	for _, ext := range exts {
		_, name, err := git.FileComparer(ext)
		if err != nil || name != "Ruby" {
			t.Fatalf(`Failed to identify Ruby filetype with extension %s`, ext)
		}
	}
}

func TestPythonExt(t *testing.T) {
	exts := []string{".py", ".pyw", ".pyde", "pyt"}
	for _, ext := range exts {
		_, name, err := git.FileComparer(ext)
		if err != nil || name != "Python" {
			t.Fatalf(`Failed to identify Python filetype with extension %s`, ext)
		}
	}
}

func TestSQLExt(t *testing.T) {
	exts := []string{
		".sql", ".pls", ".bdy", ".ddl", ".fnc", ".pck", ".pkb", ".pks",
		".pgsql", ".plb", ".plsql", ".prc", ".spc", ".sql", ".tpb", ".tps",
		".trg", ".vw"}
	for _, ext := range exts {
		_, name, err := git.FileComparer(ext)
		if err != nil || name != "SQL" {
			t.Fatalf(`Failed to identify SQL filetype with extension %s`, ext)
		}
	}
}

func TestJSExt(t *testing.T) {
	exts := []string{".js", ".jsx", ".mdx", ".cjs", ".mjs", ".es", ".es6"}
	for _, ext := range exts {
		_, name, err := git.FileComparer(ext)
		if err != nil || name != "JavaScript" {
			t.Fatalf(`Failed to identify JavaScript filetype with extension %s`, ext)
		}
	}
}

func TestJSONExt(t *testing.T) {
	exts := []string{
		".json", ".json5", ".4dform", ".4dproject", ".avsc", ".geojson", ".gltf",
		".har", ".ice", ".json-tmlanguage", ".jsonl", ".mcmeta", ".tfstate",
		".tfstate.backup", ".topojson", ".webapp", ".webmanifest", ".yy", ".yyp"}
	for _, ext := range exts {
		_, name, err := git.FileComparer(ext)
		if err != nil || name != "JSON" {
			t.Fatalf(`Failed to identify JSON filetype with extension %s`, ext)
		}
	}
}

func TestGoExt(t *testing.T) {
	exts := []string{".go", ".v"}
	for _, ext := range exts {
		_, name, err := git.FileComparer(ext)
		if err != nil || name != "Go" {
			t.Fatalf(`Failed to identify Go filetype with extension %s`, ext)
		}
	}
}

func TestFallBackExt(t *testing.T) {
	for _, ext := range manyExts {
		_, name, err := git.FileComparer(ext)
		if err == nil && !known.Contains(name) {
			t.Fatalf(`Failed to return an error when filetype is not known: %s`, ext)
		}
		if err != nil && name != "Tree-sitter?" {
			t.Fatalf(`Failed to signal tree-sitter when native parser absent: %s`, ext)
		}
	}
}
