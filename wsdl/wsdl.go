package wsdl

import (
	"encoding/xml"

	"github.com/afocus/gosoap/xsd"
)

type Definitions struct {
	XMLName  xml.Name  `xml:"http://schemas.xmlsoap.org/wsdl/ definitions"`
	SoapEnv  string    `xml:"xmlns:SOAP-ENV,attr"`
	TargetNs string    `xml:"targetNamespace,attr"`
	Tns      string    `xml:"xmlns:tns,attr"`
	Soap     string    `xml:"xmlns:soap,attr"`
	Xsd      string    `xml:"xmlns:xsd,attr"`
	Xsi      string    `xml:"xmlns:xsi,attr"`
	Wsdl     string    `xml:"xmlns:wsdl,attr"`
	Types    Type      `xml:"types"`
	Message  []Message `xml:"message"`
	PortType PortType  `xml:"portType"`
	Binding  Binding   `xml:"binding"`
	Service  Service   `xml:"service"`
}

type Type struct {
	Schemas []xsd.Schema `xml:"schema"`
}

type Message struct {
	Name string `xml:"name,attr"`
	Part []Part `xml:"part"`
}

type Part struct {
	Name    string `xml:"name,attr"`
	Type    string `xml:"type,attr,omitempty"`
	Element string `xml:"element,attr,omitempty"`
}

type PortType struct {
	Name       string              `xml:"name,attr"`
	Operations []PortTypeOperation `xml:"operation"`
}

type PortTypeOperation struct {
	Name   string                   `xml:"name,attr"`
	Input  PortTypeOperationMessage `xml:"input"`
	Output PortTypeOperationMessage `xml:"output"`
	//Fault  PortTypeOperationMessage `xml:"fault,omitempty"`
}

type PortTypeOperationMessage struct {
	Name    string `xml:"name,attr,omitempty"`
	Message string `xml:"message,attr"`
}

type Binding struct {
	Name        string             `xml:"name,attr"`
	Type        string             `xml:"type,attr"`
	SoapBinding SoapBinding        `xml:"soap:binding"`
	Operations  []BindingOperation `xml:"operation"`
}

type SoapBinding struct {
	XMLName   xml.Name `xml:"soap:binding"`
	Transport string   `xml:"transport,attr"`
	Style     string   `xml:"style,attr"`
}

type BindingOperation struct {
	Name          string        `xml:"name,attr"`
	SoapOperation SoapOperation `xml:"soap:operation"`
	Input         SoapBodyIO    `xml:"input"`
	Output        SoapBodyIO    `xml:"output"`
	//Fault         SoapBody      `xml:"fault>fault,omitempty"`
}

type SoapOperation struct {
	SoapAction string `xml:"soapAction,attr"`
	Style      string `xml:"style,attr,omitempty"`
}

type SoapBodyIO struct {
	SoapBody SoapBody `xml:"soap:body"`
}

type SoapBody struct {
	Name string `xml:"name,attr,omitempty"`
	Use  string `xml:"use,attr"`
}

type Service struct {
	Name string      `xml:"name,attr"`
	Port ServicePort `xml:"port"`
}

type ServicePort struct {
	XMLName xml.Name       `xml:"port"`
	Name    string         `xml:"name,attr"`
	Binding string         `xml:"binding,attr"`
	Address ServiceAddress `xml:"soap:address"`
}

type ServiceAddress struct {
	XMLName  xml.Name `xml:"soap:address"`
	Location string   `xml:"location,attr"`
}
