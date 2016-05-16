package main

import (
	"fmt"
	"github.com/afocus/gosoap"
	"math/rand"
	"time"
)

// 简单的
type User struct {
	In struct {
		Id int64
	}
	Out struct {
		Id       int64
		Name     string
		Sex      int
		Address  string
		Birthday time.Time
	}
}

func (u *User) Action() *gosoap.SoapFault {

	if u.In.Id != 100 {
		return gosoap.NewSoapFault("输入参数错误", "id必须是100", "")
	}

	u.Out.Id = u.In.Id
	u.Out.Name = "Afocus"
	u.Out.Birthday = time.Now()
	u.Out.Sex = 1
	u.Out.Address = "陕西西安未央区"

	return nil
}

// 返回带列表的
type DataList struct {
	In struct {
		Page    int
		PerPage int
		Search  string
	}
	Out struct {
		Total int
		Items []struct {
			Id     int64
			Name   string
			Status int
		}
	}
}

func (d *DataList) Action() *gosoap.SoapFault {
	fmt.Printf("%+v\n", d.In)
	d.Out.Total = 10
	d.Out.Items = make([]struct {
		Id     int64
		Name   string
		Status int
	}, 10)
	for index := 0; index < 10; index++ {
		d.Out.Items[index].Id = int64(index)
		//
		rand.Seed(time.Now().UnixNano())
		kinds := [][]int{[]int{10, 48}, []int{26, 97}, []int{26, 65}}
		name := make([]byte, 10)
		for i := 0; i < 10; i++ {
			ikind := rand.Intn(3)
			scope, base := kinds[ikind][0], kinds[ikind][1]
			name[i] = uint8(base + rand.Intn(scope))
		}

		d.Out.Items[index].Name = fmt.Sprintf("%s%s%d", d.In.Search, string(name), index)
		d.Out.Items[index].Status = 0
	}
	return nil
}

func main() {
	s := gosoap.NewServer("people")
	s.Register(new(User), new(DataList))

	panic(s.Service("8080"))
}
