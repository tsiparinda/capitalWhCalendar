package utils

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"reflect"

	dec "github.com/shopspring/decimal"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

func StructToMap(item interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func StructToMapNew(item interface{}) (interface{}, error) {
	data, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}

	if data[0] == '[' {
		var arr []interface{}
		err = json.Unmarshal(data, &arr)
		if err != nil {
			return nil, err
		}
		return arr, nil
	} else {
		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
}

func Utf8ToWin1251(input string) (string, error) {
	decoder := charmap.Windows1251.NewDecoder()
	output, err := decoder.String(input)
	if err != nil {
		return "", err
	}
	return output, nil
}

func Str2Dec(str string, prec int32) (summ dec.Decimal, err error) {
	summ, err = dec.NewFromString(func() string {
		if str == "" {
			return "0"
		}
		return str
	}())
	if err != nil {
		// fmt.Println("Failed to convert string to decimal:", err.Error())
		return dec.Zero, err
	}
	summ = summ.Round(prec)
	return
}

func ConvertToCP1251(data []byte) ([]byte, error) {
	reader := bytes.NewReader(data)
	transformer := charmap.Windows1251.NewEncoder()
	transformedReader := transform.NewReader(reader, transformer)
	return ioutil.ReadAll(transformedReader)
}

func EncodeWindows1251(ba []uint8) ([]uint8, error) {
	enc := charmap.Windows1251.NewEncoder()
	out, err := enc.String(string(ba))
	if err != nil {
		return []uint8(""), err
	}
	return []uint8(out), nil
}

func NS(i interface{}) string {
	if i == nil {
		return ""
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.String:
		return reflect.ValueOf(i).String()
	}
	return ""
}
