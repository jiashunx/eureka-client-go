package client

import (
    "github.com/google/uuid"
    "github.com/jiashunx/eureka-client-go/log"
    "github.com/jiashunx/eureka-client-go/meta"
    "github.com/stretchr/testify/assert"
    "testing"
    "time"
)

var serviceUrl = "http://admin:123123@127.0.0.1:20000/eureka,http://192.168.138.130:20000/eureka"

// TestEurekaClient1 客户端测试样例1(简单服务注册及服务发现客户端+手工更新服务实例状态、关闭服务注册与服务发现的客户端)
func TestEurekaClient1(t *testing.T) {
    ast := assert.New(t)

    // 创建客户端1
    client1, err := NewEurekaClient(&meta.EurekaConfig{
        InstanceConfig: &meta.InstanceConfig{
            AppName:       "eureka-client-test1",
            InstanceId:    "127.0.0.1:28081",
            NonSecurePort: 28081,
            Hostname:      "127.0.0.1",
        },
        ClientConfig: &meta.ClientConfig{
            ServiceUrlOfDefaultZone: serviceUrl,
        },
    })
    ast.Nilf(err, "%v", err)

    // 更新日志级别
    client1.logger.SetLevel(log.DebugLevel)

    // 启动客户端1
    err = client1.Start()
    ast.Nilf(err, "%v", err)

    <-time.NewTimer(time.Second).C

    // 客户端1默认注册时实例状态为STARTING，需手工修改状态为UP
    response1 := client1.ChangeStatus(meta.StatusUp)
    ast.Nilf(response1.Error, "%v", response1.Error)

    <-time.NewTimer(60 * time.Second).C

    // 停止客户端1，停止后客户端不可用，服务注册与发现相关goroutine自动停止并回收
    response1 = client1.Stop()
    ast.Nilf(response1.Error, "%v", response1.Error)

    <-time.NewTimer(time.Second).C

    // 停止客户端后可再次启动客户端（更新UUID以便于辨别输出的调试日志）
    client1.UUID = uuid.New().String()
    err = client1.Start()
    ast.Nilf(err, "%v", err)

    <-time.NewTimer(60 * time.Second).C

    // 创建客户端2（未开启服务注册与服务发现功能），该客户端无需手工关闭
    client2, err := NewEurekaClient(&meta.EurekaConfig{})
    ast.Nilf(err, "%v", err)

    // 更新日志级别
    client2.logger.SetLevel(log.DebugLevel)

    // 从客户端2获取与eureka server通讯的http客户端
    httpClient2 := client2.HttpClient()

    // 通过HttpClient与eureka server交互
    response2 := httpClient2.QueryApps(&meta.EurekaServer{ServiceUrl: serviceUrl})
    ast.Nilf(response2.Error, "%v", response2.Error)
    ast.True(len(response2.Apps) > 0)

    // 停止客户端1
    response1 = client1.Stop()
    ast.Nilf(response1.Error, "%v", response1.Error)
}

// TestEurekaClient2 客户端测试样例2(多zone服务注册与服务发现客户端+服务实例启动即可用)
func TestEurekaClient2(t *testing.T) {
    ast := assert.New(t)

    // 创建客户端
    client, err := NewEurekaClient(&meta.EurekaConfig{
        InstanceConfig: &meta.InstanceConfig{
            AppName:             "eureka-client-test2",
            InstanceId:          "127.0.0.1:28082",
            SecurePort:          28082,
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

    // 更新日志级别
    client.logger.SetLevel(log.DebugLevel)

    // 启动客户端（指定了 InstanceEnabledOnIt 参数，默认注册时服务实例状态为UP）
    err = client.Start()
    ast.Nilf(err, "%v", err)

    <-time.NewTimer(60 * time.Second).C

    // 服务发现
    app, err := client.AccessApp(client.config.AppName)
    ast.Nilf(err, "%v", err)
    ast.NotNil(app)

    // 停止客户端，停止后客户端不可用，服务注册与发现相关goroutine自动停止并回收
    response := client.Stop()
    ast.Nilf(response.Error, "%v", response.Error)
}
