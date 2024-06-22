package client

import (
    "github.com/jiashunx/eureka-client-go/log"
    "github.com/jiashunx/eureka-client-go/meta"
    "github.com/stretchr/testify/assert"
    "testing"
    "time"
)

var TestHttpServiceUrl = "http://admin:123123@127.0.0.1:20000/eureka,http://192.168.138.130:20000/eureka"

var TestHttpInstanceInfo = &meta.InstanceInfo{
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

var TestHttpClient = &HttpClient{}

func TestHttpClient_Init(t *testing.T) {
    TestHttpClient.GetLogger().SetLevel(log.InfoLevel)
}

func TestHttpClient_Register(t *testing.T) {
    ast := assert.New(t)
    response := TestHttpClient.SimpleRegister(TestHttpServiceUrl, TestHttpInstanceInfo)
    ast.Nilf(response.Error, "%v", response.Error)
    ast.Equal(204, response.StatusCode)
}

func TestHttpClient_Heartbeat(t *testing.T) {
    <-time.NewTimer(time.Second).C
    ast := assert.New(t)
    response := TestHttpClient.SimpleHeartbeat(TestHttpServiceUrl, TestHttpInstanceInfo.AppName, TestHttpInstanceInfo.InstanceId)
    ast.Nilf(response.Error, "%v", response.Error)
    ast.Equal(200, response.StatusCode)
}

func TestHttpClient_QueryApps(t *testing.T) {
    <-time.NewTimer(60 * time.Second).C
    ast := assert.New(t)
    response := TestHttpClient.SimpleQueryApps(TestHttpServiceUrl)
    ast.Nilf(response.Error, "%v", response.Error)
    ast.Equal(200, response.StatusCode)
    ast.True(len(response.Apps) > 0)
}

func TestHttpClient_QueryApp(t *testing.T) {
    <-time.NewTimer(time.Second).C
    ast := assert.New(t)
    response := TestHttpClient.SimpleQueryApp(TestHttpServiceUrl, TestHttpInstanceInfo.AppName)
    ast.Nilf(response.Error, "%v", response.Error)
    ast.Equal(200, response.StatusCode)
    ast.True(len(response.Instances) > 0)
}

func TestHttpClient_QueryAppInstance(t *testing.T) {
    <-time.NewTimer(time.Second).C
    ast := assert.New(t)
    response := TestHttpClient.SimpleQueryAppInstance(TestHttpServiceUrl, TestHttpInstanceInfo.AppName, TestHttpInstanceInfo.InstanceId)
    ast.Nilf(response.Error, "%v", response.Error)
    ast.Equal(200, response.StatusCode)
    ast.NotNil(response.Instance)
}

func TestHttpClient_QueryInstance(t *testing.T) {
    <-time.NewTimer(time.Second).C
    ast := assert.New(t)
    response := TestHttpClient.SimpleQueryInstance(TestHttpServiceUrl, TestHttpInstanceInfo.InstanceId)
    ast.Nilf(response.Error, "%v", response.Error)
    ast.Equal(200, response.StatusCode)
    ast.NotNil(response.Instance)
}

func TestHttpClient_ChangeStatus(t *testing.T) {
    <-time.NewTimer(time.Second).C
    ast := assert.New(t)
    response := TestHttpClient.SimpleChangeStatus(TestHttpServiceUrl, TestHttpInstanceInfo.AppName, TestHttpInstanceInfo.InstanceId, meta.StatusOutOfService)
    ast.Nilf(response.Error, "%v", response.Error)
    ast.Equal(200, response.StatusCode)
}

func TestHttpClient_ModifyMetadata(t *testing.T) {
    <-time.NewTimer(time.Second).C
    ast := assert.New(t)
    response := TestHttpClient.SimpleModifyMetadata(TestHttpServiceUrl, TestHttpInstanceInfo.AppName, TestHttpInstanceInfo.InstanceId, map[string]string{"hello": "world"})
    ast.Nilf(response.Error, "%v", response.Error)
    ast.Equal(200, response.StatusCode)
}

func TestHttpClient_QueryVipApps(t *testing.T) {
    <-time.NewTimer(time.Second).C
    ast := assert.New(t)
    response := TestHttpClient.SimpleQueryVipApps(TestHttpServiceUrl, TestHttpInstanceInfo.VipAddress)
    ast.Nilf(response.Error, "%v", response.Error)
    ast.Equal(200, response.StatusCode)
    ast.True(len(response.Apps) > 0)
}

func TestHttpClient_QuerySvipApps(t *testing.T) {
    <-time.NewTimer(time.Second).C
    ast := assert.New(t)
    response := TestHttpClient.SimpleQuerySvipApps(TestHttpServiceUrl, TestHttpInstanceInfo.SecureVipAddress)
    ast.Nilf(response.Error, "%v", response.Error)
    ast.Equal(200, response.StatusCode)
    ast.True(len(response.Apps) > 0)
}

func TestHttpClient_UnRegister(t *testing.T) {
    <-time.NewTimer(time.Second).C
    ast := assert.New(t)
    response := TestHttpClient.SimpleUnRegister(TestHttpServiceUrl, TestHttpInstanceInfo.AppName, TestHttpInstanceInfo.InstanceId)
    ast.Nilf(response.Error, "%v", response.Error)
    ast.Equal(200, response.StatusCode)
}
