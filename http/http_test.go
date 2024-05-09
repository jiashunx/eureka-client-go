package http

import (
    "github.com/jiashunx/eureka-client-go/meta"
    "github.com/stretchr/testify/assert"
    "testing"
)

// 所有eureka server均无安全认证
var serviceUrl = "http://127.0.0.1:20000/eureka,http://192.168.138.130:20000/eureka"

// 部分eureka server有安全认证
var serviceUrl2 = "http://admin:123123@127.0.0.1:20000/eureka,http://192.168.138.130:20000/eureka"

// 指定eureka server安全认证信息（若serviceUrl中有安全认证信息，则其优先级更高）
var eurekaServer2 = &meta.EurekaServer{
    ServiceUrl: serviceUrl2,
    Username:   "admin",
    Password:   "123123",
}

var instance = &meta.InstanceInfo{AppName: "test", InstanceId: "hello"}

func TestRegister(t *testing.T) {
    ast := assert.New(t)
    response := SimpleRegister(serviceUrl, instance)
    ast.Nilf(response.Error, "Register处理失败，失败原因：%s", response.Error)
    ast.Equal(204, response.StatusCode)
}

func TestHeartbeat(t *testing.T) {
    ast := assert.New(t)
    response := SimpleHeartbeat(serviceUrl, instance.AppName, instance.InstanceId)
    ast.Nilf(response.Error, "Heartbeat处理失败，失败原因：%s", response.Error)
    ast.Equal(200, response.StatusCode)
}

func TestChangeStatus(t *testing.T) {
    ast := assert.New(t)
    response := SimpleChangeStatus(serviceUrl, instance.AppName, instance.InstanceId, meta.StatusOutOfService)
    ast.Nilf(response.Error, "ChangeStatus处理失败，失败原因：%s", response.Error)
    ast.Equal(200, response.StatusCode)
}

func TestModifyMetadata(t *testing.T) {
    ast := assert.New(t)
    response := SimpleModifyMetadata(serviceUrl, instance.AppName, instance.InstanceId, "hello", "world")
    ast.Nilf(response.Error, "ModifyMetadata处理失败，失败原因：%s", response.Error)
    ast.Equal(200, response.StatusCode)
}

func TestUnRegister(t *testing.T) {
    ast := assert.New(t)
    response := SimpleUnRegister(serviceUrl, instance.AppName, instance.InstanceId)
    ast.Nilf(response.Error, "UnRegister处理失败，失败原因：%s", response.Error)
    ast.Equal(200, response.StatusCode)
}
