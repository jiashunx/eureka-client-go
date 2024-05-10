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
    client, err := NewEurekaClient(&meta.EurekaConfig{
        InstanceConfig: &meta.InstanceConfig{
            AppName:       "test-eureka-client",
            InstanceId:    "127.0.0.1:8081",
            NonSecurePort: 8081,
            Hostname:      "127.0.0.1",
        },
        ClientConfig: &meta.ClientConfig{
            ServiceUrlOfDefaultZone: serviceUrl,
        },
    })
    ast.Nilf(err, "创建EurekaClient失败，失败原因：%v", err)
    // 启动客户端
    err = client.Start()
    ast.Nilf(err, "启动EurekaClient失败，失败原因：%v", err)
    // client默认注册时实例状态为STARTING，需手工修改状态为UP
    response := client.ChangeStatus(meta.StatusUp)
    ast.Nilf(response.Error, "更新服务实例状态为UP，请求失败，失败原因：%v", response.Error)
    // 停止客户端，停止后客户端不可用，服务注册与发现相关goroutine自动停止并回收
    client.Stop()
    <-time.NewTimer(2 * time.Second).C

    // 停止客户端后可再次启动客户端
    err = client.Start()
    ast.Nilf(err, "再次启动EurekaClient失败，失败原因：%v", err)
    // 停止客户端
    client.Stop()
}

// TestEurekaClient2 客户端测试样例2
func TestEurekaClient2(t *testing.T) {
    ast := assert.New(t)
    client, err := NewEurekaClient(&meta.EurekaConfig{
        InstanceConfig: &meta.InstanceConfig{
            AppName:             "test-eureka-client",
            InstanceId:          "127.0.0.1:8082",
            SecurePort:          8082,
            IpAddress:           "127.0.0.1",
            PreferIpAddress:     &meta.True,
            InstanceEnabledOnIt: &meta.True,
        },
        ClientConfig: &meta.ClientConfig{
            Zone:           "zone2",
            AvailableZones: "zone1,zone2",
            ServiceUrlOfAllZone: map[string]string{
                "zone1": serviceUrl,
                "zone2": serviceUrl,
            },
        },
    })
    ast.Nilf(err, "创建EurekaClient失败，失败原因：%v", err)
    // 启动客户端（指定了 InstanceEnabledOnIt 参数，默认注册时服务实例状态为UP）
    err = client.Start()
    ast.Nilf(err, "启动EurekaClient失败，失败原因：%v", err)
    // 停止客户端，停止后客户端不可用，服务注册与发现相关goroutine自动停止并回收
    client.Stop()
    <-time.NewTimer(2 * time.Second).C
}
