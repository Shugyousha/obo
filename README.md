# Introduction to the obo package

This is a WIP for a go library for parsing [.obo files](http://www.geneontology.org/GO.format.obo-1_2.shtml). The library currently parses .obo files and returns either a channel to, or a slice of the [Term] stanzas contained in the files.

The parsing of other stanza types is still under construction as are several of the key: value-pairs associated with the [Term] stanza. Please see the TODO section further below.



# Examples

Reading a .obo file into a slice of OboTermEntries


```Go
var obolist []*OboTermEntry
var parentchildren map[string][]*OboTermEntry

reader := bufio.NewReader(os.File("/path/to/file/file.obo"))
obolist, parentchildren = ParseToSlice(*reader, parentchildren, obolist)

```

To parse the OboTermEntries and having them fed to a channel:

```Go
obochan := make(chan *OboTermEntry)
var obolist []*OboTermEntry

reader := bufio.NewReader(os.File("/path/to/file/file.obo"))
obochan = ParseToChannel(*reader, obochan)

for ent := range obochan {
	obolist = append(obolist, ent)
}

```


# TODO

* Parsing of header data
* Writing of .obo files
* Parsing of [Typedef] stanzas
* Parsing of [Instance] stanzas
* Handling of several tags is still missing

