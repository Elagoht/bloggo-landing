package prerender

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"bloggo-landing/internal/fetcher"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type Renderer struct {
	fetcher   *fetcher.Fetcher
	templates *template.Template
	outputDir string
}

func New(f *fetcher.Fetcher, templatesDir, outputDir string) (*Renderer, error) {
	// Parse templates from all subdirectories
	tmpl := template.New("")

	patterns := []string{
		filepath.Join(templatesDir, "*.html"),
		filepath.Join(templatesDir, "layouts", "*.html"),
		filepath.Join(templatesDir, "partials", "*.html"),
	}

	for _, pattern := range patterns {
		_, err := tmpl.ParseGlob(pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to parse templates from %s: %w", pattern, err)
		}
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	    return &Renderer{
        fetcher:   f,
        templates: tmpl,
        outputDir: outputDir,
    }, nil
}

func (r *Renderer) RenderAll() error {
    if err := r.RenderHome(); err != nil {
        return fmt.Errorf("failed to render home: %w", err)
    }

    if err := r.RenderDocuments(); err != nil {
        return fmt.Errorf("failed to render documents: %w", err)
    }

    documents, err := r.fetcher.GetDocuments()
    if err != nil {
        return fmt.Errorf("failed to fetch documents: %w", err)
    }

    for _, doc := range documents {
        if err := r.RenderPostBySlug(doc.Slug); err != nil {
            return fmt.Errorf("failed to render document %s: %w", doc.Slug, err)
        }
    }

    return nil
}

func (r *Renderer) RenderHome() error {
    var buf bytes.Buffer
    if err := r.templates.ExecuteTemplate(&buf, "home.html", nil); err != nil {
        return err
    }

    outputPath := filepath.Join(r.outputDir, "index.html")
    return os.WriteFile(outputPath, buf.Bytes(), 0644)
}

func (r *Renderer) RenderBlogIndex() error {
    posts, err := r.fetcher.GetAllPosts()
    if err != nil {
        return err
    }

    data := map[string]any{
        "Posts": posts,
    }

    var buf bytes.Buffer
    if err := r.templates.ExecuteTemplate(&buf, "blog-index.html", data); err != nil {
        return err
    }

    blogDir := filepath.Join(r.outputDir, "blog")
    if err := os.MkdirAll(blogDir, 0755); err != nil {
        return err
    }

    outputPath := filepath.Join(blogDir, "index.html")
    return os.WriteFile(outputPath, buf.Bytes(), 0644)
}

// RenderDocuments renders the documents landing page at "/documents" using the first document content
func (r *Renderer) RenderDocuments() error {
    documents, err := r.fetcher.GetDocuments()
    if err != nil {
        return err
    }

    data := map[string]any{
        "Documents": documents,
    }

    if len(documents) > 0 {
        first := documents[0]
        if post, err := r.fetcher.GetPostBySlug(first.Slug); err == nil && post != nil {
            data["Post"] = post
            data["HTMLContent"] = r.markdownToHTML(post.Content)
        }
    }

    var buf bytes.Buffer
    if err := r.templates.ExecuteTemplate(&buf, "post.html", data); err != nil {
        return err
    }

    outputPath := filepath.Join(r.outputDir, "documents.html")
    return os.WriteFile(outputPath, buf.Bytes(), 0644)
}

func (r *Renderer) markdownToHTML(md string) template.HTML {
    extensions :=
        parser.Tables |
            parser.FencedCode |
            parser.Autolink |
            parser.Strikethrough |
            parser.LaxHTMLBlocks |
            parser.NonBlockingSpace |
            parser.Footnotes |
            parser.HeadingIDs |
            parser.Titleblock |
            parser.AutoHeadingIDs |
            parser.BackslashLineBreak |
            parser.DefinitionLists |
            parser.MathJax |
            parser.OrderedListStart |
            parser.Attributes |
            parser.SuperSubscript |
            parser.EmptyLinesBreakList

    p := parser.NewWithExtensions(extensions)

    htmlFlags := html.CommonFlags | html.HrefTargetBlank
    opts := html.RendererOptions{Flags: htmlFlags}
    renderer := html.NewRenderer(opts)

    htmlContent := markdown.ToHTML([]byte(md), p, renderer)
    return template.HTML(htmlContent)
}

func (r *Renderer) RenderDocument(doc *fetcher.Post) error {
	// Convert markdown content to HTML
	htmlContent := r.markdownToHTML(doc.Content)
	// Get all documents for sidebar
	documents, err := r.fetcher.GetDocuments()
	if err != nil {
		return fmt.Errorf("failed to fetch documents for sidebar: %w", err)
	}

	data := map[string]any{
		"Post":        doc,
		"HTMLContent": htmlContent,
		"Documents":   documents,
	}

	var buf bytes.Buffer
	if err := r.templates.ExecuteTemplate(&buf, "post.html", data); err != nil {
		return err
	}

	documentsDir := filepath.Join(r.outputDir, "documents")
	if err := os.MkdirAll(documentsDir, 0755); err != nil {
		return err
	}

	outputPath := filepath.Join(documentsDir, doc.Slug+".html")
	return os.WriteFile(outputPath, buf.Bytes(), 0644)
}

func (r *Renderer) RenderPostBySlug(slug string) error {
	post, err := r.fetcher.GetPostBySlug(slug)
	if err != nil {
		return err
	}

	// All posts are documents
	return r.RenderDocument(post)
}

func (r *Renderer) DeletePost(slug string) error {
	documentsPath := filepath.Join(r.outputDir, "documents", slug+".html")
	if err := os.Remove(documentsPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
