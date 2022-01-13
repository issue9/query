// SPDX-License-Identifier: MIT

package query

import (
	"encoding"
	"net/url"
	"reflect"
	"strings"
	"unicode"

	"github.com/issue9/conv"
)

// Parse 将查询参数解析到一个对象中
//
// 返回的是每一个字段对应的错误信息。
// 如果 queries 中的元素，实现了 Unmarshaler 或是 encoding.TextUnmarshaler，
// 则会调用相应的接口解码。
//
// v 只能是指针，如果是指针的批针，请注意接口的实现是否符合你的预期。
func Parse(queries url.Values, v interface{}) (errors Errors) {
	rval := reflect.ValueOf(v)
	if rval.Kind() != reflect.Ptr {
		panic("v 必须为指针")
	}

	for rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	errors = make(Errors, rval.NumField())
	parseField(queries, rval, errors)

	// 接口在转换完成之后调用。
	if s, ok := v.(Sanitizer); ok {
		s.SanitizeQuery(errors)
	}

	return errors
}

func parseField(vals url.Values, rval reflect.Value, errors Errors) {
	rtype := rval.Type()
	for i := 0; i < rtype.NumField(); i++ {
		tf := rtype.Field(i)

		if tf.Anonymous {
			parseField(vals, rval.Field(i), errors)
			continue
		}

		vf := rval.Field(i)

		switch tf.Type.Kind() {
		case reflect.Slice:
			parseFieldSlice(vals, errors, tf, vf)
		case reflect.Ptr, reflect.Chan, reflect.Func, reflect.Array, reflect.Complex128, reflect.Complex64:
			// 这些类型的字段，直接忽略
		default:
			parseFieldValue(vals, errors, tf, vf)
		}
	} // end for
}

func parseFieldValue(vals url.Values, errors Errors, tf reflect.StructField, vf reflect.Value) {
	name, def := getQueryTag(tf)
	if name == "" {
		return
	}

	val := vals.Get(name)
	if val == "" {
		if vf.Interface() != reflect.Zero(tf.Type).Interface() {
			return
		}
		val = def
	}

	if val == "" { // 依然是空值
		return
	}

	unmarshal(name, vf.Addr(), val, errors)
}

func unmarshal(name string, vf reflect.Value, val string, errors Errors) (ok bool) {
	if q, ok := vf.Interface().(Unmarshaler); ok {
		if err := q.UnmarshalQuery(val); err != nil {
			errors.Add(name, err.Error())
			return false
		}
	} else if u, ok := vf.Interface().(encoding.TextUnmarshaler); ok {
		if err := u.UnmarshalText([]byte(val)); err != nil {
			errors.Add(name, err.Error())
			return false
		}
	} else if err := conv.Value(val, vf); err != nil {
		errors.Add(name, err.Error())
		return false
	}
	return true
}

func parseFieldSlice(form url.Values, errors Errors, tf reflect.StructField, vf reflect.Value) {
	name, def := getQueryTag(tf)
	if name == "" {
		return
	}

	vals := make([]string, 0, len(form[name]))
	for _, val := range form[name] {
		if val == "" {
			continue
		}
		vals = append(vals, val)
	}

	if len(vals) == 0 {
		if vf.Len() > 0 { // 实例有默认值，则采用默认值
			return
		}

		if def == "" {
			return
		}

		vals = []string{def}
	}

	if len(vals) == 1 {
		vals = strings.Split(vals[0], ",")
	}

	// 指定了参数，则舍弃 slice 中的旧值
	vf.Set(vf.Slice(0, 0))

	elemtype := tf.Type.Elem()
	for elemtype.Kind() == reflect.Ptr {
		elemtype = elemtype.Elem()
	}
	for _, v := range vals {
		elem := reflect.New(elemtype)
		if !unmarshal(name, elem, v, errors) {
			return
		}
		vf.Set(reflect.Append(vf, elem.Elem()))
	}
}

// 返回值中的 name 如果为空，表示忽略这个字段的内容。
func getQueryTag(field reflect.StructField) (name, def string) {
	if field.Name != "" && unicode.IsLower(rune(field.Name[0])) {
		return "", ""
	}
	tag := field.Tag.Get(Tag)
	if tag == "-" {
		return "", ""
	}

	tags := strings.SplitN(tag, ",", 2)

	switch len(tags) {
	case 0: // 都采用默认值
	case 1:
		name = strings.TrimSpace(tags[0])
	case 2:
		name = strings.TrimSpace(tags[0])
		def = strings.TrimSpace(tags[1])
	}

	if name == "" {
		name = field.Name
	}

	return name, def
}
