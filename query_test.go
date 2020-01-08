// SPDX-License-Identifier: MIT

package query

import "fmt"

type State int

const (
	StateNormal State = iota + 1 // 正常
	StateLocked                  // 锁定
	StateLeft                    // 离职
)

// UnmarshalQuery 解码
func (s *State) UnmarshalQuery(data string) error {
	switch data {
	case "normal":
		*s = StateNormal
	case "locked":
		*s = StateLocked
	case "left":
		*s = StateLeft
	default:
		return fmt.Errorf("无效的值：%s", string(data))
	}

	return nil
}

type testQueryString struct {
	String  string   `query:"string,str1,str2"`
	Strings []string `query:"strings,str1,str2"`
	State   State    `query:"state,normal"`
}

// 带中文的字段
type testCNQueryString struct {
	String  string   `query:"字符串,str1,str2"`
	Strings []string `query:"字符串列表,str1,str2"`
	State   State    `query:"state,normal"`
}

type testQueryObject struct {
	testQueryString
	Int    int       `query:"int,1"`
	Floats []float64 `query:"floats,1.1,2.2"`
	States []State   `query:"states,normal,left"`

	Array [5]int  // 即使不指定 query:"-" 也将被忽略
	Ints  []int   `query:"-"`
	Float float32 `query:"-"`
}

func (obj *testQueryString) SanitizeQuery(errors map[string]string) {
	if obj.State == -1 {
		errors["state"] = "取值错误"
	}
}

func (obj *testQueryObject) SanitizeQuery(errors map[string]string) {
	obj.testQueryString.SanitizeQuery(errors)

	if obj.Int == 0 {
		errors["int"] = "取值错误"
	}
}
