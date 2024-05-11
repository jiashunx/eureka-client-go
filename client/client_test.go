package client

import (
    "github.com/jiashunx/eureka-client-go/meta"
    "github.com/stretchr/testify/assert"
    "testing"
    "time"
)

var serviceUrl = "http://admin:123123@127.0.0.1:20000/eureka,http://192.168.138.130:20000/eureka"

// TestEurekaClient1 客户端测试样例1
func TestEurekaClient1(t *testing.T) {
    ast := assert.New(t)

    // 创建客户端
    client, err := NewEurekaClient(&meta.EurekaConfig{
        InstanceConfig: &meta.InstanceConfig{
            AppName:       "eureka-client-test1",
            InstanceId:    "127.0.0.1:8080",
            NonSecurePort: 8081,
            Hostname:      "127.0.0.1",
        },
        ClientConfig: &meta.ClientConfig{
            ServiceUrlOfDefaultZone: serviceUrl,
        },
    })
    ast.Nilf(err, "%v", err)

    // 启动客户端
    err = client.Start()
    ast.Nilf(err, "%v", err)

    <-time.NewTimer(time.Second).C

    // client默认注册时实例状态为STARTING，需手工修改状态为UP
    response := client.ChangeStatus(meta.StatusUp)
    ast.Nilf(response.Error, "%v", response.Error)

    <-time.NewTimer(time.Second).C

    // 停止客户端，停止后客户端不可用，服务注册与发现相关goroutine自动停止并回收
    response = client.Stop()
    ast.Nilf(response.Error, "%v", response.Error)

    <-time.NewTimer(time.Second).C

    // 停止客户端后可再次启动客户端
    err = client.Start()
    ast.Nilf(err, "%v", err)

    <-time.NewTimer(time.Second).C

    // 停止客户端
    response = client.Stop()
    ast.Nilf(response.Error, "%v", response.Error)
}

// TestEurekaClient2 客户端测试样例2
func TestEurekaClient2(t *testing.T) {
    ast := assert.New(t)

    // 创建客户端21
    client21, err := NewEurekaClient(&meta.EurekaConfig{
        InstanceConfig: &meta.InstanceConfig{
            AppName:             "eureka-client-test2",
            InstanceId:          "127.0.0.1:8081",
            NonSecurePort:       8081,
            Hostname:            "127.0.0.1",
            InstanceEnabledOnIt: &meta.True,
        },
        ClientConfig: &meta.ClientConfig{
            ServiceUrlOfDefaultZone: serviceUrl,
        },
    })
    ast.Nilf(err, "%v", err)

    // 创建客户端22
    client22, err := NewEurekaClient(&meta.EurekaConfig{
        InstanceConfig: &meta.InstanceConfig{
            AppName:             "eureka-client-test2",
            InstanceId:          "127.0.0.1:8082",
            SecurePort:          8082,
            IpAddress:           "127.0.0.1",
            PreferIpAddress:     &meta.True,
            InstanceEnabledOnIt: &meta.True,
        },
        ClientConfig: &meta.ClientConfig{
            PreferSameZoneEureka: &meta.False,
            Region:               "cn",
            Zone:                 "zone1",
            AvailableZones: map[string]string{
                "cn":  "zone1,zone2",
                "usa": "zone3",
                "uk":  "zone4,zone5,zone6",
            },
            ServiceUrlOfAllZone: map[string]string{
                "zone1": serviceUrl,
                "zone2": "",
            },
        },
    })
    ast.Nilf(err, "%v", err)

    // 启动客户端（指定了 InstanceEnabledOnIt 参数，默认注册时服务实例状态为UP）
    err = client21.Start()
    ast.Nilf(err, "%v", err)
    err = client22.Start()
    ast.Nilf(err, "%v", err)

    <-time.NewTimer(60 * time.Second).C

    // 服务发现
    app, err := client21.AccessApp(client21.config.AppName)
    ast.Nilf(err, "%v", err)
    ast.NotNil(app)

    // 停止客户端，停止后客户端不可用，服务注册与发现相关goroutine自动停止并回收
    response := client21.Stop()
    ast.Nilf(response.Error, "%v", response.Error)
    response = client22.Stop()
    ast.Nilf(response.Error, "%v", response.Error)
}

// TestEurekaClient3 客户端测试样例3
func TestEurekaClient3(t *testing.T) {
    ast := assert.New(t)

    // 创建客户端（未开启服务注册与服务发现功能）
    client, err := NewEurekaClient(&meta.EurekaConfig{})
    ast.Nilf(err, "%v", err)

    // 获取与eureka server通讯的http客户端
    httpClient := client.HttpClient

    // 通过HttpClient与eureka server交互
    response := httpClient.QueryApps(&meta.EurekaServer{ServiceUrl: serviceUrl})
    ast.Nilf(response.Error, "%v", response.Error)
    ast.True(len(response.Apps) > 0)

    // 客户端无需关闭
    // client.Stop()
}
