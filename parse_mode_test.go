package tg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseModeHTML(t *testing.T) {
	assert.Equal(t, "HTML", HTML.Name())
	assert.Equal(t, "Hello World", HTML.Line("Hello", "World"))
	assert.Equal(t, "Hello\nWorld", HTML.Text("Hello", "World"))
	assert.Equal(t, "<b>Hello World</b>", HTML.Bold("Hello", "World"))
	assert.Equal(t, "<i>Hello World</i>", HTML.Italic("Hello", "World"))
	assert.Equal(t, "<u>Hello World</u>", HTML.Underline("Hello", "World"))
	assert.Equal(t, "<s>Hello World</s>", HTML.Strike("Hello", "World"))
	assert.Equal(t, "<tg-spoiler>Hello World</tg-spoiler>", HTML.Spoiler("Hello", "World"))
	assert.Equal(t, "<a href=\"https://telegram.org\">Hello World</a>", HTML.Link("Hello World", "https://telegram.org"))
	assert.Equal(t, "<code>Hello World</code>", HTML.Code("Hello World"))
	assert.Equal(t, "<pre>Hello World</pre>", HTML.Pre("Hello World"))
	assert.Equal(t, "<b>Hello, World</b>", HTML.Sep(", ").Bold("Hello", "World"))
	assert.Equal(t, "Me &amp; You", HTML.Escape("Me & You"))
}

func TestParseModeMarkdown(t *testing.T) {
	assert.Equal(t, "Markdown", MD.Name())
	assert.Equal(t, "Hello World", MD.Line("Hello", "World"))
	assert.Equal(t, "Hello\nWorld", MD.Text("Hello", "World"))
	assert.Equal(t, "*Hello World*", MD.Bold("Hello", "World"))
	assert.Equal(t, "_Hello World_", MD.Italic("Hello", "World"))
	assert.Equal(t, "Hello World", MD.Underline("Hello", "World"))
	assert.Equal(t, "Hello World", MD.Strike("Hello", "World"))
	assert.Equal(t, "Hello World", MD.Spoiler("Hello", "World"))
	assert.Equal(t, "[Hello World](https://telegram.org)", MD.Link("Hello World", "https://telegram.org"))
	assert.Equal(t, "`Hello World`", MD.Code("Hello World"))
	assert.Equal(t, "```Hello World```", MD.Pre("Hello World"))
	assert.Equal(t, "*Hello, World*", MD.Sep(", ").Bold("Hello", "World"))
	assert.Equal(t, "\\*go\\_tg\\*", MD.Escape("*go_tg*"))
}

func TestParseModeMarkdownV2(t *testing.T) {
	assert.Equal(t, "MarkdownV2", MD2.Name())
	assert.Equal(t, "Hello World", MD2.Line("Hello", "World"))
	assert.Equal(t, "Hello\nWorld", MD2.Text("Hello", "World"))

	assert.Equal(t, "*Hello World*", MD2.Bold("Hello", "World"))
	assert.Equal(t, "_Hello World_", MD2.Italic("Hello", "World"))
	assert.Equal(t, "__Hello World__", MD2.Underline("Hello", "World"))
	assert.Equal(t, "~Hello World~", MD2.Strike("Hello", "World"))
	assert.Equal(t, "||Hello World||", MD2.Spoiler("Hello", "World"))
	assert.Equal(t, "[Hello World](https://telegram.org)", MD2.Link("Hello World", "https://telegram.org"))
	assert.Equal(t, "`Hello World`", MD2.Code("Hello World"))
	assert.Equal(t, "```Hello World```", MD2.Pre("Hello World"))
	assert.Equal(t, "*Hello, World*", MD2.Sep(", ").Bold("Hello", "World"))

	assert.Equal(t, "\\[\\*go\\_tg\\*\\]", MD2.Escape("[*go_tg*]"))
	assert.Equal(t, "go\\.tg", MD2.Escape("go.tg"))
}
