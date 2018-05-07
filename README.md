query [![Build Status](https://travis-ci.org/issue9/query.svg?branch=master)](https://travis-ci.org/issue9/query)
[![Go version](https://img.shields.io/badge/Go-1.10-brightgreen.svg?style=flat)](https://golang.org)
[![Go Report Card](https://goreportcard.com/badge/github.com/issue9/query)](https://goreportcard.com/report/github.com/issue9/query)
[![license](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat)](https://opensource.org/licenses/MIT)
[![codecov](https://codecov.io/gh/issue9/query/branch/master/graph/badge.svg)](https://codecov.io/gh/issue9/query)
======


提供了将 web 请求中的查询参数解析到结构体的操作。

```go
type State int8

const (
    StateLocked State = iota+1
    StateDelete
)

// 实现 query.UnmarshalQueryer 接口
func (s *State) UnmarshalQuery(data string) error {
    switch data {
    case "locked":
        *s = StateLocked
    case "delete":
        *s = StateDelete
    default:
        return errors.New("无效的值")
    }
}

type struct Query {
    Page int `query:"page,1"`
    Size int `query:"size,20"`
    States []State `query:"state,normal"`
}

func (q *Query) SanitizeQuery(errors map[string]string) {
    if q.Page < 0 {
        errors["page"] = "不能小于零"
    }

    // 其它字段的验证
}


func handle(w http.ResponseWriter, r *http.Request) {
    q := &Query{}
    errors := query.Parse(r, q)
    if len(errors) > 0 {
        // TODO
        return
    }

    // 请求参数为 /?page=1&size=2&state=normal,delete
    // 则 q 的值为
    // page = 1
    // size = 2
    // states = []State{StateLocked, StateDelete}
    //
    // 参数 state 也可使用以下方式
    // /?page=1&size=2&state=normal&normal=delete
}
```


### 安装

```shell
go get github.com/issue9/query
```


### 文档

[![Go Walker](https://gowalker.org/api/v1/badge)](http://gowalker.org/github.com/issue9/query)
[![GoDoc](https://godoc.org/github.com/issue9/query?status.svg)](https://godoc.org/github.com/issue9/query)



### 版权

本项目采用 [MIT](https://opensource.org/licenses/MIT) 开源授权许可证，完整的授权说明可在 [LICENSE](LICENSE) 文件中找到。
