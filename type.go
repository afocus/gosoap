package gosoap

import (
	"errors"
	"reflect"
	"strings"

	"github.com/afocus/gosoap/xsd"
)

func checkBaseTypeKind(k reflect.Kind) (string, error) {

	switch k {
	case reflect.String:
		return xsd.String, nil

	case reflect.Int, reflect.Int32:
		return xsd.Int32, nil

	case reflect.Int64:
		return xsd.Int64, nil

	case reflect.Bool:
		return xsd.Bool, nil

	case reflect.Float32:
		return xsd.Float32, nil

	case reflect.Float64:
		return xsd.Float64, nil

	default:
		return "", errors.New("xxx")
	}

}

func getTagsInfo(t reflect.StructField) (string, bool) {
	required := false
	name := t.Name
	tags := strings.Split(t.Tag.Get("wsdl"), ",")
	for k, v := range tags {
		tag := strings.TrimSpace(v)
		if k == 0 {
			if tag != "" {
				name = tag
			}
		} else {
			if tag == "required" {
				required = true
				break
			}
		}
	}
	return name, required
}
