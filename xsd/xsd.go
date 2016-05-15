package xsd

import (
	"encoding/xml"
)

const (
	String  string = "xsd:string"
	Int32          = "xsd:int"
	Int64          = "xsd:long"
	Float32        = "xsd:float"
	Float64        = "Xsd:double"
	Bool           = "xsd:boolean"
)

type Schema struct {
	XMLName            xml.Name      `xml:"http://www.w3.org/2001/XMLSchema xsd:schema"`
	TNS                string        `xml:"xmlns tns,attr,omitempty"`
	XS                 string        `xml:"xmlns xs,attr,omitempty"`
	TargetNamespace    string        `xml:"targetNamespace,attr,omitempty"`
	ElementFormDefault string        `xml:"elementFormDefault,attr,omitempty"`
	Version            string        `xml:"version,attr,omitempty"`
	Elements           []Element     `xml:"http://www.w3.org/2001/XMLSchema element"`
	ComplexTypes       []ComplexType `xml:"http://www.w3.org/2001/XMLSchema complexType"`
	Import             []Import      `xml:"xsd:import"`
}

type Element struct {
	XMLName      xml.Name     `xml:"http://www.w3.org/2001/XMLSchema element"`
	Type         string       `xml:"type,attr,omitempty"`
	Nillable     bool         `xml:"nillable,attr"`
	MinOccurs    int          `xml:"minOccurs,attr"`
	MaxOccurs    int          `xml:"maxOccurs,attr,omitempty"`
	Form         string       `xml:"form,attr,omitempty"`
	Name         string       `xml:"name,attr"`
	ComplexTypes *ComplexType `xml:"http://www.w3.org/2001/XMLSchema complexType"`
}

type ComplexType struct {
	XMLName  xml.Name        `xml:"http://www.w3.org/2001/XMLSchema complexType"`
	Name     string          `xml:"name,attr,omitempty"`
	Abstract bool            `xml:"abstract,attr"`
	Sequence []Element       `xml:"sequence>element"`
	Content  *ComplexContent `xml:"http://www.w3.org/2001/XMLSchema complexContent"`
}

type ComplexContent struct {
	XMLName   xml.Name  `xml:"http://www.w3.org/2001/XMLSchema complexContent"`
	Extension Extension `xml:"http://www.w3.org/2001/XMLSchema extension"`
}

type Extension struct {
	XMLName  xml.Name  `xml:"http://www.w3.org/2001/XMLSchema extension"`
	Base     string    `xml:"base,attr"`
	Sequence []Element `xml:"sequence>element"`
}

type Import struct {
	XMLName        xml.Name `xml:"xsd:import"`
	SchemaLocation string   `xml:"schemaLocation,attr,omitempty"`
	Namespace      string   `xml:"namespace,attr"`
}
