// SPDX-License-Identifier: MIT

package query

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/issue9/assert/v2"
)

func TestParse(t *testing.T) {
	a := assert.New(t, false)
	r := httptest.NewRequest(http.MethodGet, "/q?string=str&strings=s1,s2&int=0", nil)
	data := &testQueryObject{}

	errors := Parse(r.URL.Query(), data)
	a.Equal(len(errors["int"]), 2)
	a.Equal(data.Int, 0).
		Equal(data.String, "str").
		Equal(data.Strings, []string{"s1", "s2"})
}

func TestParseField(t *testing.T) {
	a := assert.New(t, false)

	errors := map[string][]string{}
	r := httptest.NewRequest(http.MethodGet, "/q?string=str&strings=s1,s2&text=deleted", nil)
	data := &testQueryObject{}
	parseField(r.URL.Query(), reflect.ValueOf(data).Elem(), errors)
	a.Empty(errors)
	a.Equal(data.String, "str").
		Equal(data.State, StateNormal).
		Equal(data.Strings, []string{"s1", "s2"}).
		Equal(data.Int, 1). // 默认值
		Equal(data.Floats, []float64{1.1, 2.2}).
		Equal(data.Text, TextDeleted)

	errors = map[string][]string{}
	r = httptest.NewRequest(http.MethodGet, "/q?floats=1,1.1&int=5&strings=s1", nil)
	data = &testQueryObject{}
	parseField(r.URL.Query(), reflect.ValueOf(data).Elem(), errors)
	a.Empty(errors)
	a.Equal(data.String, "str1,str2").
		Equal(data.Floats, []float64{1.0, 1.1}).
		Equal(data.Strings, []string{"s1"})

	// 非英文内容
	errors = map[string][]string{}
	r = httptest.NewRequest(http.MethodGet, "/q?字符串=字符串1&字符串列表=字符串2&字符串列表=字符串3", nil)
	cnobj := &testCNQueryString{}
	parseField(r.URL.Query(), reflect.ValueOf(cnobj).Elem(), errors)
	a.Empty(errors)
	a.Equal(cnobj.String, "字符串1").
		Equal(cnobj.Strings, []string{"字符串2", "字符串3"})

	// 出错时的处理
	errors = map[string][]string{}
	r = httptest.NewRequest(http.MethodGet, "/q?floats=str,1.1&array=10&int=5&strings=s1", nil)
	data = &testQueryObject{
		Floats: []float64{3.3, 4.4},
	}
	parseField(r.URL.Query(), reflect.ValueOf(data).Elem(), errors)
	a.NotEmpty(errors["floats"]) // floats 解析会出错
	a.Equal(data.String, "str1,str2").
		Empty(data.Floats).
		Equal(data.Strings, []string{"s1"})
}

func TestParseField_slice(t *testing.T) {
	a := assert.New(t, false)

	// 指定了默认值，也指定了参数。则以参数优先
	errors := map[string][]string{}
	r := httptest.NewRequest(http.MethodGet, "/q?floats=11.1", nil)
	data := &testQueryObject{
		Floats: []float64{3.3, 4.4},
	}
	parseField(r.URL.Query(), reflect.ValueOf(data).Elem(), errors)
	a.Empty(errors)
	a.Equal(data.Floats, []float64{11.1})

	// 指定了默认值，指定空参数，则使用默认值
	errors = map[string][]string{}
	r = httptest.NewRequest(http.MethodGet, "/q?floats=", nil)
	data = &testQueryObject{
		Floats: []float64{3.3, 4.4},
	}
	parseField(r.URL.Query(), reflect.ValueOf(data).Elem(), errors)
	a.Empty(errors)
	a.Equal(data.Floats, []float64{3.3, 4.4})

	// 指定了默认值，未指定参数，则使用默认值
	errors = map[string][]string{}
	r = httptest.NewRequest(http.MethodGet, "/q", nil)
	data = &testQueryObject{
		Floats: []float64{3.3, 4.4},
	}
	parseField(r.URL.Query(), reflect.ValueOf(data).Elem(), errors)
	a.Empty(errors)
	a.Equal(data.Floats, []float64{3.3, 4.4})

	// 都未指定，则使用 struct tag 中的默认值
	errors = map[string][]string{}
	r = httptest.NewRequest(http.MethodGet, "/q", nil)
	data = &testQueryObject{}
	parseField(r.URL.Query(), reflect.ValueOf(data).Elem(), errors)
	a.Empty(errors)
	a.Equal(data.Floats, []float64{1.1, 2.2})

	// 采用 x=1&x=2的方式传递数组
	errors = map[string][]string{}
	r = httptest.NewRequest(http.MethodGet, "/q?floats=3.3&floats=4.4", nil)
	data = &testQueryObject{}
	parseField(r.URL.Query(), reflect.ValueOf(data).Elem(), errors)
	a.Empty(errors)
	a.Equal(data.Floats, []float64{3.3, 4.4})

	// 采用 x=1&x=2的方式传递数组，且值中带逗号
	errors = map[string][]string{}
	r = httptest.NewRequest(http.MethodGet, "/q?strings=str1&strings=str2,str3", nil)
	data = &testQueryObject{}
	parseField(r.URL.Query(), reflect.ValueOf(data).Elem(), errors)
	a.Empty(errors)
	a.Equal(data.Strings, []string{"str1", "str2,str3"})

	// 无法解析的参数
	errors = map[string][]string{}
	r = httptest.NewRequest(http.MethodGet, "/q?floats=3x.5,bb", nil)
	data = &testQueryObject{}
	parseField(r.URL.Query(), reflect.ValueOf(data).Elem(), errors)
	a.True(len(errors["floats"]) > 0)
	a.Empty(data.Floats)
}

func TestGetQueryTag(t *testing.T) {
	a := assert.New(t, false)

	test := func(tag, name, def string) {
		field := reflect.StructField{
			Tag: reflect.StructTag(tag),
		}
		n, d := getQueryTag(field)
		a.Equal(n, name).
			Equal(d, def)
	}

	test(Tag+`:"name,def"`, "name", "def")
	test(Tag+`:",def"`, "", "def")
	test(Tag+`:"name,"`, "name", "")
	test(Tag+`:"name"`, "name", "")
	test(Tag+`:"name,1,2"`, "name", "1,2")
	test(Tag+`:"name,1,2,"`, "name", "1,2,")
	test(Tag+`:"-"`, "", "")
}
