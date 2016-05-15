package main

import (
	"fmt"

	"github.com/afocus/gosoap"
)

type User struct {
	In struct {
		Name string
		Age  int
		Sex  bool
	}
	Out struct {
		Id   int
		Info string
	}
}

func (u *User) Action() *gosoap.SoapFault {
	fmt.Printf("%+v\n", u.In)
	u.Out.Id = 100
	u.Out.Info = "收到结果"
	return nil
}

type Error struct {
	In  string
	Out int
}

func (e *Error) Action() *gosoap.SoapFault {
	return gosoap.NewSoapFault("500", "内部错误", "")
}

func main() {
	s := gosoap.NewServer("people")
	s.Register(new(User), new(Error))
	panic(s.Service("8080"))
}
