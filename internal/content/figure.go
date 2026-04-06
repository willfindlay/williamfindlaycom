package content

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// Figure is a block node representing an HTML <figure> element.
type Figure struct {
	ast.BaseBlock
}

// KindFigure is the NodeKind for Figure.
var KindFigure = ast.NewNodeKind("Figure")

func (n *Figure) Kind() ast.NodeKind { return KindFigure }

func (n *Figure) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// FigCaption is a block node representing an HTML <figcaption> element.
type FigCaption struct {
	ast.BaseBlock
}

// KindFigCaption is the NodeKind for FigCaption.
var KindFigCaption = ast.NewNodeKind("FigCaption")

func (n *FigCaption) Kind() ast.NodeKind { return KindFigCaption }

func (n *FigCaption) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// figureTransformer converts standalone image paragraphs into <figure> elements.
// A paragraph containing exactly one image (and nothing else) is replaced with
// a Figure node. If the image has a title, the title becomes a FigCaption child
// and is cleared from the image so it doesn't also render as an HTML title attribute.
type figureTransformer struct{}

func (t *figureTransformer) Transform(node *ast.Document, _ text.Reader, _ parser.Context) {
	type replacement struct {
		paragraph *ast.Paragraph
		image     *ast.Image
	}

	var replacements []replacement

	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		p, ok := n.(*ast.Paragraph)
		if !ok {
			return ast.WalkContinue, nil
		}
		// Paragraph must have exactly one child, and it must be an image.
		if p.ChildCount() != 1 {
			return ast.WalkContinue, nil
		}
		img, ok := p.FirstChild().(*ast.Image)
		if !ok {
			return ast.WalkContinue, nil
		}
		replacements = append(replacements, replacement{paragraph: p, image: img})
		return ast.WalkSkipChildren, nil
	})

	for _, r := range replacements {
		fig := &Figure{}
		r.paragraph.Parent().ReplaceChild(r.paragraph.Parent(), r.paragraph, fig)

		// Move the image into the figure.
		fig.AppendChild(fig, r.image)

		// If the image has a title, create a figcaption from it.
		if len(r.image.Title) > 0 {
			caption := &FigCaption{}
			captionText := ast.NewString(r.image.Title)
			caption.AppendChild(caption, captionText)
			fig.AppendChild(fig, caption)
			r.image.Title = nil
		}
	}
}

// figureRenderer renders Figure and FigCaption nodes to HTML.
type figureRenderer struct{}

func (r *figureRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindFigure, r.renderFigure)
	reg.Register(KindFigCaption, r.renderFigCaption)
}

func (r *figureRenderer) renderFigure(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("<figure>\n")
	} else {
		_, _ = w.WriteString("</figure>\n")
	}
	return ast.WalkContinue, nil
}

func (r *figureRenderer) renderFigCaption(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("<figcaption>")
	} else {
		_, _ = w.WriteString("</figcaption>\n")
	}
	return ast.WalkContinue, nil
}

// figureExtender registers the figure transformer and renderer with goldmark.
type figureExtender struct{}

func (e *figureExtender) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(&figureTransformer{}, 500),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(&figureRenderer{}, 500),
		),
	)
}
