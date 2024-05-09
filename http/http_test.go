package http

import (
    "github.com/jiashunx/eureka-client-go/meta"
    "github.com/stretchr/testify/assert"
    "testing"
)

var serviceUrl = "http://127.0.0.1:20000/eureka,http://192.168.138.130:20000/eureka"

var instance = &meta.InstanceInfo{AppName: "test", InstanceId: "hello"}

func TestRegister(t *testing.T) {
    ast := assert.New(t)
    code, err := SimpleRegister(serviceUrl, instance)
    ast.Nilf(err, "Register处理失败，失败原因：%s", err)
    ast.Equal(204, code)
}

func TestHeartbeat(t *testing.T) {
    ast := assert.New(t)
    code, err := SimpleHeartbeat(serviceUrl, instance.AppName, instance.InstanceId)
    ast.Nilf(err, "Heartbeat处理失败，失败原因：%s", err)
    ast.Equal(200, code)
}

func TestUnRegister(t *testing.T) {
    ast := assert.New(t)
    code, err := SimpleUnRegister(serviceUrl, instance.AppName, instance.InstanceId)
    ast.Nilf(err, "UnRegister处理失败，失败原因：%s", err)
    ast.Equal(200, code)
}
