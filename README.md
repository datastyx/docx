## Simple golang library to replace text in WordprocessingML Office Open XML Format(.docx) files

The following constitutes the bare minimum required to replace text in DOCX document.
``` go

import (
	"github.com/datastyx/docx"
)

func main() {
	// Read from docx file
	r, err := docx.ReadDocxFile("./template.docx")
	// Or read from memory
	// r, err := docx.ReadDocxFromMemory(data io.ReaderAt, size int64)
	if err != nil {
		panic(err)
	}
	docx1 := r.Editable()
	// Replace like https://golang.org/pkg/strings/#Replace
	docx1.Replace("old_1_1", "new_1_1", -1)
	docx1.Replace("old_1_2", "new_1_2", -1)
	docx1.ReplaceLink("http://example.com/", "https://github.com/nguyenthenguyen/docx")
	docx1.ReplaceHeader("out with the old", "in with the new")
	docx1.ReplaceFooter("Change This Footer", "new footer")
	docx1.WriteToFile("./new_result_1.docx")

	docx2 := r.Editable()
	docx2.Replace("old_2_1", "new_2_1", -1)
	docx2.Replace("old_2_2", "new_2_2", -1)
	docx2.WriteToFile("./new_result_2.docx")

	// Or write to ioWriter
	// docx2.Write(ioWriter io.Writer)

	r.Close()
}

```

This fork's main goal is to add two functionalities to nguyenthenguyen/docx library :
1. adds functionality to remove WordprocessingML content based on an XPath referencing the nodes to be removed, and;
``` go
// parse OPC
rodoc, err := docx.ReadDocxFile("./template.docx")
if err != nil {
	panic(err)
}
docx := rodoc.Editable()
err = docx.RemoveElementsFromMainDocument(someXPath)
if err != nil {
	panic(err)
}
```

2. Remove CustomXML Parts based on a path referencing it, the paths root shall be the root of the OPC container e.g. 'customXml/item1.xml'.
``` go
// parse OPC
rodoc, err := docx.ReadDocxFile("./template.docx")
if err != nil {
	panic(err)
}
docx := rodoc.Editable()
err = docx.RemoveCustomXML("customXml/item1.xml")
if err != nil {
	panic(err)
}
```
