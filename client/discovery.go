package client

import (
    "errors"
    "github.com/jiashunx/eureka-client-go/meta"
    "time"
)

// discoveryClient eureka服务发现客户端
type discoveryClient struct {
    client *EurekaClient
    Apps   map[string][]*meta.AppInfo // zone与服务列表映射
}

// start 启动eureka服务发现客户端
func (discovery *discoveryClient) start() {
    go discovery.discovery0()
    go discovery.discovery()
}

// discovery 具体服务发现处理逻辑
func (discovery *discoveryClient) discovery() {
    client := discovery.client
    ticker := time.NewTicker(time.Duration(client.config.RegistryFetchIntervalSeconds) * time.Second)
FL:
    for {
        <-ticker.C
        select {
        case <-client.ctx.Done():
            ticker.Stop()
            break FL
        default:
            if b, _ := discovery.isEnabled(); b {
                go discovery.discovery0()
            }
        }
    }
}

// discovery 具体服务发现处理逻辑
func (discovery *discoveryClient) discovery0() {
    client := discovery.client
    servers, err := client.config.GetAllZoneEurekaServers()
    if err != nil {
        return
    }
    c := make(chan map[string][]*meta.AppInfo)
    for zone, server := range servers {
        go func(zone string, server *meta.EurekaServer) {
            response := client.HttpClient.QueryApps(server)
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
    discovery.Apps = apps
}

// isEnabled 服务发现功能是否开启
func (discovery *discoveryClient) isEnabled() (bool, error) {
    client := discovery.client
    if !*client.config.DiscoveryEnabled {
        return false, errors.New("eureka client's service discovery feature is not enabled")
    }
    return true, nil
}

// getApp 查询服务信息
func (discovery *discoveryClient) getApp(appName string) (*meta.AppInfo, error) {
    if _, err := discovery.isEnabled(); err != nil {
        return nil, err
    }
    for _, apps := range discovery.Apps {
        return GetApp(apps, appName), nil
    }
    return nil, nil
}

// getAppInstance 查询服务实例信息
func (discovery *discoveryClient) getAppInstance(appName, instanceId string) (*meta.InstanceInfo, error) {
    if _, err := discovery.isEnabled(); err != nil {
        return nil, err
    }
    for _, apps := range discovery.Apps {
        return GetAppInstance(apps, appName, instanceId), nil
    }
    return nil, nil
}

// getInstance 查询服务实例信息
func (discovery *discoveryClient) getInstance(instanceId string) (*meta.InstanceInfo, error) {
    if _, err := discovery.isEnabled(); err != nil {
        return nil, err
    }
    for _, apps := range discovery.Apps {
        return GetInstance(apps, instanceId), nil
    }
    return nil, nil
}

// getAppsByVip 查询有相同vip的服务信息列表
func (discovery *discoveryClient) getAppsByVip(vip string) ([]*meta.AppInfo, error) {
    if _, err := discovery.isEnabled(); err != nil {
        return nil, err
    }
    for _, apps := range discovery.Apps {
        return GetAppsByVip(apps, vip), nil
    }
    return nil, nil
}

// getAppsBySvip 查询有相同svip的服务信息列表
func (discovery *discoveryClient) getAppsBySvip(svip string) ([]*meta.AppInfo, error) {
    if _, err := discovery.isEnabled(); err != nil {
        return nil, err
    }
    for _, apps := range discovery.Apps {
        return GetAppsBySvip(apps, svip), nil
    }
    return nil, nil
}
