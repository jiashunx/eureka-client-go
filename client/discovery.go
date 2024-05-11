package client

import (
    "errors"
    "fmt"
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
    var app *meta.AppInfo
    config := discovery.client.config
    if *config.PreferSameZoneEureka {
        if apps, ok := discovery.Apps[config.Zone]; ok {
            app = GetApp(apps, appName)
            if app != nil {
                return app, nil
            }
        }
    }
    anyMap := make(map[string]interface{})
    for k, v := range discovery.Apps {
        anyMap[k] = v
    }
    err := RandomLoopMap(anyMap, func(k string, v interface{}) (bool, error) {
        if v != nil {
            app = GetApp(v.([]*meta.AppInfo), appName)
            if app != nil {
                return false, nil
            }
        }
        return true, nil
    })
    if err == nil && app == nil {
        err = errors.New(fmt.Sprintf("no service found, appName: %s", appName))
    }
    return app, err
}

// getAppInstance 查询服务实例信息
func (discovery *discoveryClient) getAppInstance(appName, instanceId string) (*meta.InstanceInfo, error) {
    if _, err := discovery.isEnabled(); err != nil {
        return nil, err
    }
    var instance *meta.InstanceInfo
    config := discovery.client.config
    if *config.PreferSameZoneEureka {
        if apps, ok := discovery.Apps[config.Zone]; ok {
            instance = GetAppInstance(apps, appName, instanceId)
            if instance != nil {
                return instance, nil
            }
        }
    }
    anyMap := make(map[string]interface{})
    for k, v := range discovery.Apps {
        anyMap[k] = v
    }
    err := RandomLoopMap(anyMap, func(k string, v interface{}) (bool, error) {
        if v != nil {
            instance = GetAppInstance(v.([]*meta.AppInfo), appName, instanceId)
            if instance != nil {
                return false, nil
            }
        }
        return true, nil
    })
    if err == nil && instance == nil {
        err = errors.New(fmt.Sprintf("no service instance found, appName: %s, instanceId: %s", appName, instanceId))
    }
    return instance, err
}

// getInstance 查询服务实例信息
func (discovery *discoveryClient) getInstance(instanceId string) (*meta.InstanceInfo, error) {
    if _, err := discovery.isEnabled(); err != nil {
        return nil, err
    }
    var instance *meta.InstanceInfo
    config := discovery.client.config
    if *config.PreferSameZoneEureka {
        if apps, ok := discovery.Apps[config.Zone]; ok {
            instance = GetInstance(apps, instanceId)
            if instance != nil {
                return instance, nil
            }
        }
    }
    anyMap := make(map[string]interface{})
    for k, v := range discovery.Apps {
        anyMap[k] = v
    }
    err := RandomLoopMap(anyMap, func(k string, v interface{}) (bool, error) {
        if v != nil {
            instance = GetInstance(v.([]*meta.AppInfo), instanceId)
            if instance != nil {
                return false, nil
            }
        }
        return true, nil
    })
    if err == nil && instance == nil {
        err = errors.New(fmt.Sprintf("no service instance found, instanceId: %s", instanceId))
    }
    return instance, err
}

// getAppsByVip 查询指定vip的服务列表
func (discovery *discoveryClient) getAppsByVip(vip string) ([]*meta.AppInfo, error) {
    if _, err := discovery.isEnabled(); err != nil {
        return nil, err
    }
    var vipApps []*meta.AppInfo
    config := discovery.client.config
    if *config.PreferSameZoneEureka {
        if apps, ok := discovery.Apps[config.Zone]; ok {
            vipApps = GetAppsByVip(apps, vip)
            if vipApps != nil && len(vipApps) > 0 {
                return vipApps, nil
            }
        }
    }
    anyMap := make(map[string]interface{})
    for k, v := range discovery.Apps {
        anyMap[k] = v
    }
    err := RandomLoopMap(anyMap, func(k string, v interface{}) (bool, error) {
        if v != nil {
            vipApps = GetAppsByVip(v.([]*meta.AppInfo), vip)
            if vipApps != nil && len(vipApps) > 0 {
                return false, nil
            }
        }
        return true, nil
    })
    if err == nil && (vipApps == nil || len(vipApps) == 0) {
        err = errors.New(fmt.Sprintf("no service found, vip: %s", vip))
    }
    return vipApps, err
}

// getAppsBySvip 查询指定svip的服务列表
func (discovery *discoveryClient) getAppsBySvip(svip string) ([]*meta.AppInfo, error) {
    if _, err := discovery.isEnabled(); err != nil {
        return nil, err
    }
    var svipApps []*meta.AppInfo
    config := discovery.client.config
    if *config.PreferSameZoneEureka {
        if apps, ok := discovery.Apps[config.Zone]; ok {
            svipApps = GetAppsBySvip(apps, svip)
            if svipApps != nil && len(svipApps) > 0 {
                return svipApps, nil
            }
        }
    }
    anyMap := make(map[string]interface{})
    for k, v := range discovery.Apps {
        anyMap[k] = v
    }
    err := RandomLoopMap(anyMap, func(k string, v interface{}) (bool, error) {
        if v != nil {
            svipApps = GetAppsBySvip(v.([]*meta.AppInfo), svip)
            if svipApps != nil && len(svipApps) > 0 {
                return false, nil
            }
        }
        return true, nil
    })
    if err == nil && (svipApps == nil || len(svipApps) == 0) {
        err = errors.New(fmt.Sprintf("no service found, svip: %s", svip))
    }
    return svipApps, err
}

// getInstancesByVip 查询指定vip的服务实例列表
func (discovery *discoveryClient) getInstancesByVip(vip string) ([]*meta.InstanceInfo, error) {
    if _, err := discovery.isEnabled(); err != nil {
        return nil, err
    }
    var instances []*meta.InstanceInfo
    config := discovery.client.config
    if *config.PreferSameZoneEureka {
        if apps, ok := discovery.Apps[config.Zone]; ok {
            instances = GetInstancesByVip(apps, vip)
            if instances != nil && len(instances) > 0 {
                return instances, nil
            }
        }
    }
    anyMap := make(map[string]interface{})
    for k, v := range discovery.Apps {
        anyMap[k] = v
    }
    err := RandomLoopMap(anyMap, func(k string, v interface{}) (bool, error) {
        if v != nil {
            instances = GetInstancesByVip(v.([]*meta.AppInfo), vip)
            if instances != nil && len(instances) > 0 {
                return false, nil
            }
        }
        return true, nil
    })
    if err == nil && (instances == nil || len(instances) == 0) {
        err = errors.New(fmt.Sprintf("no service instance found, vip: %s", vip))
    }
    return instances, err
}

// getInstancesBySvip 查询指定svip的服务实例列表
func (discovery *discoveryClient) getInstancesBySvip(svip string) ([]*meta.InstanceInfo, error) {
    if _, err := discovery.isEnabled(); err != nil {
        return nil, err
    }
    var instances []*meta.InstanceInfo
    config := discovery.client.config
    if *config.PreferSameZoneEureka {
        if apps, ok := discovery.Apps[config.Zone]; ok {
            instances = GetInstancesBySvip(apps, svip)
            if instances != nil && len(instances) > 0 {
                return instances, nil
            }
        }
    }
    anyMap := make(map[string]interface{})
    for k, v := range discovery.Apps {
        anyMap[k] = v
    }
    err := RandomLoopMap(anyMap, func(k string, v interface{}) (bool, error) {
        if v != nil {
            instances = GetInstancesBySvip(v.([]*meta.AppInfo), svip)
            if instances != nil && len(instances) > 0 {
                return false, nil
            }
        }
        return true, nil
    })
    if err == nil && (instances == nil || len(instances) == 0) {
        err = errors.New(fmt.Sprintf("no service instance found, svip: %s", svip))
    }
    return instances, err
}
