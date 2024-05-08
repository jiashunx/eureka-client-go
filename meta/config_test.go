package meta

import (
    "fmt"
    "github.com/stretchr/testify/assert"
    "testing"
)

func TestNewEurekaConfig(t *testing.T) {
    ast := assert.New(t)
    ec, err := NewEurekaConfig(nil, nil)
    ast.Nil(err, "创建EurekaConfig实例失败")
    ast.NotNil(ec, "创建EurekaConfig实例失败")
    fmt.Println(ec)
}
