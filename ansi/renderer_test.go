package ansi

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/x/exp/golden"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

const (
	examplesDir = "../styles/examples/"
	issuesDir   = "../testdata/issues/"
)

func TestCodeBlockPreservesSourceSegments(t *testing.T) {
	for _, tc := range []struct {
		name, source, code, language string
	}{
		{"fenced", "```go\nleft\t世界\nright\n```\n", "left\t世界\nright\n", "go"},
		{"open fence", "```go\nfirst\nunfinished", "first\nunfinished\n", "go"},
		{"indented", "    left\n      right\n", "left\n  right\n", ""},
		{"tabs", "\t  left\n\t\tright\n", "  left\n\tright\n", ""},
		{"quoted fence", "> ```go\n> left\n> right\n> ```\n", "left\nright\n", "go"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := []byte(tc.source)
			doc := goldmark.New().Parser().Parse(text.NewReader(source))
			blocks := 0
			err := ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
				if entering && (node.Kind() == ast.KindFencedCodeBlock || node.Kind() == ast.KindCodeBlock) {
					blocks++
					code := NewRenderer(Options{}).NewElement(node, source).Renderer.(*CodeBlockElement)
					if code.Code != tc.code || code.Language != tc.language {
						t.Errorf("code = %q, language = %q; want %q, %q", code.Code, code.Language, tc.code, tc.language)
					}
				}
				return ast.WalkContinue, nil
			})
			if err != nil || blocks != 1 {
				t.Fatalf("walk error = %v, code blocks = %d; want one", err, blocks)
			}
		})
	}
}

func TestRenderer(t *testing.T) {
	files, err := filepath.Glob(examplesDir + "*.md")
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		bn := strings.TrimSuffix(filepath.Base(f), ".md")
		t.Run(bn, func(t *testing.T) {
			sn := filepath.Join(examplesDir, bn+".style")

			in, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			b, err := os.ReadFile(sn)
			if err != nil {
				t.Fatal(err)
			}

			options := Options{
				WordWrap: 80,
			}
			err = json.Unmarshal(b, &options.Styles)
			if err != nil {
				t.Fatal(err)
			}

			switch bn {
			case "table_wrap":
				tableWrap := true
				options.TableWrap = &tableWrap
			case "table_truncate":
				tableWrap := false
				options.TableWrap = &tableWrap
			case "table_with_inline_links":
				options.InlineTableLinks = true
			case "table_with_footer_links", "table_with_footer_links_no_color":
				options.InlineTableLinks = false
			}

			md := goldmark.New(
				goldmark.WithExtensions(
					extension.GFM,
					extension.DefinitionList,
					emoji.Emoji,
				),
				goldmark.WithParserOptions(
					parser.WithAutoHeadingID(),
				),
			)

			ar := NewRenderer(options)
			md.SetRenderer(
				renderer.NewRenderer(
					renderer.WithNodeRenderers(util.Prioritized(ar, 1000))))

			var buf bytes.Buffer
			if err := md.Convert(in, &buf); err != nil {
				t.Error(err)
			}

			golden.RequireEqual(t, buf.Bytes())
		})
	}
}

func TestRendererIssues(t *testing.T) {
	files, err := filepath.Glob(issuesDir + "*.md")
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		bn := strings.TrimSuffix(filepath.Base(f), ".md")
		t.Run(bn, func(t *testing.T) {
			in, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			b, err := os.ReadFile("../styles/dark.json")
			if err != nil {
				t.Fatal(err)
			}

			options := Options{
				WordWrap: 80,
			}
			err = json.Unmarshal(b, &options.Styles)
			if err != nil {
				t.Fatal(err)
			}
			if bn == "493" {
				tableWrap := false
				options.TableWrap = &tableWrap
			}

			md := goldmark.New(
				goldmark.WithExtensions(
					extension.GFM,
					extension.DefinitionList,
					emoji.Emoji,
				),
				goldmark.WithParserOptions(
					parser.WithAutoHeadingID(),
				),
			)

			ar := NewRenderer(options)
			md.SetRenderer(
				renderer.NewRenderer(
					renderer.WithNodeRenderers(util.Prioritized(ar, 1000))))

			var buf bytes.Buffer
			if err := md.Convert(in, &buf); err != nil {
				t.Error(err)
			}

			golden.RequireEqual(t, buf.Bytes())
		})
	}
}
