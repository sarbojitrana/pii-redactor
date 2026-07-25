package docx

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
)

type TextNode struct {
	doc     *Document
	indices []int
}

func (n *TextNode) Text() string {
	var buf bytes.Buffer
	for _, i := range n.indices {
		if cd, ok := n.doc.tokens[i].(xml.CharData); ok {
			buf.Write(cd)
		}
	}
	return buf.String()
}

func (n *TextNode) SetText(s string) {
	if len(n.indices) == 0 {
		return
	}
	n.doc.tokens[n.indices[0]] = xml.CharData([]byte(s))
	for _, i := range n.indices[1:] {
		n.doc.tokens[i] = xml.CharData([]byte{})
	}
}

type Paragraph struct {
	TextNodes  []*TextNode
	breakAfter map[int]bool
}

func (p Paragraph) Text() string {
	var buf bytes.Buffer
	for i, n := range p.TextNodes {
		if i > 0 {
			if p.breakAfter[i-1] {
				buf.WriteByte('\n')
			} else {
				buf.WriteByte(' ')
			}
		}
		buf.WriteString(n.Text())
	}
	return buf.String()
}

type Document struct {
	tokens     []xml.Token
	paragraphs []Paragraph
}

func ParseDocumentXML(data []byte) (*Document, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	doc := &Document{}

	var paraStack []*Paragraph
	var textDepth int
	var curTextIndices []int

	for {
		tok, err := decoder.RawToken()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("decoding document.xml: %w", err)
		}

		tok = xml.CopyToken(tok)
		doc.tokens = append(doc.tokens, tok)
		idx := len(doc.tokens) - 1

		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "p":
				paraStack = append(paraStack, &Paragraph{})
			case "t":
				textDepth++
				curTextIndices = nil
			case "br", "cr":
				if len(paraStack) > 0 {
					cur := paraStack[len(paraStack)-1]
					if len(cur.TextNodes) > 0 {
						if cur.breakAfter == nil {
							cur.breakAfter = make(map[int]bool)
						}
						cur.breakAfter[len(cur.TextNodes)-1] = true
					}
				}
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "p":
				if len(paraStack) > 0 {
					last := paraStack[len(paraStack)-1]
					doc.paragraphs = append(doc.paragraphs, *last)
					paraStack = paraStack[:len(paraStack)-1]
				}
			case "t":
				if textDepth > 0 {
					textDepth--
					if len(curTextIndices) > 0 && len(paraStack) > 0 {
						node := &TextNode{doc: doc, indices: curTextIndices}
						cur := paraStack[len(paraStack)-1]
						cur.TextNodes = append(cur.TextNodes, node)
					}
					curTextIndices = nil
				}
			}
		case xml.CharData:
			if textDepth > 0 {
				curTextIndices = append(curTextIndices, idx)
			}
		}
	}

	return doc, nil
}

func (d *Document) Paragraphs() []Paragraph {
	return d.paragraphs
}

func (d *Document) Serialize() ([]byte, error) {
	var buf bytes.Buffer
	tokens := d.tokens
	depth := 0

	for i := 0; i < len(tokens); i++ {
		switch t := tokens[i].(type) {
		case xml.ProcInst:
			buf.WriteString("<?")
			buf.WriteString(t.Target)
			if len(t.Inst) > 0 {
				buf.WriteByte(' ')
				buf.Write(t.Inst)
			}
			buf.WriteString("?>")
		case xml.Directive:
			buf.WriteString("<!")
			buf.Write(t)
			buf.WriteByte('>')
		case xml.Comment:
			buf.WriteString("<!--")
			buf.Write(t)
			buf.WriteString("-->")
		case xml.CharData:
			if depth > 0 {
				if err := xml.EscapeText(&buf, t); err != nil {
					return nil, err
				}
			} else {
				buf.Write(t)
			}
		case xml.StartElement:
			selfClosing := false
			if i+1 < len(tokens) {
				if end, ok := tokens[i+1].(xml.EndElement); ok && end.Name == t.Name {
					selfClosing = true
				}
			}
			if err := writeElementOpen(&buf, t, selfClosing); err != nil {
				return nil, err
			}
			if selfClosing {
				i++
			} else {
				depth++
			}
		case xml.EndElement:
			writeElementClose(&buf, t)
			depth--
		}
	}

	return buf.Bytes(), nil
}

func writeQName(buf *bytes.Buffer, name xml.Name) {
	if name.Space != "" {
		buf.WriteString(name.Space)
		buf.WriteByte(':')
	}
	buf.WriteString(name.Local)
}

func writeElementOpen(buf *bytes.Buffer, t xml.StartElement, selfClosing bool) error {
	buf.WriteByte('<')
	writeQName(buf, t.Name)
	for _, attr := range t.Attr {
		buf.WriteByte(' ')
		writeQName(buf, attr.Name)
		buf.WriteString(`="`)
		if err := xml.EscapeText(buf, []byte(attr.Value)); err != nil {
			return err
		}
		buf.WriteByte('"')
	}
	if selfClosing {
		buf.WriteString("/>")
	} else {
		buf.WriteByte('>')
	}
	return nil
}

func writeElementClose(buf *bytes.Buffer, t xml.EndElement) {
	buf.WriteString("</")
	writeQName(buf, t.Name)
	buf.WriteByte('>')
}