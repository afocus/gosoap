package gosoap

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"
	"strings"

	"github.com/afocus/gosoap/soap"
	"github.com/afocus/gosoap/wsdl"
)

// Methoder ...
type Methoder interface {
	Action() *SoapFault
}

// Server ...
type Server struct {
	location  string
	methods   map[string]reflect.Type
	wsdl      *wsdl.Definitions
	wsdlcache []byte
	ip        string
}

type SoapFault soap.Fault

func NewSoapFault(faultcode, faultstring, detail string) *SoapFault {
	f := new(SoapFault)
	f.FaultCode = faultcode
	f.FaultString = faultstring
	f.Detail = detail
	return f
}

func (s *Server) parseMessage(name string, t reflect.StructField) wsdl.Message {
	msg := wsdl.Message{Name: name}

	retype := t.Type
	partname, _ := getTagsInfo(t)

	k := retype.Kind()
	switch k {
	case reflect.Struct:
		for i := 0; i < retype.NumField(); i++ {
			name, _ := getTagsInfo(retype.Field(i))
			ik := retype.Field(i).Type.Kind()
			ts, erro := checkBaseTypeKind(ik)
			if erro != nil {
				ts = "tns:" + name + ik.String()
			}
			msg.Part = append(msg.Part, wsdl.Part{Name: name, Type: ts})
		}
	case reflect.Slice, reflect.Map, reflect.Array:
		msg.Part = append(msg.Part, wsdl.Part{Name: partname, Type: "tns:" + partname + k.String()})
	default:
		ts, erro := checkBaseTypeKind(k)
		if erro != nil {
			ts = "tns:" + name + k.String()
		}
		msg.Part = append(msg.Part, wsdl.Part{Name: partname, Type: ts})
	}
	return msg
}

func (s *Server) register(a Methoder) error {

	t := reflect.TypeOf(a)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return errors.New("method 不是struct")
	}
	// 方法名
	tname := t.Name()
	if _, ok := s.methods[tname]; ok {
		return errors.New("方法重复注册:" + tname)
	}
	// 解析方法参数
	in, has := t.FieldByName("In")
	if !has {
		return errors.New("缺少入参In")
	}
	// 解析返回值类类型
	out, has := t.FieldByName("Out")
	if !has {
		return errors.New("缺少出参Out")
	}
	// message
	inmsg := s.parseMessage(tname+"Request", in)
	outmsg := s.parseMessage(tname+"Respone", out)
	s.wsdl.Message = append(s.wsdl.Message, inmsg, outmsg)
	// portype
	op := wsdl.PortTypeOperation{
		Name:   tname,
		Input:  wsdl.PortTypeOperationMessage{Message: "tns:" + tname + "Request"},
		Output: wsdl.PortTypeOperationMessage{Message: "tns:" + tname + "Respone"},
	}
	s.wsdl.PortType.Operations = append(s.wsdl.PortType.Operations, op)
	// binding
	bindop := wsdl.BindingOperation{
		Name:          tname,
		SoapOperation: wsdl.SoapOperation{Style: "rpc", SoapAction: s.wsdl.Tns + "/" + tname},
		Input:         wsdl.SoapBodyIO{SoapBody: wsdl.SoapBody{Use: "encoded"}},
		Output:        wsdl.SoapBodyIO{SoapBody: wsdl.SoapBody{Use: "encoded"}},
	}
	s.wsdl.Binding.Operations = append(s.wsdl.Binding.Operations, bindop)

	reflectVal := reflect.ValueOf(a)
	mt := reflect.Indirect(reflectVal).Type()

	s.methods[tname] = mt
	return nil
}

func (s *Server) Register(method ...Methoder) error {
	for _, v := range method {
		if erro := s.register(v); erro != nil {
			return erro
		}
	}
	return nil
}

func (s *Server) httpHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		// 网址带参数wsdl则显示wsdl文件
		if strings.EqualFold("wsdl", r.URL.RawQuery) {
			w.Header().Set("Content-type", "text/xml")
			w.Write(s.wsdlcache)

		} else {
			// 其他情况返回一个提示信息
			w.Write([]byte("welcome"))
		}
		return
	}

	// post请求则处理接受的xml
	// 读取post的body信息
	b, erro := ioutil.ReadAll(r.Body)
	if erro != nil {
		serverSoapFault(w, NewSoapFault("Server", "读取body出错", erro.Error()))
		return
	}
	defer r.Body.Close()

	// 转化为Envelope对象
	env := soap.Envelope{}
	xml.Unmarshal(b, &env)
	// 解析请求的方法名字
	var startEle *xml.StartElement
	reader := bytes.NewReader(env.Body.Content)
	de := xml.NewDecoder(reader)
	for {
		t, erro := de.Token()
		if erro != nil {
			break
		}
		if x, ok := t.(xml.StartElement); ok {
			startEle = &x
			break
		}
	}
	if startEle == nil {
		serverSoapFault(w, NewSoapFault("Server", "接受到的data无效", ""))
		return
	}

	s.Request(w, de, startEle)

}

func serverSoapFault(w http.ResponseWriter, fault *SoapFault) {
	if w.Header().Get("Content-type") != "text/xml" {
		w.Header().Set("Content-type", "text/xml")
	}
	data, _ := xml.Marshal(fault)
	b, _ := xml.Marshal(soap.NewEnvelope(data))
	w.Write(b)
}

func (s *Server) Request(w http.ResponseWriter, de *xml.Decoder, startEle *xml.StartElement) {
	w.Header().Set("Content-type", "text/xml")

	mname := startEle.Name.Local
	t, has := s.methods[mname]
	if !has {
		serverSoapFault(w, NewSoapFault("Server", "没有这个方法:"+mname, ""))
		return
	}

	v := reflect.New(t)
	// 解析入参
	params := v.Elem().FieldByName("In").Addr().Interface()
	erro := de.DecodeElement(params, startEle)
	if erro != nil {
		serverSoapFault(w, NewSoapFault("Client", "接受参数错误", erro.Error()))
		return
	}
	// 调用action方法
	rets := v.MethodByName("Action").Call([]reflect.Value{})
	// 处理返回值
	fault := rets[0].Interface().(*SoapFault)
	if fault != nil {
		serverSoapFault(w, fault)
		return
	}

	name := v.Elem().Type().Name() + "Respone"
	returns := v.Elem().FieldByName("Out").Interface()

	buf := bytes.NewBuffer([]byte{})

	en := xml.NewEncoder(buf)
	soapbody := xml.StartElement{
		Name: xml.Name{Local: name},
		Attr: []xml.Attr{{
			Name:  xml.Name{Local: "xmlns"},
			Value: s.location,
		}},
	}
	erro = en.EncodeElement(returns, soapbody)
	if erro != nil {
		serverSoapFault(w, NewSoapFault("Server", "Respone错误", erro.Error()))
		return
	}

	b, _ := xml.Marshal(soap.NewEnvelope(buf.Bytes()))
	w.Write(b)

}

func (s *Server) Service(port string) error {
	s.wsdl.Service.Port.Address.Location = "http://" + s.ip + ":" + port + "/" + s.location
	b, erro := xml.Marshal(s.wsdl)
	if erro != nil {
		return erro
	}
	s.wsdlcache = []byte(xml.Header)
	s.wsdlcache = append(s.wsdlcache, b...)
	http.HandleFunc("/"+s.location, s.httpHandle)
	return http.ListenAndServe(":"+port, nil)

}

func get_internal() string {
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}
	return ""
}

func NewServer(name string) *Server {
	s := &Server{
		location: name,
		methods:  make(map[string]reflect.Type),
		ip:       get_internal(),
	}
	namespace := fmt.Sprintf("http://%s/%s", s.ip, s.location)
	s.wsdl = wsdl.NewDefinitions(namespace, name)
	return s
}
