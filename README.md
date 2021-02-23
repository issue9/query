query
[![Go](https://github.com/issue9/query/workflows/Go/badge.svg)](https://github.com/issue9/query/actions?query=workflow%3AGo)
![Go version](https://img.shields.io/github/go-mod/go-version/issue9/query)
[![Go Report Card](https://goreportcard.com/badge/github.com/issue9/query)](https://goreportcard.com/report/github.com/issue9/query)
![License](https://img.shields.io/github/license/issue9/query)
[![codecov](https://codecov.io/gh/issue9/query/branch/master/graph/badge.svg)](https://codecov.io/gh/issue9/query)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/issue9/query/v2)](https://pkg.go.dev/github.com/issue9/query/v2)
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
    errors := query.Parse(r.URL.Query(), q)
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

版权
----

本项目采用 [MIT](https://opensource.org/licenses/MIT) 开源授权许可证，完整的授权说明可在 [LICENSE](LICENSE) 文件中找到。
