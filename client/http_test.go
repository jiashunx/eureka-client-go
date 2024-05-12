package client

import (
    "github.com/jiashunx/eureka-client-go/log"
    "github.com/jiashunx/eureka-client-go/meta"
    "github.com/stretchr/testify/assert"
    "testing"
    "time"
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

var httpInstance = &meta.InstanceInfo{
    AppName:    "http-client-test",
    InstanceId: "127.0.0.1:18080",
    IpAddr:     "127.0.0.1",
    Status:     meta.StatusUp,
    SecurePort: &meta.PortWrapper{
        Enabled: meta.StrTrue,
        Port:    18080,
    },
    VipAddress:       "http-client-test-A",
    SecureVipAddress: "http-client-test-B",
}

var httpClient = &HttpClient{logger: log.DefaultLogger()}

func TestInit(t *testing.T) {
    httpClient.logger.SetLevel(log.DebugLevel)
}

func TestRegister(t *testing.T) {
    ast := assert.New(t)
    response := httpClient.SimpleRegister(serviceUrl1, httpInstance)
    ast.Nilf(response.Error, "%v", response.Error)
    ast.Equal(204, response.StatusCode)
}

func TestHeartbeat(t *testing.T) {
    <-time.NewTimer(time.Second).C
    ast := assert.New(t)
    response := httpClient.SimpleHeartbeat(serviceUrl1, httpInstance.AppName, httpInstance.InstanceId)
    ast.Nilf(response.Error, "%v", response.Error)
    ast.Equal(200, response.StatusCode)
}

func TestQueryApps(t *testing.T) {
    <-time.NewTimer(60 * time.Second).C
    ast := assert.New(t)
    response := httpClient.SimpleQueryApps(serviceUrl1)
    ast.Nilf(response.Error, "%v", response.Error)
    ast.Equal(200, response.StatusCode)
    ast.True(len(response.Apps) > 0)
}

func TestQueryApp(t *testing.T) {
    <-time.NewTimer(time.Second).C
    ast := assert.New(t)
    response := httpClient.SimpleQueryApp(serviceUrl1, httpInstance.AppName)
    ast.Nilf(response.Error, "%v", response.Error)
    ast.Equal(200, response.StatusCode)
    ast.True(len(response.Instances) > 0)
}

func TestQueryAppInstance(t *testing.T) {
    <-time.NewTimer(time.Second).C
    ast := assert.New(t)
    response := httpClient.SimpleQueryAppInstance(serviceUrl1, httpInstance.AppName, httpInstance.InstanceId)
    ast.Nilf(response.Error, "%v", response.Error)
    ast.Equal(200, response.StatusCode)
    ast.NotNil(response.Instance)
}

func TestQueryInstance(t *testing.T) {
    <-time.NewTimer(time.Second).C
    ast := assert.New(t)
    response := httpClient.SimpleQueryInstance(serviceUrl1, httpInstance.InstanceId)
    ast.Nilf(response.Error, "%v", response.Error)
    ast.Equal(200, response.StatusCode)
    ast.NotNil(response.Instance)
}

func TestChangeStatus(t *testing.T) {
    <-time.NewTimer(time.Second).C
    ast := assert.New(t)
    response := httpClient.SimpleChangeStatus(serviceUrl1, httpInstance.AppName, httpInstance.InstanceId, meta.StatusOutOfService)
    ast.Nilf(response.Error, "%v", response.Error)
    ast.Equal(200, response.StatusCode)
}

func TestModifyMetadata(t *testing.T) {
    <-time.NewTimer(time.Second).C
    ast := assert.New(t)
    response := httpClient.SimpleModifyMetadata(serviceUrl1, httpInstance.AppName, httpInstance.InstanceId, map[string]string{"hello": "world"})
    ast.Nilf(response.Error, "%v", response.Error)
    ast.Equal(200, response.StatusCode)
}

func TestQueryVipApps(t *testing.T) {
    <-time.NewTimer(time.Second).C
    ast := assert.New(t)
    response := httpClient.SimpleQueryVipApps(serviceUrl1, httpInstance.VipAddress)
    ast.Nilf(response.Error, "%v", response.Error)
    ast.Equal(200, response.StatusCode)
    ast.True(len(response.Apps) > 0)
}

func TestQuerySvipApps(t *testing.T) {
    <-time.NewTimer(time.Second).C
    ast := assert.New(t)
    response := httpClient.SimpleQuerySvipApps(serviceUrl1, httpInstance.SecureVipAddress)
    ast.Nilf(response.Error, "%v", response.Error)
    ast.Equal(200, response.StatusCode)
    ast.True(len(response.Apps) > 0)
}

func TestUnRegister(t *testing.T) {
    <-time.NewTimer(time.Second).C
    ast := assert.New(t)
    response := httpClient.SimpleUnRegister(serviceUrl1, httpInstance.AppName, httpInstance.InstanceId)
    ast.Nilf(response.Error, "%v", response.Error)
    ast.Equal(200, response.StatusCode)
}
