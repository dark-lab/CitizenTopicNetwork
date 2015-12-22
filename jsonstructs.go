package main

import "encoding/xml"

type Graph struct {
	XMLName         xml.Name `xml:"graph"`
	Nodes           []Node   `json:"nodes" xml:"nodes>node"`
	Links           []Link   `json:"links" xml:"edges>edge"`
	Mode            string   `json:"-" xml:"mode,attr"`
	Defaultedgetype string   `json:"-" xml:"defaultedgetype,attr"`
}

type Node struct {
	XMLName xml.Name `xml:"node"`
	Id      int      `json:"id" xml:"id,attr"`
	Name    string   `json:"name" xml:"label,attr"`
	Group   int      `json:"group" xml:"-"`
}

type Link struct {
	XMLName xml.Name `xml:"edge"`
	Id      int      `json:"id" xml:"id,attr"`
	Source  int      `json:"source" xml:"source,attr"`
	Target  int      `json:"target" xml:"target,attr"`
	Value   float32  `json:"value" xml:"weight,attr"`
}
