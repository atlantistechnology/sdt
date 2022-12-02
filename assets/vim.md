---
title: Vimscript function and key binding
date: November 28, 2022
---

# A custom vim shortcut

A full plugin is not really needed for vim integration. Rather, a simple
command and a keystroke binding should be sufficient.  We welcome
contribution of a full plug-in that adds enhanced capabilities.

The main author of SDT has the following in his `.vimrc`

```vimscript
" Simply use a dedicated temporary file to hold the latest
" analysis by Semantic Diff Tool. Load as necessary.
function! SemanticDiffTool()
    execute "! sdt semantic -d -m > /tmp/sdt"
    execute "e /tmp/sdt"
endfunction
```

Binding a keystroke is probably useful:

```vimscript
" David has a preference for shortcuts beginning with comma.
" Choose whatever binding matches your way of working.
map ,s :call SemanticDiffTool()<CR> " run sdt semantic to buffer
```
