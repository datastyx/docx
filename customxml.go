package docx

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/antchfx/xmlquery"
)

const documentRelsPath = "word/_rels/document.xml.rels"
const iso29500RelationshipsNS = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
const documentRelsNS = "http://schemas.openxmlformats.org/package/2006/relationships"

var retrieveDocumentRelationshipsXPath = fmt.Sprintf("/*[local-name()='Relationships' and namespace-uri()='%s']/*[ local-name()='Relationship' and namespace-uri()='%s' ]", documentRelsNS, documentRelsNS)

// GetCustomXML returns a map with key values being a path to a customXML
// file and the map value the respective content as a string.
// Result is nil if no customXML parts were retrieved
func (d *Docx) GetCustomXML() map[string]string {
	return d.customXML
}

// GetCustomXMLByRootNs returns a map with key values being a path to a customXML
// Part file in the OPC and the map value the respective content as a string.
// The map will only contain customXML Parts which XML root namespace matches the given parameter one.
// Result is nil if no customXML parts were retrieved
func (d *Docx) GetCustomXMLByRootNs(rootNS string) (customXmlmap map[string]string, err error) {
	// recover in case of xpath panic
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from error : %s", r)
		}
	}()
	customXmlmap = make(map[string]string)
	for path := range d.customXML {
		content := d.customXML[path]
		// parse and retrive root NS
		documentRelsContent, parseErr := xmlquery.Parse(strings.NewReader(content))
		if parseErr != nil {
			return nil, parseErr
		}
		if documentRelsContent.FirstChild.Type == xmlquery.DeclarationNode && documentRelsContent.LastChild.NamespaceURI == rootNS {
			customXmlmap[path] = content
		}
	}
	if len(customXmlmap) == 0 {
		return nil, nil
	}
	return customXmlmap, nil
}

// NoCustomXmlPartError is returned when trying to remove a CustomXML Part that can't be found
type NoCustomXmlPartError struct {
	err string
}

func (e NoCustomXmlPartError) Error() string {
	return e.err
}

//ExistingRelationshipsError is returned when trying to remove a customXML Part that still has existing relationships
type ExistingRelationshipsError struct {
	err string
}

func (e ExistingRelationshipsError) Error() string {
	return e.err
}

// RemoveCustomXML tries to remove a customXml Part from the OPC container.
// The Part to be removed is designated by the path to 'itemX.xml' file under the 'customXml' directory in the OPC.
// e.g. to remove 'item1.xml' the given path should be 'customXml/item1.xml'.
// If the file is refered to from the document.xml the removal is canceled and the method returns an error.
// This method also removes the following references : the related 'customXml/itemPropsX.xml' file, the related 'customXml/_rels/itemX.xml.rels' and the relationship reference in the 'word/_rels/document.xml.rels'
// path argument shall be given as an absolute path where the root is that of the OPC.
// When an error is returned customXml Part is not removed. If the reason is that no customXml part was found for the given path then a  'docx.NoCustomXmlPartError' is returned.
// If the reason is that the designated customXml Part has existing relationships then a 'docx.ExistingRelationshipsError' is returned
func (d *Docx) RemoveCustomXML(path string) (err error) {
	// validate existance of a customXml Part of given path
	if _, ok := d.customXML[path]; !ok {
		return NoCustomXmlPartError{"no customXml Part is found for the given path"}
	}

	// validate there are no references to the customXml part from the document.xml
	// TODO update library to deal with docx files have a main document.xml and section documents.
	relsIdMap, relsElemMap, err := d.relationshipIdsToGivenOPCPart(path)
	if err != nil {
		return err
	}
	if len(relsIdMap) > 0 {
		// has relationships so document.xml must be parsed to check if they are actually referenced
		docHasRefsToRelationships, err := d.docReferencesRelsIds(relsIdMap)
		if err != nil {
			return err
		}
		// if rels exist //
		if docHasRefsToRelationships {
			return ExistingRelationshipsError{fmt.Sprintf("At least an existing relationships was found to the customXmlPart of path '%s' ", path)}
		}
		// else : backup "word/_rels/document.xml.rels" before modification for eventual restoring in case of failure further on
		defer revertOnError(d.links, &d.links, err)
		// remove relationships in "word/_rels/document.xml.rels"
		arrayOfElements := []string{}
		for _, elem := range relsElemMap {
			arrayOfElements = append(arrayOfElements, elem)
		}
		// relationships file needs to be parsed to have same represenstations of the relationship elements
		parsedRelationships, parseErr := xmlquery.Parse(strings.NewReader(d.links))
		if parseErr != nil {
			return parseErr
		}
		d.links = parsedRelationships.OutputXML(false)
		d.links, err = remove(d.links, arrayOfElements)
		if err != nil {
			return err
		}

	}

	// check if customXml Part has a relationship in 'customXml/_rels/itemX.xml.rels' : stack references for removal
	// if yes, backup 'customXml/_resl/itemX.xml.rels'  for eventual restoring in case of failure further on
	// remove references from 'customXml/_rels/itemX.xml.rels'
	// removed referenced itemPropsX.xml

	//remove customXml Part
	for index := range d.files {
		if d.files[index].Name == path {
			d.files[index] = d.files[len(d.files)-1]
			d.files = d.files[:len(d.files)-1]
			break
		}
	}

	return nil
}

// revertOnError is to be used as a defered call. When unpiled, if err is not nil backup replaces the target data.
func revertOnError(backup string, targetData *string, err error) {
	if err != nil {
		*targetData = backup
	}
}

func remove(inputString string, toBeRemoved []string) (emptiedString string, err error) {
	// create an array of replacer elems
	replacerArray := []string{}
	for _, relsElement := range toBeRemoved {

		replacerArray = append(replacerArray, relsElement, "") // creates an array as : {'a','','b','' ...} so each instance gets deleted instead of replaced
	}
	replacer := strings.NewReplacer(replacerArray...)

	emptiedString = replacer.Replace(inputString)
	return emptiedString, nil
}

// docReferencesRels for each given ID check if exist in document
func (d *Docx) docReferencesRelsIds(relIds []string) (bool, error) {
	if len(relIds) == 0 {
		return false, nil
	}
	parsedDoc, parseErr := xmlquery.Parse(strings.NewReader(d.content))
	if parseErr != nil {
		return false, parseErr
	}
	//@*[local-name()='val' and namespace-uri()='http://schemas.openxmlformats.org/wordprocessingml/2006/main'  and .='MOCK UNCLASSIFIED']
	var xpath string = "//@*["
	for _, val := range relIds {
		xpath += fmt.Sprintf("(local-name()='id' and namespace-uri()='%s' and .='%s')or", iso29500RelationshipsNS, val)
	}
	xpath = strings.TrimSuffix(xpath, "or")
	xpath += "]"
	// xpath on content to retrieve rels ids
	list := xmlquery.Find(parsedDoc, retrieveDocumentRelationshipsXPath)
	if len(list) > 0 {
		return true, nil
	}
	return false, nil

}

// relationshipIdsToGivenOPCPart returns true if no relationships are found in the "word/_rels/document.xml.rels" file towards a part which path ends with the given argument
func (d *Docx) relationshipIdsToGivenOPCPart(partPathEndinWith string) (relationshipIds []string, relationshipElementMap map[string]string, err error) {
	relationshipIDMap, relationshipElementMap, err := d.getCustomXMLRelationshipIDs()
	if err != nil {
		return nil, nil, err
	}

	for id, val := range relationshipIDMap {
		if strings.HasSuffix(val, partPathEndinWith) {
			relationshipIds = append(relationshipIds, id)
		} else {
			delete(relationshipElementMap, id)
		}
	}
	return relationshipIds, relationshipElementMap, nil

}

// getCustomXMLRelationshipID get the relationship Id in the 'word/_rels/document.xml.rels' document for a given customXml Part ('customXml/itemX.xml')
// key of the map is the relationship ID and the value is the path to the relationship target
func (d *Docx) getCustomXMLRelationshipIDs() (relationshipIDmap map[string]string, relationshipElementMap map[string]string, err error) {
	// recover in case of xpath panic
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint("Recovered from XPath error :", r))
		}
	}()
	documentRelsContent, parseErr := xmlquery.Parse(strings.NewReader(d.links))
	if parseErr != nil {
		return nil, nil, parseErr
	}
	// xpath on content to retrieve rels ids
	list := xmlquery.Find(documentRelsContent, retrieveDocumentRelationshipsXPath)
	if len(list) == 0 {
		log.Printf("No relationship elements found in %s", documentRelsPath)
		return
	}

	// make sure all result are attribute nodes and append to result
	relationshipIDmap = make(map[string]string)
	relationshipElementMap = make(map[string]string)
	for i := range list {
		if list[i].Type != xmlquery.ElementNode {
			return nil, nil, fmt.Errorf("While retrieving relationships only Element Nodes where expected but found : '%s'", getXMLNodeType(list[i].Type))
		}
		relationshipIDmap[list[i].SelectAttr("Id")] = list[i].SelectAttr("Target")
		relationshipElementMap[list[i].SelectAttr("Id")] = list[i].OutputXML(true)
	}

	return relationshipIDmap, relationshipElementMap, nil
}

// getXMLNodeType returns the type of XML node in a textual form
func getXMLNodeType(xmlqueryNodeType xmlquery.NodeType) (nodeType string) {
	switch xmlqueryNodeType {
	case xmlquery.AttributeNode:
		return "AttributeNode"
	case xmlquery.DocumentNode:
		return "DocumentNode"
	case xmlquery.DeclarationNode:
		return "DeclarationNode"
	case xmlquery.ElementNode:
		return "ElementNode"
	case xmlquery.TextNode:
		return "TextNode"
	case xmlquery.CharDataNode:
		return "CharDataNode"
	case xmlquery.CommentNode:
		return "CommentNode"
	default:
		return
	}

}
