package main

import (
	"embed"
	"path"
	"strings"
)

// The backgrounds are scenes with no one in them, public domain (CC0) photographs the app ships (NOTICE.md), which a look may set the person in.
//
//go:embed backgrounds/*.jpg
var backgroundFiles embed.FS

// backgroundScheme marks an item's picture as one of the shipped backgrounds rather than a URL or a file the app keeps.
const backgroundScheme = "background:"

// The backgrounds' groups, the chips of their gallery.
const (
	groupIndoor = "indoor"
	groupCity   = "city"
	groupNature = "nature"
)

// backgroundGroups are the groups in the order the gallery shows them.
var backgroundGroups = []string{groupIndoor, groupCity, groupNature}

// backgroundCatalog is the shipped backgrounds, each file named after its group: indoor-brick-wall.jpg.
func backgroundCatalog() Catalog {
	entries, _ := backgroundFiles.ReadDir("backgrounds") // the directory is embedded, so it reads
	var c Catalog
	for _, g := range backgroundGroups {
		for _, e := range entries {
			name := strings.TrimSuffix(e.Name(), path.Ext(e.Name()))
			if !strings.HasPrefix(name, g+"-") {
				continue
			}
			uri := backgroundScheme + e.Name()
			c.Items = append(c.Items, Item{Kind: KindBackground, ID: name, Group: g, Thumb: uri, Full: uri})
		}
	}
	c.Groups = backgroundGroups
	return c
}

// backgroundData is the picture of a shipped background's URI, false for any other.
func backgroundData(uri string) ([]byte, bool) {
	name, ok := strings.CutPrefix(uri, backgroundScheme)
	if !ok {
		return nil, false
	}
	data, err := backgroundFiles.ReadFile("backgrounds/" + path.Base(name))
	return data, err == nil
}
