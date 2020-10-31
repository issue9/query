// SPDX-License-Identifier: MIT

// Package query 提供将查询参数解析到结构体的相关操作
//
//
// struct tag
//
// 通过 struct tag 的方式将查询参数与结构体中的字段进行关联。
// struct tag 的格式如下：
//  `query:"name,default"`
// 其中 name 为对应查询参数的名称，若是为空则采用字段本身的名称；
// default 表示在没有参数的情况下，采用的默认值，可以为空。
// 若是将整个值设置为 -，则表示忽略当前字段。
//
//
// 数组：
//
// 如果字段表示的是切片，那么查询参数的值，将以半角逗号作为分隔符进行转换写入到切片中。
// struct tag 中的默认值，也可以指定多个：
//  type Object struct {
//      Slice []string `query:"slices,1,2"`
//  }
// 以上内容，在没有指定参数的情况下，Slice 会被指定为 []string{"1", "2"}
//
// 若 URL 中指定了 /?slices=4,5,6，则 Slice 的值会被设置为 []string{"4", "5", "6"}
//
// 如果值中有逗号，则可以使用 slices=v1&slices=v2,v3 的方式将值解析成 []string{"v1", "v2,v3"}
//
//
// 默认值：
//
// 默认值可以通过 struct tag 指定，也可以通过在初始化对象时，另外指定：
//  obj := &Object{
//      Slice: []int{3,4,5}
//  }
// 以上内容，在不传递参数时，会采用 []int{3,4,5} 作为其默认值，而不是 struct tag
// 中指定的 []int{1,2}。
package query

import "github.com/issue9/validation"

// Tag 在 struct tag 的标签名称
const Tag = "query"

// Sanitizer 表示对一个查询参数构成的结构体进行数据验证和内容修正的接口
type Sanitizer interface {
	// 参数 errors 用来保存由函数中发现的错误信息。
	//
	// 其中的键名为错误字段名称，键值为错误信息。
	SanitizeQuery(errors Errors)
}

// Unmarshaler 该接口实现在了将一些特定的查询参数格式转换成其它类型的接口
//
// 比如一个查询参数格式如下：
//  /path?state=locked
// 而实际上后端将 state 表示为一个数值：
//  type State int8
//  const StateLocked State = 1
// 那么只要 State 实现 Unmarshaler 接口，就可以实现将 locked 转换成 1 的能力。
//  func (s *State) UnmarshalQuery(data string) error {
//      if data == "locked" {
//          *s = StateLocked
//      }
//  }
//
// NOTE: 空值不会调用该接口。
type Unmarshaler interface {
	// data 表示由查询参数传递过来的单个值。
	UnmarshalQuery(data string) error
}

// Errors 表示一组错误信息的集合
//
// 键名查询参数名称，键值则为在解析和验证过种中返回的错误信息。
type Errors = validation.Messages
