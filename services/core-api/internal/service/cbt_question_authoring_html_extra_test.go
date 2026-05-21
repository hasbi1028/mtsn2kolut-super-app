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

func TestSanitizeHTMLPreservesBankSoalTableStructure(t *testing.T) {
	input := `<table><thead><tr><th>Mapel</th><th>Nilai</th></tr></thead><tbody><tr><td>Matematika</td><td>90</td></tr></tbody></table>`

	got := sanitizeHTML(input)

	for _, want := range []string{
		"<table>", "<thead>", "<tr>", "<th>Mapel</th>", "<th>Nilai</th>",
		"<tbody>", "<td>Matematika</td>", "<td>90</td>", "</table>",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("sanitizeHTML(table) = %q, want to contain %q", got, want)
		}
	}
}

func TestSanitizeHTMLRemovesDangerousPayloadInsideBankSoalTable(t *testing.T) {
	input := `<table><tbody><tr onclick="alert(1)"><td style="color:#112233; background:url(javascript:alert(7))" onmouseover="alert(2)">A<script>alert(3)</script><a href="javascript:alert(4)">tautan</a><img src="javascript:alert(5)" onerror="alert(6)"></td></tr></tbody></table>`

	got := sanitizeHTML(input)

	for _, want := range []string{"<table>", "<tbody>", "<tr>", "<td", "A", "tautan"} {
		if !strings.Contains(got, want) {
			t.Fatalf("sanitizeHTML(dangerous table) = %q, want to contain %q", got, want)
		}
	}
	for _, forbidden := range []string{"onclick", "onmouseover", "onerror", "javascript:", "<script", "alert("} {
		if strings.Contains(strings.ToLower(got), forbidden) {
			t.Fatalf("sanitizeHTML(dangerous table) = %q, must not contain %q", got, forbidden)
		}
	}
}

func TestSanitizeHTMLDocumentsTableHeaderSpanAttributes(t *testing.T) {
	input := `<table><thead><tr><th colspan="2" rowspan="1" data-extra="drop" style="text-align:center;color:#ff0000">Kompetensi</th></tr></thead><tbody><tr><td colspan="3" rowspan="2">Baris 1</td></tr></tbody></table>`

	got := sanitizeHTML(input)

	for _, want := range []string{
		"<th", "colspan=\"2\"", "rowspan=\"1\"", "text-align: center", "color: #ff0000", "Kompetensi</th>",
		"<td", "colspan=\"3\"", "rowspan=\"2\"", "Baris 1</td>",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("sanitizeHTML(table spans) = %q, want to contain %q", got, want)
		}
	}
	if strings.Contains(got, "data-extra") {
		t.Fatalf("sanitizeHTML(table spans) = %q, data-extra should not be preserved", got)
	}
}
