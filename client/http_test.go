package client

import (
    "fmt"
    "github.com/jiashunx/eureka-client-go/meta"
    "github.com/stretchr/testify/assert"
    "testing"
)

// 所有eureka server均无安全认证
var serviceUrl1 = "http://127.0.0.1:20000/eureka,http://192.168.138.130:20000/eureka"

// 部分eureka server有安全认证
var serviceUrl2 = "http://admin:123123@127.0.0.1:20000/eureka,http://192.168.138.130:20000/eureka"

// 指定eureka server安全认证信息（若serviceUrl中有安全认证信息，则其优先级更高）
var eurekaServer2 = &meta.EurekaServer{
    ServiceUrl: serviceUrl2,
    Username:   "admin",
    Password:   "123123",
}

var instance = &meta.InstanceInfo{
    AppName:          "test",
    InstanceId:       "hello",
    VipAddress:       meta.DefaultVirtualHostname,
    SecureVipAddress: meta.DefaultSecureVirtualHostname,
}

var httpClient = &HttpClient{}

func TestRegister(t *testing.T) {
    ast := assert.New(t)
    response := httpClient.SimpleRegister(serviceUrl1, instance)
    ast.Nilf(response.Error, "Register处理失败，失败原因：%s", response.Error)
    ast.Equal(204, response.StatusCode)
    fmt.Println("Register请求方法:", response.Response.Request.Method)
    fmt.Println("Register请求URL:", response.Response.Request.RequestUrl)
    fmt.Println("Register请求报文:", response.Response.Request.Body)
    fmt.Println("Register响应状态:", response.StatusCode)
}

func TestHeartbeat(t *testing.T) {
    ast := assert.New(t)
    response := httpClient.SimpleHeartbeat(serviceUrl1, instance.AppName, instance.InstanceId)
    ast.Nilf(response.Error, "Heartbeat处理失败，失败原因：%s", response.Error)
    ast.Equal(200, response.StatusCode)
    fmt.Println("Heartbeat请求方法:", response.Response.Request.Method)
    fmt.Println("Heartbeat请求URL:", response.Response.Request.RequestUrl)
    fmt.Println("Heartbeat响应状态:", response.StatusCode)
}

func TestQueryApps(t *testing.T) {
    ast := assert.New(t)
    response := httpClient.SimpleQueryApps(serviceUrl1)
    ast.Nilf(response.Error, "QueryApps处理失败，失败原因：%s", response.Error)
    ast.Equal(200, response.StatusCode)
    fmt.Println("QueryApps请求方法:", response.Response.Request.Method)
    fmt.Println("QueryApps请求URL:", response.Response.Request.RequestUrl)
    fmt.Println("QueryApps响应状态:", response.StatusCode)
    fmt.Println("QueryApps响应报文:", response.Response.Body)
    fmt.Println("QueryApps返回结果:", fmt.Sprintf("%#v", *response.Apps[0]))
}

func TestQueryApp(t *testing.T) {
    ast := assert.New(t)
    response := httpClient.SimpleQueryApp(serviceUrl1, instance.AppName)
    ast.Nilf(response.Error, "QueryApp处理失败，失败原因：%s", response.Error)
    ast.Equal(200, response.StatusCode)
    fmt.Println("QueryApp请求方法:", response.Response.Request.Method)
    fmt.Println("QueryApp请求URL:", response.Response.Request.RequestUrl)
    fmt.Println("QueryApp响应状态:", response.StatusCode)
    fmt.Println("QueryApp响应报文:", response.Response.Body)
    fmt.Println("QueryApp返回结果:", fmt.Sprintf("%#v", *response.Instances[0]))
}

func TestQueryAppInstance(t *testing.T) {
    ast := assert.New(t)
    response := httpClient.SimpleQueryAppInstance(serviceUrl1, instance.AppName, instance.InstanceId)
    ast.Nilf(response.Error, "QueryAppInstance处理失败，失败原因：%s", response.Error)
    ast.Equal(200, response.StatusCode)
    fmt.Println("QueryAppInstance请求方法:", response.Response.Request.Method)
    fmt.Println("QueryAppInstance请求URL:", response.Response.Request.RequestUrl)
    fmt.Println("QueryAppInstance响应状态:", response.StatusCode)
    fmt.Println("QueryAppInstance响应报文:", response.Response.Body)
    fmt.Println("QueryAppInstance返回结果:", fmt.Sprintf("%#v", *response.Instance))
}

func TestQueryInstance(t *testing.T) {
    ast := assert.New(t)
    response := httpClient.SimpleQueryInstance(serviceUrl1, instance.InstanceId)
    ast.Nilf(response.Error, "QueryInstance处理失败，失败原因：%s", response.Error)
    ast.Equal(200, response.StatusCode)
    fmt.Println("QueryInstance请求方法:", response.Response.Request.Method)
    fmt.Println("QueryInstance请求URL:", response.Response.Request.RequestUrl)
    fmt.Println("QueryInstance响应状态:", response.StatusCode)
    fmt.Println("QueryInstance响应报文:", response.Response.Body)
    fmt.Println("QueryInstance返回结果:", fmt.Sprintf("%#v", *response.Instance))
}

func TestChangeStatus(t *testing.T) {
    ast := assert.New(t)
    response := httpClient.SimpleChangeStatus(serviceUrl1, instance.AppName, instance.InstanceId, meta.StatusOutOfService)
    ast.Nilf(response.Error, "ChangeStatus处理失败，失败原因：%s", response.Error)
    ast.Equal(200, response.StatusCode)
    fmt.Println("ChangeStatus请求方法:", response.Response.Request.Method)
    fmt.Println("ChangeStatus请求URL:", response.Response.Request.RequestUrl)
    fmt.Println("ChangeStatus响应状态:", response.StatusCode)
}

func TestModifyMetadata(t *testing.T) {
    ast := assert.New(t)
    response := httpClient.SimpleModifyMetadata(serviceUrl1, instance.AppName, instance.InstanceId, map[string]string{"hello": "world"})
    ast.Nilf(response.Error, "ModifyMetadata处理失败，失败原因：%s", response.Error)
    ast.Equal(200, response.StatusCode)
    fmt.Println("ModifyMetadata请求方法:", response.Response.Request.Method)
    fmt.Println("ModifyMetadata请求URL:", response.Response.Request.RequestUrl)
    fmt.Println("ModifyMetadata响应状态:", response.StatusCode)
}

func TestQueryVipApps(t *testing.T) {
    ast := assert.New(t)
    response := httpClient.SimpleQueryVipApps(serviceUrl1, instance.VipAddress)
    ast.Nilf(response.Error, "QueryVipApps处理失败，失败原因：%s", response.Error)
    ast.Equal(200, response.StatusCode)
    fmt.Println("QueryVipApps请求方法:", response.Response.Request.Method)
    fmt.Println("QueryVipApps请求URL:", response.Response.Request.RequestUrl)
    fmt.Println("QueryVipApps响应状态:", response.StatusCode)
    fmt.Println("QueryVipApps响应报文:", response.Response.Body)
    fmt.Println("QueryVipApps返回结果:", fmt.Sprintf("%#v", *response.Apps[0]))
}

func TestQuerySvipApps(t *testing.T) {
    ast := assert.New(t)
    response := httpClient.SimpleQuerySvipApps(serviceUrl1, instance.SecureVipAddress)
    ast.Nilf(response.Error, "QuerySvipApps处理失败，失败原因：%s", response.Error)
    ast.Equal(200, response.StatusCode)
    fmt.Println("QuerySvipApps请求方法:", response.Response.Request.Method)
    fmt.Println("QuerySvipApps请求URL:", response.Response.Request.RequestUrl)
    fmt.Println("QuerySvipApps响应状态:", response.StatusCode)
    fmt.Println("QuerySvipApps响应报文:", response.Response.Body)
    fmt.Println("QuerySvipApps返回结果:", fmt.Sprintf("%#v", *response.Apps[0]))
}

func TestUnRegister(t *testing.T) {
    ast := assert.New(t)
    response := httpClient.SimpleUnRegister(serviceUrl1, instance.AppName, instance.InstanceId)
    ast.Nilf(response.Error, "UnRegister处理失败，失败原因：%s", response.Error)
    ast.Equal(200, response.StatusCode)
    fmt.Println("UnRegister请求方法:", response.Response.Request.Method)
    fmt.Println("UnRegister请求URL:", response.Response.Request.RequestUrl)
    fmt.Println("UnRegister响应状态:", response.StatusCode)
}
