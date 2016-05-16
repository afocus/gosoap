package gosoap

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"github.com/afocus/gosoap/soap"
	"github.com/afocus/gosoap/wsdl"
	"github.com/afocus/gosoap/xsd"
)

// Methoder 方法接口
// 所有注册的服务方法必须实现此接口
type Methoder interface {
	Action() *SoapFault
}

// BindHost 绑定的访问地址
// 不带http:// 以及端口
// 例如 www.baidu.com 、180.97.81.222、 webservice.demo.io
var BindHost = "127.0.0.1"

// Server 结构
type Server struct {
	// 服务名称
	name string
	// 映射的方法列表
	methods map[string]reflect.Type
	// wsdl对象
	wsdl *wsdl.Definitions
	// 缓存生成的wsdl文件
	wsdlcache []byte
	// wsdl文件的自主命名空间
	namespace string
}

// SoapFault
// 响应时返回的错误类型
type SoapFault soap.Fault

// NewSoapFault 创建一个soap错误
// 主要用在Methoder接口的Action返回值
func NewSoapFault(faultcode, faultstring, detail string) *SoapFault {
	f := new(SoapFault)
	f.FaultCode = faultcode
	f.FaultString = faultstring
	f.Detail = detail
	return f
}

// NewServer 创建一个服务
// 创建后如果只有一个服务可以使用Server.Service监听服务
func NewServer(name string) *Server {
	namespace := fmt.Sprintf("http://%s/%s", BindHost, name)
	s := &Server{
		name:      name,
		namespace: namespace,
		methods:   make(map[string]reflect.Type),
	}
	s.buildWsdl()
	return s
}

// Register 向服务中注册方法
// 可以同时注册多个服务
// 多服务的监听需要调用 MulitService
func (s *Server) Register(method ...Methoder) error {
	for _, v := range method {
		if erro := s.register(v); erro != nil {
			return erro
		}
	}
	return nil
}

// Service 单服务监听
func (s *Server) Service(port string) error {
	if erro := s.bind(port); erro != nil {
		return erro
	}
	return listen(port)
}

// MulitService 多服务监听
func MulitService(port string, server ...*Server) error {
	for _, v := range server {
		if erro := v.bind(port); erro != nil {
			return erro
		}
	}
	return listen(port)
}

/////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////

// 监听服务
// 这里可以设置http 比如超时时间 最大连接数等 后续再做
func listen(port string) error {
	return http.ListenAndServe(":"+port, nil)
}

// 解析参数并转化为对应的wsdl message
func (s *Server) parseMessage(name string, t reflect.Type, field string) error {
	msg := wsdl.Message{Name: name}
	if f, has := t.FieldByName(field); !has {
		return errors.New("method结构体缺少必要参数" + field)
	} else {
		retype := f.Type
		if retype.Kind() != reflect.Struct {
			return errors.New("method结构体 In,Out参数必须是结构体")
		}
		// 遍历结构体参数列表
		for i := 0; i < retype.NumField(); i++ {
			name, _ := getTagsInfo(retype.Field(i))
			ik := retype.Field(i).Type.Kind()
			ts, erro := checkBaseTypeKind(ik)
			// 如果非基本类型则转为自有命名空间的自定义类型
			if erro != nil {
				ts = "tns:" + name + ik.String()
			}
			msg.Part = append(msg.Part, wsdl.Part{Name: name, Type: ts})
		}
		s.wsdl.Message = append(s.wsdl.Message, msg)
		return nil
	}
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
	// message
	erro := s.parseMessage(tname+"Request", t, "In")
	if erro != nil {
		return erro
	}
	erro = s.parseMessage(tname+"Respone", t, "Out")
	if erro != nil {
		return erro
	}
	s.regWsdlBindPort(tname)
	// 方法类型映射到server的methods中
	// 接受处理时可以反射出In Out参数的具体值
	reflectVal := reflect.ValueOf(a)
	mt := reflect.Indirect(reflectVal).Type()
	s.methods[tname] = mt
	return nil
}

func (s *Server) regWsdlBindPort(name string) {
	// portype
	op := wsdl.PortTypeOperation{
		Name:   name,
		Input:  wsdl.PortTypeOperationMessage{Message: "tns:" + name + "Request"},
		Output: wsdl.PortTypeOperationMessage{Message: "tns:" + name + "Respone"},
	}
	s.wsdl.PortType.Operations = append(s.wsdl.PortType.Operations, op)
	// binding
	soapio := wsdl.SoapBodyIO{SoapBody: wsdl.SoapBody{Use: "encoded"}}
	bindop := wsdl.BindingOperation{
		Name: name, Input: soapio,
		Output: soapio,
		SoapOperation: wsdl.SoapOperation{
			Style: "rpc", SoapAction: s.wsdl.Tns + "/" + name,
		},
	}
	s.wsdl.Binding.Operations = append(s.wsdl.Binding.Operations, bindop)
}

func (s *Server) handleFunc(w http.ResponseWriter, r *http.Request) {
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
		serverSoapFault(
			w, NewSoapFault("Server", "读取body出错", erro.Error()),
		)
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
		serverSoapFault(
			w, NewSoapFault("Server", "接受到的data无效", ""),
		)
		return
	}
	s.request(w, de, startEle)
}

func serverSoapFault(w http.ResponseWriter, fault *SoapFault) {
	w.Header().Set("Content-type", "text/xml")
	data, _ := xml.Marshal(fault)
	b, _ := xml.Marshal(soap.NewEnvelope(data))
	w.Write(b)
}

func (s *Server) request(w http.ResponseWriter, de *xml.Decoder, startEle *xml.StartElement) {

	mname := startEle.Name.Local
	t, has := s.methods[mname]
	if !has {
		serverSoapFault(
			w, NewSoapFault("Server", "没有这个方法:"+mname, ""),
		)
		return
	}
	v := reflect.New(t)
	// 解析入参
	params := v.Elem().FieldByName("In").Addr().Interface()
	erro := de.DecodeElement(params, startEle)
	if erro != nil {
		serverSoapFault(
			w, NewSoapFault("Client", "接受参数错误", erro.Error()),
		)
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
	// 解析Out出参
	name := v.Elem().Type().Name() + "Respone"
	returns := v.Elem().FieldByName("Out").Interface()
	buf := bytes.NewBuffer([]byte{})
	en := xml.NewEncoder(buf)
	//
	soapbody := xml.StartElement{
		Name: xml.Name{Local: name},
		Attr: []xml.Attr{{
			Name:  xml.Name{Local: "xmlns"},
			Value: s.namespace,
		}},
	}
	erro = en.EncodeElement(returns, soapbody)
	if erro != nil {
		serverSoapFault(
			w, NewSoapFault("Server", "Respone错误", erro.Error()),
		)
		return
	}
	w.Header().Set("Content-type", "text/xml")
	b, _ := xml.Marshal(soap.NewEnvelope(buf.Bytes()))
	w.Write(b)

}

func (s *Server) bind(port string) error {
	// 本服务的绑定地址
	adr := fmt.Sprintf("http://%s:%s/%s", BindHost, port, s.name)
	s.wsdl.Service.Port.Address.Location = adr
	b, erro := xml.Marshal(s.wsdl)
	if erro != nil {
		return erro
	}
	s.wsdlcache = []byte(xml.Header)
	s.wsdlcache = append(s.wsdlcache, b...)
	http.HandleFunc("/"+s.name, s.handleFunc)
	return nil
}

// func get_internal() string {
// 	addrs, err := net.InterfaceAddrs()
// 	if err == nil {
// 		for _, a := range addrs {
// 			if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
// 				if ipnet.IP.To4() != nil {
// 					return ipnet.IP.String()
// 				}
// 			}
// 		}
// 	}
// 	return ""
// }

func (s *Server) buildWsdl() {
	def := &wsdl.Definitions{
		Tns:      s.namespace,
		TargetNs: s.namespace,
		Soap:     "http://schemas.xmlsoap.org/wsdl/soap/",
		SoapEnv:  "http://schemas.xmlsoap.org/soap/envelope/",
		Wsdl:     "http://schemas.xmlsoap.org/wsdl/",
		Xsd:      "http://www.w3.org/2001/XMLSchema",
		Xsi:      "http://www.w3.org/2001/XMLSchema-instance",
	}
	sch := xsd.Schema{
		TargetNamespace: s.namespace,
		Import: []xsd.Import{
			{Namespace: "http://schemas.xmlsoap.org/soap/encoding/"},
			{Namespace: "http://schemas.xmlsoap.org/wsdl/"}},
	}
	def.Types.Schemas = append(def.Types.Schemas, sch)

	def.PortType.Name = s.name + "PortType"

	def.Binding.Name = s.name + "Binding"
	def.Binding.Type = "tns:" + def.PortType.Name
	def.Binding.SoapBinding.Style = "rpc"
	def.Binding.SoapBinding.Transport = "http://schemas.xmlsoap.org/soap/http"

	def.Service.Name = s.name
	def.Service.Port = wsdl.ServicePort{
		Name:    s.name + "Port",
		Binding: "tns:" + def.Binding.Name,
		// Address: ServiceAddress{Location: location},
	}
	s.wsdl = def
}
