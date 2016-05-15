package soap

import (
	"encoding/xml"
)

type Envelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	XSI     string   `xml:"xmlns:xsi,attr"`
	XSD     string   `xml:"xmlns:xsd,attr"`
	Soap    string   `xml:"xmlns:soap,attr"`
	Body    Body
}

func NewEnvelope(content []byte) *Envelope {
	se := &Envelope{}
	se.XSI = "http://www.w3.org/2001/XMLSchema-instance"
	se.XSD = "http://www.w3.org/2001/XMLSchema"
	se.Soap = "http://schemas.xmlsoap.org/soap/envelope/"
	se.Body.Content = content
	return se
}

type Body struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
	Content []byte   `xml:",innerxml"`
}

type Fault struct {
	XMLName     xml.Name `xml:"Fault"`
	FaultCode   string   `xml:"faultcode"`
	FaultString string   `xml:"faultstring"`
	Detail      string   `xml:"detail"`
}
