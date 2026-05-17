package service

import (
	"strings"
	"testing"

	nethtml "golang.org/x/net/html"
)

func TestAppendHTMLTextDocumentsCurrentSpacingAndSkipRules(t *testing.T) {
	t.Run("nil node is ignored", func(t *testing.T) {
		var builder strings.Builder
		appendHTMLText(&builder, nil)
		if got := builder.String(); got != "" {
			t.Fatalf("appendHTMLText(nil) = %q, want empty", got)
		}
	})

	t.Run("text nodes receive trailing space", func(t *testing.T) {
		var builder strings.Builder
		appendHTMLText(&builder, &nethtml.Node{Type: nethtml.TextNode, Data: "Alpha"})
		if got := builder.String(); got != "Alpha " {
			t.Fatalf("appendHTMLText(text) = %q, want current trailing-space behavior", got)
		}
	})

	t.Run("br writes both immediate and block trailing spaces", func(t *testing.T) {
		var builder strings.Builder
		appendHTMLText(&builder, &nethtml.Node{Type: nethtml.ElementNode, Data: "br"})
		if got := builder.String(); got != "  " {
			t.Fatalf("appendHTMLText(br) = %q, want two spaces", got)
		}
	})

	t.Run("block and inline elements preserve current flattened spacing", func(t *testing.T) {
		paragraph := &nethtml.Node{Type: nethtml.ElementNode, Data: "p"}
		paragraph.AppendChild(&nethtml.Node{Type: nethtml.TextNode, Data: "Alpha"})
		span := &nethtml.Node{Type: nethtml.ElementNode, Data: "span"}
		span.AppendChild(&nethtml.Node{Type: nethtml.TextNode, Data: "Beta"})
		paragraph.AppendChild(span)

		var builder strings.Builder
		appendHTMLText(&builder, paragraph)
		if got := builder.String(); got != "Alpha Beta  " {
			t.Fatalf("appendHTMLText(paragraph) = %q, want current flattened spacing", got)
		}
	})

	t.Run("script style and svg subtrees are skipped", func(t *testing.T) {
		for _, tag := range []string{"script", "style", "svg"} {
			node := &nethtml.Node{Type: nethtml.ElementNode, Data: tag}
			node.AppendChild(&nethtml.Node{Type: nethtml.TextNode, Data: "hidden"})

			var builder strings.Builder
			appendHTMLText(&builder, node)
			if got := builder.String(); got != "" {
				t.Fatalf("appendHTMLText(%s subtree) = %q, want skipped", tag, got)
			}
		}
	})
}

func TestPlainTextFromHTMLDocumentsAdditionalCurrentBranches(t *testing.T) {
	t.Run("trims blank input before parsing", func(t *testing.T) {
		if got := plainTextFromHTML(" \n\t "); got != "" {
			t.Fatalf("plainTextFromHTML(blank) = %q, want empty", got)
		}
	})

	t.Run("fallback sanitizer ignores comments and leaves escaped entities", func(t *testing.T) {
		input := `Alpha &amp; <strong>Beta</strong><!-- hidden --><p>Gamma&nbsp;Delta</p>`
		want := "Alpha &amp; BetaGamma\u00a0Delta"
		if got := plainTextFromHTML(input); got != want {
			t.Fatalf("plainTextFromHTML() = %q, want %q", got, want)
		}
	})

	t.Run("fallback sanitizer skips script and style but keeps svg text", func(t *testing.T) {
		input := `<div>Visible<script>alert(1)</script><style>.x{}</style><svg><text>icon</text></svg><span>Tail</span></div>`
		want := "VisibleiconTail"
		if got := plainTextFromHTML(input); got != want {
			t.Fatalf("plainTextFromHTML(skip subtrees) = %q, want %q", got, want)
		}
	})

	t.Run("malformed fragments are fallback-sanitized", func(t *testing.T) {
		input := `<p>One<span>Two</p>Three`
		want := "OneTwoThree"
		if got := plainTextFromHTML(input); got != want {
			t.Fatalf("plainTextFromHTML(malformed) = %q, want current fallback-sanitized text %q", got, want)
		}
	})
}
