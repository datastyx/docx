package docx

import (
	"errors"
	"fmt"
	"strings"

	"github.com/antchfx/xmlquery"
)

const wordprocessingmlNS = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"

type NoElementsFoundAtXPath struct {
	err string
}

func (e NoElementsFoundAtXPath) Error() string {
	return e.err
}

// removeElementFromMainDocument finds an element defined by the 'elementXpath' ( if no corresponding element Node found returns a 'NoElementsFoundAtXPath' error).
func (d *Docx) RemoveElementsFromMainDocument(elementXpath string) (err error) {

	parsedDocumentContent, parseErr := xmlquery.Parse(strings.NewReader(d.content))
	if parseErr != nil {
		return parseErr
	}
	// xpath on content to retrieve rels ids
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint("Recovered from XPath error :", r))
		}
	}()
	elems := xmlquery.Find(parsedDocumentContent, elementXpath)

	if len(elems) == 0 {
		return fmt.Errorf("No elements found at %s", elementXpath)
	}

	// make sure all result are attribute nodes and append to removal array
	removalArray := []string{}
	for i := range elems {
		if elems[i].Type != xmlquery.ElementNode {
			return NoCustomXmlPartError{fmt.Sprintf("Expected Element Nodes for given Xpath '%s' but found '%s' ", elementXpath, getXMLNodeType(elems[i].Type))}
		}
		removalArray = append(removalArray, elems[i].OutputXML(true))
	}

	d.content, err = remove(parsedDocumentContent.OutputXML(false), removalArray)
	if err != nil {
		return err
	}

	return nil
}
