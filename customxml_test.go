package docx

import (
	"fmt"
	"testing"
)

const customXMLTestFile = "./demoFile.docx"

func TestRetrieveCustomXML(t *testing.T) {
	d := LoadFile(customXMLTestFile)
	customXMLmap := d.GetCustomXML()
	if _, ok := customXMLmap["customXml/item1.xml"]; ok {
		return
	}
	t.Error("Expected a customXML file of path : 'customXml/item1.xml' but it was not found")
}

// testReadWriteFileWithCustomXML parses file then writes it again then parses again to check if writing didn't delete the custom XML parts.
func TestReadWriteFileWithCustomXML(t *testing.T) {
	d := LoadFile(customXMLTestFile)

	d.WriteToFile(testFileResult)

	TestRetrieveCustomXML(t)
}

// try to retrieve customXml IDs from test file
func TestGetCustomXMLRelationshipIDs(t *testing.T) {
	d := LoadFile(customXMLTestFile)
	relsMap, _, err := d.getCustomXMLRelationshipIDs()
	if err != nil {
		t.Error(err)
		return
	}
	if val, exists := relsMap["rId1"]; exists {
		if val == "../customXml/item1.xml" {
			// expect relationship found
			return
		}
	}
	t.Error("Expected a customXML file of path : 'customXml/item1.xml' but it was not found")
	return
}

func TestInsureRelationshipsToOPCPart(t *testing.T) {
	d := LoadFile(customXMLTestFile)
	rels, _, err := d.relationshipIdsToGivenOPCPart("customXml/item1.xml")
	if err != nil {
		t.Error(err)
		return
	}
	if len(rels) < 0 {
		t.Error("A relationship was expected to be found")
		return
	}
	return
}

func TestInsureNoRelationshipsToOPCPart(t *testing.T) {
	d := LoadFile(customXMLTestFile)
	rels, _, err := d.relationshipIdsToGivenOPCPart("customXml/item99.xml")
	if err != nil {
		t.Error(err)
		return
	}
	if len(rels) > 0 {
		t.Error("No relationship was expected to be found")
		return
	}
	return
}

func TestRemoveCustomXML(t *testing.T) {
	// check befor modifications

	d := LoadFile(customXMLTestFile)
	customXMLmap := d.GetCustomXML()
	if _, ok := customXMLmap["customXml/item1.xml"]; !ok {
		t.Error("Expected a customXML file of path : 'customXml/item1.xml' but it was not found")
		return
	}
	rels, _, err := d.relationshipIdsToGivenOPCPart("customXml/item1.xml")
	if err != nil {
		t.Error(err)
		return
	}
	if len(rels) < 0 {
		t.Error("A relationship was expected to be found")
		return
	}
	// remove
	err = d.RemoveCustomXML("customXml/item1.xml")
	if err != nil {
		t.Errorf("Failed to remove CustomXml Part :\n%s\n", err.Error())
		return
	}
	d.WriteToFile(testFileResult)
	// check after modifications
	d = LoadFile(testFileResult)
	customXMLmap = d.GetCustomXML()
	if _, isPresent := customXMLmap["customXml/item1.xml"]; isPresent {
		t.Error("Expected a customXML file of path : 'customXml/item1.xml' but it was not found")
		return
	}
	rels, _, err = d.relationshipIdsToGivenOPCPart("customXml/item1.xml")
	if err != nil {
		t.Error(err)
		return
	}
	if len(rels) > 0 {
		t.Error("No relationship was expected to be found")
		return
	}
}

func TestRemove(t *testing.T) {
	text := "sometext abc abcdefg"
	fmt.Printf("text befor removal :\n%s\n", text)
	text, err := remove(text, []string{"abc", "de"})
	if err != nil {
		t.Errorf("error raised while removing :\n%s\n", err.Error())
		return
	}
	fmt.Printf("text after removal :\n%s\n", text)
	expectedResult := "sometext  fg"
	if text != expectedResult {
		t.Errorf("Expected : %s\nFound : %s\n", expectedResult, text)
		return
	}
}

func TestDocx_GetCustomXMLByRootNs(t *testing.T) {
	doc := LoadFile(customXMLTestFile)
	tests := []struct {
		name                string
		d                   *Docx
		rootNS              string
		wantCustomXmlmapLen int
		wantErr             bool
	}{
		{"bindingNs", doc, "urn:nato:stanag:4778:bindinginformation:1:0", 4, false},
		{"UnexistingNs", doc, "nothingThatExists", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCustomXmlmap, err := tt.d.GetCustomXMLByRootNs(tt.rootNS)
			if (err != nil) != tt.wantErr {
				t.Errorf("Docx.GetCustomXMLByRootNs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(gotCustomXmlmap) != tt.wantCustomXmlmapLen {
				t.Errorf(" length Docx.GetCustomXMLByRootNs() = %v, want %v", len(gotCustomXmlmap), tt.wantCustomXmlmapLen)
			}
		})
	}
}
