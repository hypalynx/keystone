# Keystone

A small Golang library that makes working with the stdlib `html/template` easy!

## Installation

```bash
go get github.com/hypalynx.com/keystone
```

## Description

There are many other cool templating projects out there, I've found that
`html/template` does the job well and offers some advantages:

- It's simple, when something breaks it's easier to debug/maintain.
- Hot-reloading is much faster when you don't have to re-compile your whole service.

Keystone builds on this by providing a way to setup `html/template` in a way that supports:

- Hot-reloading, using embed for production but reading from disk in
  development (or as you prefer to configure it).
- Categorising templates into `layouts`, `pages`, `components` and `partials`
  so that you can use `{{ define "content" }}` in your pages and reuse layouts
  easily while also making sure nothing else is overwritten.
- Subdir support for the categories too e.g `components/product/card.tmpl`

## How I work with templates

_This is how I setup my editor to work better with `html/template` and make the
experience that little bit nicer._

- I use neovim (btw) with `gotmpl` and `html` treesitter plugins configured
  with this injection, which allows both to work together in the same file:
```bash
mkdir -p "$NEOVIM_PATH/queries/gotmpl" # probably ~/.config/nvim
echo '((text) @injection.content
 (#set! injection.language "html")
 (#set! injection.combined))' > "$NEOVIM_PATH/queries/gotmpl/injections.scm"
```
- Treesitter also supports formatting the file which you can configure to do on save.
- I use neovim with the `html`, `htmx` & `tailwindcss` lsps too.
