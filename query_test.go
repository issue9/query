// SPDX-License-Identifier: MIT

package query

import (
	"bytes"
	"encoding"
	"fmt"
)

type State int

type Text int

const (
	StateNormal State = iota + 1
	StateLocked
	StateLeft
)

const (
	TextNormal Text = iota
	TextDeleted
)

var (
	x State       = StateLeft
	_ Unmarshaler = (*State)(&x)

	y Text                     = TextDeleted
	_ encoding.TextUnmarshaler = (*Text)(&y)
)

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

func (t *Text) UnmarshalText(bs []byte) error {
	if bytes.Equal(bs, []byte("normal")) {
		*t = TextNormal
	} else if bytes.Equal(bs, []byte("deleted")) {
		*t = TextDeleted
	} else {
		return fmt.Errorf("无效的值")
	}
	return nil
}

type testQueryString struct {
	String     string   `query:"string,str1,str2"`
	Strings    []string `query:"strings,str1,str2"`
	State      State    `query:"state,normal"`
	Text       Text     `query:"text,normal"`
	unExported string
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
