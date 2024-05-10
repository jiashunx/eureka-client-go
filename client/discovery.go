package client

import (
    "context"
    "errors"
    "github.com/jiashunx/eureka-client-go/http"
    "github.com/jiashunx/eureka-client-go/meta"
    "time"
)

// DiscoveryClient eureka服务发现客户端
type DiscoveryClient struct {
    config *meta.EurekaConfig
    ctx    context.Context
    Apps   map[string][]*meta.AppInfo // zone与服务列表映射
}

// start 启动eureka服务发现客户端
func (client *DiscoveryClient) start() {
    go client.discovery0()
    go client.discovery()
}

// discovery 具体服务发现处理逻辑
func (client *DiscoveryClient) discovery() {
    ticker := time.NewTicker(time.Duration(client.config.RegistryFetchIntervalSeconds) * time.Second)
FL:
    for {
        <-ticker.C
        select {
        case <-client.ctx.Done():
            ticker.Stop()
            break FL
        default:
            if b, _ := client.isEnabled(); b {
                go client.discovery0()
            }
        }
    }
}

// discovery 具体服务发现处理逻辑
func (client *DiscoveryClient) discovery0() {
    servers, err := client.config.GetAllZoneEurekaServers()
    if err != nil {
        return
    }
    c := make(chan map[string][]*meta.AppInfo)
    for zone, server := range servers {
        go func(zone string, server *meta.EurekaServer) {
            response := http.QueryApps(server)
            if response.Error != nil {
                c <- map[string][]*meta.AppInfo{}
                return
            }
            c <- map[string][]*meta.AppInfo{zone: response.Apps}
        }(zone, server)
    }
    apps := make(map[string][]*meta.AppInfo)
    for i, size := 0, len(servers); i < size; i++ {
        for key, value := range <-c {
            apps[key] = value
        }
    }
    close(c)
    client.Apps = apps
}

// isEnabled 服务发现功能是否开启
func (client *DiscoveryClient) isEnabled() (bool, error) {
    if !*client.config.DiscoveryEnabled {
        return false, errors.New("service discovery feature is not enabled")
    }
    return true, nil
}
