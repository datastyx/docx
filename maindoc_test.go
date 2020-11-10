package docx

import (
	"fmt"
	"strings"
	"testing"

	"github.com/antchfx/xmlquery"
)

func TestDocx_removeElementsFromMainDocument(t *testing.T) {
	doc := LoadFile(testFile)

	tests := []struct {
		name         string
		d            *Docx
		elementXpath string
		wantErr      bool
	}{
		{"noMatchingELement", doc, "/something", true}, // fails because xpath dosn't target existing element
		{"notAnElementNode", doc, fmt.Sprintf("//@*[local-name()='id' and namespace-uri()='%s']", iso29500RelationshipsNS), true}, // fails because matches an attribute node and note an element node, TODO Xpath seems unsupported
		{"hyperlinkMatch", doc, fmt.Sprintf("//*[local-name()='hyperlink' and namespace-uri()='%s']", wordprocessingmlNS), false}, // matches one hyperlink
		{"threeParagraphMatch", doc, fmt.Sprintf("//*[local-name()='p' and namespace-uri()='%s']", wordprocessingmlNS), false},    // matches three paragraphs
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.RemoveElementsFromMainDocument(tt.elementXpath); (err != nil) != tt.wantErr {
				t.Errorf("Docx.removeElementsFromMainDocument() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				tt.d.WriteToFile(testFileResult)
				docx := LoadFile(testFileResult)
				parsedDocumentContent, parseErr := xmlquery.Parse(strings.NewReader(docx.content))
				if parseErr != nil {
					t.Error(parseErr)
				}
				elems := xmlquery.Find(parsedDocumentContent, tt.elementXpath)

				if len(elems) > 0 {
					t.Errorf("Elements where expected to be removed for xpath : %s", tt.elementXpath)
				}
			}
		})
	}
}
