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
	enc := xml.NewEncoder(&buf)
	for _, tok := range d.tokens {
		if err := enc.EncodeToken(tok); err != nil {
			return nil, err
		}
	}
	if err := enc.Flush(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
