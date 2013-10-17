/* See LICENSE file for copyright and license details. */

package obo

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type OboTermEntry struct {
	Id       string
	Altids   []string
	Name     string
	Def      string
	Xrefs    []string
	Synonyms []string
	IsA      []string
	Obsolete bool
}

// Use this function to filter the OboTermEntry pointer slice
func Filter(s []*OboTermEntry, fn func(*OboTermEntry) bool) []*OboTermEntry {
	var p []*OboTermEntry
	for _, i := range s {
		if fn(i) {
			p = append(p, i)
		}
	}
	return p
}

// This function prints a simple representation of the parsed Obo
// data. It is used for debugging purposes and may be removed at a later
// stage.
func Dump(oboent []*OboTermEntry, parentchildrenmap map[string][]*OboTermEntry) {
	var potentialroots []string
	for _, entry := range oboent {
		if len(entry.IsA) == 0 {
			potentialroots = append(potentialroots, entry.Id)
		}

		fmt.Printf("%s\n\tPT %s\n", entry.Id, entry.Name)
		if len(entry.Synonyms) > 0 {
			fmt.Print("\tSYN ")
			fmt.Print(strings.Join(entry.Synonyms, "\n\tSYN "))
			fmt.Print("\n")
		}
		if children, ok := parentchildrenmap[entry.Id]; ok {
			fmt.Print("\tNT ")
			for _, child := range children {
				fmt.Print(child.Id, "\n\tNT ")
			}
			fmt.Print("\n")
		}
		fmt.Print("\n")
	}
	fmt.Fprintf(os.Stderr, "Number of entries in the list: %d\n", len(oboent))
	fmt.Fprintf(os.Stderr, "Number of entries with children: %d\n", len(parentchildrenmap))
	fmt.Fprintf(os.Stderr, "Number of orphan nodes: %d\n", len(potentialroots))

	fmt.Print("root\n\tPT YourOntologyNameHere\n")
	for _, potroot := range potentialroots {
		fmt.Printf("\tNT %s\n", potroot)
	}
	fmt.Print("\n")
}

func parseObo(oboinput bufio.Reader, obochan chan *OboTermEntry, parentchildren map[string][]*OboTermEntry) {
	lineno := 0
	var entry *OboTermEntry
	var termsstarted bool
	defer close(obochan)

	for {
		line, err := oboinput.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("Error while reading obo file at line nr. %d: %v\n", lineno, err)
			os.Exit(1)
		}
		lineno++
		line = line[:len(line)-1] // chop \n
		if lineno%1000000 == 0 {
			fmt.Fprintf(os.Stderr, "Chopped line number: %d\n", lineno)
		}

		if line == "[Term]" {
			termsstarted = true
			if entry != nil {
				obochan <- entry
			}

			entry = new(OboTermEntry)
			continue
		} else if line == "\n" {
			continue
		} else if line == "[Typedef]" {
			continue
		} else if line == "" {
			continue
		} else if line[0] == '!' {
			continue
		}

		if termsstarted {
			splitline := strings.SplitN(line, ":", 2)
			trimmedvalue := strings.Trim(splitline[1], " ")
			field := strings.Trim(splitline[0], " ")
			switch field {
			case "id":
				entry.Id = trimmedvalue
			case "name":
				entry.Name = trimmedvalue
			case "def":
				entry.Def = trimmedvalue
			case "alt_id":
				entry.Altids = append(entry.Altids, trimmedvalue)
			case "xref":
				entry.Xrefs = append(entry.Xrefs, trimmedvalue)
			case "synonym":
				syn := strings.SplitN(trimmedvalue, "\" ", 2)
				r := strings.NewReplacer("\"", "")
				entry.Synonyms = append(entry.Synonyms, r.Replace(syn[0]))
			case "is_a":
				isa := strings.SplitN(trimmedvalue, "!", 2)
				trimmedisa := strings.Trim(isa[0], " ")
				entry.IsA = append(entry.IsA, trimmedisa)
				if parentchildren != nil {
					parentchildren[trimmedisa] = append(parentchildren[trimmedisa], entry)
				}
			case "is_obsolete":
				entry.Obsolete = true
			}
		}
	}
	obochan <- entry
}

// Parses a .obo file given as a bufio.Reader into a slice of
// OboTermEntry's. Hierarchical information is saved in a map and
// returned together with the slice.
func ParseToSlice(oboinput bufio.Reader, parentchildren map[string][]*OboTermEntry, obolist []*OboTermEntry) ([]*OboTermEntry, map[string][]*OboTermEntry) {
	var ent *OboTermEntry
	obochan := make(chan *OboTermEntry, 100)

	go parseObo(oboinput, obochan, parentchildren)

	for ent = range obochan {
		obolist = append(obolist, ent)
	}
	return obolist, parentchildren
}

// This function returns a channel on which pointers to the parsed
// OboTermEntry structs will be sent. Please note that this function
// does not return the hierarchy map. If you want to parse the .obo
// file asynchronously while still having access to the hierarchical
// information you will have to build the structure containing the
// hierarchical information yourself.
func ParseToChannel(oboinput bufio.Reader, obochan chan *OboTermEntry) chan *OboTermEntry {

	go parseObo(oboinput, obochan, nil)

	return obochan
}
