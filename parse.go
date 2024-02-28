// SPDX-FileCopyrightText: 2018-2024 caixw
//
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

// Parse 将查询参数解析到 v
func Parse(queries url.Values, v any) map[string]error {
	errors := make(map[string]error, 10)
	ParseWithLog(queries, v, func(name string, err error) { errors[name] = err })
	return errors
}

// ParseWithLog 将查询参数解析至 v 中
//
// 如果 queries 中的元素，实现了 [Unmarshaler] 或是 [encoding.TextUnmarshaler]，
// 则会调用相应的接口解码。
//
// 如果有错误，则调用 log 方法进行处理，log 原型如下：
//
//	func(name string, err error)
//
// 其中的 name 表示查询参数名称，err 表示解析该参数时的错误信息。
//
// v 只能是指针，如果是指针的批针，请注意接口的实现是否符合你的预期。
//
// NOTE: ParseWithLog 适合在已经有错误处理方式的代码中使用，比如库的作者。
// 一般情况下 [Parse] 会更佳合适。
func ParseWithLog(queries url.Values, v any, log func(string, error)) {
	rval := reflect.ValueOf(v)
	if rval.Kind() != reflect.Ptr {
		panic("v 必须为指针")
	}

	for rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	// NOTE: 不应该由 Parse 实现对整个对象内容的检测，那应该是 v 的实现应当做的事。

	parseField(queries, rval, log)
}

func parseField(vals url.Values, rval reflect.Value, log func(string, error)) {
	rtype := rval.Type()
	for i := 0; i < rtype.NumField(); i++ {
		tf := rtype.Field(i)

		if tf.Anonymous {
			parseField(vals, rval.Field(i), log)
			continue
		}

		vf := rval.Field(i)

		switch tf.Type.Kind() {
		case reflect.Slice:
			parseSliceFieldValue(vals, log, tf, vf)
		case reflect.Ptr, reflect.Chan, reflect.Func, reflect.Array, reflect.Complex128, reflect.Complex64:
			// 这些类型的字段，直接忽略
		default:
			parseFieldValue(vals, log, tf, vf)
		}
	}
}

func parseFieldValue(vals url.Values, log func(string, error), tf reflect.StructField, vf reflect.Value) {
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

	unmarshal(name, vf.Addr(), val, log)
}

func parseSliceFieldValue(form url.Values, log func(string, error), tf reflect.StructField, vf reflect.Value) {
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

	elemType := tf.Type.Elem()
	for elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}
	for _, v := range vals {
		elem := reflect.New(elemType)
		if !unmarshal(name, elem, v, log) {
			return
		}
		vf.Set(reflect.Append(vf, elem.Elem()))
	}
}

func unmarshal(name string, vf reflect.Value, val string, log func(string, error)) (ok bool) {
	if q, ok := vf.Interface().(Unmarshaler); ok {
		if err := q.UnmarshalQuery(val); err != nil {
			log(name, err)
			return false
		}
	} else if u, ok := vf.Interface().(encoding.TextUnmarshaler); ok {
		if err := u.UnmarshalText([]byte(val)); err != nil {
			log(name, err)
			return false
		}
	} else if err := conv.Value(val, vf); err != nil {
		log(name, err)
		return false
	}
	return true
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
