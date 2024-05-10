package client

import (
    "fmt"
    "github.com/jiashunx/eureka-client-go/meta"
    "github.com/stretchr/testify/assert"
    "runtime"
    "testing"
    "time"
)

var serviceUrl = "http://127.0.0.1:20000/eureka,http://192.168.138.130:20000/eureka"

func TestEurekaClient(t *testing.T) {
    ast := assert.New(t)
    client, err := NewEurekaClient(nil, &meta.ClientConfig{
        ServiceUrlOfDefaultZone: serviceUrl,
    })
    ast.Nilf(err, "创建EurekaClient失败，失败原因：%v", err)
    fmt.Println("BeforeStart:NumGoroutine:", runtime.NumGoroutine())
    err = client.Start()
    ast.Nilf(err, "启动EurekaClient失败，失败原因：%v", err)
    ticker := time.NewTicker(time.Second)
    for i := 0; i < 120; i++ {
        fmt.Println("Started:NumGoroutine:", runtime.NumGoroutine())
        <-ticker.C
    }
    fmt.Println("BeforeStop:NumGoroutine:", runtime.NumGoroutine())
    client.Stop()
    ticker.Reset(time.Second * 10)
    for i := 0; i < 120; i++ {
        fmt.Println("Stopped:NumGoroutine:", runtime.NumGoroutine())
        <-ticker.C
    }
}
