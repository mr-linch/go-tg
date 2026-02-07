package tg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseModeHTML(t *testing.T) {
	assert.Equal(t, "HTML", HTML.String())
	marshaled, err := HTML.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, "HTML", string(marshaled))

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
	assert.Equal(t, "<blockquote>Hello, World</blockquote>", HTML.Sep(", ").Blockquote("Hello", "World"))
	assert.Equal(t, "Me &amp; You", HTML.Escape("Me & You"))
	assert.Equal(t, "Hello, <b>World</b>!", HTML.Escapef("Hello, %s!", HTML.Bold("World")))

	assert.Equal(t, `<a href="tg://user?id=123456789">John</a>`, HTML.Mention("John", 123456789))
	assert.Equal(t, `<tg-emoji emoji-id="5368324170671202286">👍</tg-emoji>`, HTML.CustomEmoji("👍", "5368324170671202286"))
	assert.Equal(t, `<pre><code class="language-python">print("hi")</code></pre>`, HTML.PreLanguage("python", `print("hi")`))
	assert.Equal(t, `<blockquote expandable>Hello World</blockquote>`, HTML.ExpandableBlockquote("Hello World"))
	assert.Equal(t, `<blockquote expandable>Hello, World</blockquote>`, HTML.Sep(", ").ExpandableBlockquote("Hello", "World"))
}

func TestParseModeMarkdown(t *testing.T) {
	assert.Equal(t, "Markdown", MD.String())
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
	assert.Equal(t, "Hello, *World*!", MD.Escapef("Hello, %s!", MD.Bold("World")))

	assert.Equal(t, `[John](tg://user?id=123456789)`, MD.Mention("John", 123456789))
	assert.Equal(t, "👍", MD.CustomEmoji("👍", "5368324170671202286"))
	assert.Equal(t, "```python\nprint(\"hi\")```", MD.PreLanguage("python", `print("hi")`))
	assert.Equal(t, "Hello World", MD.ExpandableBlockquote("Hello World"))
}

func TestParseModeMarkdownV2(t *testing.T) {
	assert.Equal(t, "MarkdownV2", MD2.String())
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
	assert.Equal(t, ">Hello World", MD2.Blockquote("Hello World"))
	assert.Equal(t, ">line1\n>line2", MD2.Blockquote("line1\nline2"))
	assert.Equal(t, ">line1\n>line2", MD2.Sep("\n").Blockquote("line1", "line2"))
	assert.Equal(t, "*Hello, World*", MD2.Sep(", ").Bold("Hello", "World"))

	assert.Equal(t, "\\[\\*go\\_tg\\*\\]", MD2.Escape("[*go_tg*]"))
	assert.Equal(t, "go\\.tg", MD2.Escape("go.tg"))
	assert.Equal(t, "*bold* \\| _italic_", MD2.Escapef("%s | %s", MD2.Bold("bold"), MD2.Italic("italic")))
	assert.Equal(t, "Hello\\.", MD2.Escapef("Hello."))

	assert.Equal(t, `[John](tg://user?id=123456789)`, MD2.Mention("John", 123456789))
	assert.Equal(t, `![👍](tg://emoji?id=5368324170671202286)`, MD2.CustomEmoji("👍", "5368324170671202286"))
	assert.Equal(t, "```python\nprint(\"hi\")```", MD2.PreLanguage("python", `print("hi")`))
	assert.Equal(t, ">Hello World||", MD2.ExpandableBlockquote("Hello World"))
	assert.Equal(t, ">line1\n>line2||", MD2.ExpandableBlockquote("line1\nline2"))
}
