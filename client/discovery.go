package client

import (
    "context"
    "errors"
    "fmt"
    "github.com/jiashunx/eureka-client-go/log"
    "github.com/jiashunx/eureka-client-go/meta"
    "math/rand"
    "time"
)

// discoveryClient eureka服务发现客户端
type discoveryClient struct {
    client *EurekaClient
    logger log.Logger
    Apps   map[string][]*meta.AppInfo // zone与服务列表映射
}

// start 启动eureka服务发现客户端
func (discovery *discoveryClient) start(ctx context.Context) *CommonResponse {
    go discovery.discovery(ctx)
    return &CommonResponse{Error: nil}
}

// discovery 具体服务发现处理逻辑
func (discovery *discoveryClient) discovery(ctx context.Context) {
    ticker := time.NewTicker(time.Duration(discovery.client.config.RegistryFetchIntervalSeconds) * time.Second)
FL:
    for {
        select {
        case <-ctx.Done():
            ticker.Stop()
            break FL
        default:
            if b, _ := discovery.isEnabled(); b {
                go discovery.discovery0(ctx)
            }
        }
        <-ticker.C
    }
}

// discovery 具体服务发现处理逻辑
func (discovery *discoveryClient) discovery0(ctx context.Context) (err error) {
    defer func() {
        if rc := recover(); rc != nil {
            err = errors.New(fmt.Sprintf("discoveryClient.discovery0, recover error: %v", rc))
        }
        if err != nil {
            discovery.logger.Tracef("discoveryClient.discovery0, FAILED >>> error: %v", err)
        }
    }()
    client := discovery.client
    var servers map[string]*meta.EurekaServer
    servers, err = client.config.GetAllZoneEurekaServers()
    if err != nil {
        return
    }
    c := make(chan map[string][]*meta.AppInfo)
    for zone, server := range servers {
        go func(zone string, server *meta.EurekaServer) {
            response := client.HttpClient().QueryApps(server)
            if response.Error != nil {
                c <- map[string][]*meta.AppInfo{zone: make([]*meta.AppInfo, 0)}
                return
            }
            for _, app := range response.Apps {
                app.Region = client.config.Region
                app.Zone = zone
                for _, instance := range app.Instances {
                    instance.Region = app.Region
                    instance.Zone = app.Zone
                }
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
    discovery.logger.Tracef("discoveryClient.discovery0, OK >>> apps: %v", SummaryAppsMap(apps))
    return nil
}

// isEnabled 服务发现功能是否开启
func (discovery *discoveryClient) isEnabled() (bool, error) {
    client := discovery.client
    if !*client.config.DiscoveryEnabled {
        return false, errors.New("eureka client's service discovery feature is not enabled")
    }
    return true, nil
}

// accessApp 查询可用服务
func (discovery *discoveryClient) accessApp(appName string) (app *meta.AppInfo, err error) {
    if _, err = discovery.isEnabled(); err != nil {
        return nil, err
    }
    config := discovery.client.config
    if *config.PreferSameZoneEureka {
        if apps, ok := discovery.Apps[config.Zone]; ok {
            app = FilterApp(apps, appName)
            instances := app.AvailableInstances()
            if instances != nil && len(instances) > 0 {
                return app.CopyWithInstances(instances), nil
            }
            app = nil
        }
    }
    anyMap := make(map[string]interface{})
    for k, v := range discovery.Apps {
        anyMap[k] = v
    }
    err = RandomLoopMap(anyMap, func(k string, v interface{}) (bool, error) {
        if v != nil {
            app = FilterApp(v.([]*meta.AppInfo), appName)
            instances := app.AvailableInstances()
            if instances != nil && len(instances) > 0 {
                app = app.CopyWithInstances(instances)
                return false, nil
            }
            app = nil
        }
        return true, nil
    })
    if err == nil && app == nil {
        err = errors.New("no available service found")
    }
    return app, err
}

// accessAppInstance 查询可用服务实例（随机选择）
func (discovery *discoveryClient) accessAppInstance(appName string) (*meta.InstanceInfo, error) {
    app, err := discovery.accessApp(appName)
    if err != nil {
        return nil, err
    }
    return app.Instances[rand.Intn(len(app.Instances))], nil
}

// accessAppsByVip 查询指定vip的可用服务列表
func (discovery *discoveryClient) accessAppsByVip(vip string) (vipApps []*meta.AppInfo, err error) {
    if _, err = discovery.isEnabled(); err != nil {
        return nil, err
    }
    config := discovery.client.config
    if *config.PreferSameZoneEureka {
        if apps, ok := discovery.Apps[config.Zone]; ok {
            vipApps = FilterAppsByVip(apps, vip)
            if vipApps != nil && len(vipApps) > 0 {
                accessApps := make([]*meta.AppInfo, 0)
                for _, vipApp := range vipApps {
                    instances := vipApp.AvailableInstances()
                    if instances != nil && len(instances) > 0 {
                        accessApps = append(accessApps, vipApp.CopyWithInstances(instances))
                    }
                }
                if len(accessApps) > 0 {
                    return accessApps, nil
                }
            }
            vipApps = nil
        }
    }
    anyMap := make(map[string]interface{})
    for k, v := range discovery.Apps {
        anyMap[k] = v
    }
    err = RandomLoopMap(anyMap, func(k string, v interface{}) (bool, error) {
        if v != nil {
            vipApps = FilterAppsByVip(v.([]*meta.AppInfo), vip)
            if vipApps != nil && len(vipApps) > 0 {
                accessApps := make([]*meta.AppInfo, 0)
                for _, vipApp := range vipApps {
                    instances := vipApp.AvailableInstances()
                    if instances != nil && len(instances) > 0 {
                        accessApps = append(accessApps, vipApp.CopyWithInstances(instances))
                    }
                }
                if len(accessApps) > 0 {
                    vipApps = accessApps
                    return false, nil
                }
            }
            vipApps = nil
        }
        return true, nil
    })
    if err == nil && (vipApps == nil || len(vipApps) == 0) {
        err = errors.New("no available service found")
    }
    return vipApps, err
}

// accessAppInstanceByVip 查询指定vip的可用服务实例（随机选择）
func (discovery *discoveryClient) accessAppInstanceByVip(vip string) (*meta.InstanceInfo, error) {
    apps, err := discovery.accessAppsByVip(vip)
    if err != nil {
        return nil, err
    }
    app := apps[rand.Intn(len(apps))]
    return app.Instances[rand.Intn(len(app.Instances))], nil
}

// accessAppsBySvip 查询指定svip的可用服务列表
func (discovery *discoveryClient) accessAppsBySvip(svip string) (svipApps []*meta.AppInfo, err error) {
    if _, err = discovery.isEnabled(); err != nil {
        return nil, err
    }
    config := discovery.client.config
    if *config.PreferSameZoneEureka {
        if apps, ok := discovery.Apps[config.Zone]; ok {
            svipApps = FilterAppsBySvip(apps, svip)
            if svipApps != nil && len(svipApps) > 0 {
                accessApps := make([]*meta.AppInfo, 0)
                for _, svipApp := range svipApps {
                    instances := svipApp.AvailableInstances()
                    if instances != nil && len(instances) > 0 {
                        accessApps = append(accessApps, svipApp.CopyWithInstances(instances))
                    }
                }
                if len(accessApps) > 0 {
                    return accessApps, nil
                }
            }
            svipApps = nil
        }
    }
    anyMap := make(map[string]interface{})
    for k, v := range discovery.Apps {
        anyMap[k] = v
    }
    err = RandomLoopMap(anyMap, func(k string, v interface{}) (bool, error) {
        if v != nil {
            svipApps = FilterAppsBySvip(v.([]*meta.AppInfo), svip)
            if svipApps != nil && len(svipApps) > 0 {
                accessApps := make([]*meta.AppInfo, 0)
                for _, svipApp := range svipApps {
                    instances := svipApp.AvailableInstances()
                    if instances != nil && len(instances) > 0 {
                        accessApps = append(accessApps, svipApp.CopyWithInstances(instances))
                    }
                }
                if len(accessApps) > 0 {
                    svipApps = accessApps
                    return false, nil
                }
            }
            svipApps = nil
        }
        return true, nil
    })
    if err == nil && (svipApps == nil || len(svipApps) == 0) {
        err = errors.New("no available service found")
    }
    return svipApps, err
}

// accessAppInstanceBySvip 查询指定svip的可用服务实例列表（随机选择）
func (discovery *discoveryClient) accessAppInstanceBySvip(svip string) (*meta.InstanceInfo, error) {
    apps, err := discovery.accessAppsBySvip(svip)
    if err != nil {
        return nil, err
    }
    app := apps[rand.Intn(len(apps))]
    return app.Instances[rand.Intn(len(app.Instances))], nil
}

// accessInstancesByVip 查询指定vip的可用服务实例列表
func (discovery *discoveryClient) accessInstancesByVip(vip string) (instances []*meta.InstanceInfo, err error) {
    if _, err = discovery.isEnabled(); err != nil {
        return nil, err
    }
    config := discovery.client.config
    if *config.PreferSameZoneEureka {
        if apps, ok := discovery.Apps[config.Zone]; ok {
            instances = FilterInstancesByVip(apps, vip)
            if instances != nil && len(instances) > 0 {
                tmpApp := &meta.AppInfo{Instances: instances}
                instances = tmpApp.AvailableInstances()
                if instances != nil && len(instances) > 0 {
                    return instances, nil
                }
            }
            instances = nil
        }
    }
    anyMap := make(map[string]interface{})
    for k, v := range discovery.Apps {
        anyMap[k] = v
    }
    err = RandomLoopMap(anyMap, func(k string, v interface{}) (bool, error) {
        if v != nil {
            instances = FilterInstancesByVip(v.([]*meta.AppInfo), vip)
            if instances != nil && len(instances) > 0 {
                tmpApp := &meta.AppInfo{Instances: instances}
                instances = tmpApp.AvailableInstances()
                if instances != nil && len(instances) > 0 {
                    return false, nil
                }
            }
            instances = nil
        }
        return true, nil
    })
    if err == nil && (instances == nil || len(instances) == 0) {
        err = errors.New("no available service instance found")
    }
    return instances, err
}

// accessInstanceByVip 查询指定vip的可用服务实例（随机选择）
func (discovery *discoveryClient) accessInstanceByVip(vip string) (*meta.InstanceInfo, error) {
    instances, err := discovery.accessInstancesByVip(vip)
    if err != nil {
        return nil, err
    }
    return instances[rand.Intn(len(instances))], nil
}

// accessInstancesBySvip 查询指定svip的可用服务实例列表
func (discovery *discoveryClient) accessInstancesBySvip(svip string) (instances []*meta.InstanceInfo, err error) {
    if _, err = discovery.isEnabled(); err != nil {
        return nil, err
    }
    config := discovery.client.config
    if *config.PreferSameZoneEureka {
        if apps, ok := discovery.Apps[config.Zone]; ok {
            instances = FilterInstancesBySvip(apps, svip)
            if instances != nil && len(instances) > 0 {
                tmpApp := &meta.AppInfo{Instances: instances}
                instances = tmpApp.AvailableInstances()
                if instances != nil && len(instances) > 0 {
                    return instances, nil
                }
            }
            instances = nil
        }
    }
    anyMap := make(map[string]interface{})
    for k, v := range discovery.Apps {
        anyMap[k] = v
    }
    err = RandomLoopMap(anyMap, func(k string, v interface{}) (bool, error) {
        if v != nil {
            instances = FilterInstancesBySvip(v.([]*meta.AppInfo), svip)
            if instances != nil && len(instances) > 0 {
                tmpApp := &meta.AppInfo{Instances: instances}
                instances = tmpApp.AvailableInstances()
                if instances != nil && len(instances) > 0 {
                    return false, nil
                }
            }
            instances = nil
        }
        return true, nil
    })
    if err == nil && (instances == nil || len(instances) == 0) {
        err = errors.New("no available service instance")
    }
    return instances, err
}

// accessInstanceBySvip 查询指定svip的可用服务实例（随机选择）
func (discovery *discoveryClient) accessInstanceBySvip(svip string) (*meta.InstanceInfo, error) {
    instances, err := discovery.accessInstancesBySvip(svip)
    if err != nil {
        return nil, err
    }
    return instances[rand.Intn(len(instances))], nil
}
