package client

import (
    "context"
    "errors"
    "fmt"
    "github.com/jiashunx/eureka-client-go/log"
    "github.com/jiashunx/eureka-client-go/meta"
    "math/rand"
    "strconv"
    "strings"
    "time"
)

// DiscoveryClient eureka服务发现客户端
type DiscoveryClient struct {
    HttpClient *HttpClient
    Config     *meta.EurekaConfig
    Logger     log.Logger
    // zone与服务列表映射
    Apps map[string][]*meta.AppInfo
}

// start 启动eureka服务发现客户端
func (discovery *DiscoveryClient) start(ctx context.Context) *CommonResponse {
    go discovery.discovery(ctx)
    return &CommonResponse{Error: nil}
}

// discovery 具体服务发现处理逻辑
func (discovery *DiscoveryClient) discovery(ctx context.Context) {
    ticker := time.NewTicker(time.Duration(discovery.Config.RegistryFetchIntervalSeconds) * time.Second)
FL:
    for {
        select {
        case <-ctx.Done():
            ticker.Stop()
            break FL
        default:
            if b, _ := discovery.isEnabled(); b {
                go discovery.Discovery0()
            }
        }
        <-ticker.C
    }
}

// Discovery0 具体服务发现处理逻辑
func (discovery *DiscoveryClient) Discovery0() (Apps map[string][]*meta.AppInfo, err error) {
    defer func() {
        if rc := recover(); rc != nil {
            err = errors.New(fmt.Sprintf("DiscoveryClient.Discovery0, recover error: %v", rc))
        }
        if err != nil {
            discovery.Logger.Tracef("DiscoveryClient.Discovery0, FAILED >>> error: %v", err)
        }
    }()
    var servers map[string]*meta.EurekaServer
    servers, err = discovery.Config.GetAllZoneEurekaServers()
    if err != nil {
        return
    }
    c := make(chan map[string][]*meta.AppInfo)
    for zone, server := range servers {
        go func(zone string, server *meta.EurekaServer) {
            response := discovery.HttpClient.QueryApps(server)
            if response.Error != nil {
                c <- map[string][]*meta.AppInfo{zone: make([]*meta.AppInfo, 0)}
                return
            }
            for _, app := range response.Apps {
                app.Region = discovery.Config.Region
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
    discovery.Logger.Tracef("DiscoveryClient.Discovery0, OK >>> apps: %v", SummaryAppsMap(apps))
    return apps, nil
}

// isEnabled 服务发现功能是否开启
func (discovery *DiscoveryClient) isEnabled() (bool, error) {
    if !*discovery.Config.DiscoveryEnabled {
        return false, errors.New("eureka client's service discovery feature is not enabled")
    }
    return true, nil
}

// publicQuery 对外查询api公共检查处理
func (discovery *DiscoveryClient) publicQuery(name string, r func(params ...any) (any, error), params ...any) (ret any, err error) {
    defer func() {
        if rc := recover(); rc != nil {
            err = errors.New(fmt.Sprintf("DiscoveryClient.%s, recover error: %v", name, rc))
        }
        if err != nil {
            discovery.Logger.Errorf("DiscoveryClient.%s, FAILED >>> error: %v", name, err)
        }
        if err == nil {
            discovery.Logger.Tracef("DiscoveryClient.%s, OK >>> ret: %v", name, ret)
        }
    }()
    if len(params) > 0 {
        sp := make([]any, 0)
        sp = append(sp, name)
        sl := make([]string, 0)
        for idx, param := range params {
            sl = append(sl, "arg"+strconv.Itoa(idx)+": %v")
            sp = append(sp, param)
        }
        discovery.Logger.Tracef("DiscoveryClient.%s, PARAMS >>> "+strings.Join(sl, ", "), sp...)
    }
    if _, err = discovery.isEnabled(); err != nil {
        return nil, err
    }
    if _, err = discovery.Discovery0(); err != nil {
        return nil, err
    }
    return r(params...)
}

// AccessApp 查询可用服务
func (discovery *DiscoveryClient) AccessApp(appName string) (*meta.AppInfo, error) {
    ret, err := discovery.publicQuery("AccessApp", func(params ...any) (any, error) {
        return discovery.FilterApp(discovery.Apps, params[0].(string))
    }, appName)
    if err != nil {
        return nil, err
    }
    return ret.(*meta.AppInfo), nil
}

// AccessAppInstance 查询可用服务实例（随机选择）
func (discovery *DiscoveryClient) AccessAppInstance(appName string) (*meta.InstanceInfo, error) {
    ret, err := discovery.publicQuery("AccessAppInstance", func(params ...any) (any, error) {
        return discovery.FilterAppInstance(discovery.Apps, params[0].(string))
    }, appName)
    if err != nil {
        return nil, err
    }
    return ret.(*meta.InstanceInfo), nil
}

// AccessAppsByVip 查询指定vip的可用服务列表
func (discovery *DiscoveryClient) AccessAppsByVip(vip string) (vipApps []*meta.AppInfo, err error) {
    ret, err := discovery.publicQuery("AccessAppsByVip", func(params ...any) (any, error) {
        return discovery.FilterAppsByVip(discovery.Apps, params[0].(string))
    }, vip)
    if err != nil {
        return nil, err
    }
    return ret.([]*meta.AppInfo), nil
}

// AccessAppInstanceByVip 查询指定vip的可用服务实例（随机选择）
func (discovery *DiscoveryClient) AccessAppInstanceByVip(vip string) (*meta.InstanceInfo, error) {
    ret, err := discovery.publicQuery("AccessAppInstanceByVip", func(params ...any) (any, error) {
        return discovery.FilterAppInstanceByVip(discovery.Apps, params[0].(string))
    }, vip)
    if err != nil {
        return nil, err
    }
    return ret.(*meta.InstanceInfo), nil
}

// AccessAppsBySvip 查询指定svip的可用服务列表
func (discovery *DiscoveryClient) AccessAppsBySvip(svip string) (svipApps []*meta.AppInfo, err error) {
    ret, err := discovery.publicQuery("AccessAppsBySvip", func(params ...any) (any, error) {
        return discovery.FilterAppsBySvip(discovery.Apps, params[0].(string))
    }, svip)
    if err != nil {
        return nil, err
    }
    return ret.([]*meta.AppInfo), nil
}

// AccessAppInstanceBySvip 查询指定svip的可用服务实例列表（随机选择）
func (discovery *DiscoveryClient) AccessAppInstanceBySvip(svip string) (*meta.InstanceInfo, error) {
    ret, err := discovery.publicQuery("AccessAppInstanceBySvip", func(params ...any) (any, error) {
        return discovery.FilterAppInstanceBySvip(discovery.Apps, params[0].(string))
    }, svip)
    if err != nil {
        return nil, err
    }
    return ret.(*meta.InstanceInfo), nil
}

// AccessInstancesByVip 查询指定vip的可用服务实例列表
func (discovery *DiscoveryClient) AccessInstancesByVip(vip string) (instances []*meta.InstanceInfo, err error) {
    ret, err := discovery.publicQuery("AccessInstancesByVip", func(params ...any) (any, error) {
        return discovery.FilterInstancesByVip(discovery.Apps, params[0].(string))
    }, vip)
    if err != nil {
        return nil, err
    }
    return ret.([]*meta.InstanceInfo), nil
}

// AccessInstanceByVip 查询指定vip的可用服务实例（随机选择）
func (discovery *DiscoveryClient) AccessInstanceByVip(vip string) (*meta.InstanceInfo, error) {
    ret, err := discovery.publicQuery("AccessInstanceByVip", func(params ...any) (any, error) {
        return discovery.FilterInstanceByVip(discovery.Apps, params[0].(string))
    }, vip)
    if err != nil {
        return nil, err
    }
    return ret.(*meta.InstanceInfo), nil
}

// AccessInstancesBySvip 查询指定svip的可用服务实例列表
func (discovery *DiscoveryClient) AccessInstancesBySvip(svip string) (instances []*meta.InstanceInfo, err error) {
    ret, err := discovery.publicQuery("AccessInstancesBySvip", func(params ...any) (any, error) {
        return discovery.FilterInstancesBySvip(discovery.Apps, params[0].(string))
    }, svip)
    if err != nil {
        return nil, err
    }
    return ret.([]*meta.InstanceInfo), nil
}

// AccessInstanceBySvip 查询指定svip的可用服务实例（随机选择）
func (discovery *DiscoveryClient) AccessInstanceBySvip(svip string) (*meta.InstanceInfo, error) {
    ret, err := discovery.publicQuery("AccessInstanceBySvip", func(params ...any) (any, error) {
        return discovery.FilterInstanceBySvip(discovery.Apps, params[0].(string))
    }, svip)
    if err != nil {
        return nil, err
    }
    return ret.(*meta.InstanceInfo), nil
}

// FilterApp 查询可用服务
func (discovery *DiscoveryClient) FilterApp(Apps map[string][]*meta.AppInfo, appName string) (app *meta.AppInfo, err error) {
    if _, err = discovery.isEnabled(); err != nil {
        return nil, err
    }
    Config := discovery.Config
    if *Config.PreferSameZoneEureka && Apps != nil {
        if apps, ok := Apps[Config.Zone]; ok {
            app = FilterApp(apps, appName)
            instances := app.AvailableInstances()
            if instances != nil && len(instances) > 0 {
                return app.CopyWithInstances(instances), nil
            }
            app = nil
        }
    }
    anyMap := make(map[string]interface{})
    if Apps != nil {
        for k, v := range Apps {
            anyMap[k] = v
        }
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

// FilterAppInstance 查询可用服务实例（随机选择）
func (discovery *DiscoveryClient) FilterAppInstance(Apps map[string][]*meta.AppInfo, appName string) (*meta.InstanceInfo, error) {
    app, err := discovery.FilterApp(Apps, appName)
    if err != nil {
        return nil, err
    }
    return app.Instances[rand.Intn(len(app.Instances))], nil
}

// FilterAppsByVip 查询指定vip的可用服务列表
func (discovery *DiscoveryClient) FilterAppsByVip(Apps map[string][]*meta.AppInfo, vip string) (vipApps []*meta.AppInfo, err error) {
    if _, err = discovery.isEnabled(); err != nil {
        return nil, err
    }
    Config := discovery.Config
    if *Config.PreferSameZoneEureka && Apps != nil {
        if apps, ok := Apps[Config.Zone]; ok {
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
    if Apps != nil {
        for k, v := range Apps {
            anyMap[k] = v
        }
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

// FilterAppInstanceByVip 查询指定vip的可用服务实例（随机选择）
func (discovery *DiscoveryClient) FilterAppInstanceByVip(Apps map[string][]*meta.AppInfo, vip string) (*meta.InstanceInfo, error) {
    apps, err := discovery.FilterAppsByVip(Apps, vip)
    if err != nil {
        return nil, err
    }
    app := apps[rand.Intn(len(apps))]
    return app.Instances[rand.Intn(len(app.Instances))], nil
}

// FilterAppsBySvip 查询指定svip的可用服务列表
func (discovery *DiscoveryClient) FilterAppsBySvip(Apps map[string][]*meta.AppInfo, svip string) (svipApps []*meta.AppInfo, err error) {
    if _, err = discovery.isEnabled(); err != nil {
        return nil, err
    }
    Config := discovery.Config
    if *Config.PreferSameZoneEureka && Apps != nil {
        if apps, ok := Apps[Config.Zone]; ok {
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
    if Apps != nil {
        for k, v := range Apps {
            anyMap[k] = v
        }
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

// FilterAppInstanceBySvip 查询指定svip的可用服务实例列表（随机选择）
func (discovery *DiscoveryClient) FilterAppInstanceBySvip(Apps map[string][]*meta.AppInfo, svip string) (*meta.InstanceInfo, error) {
    apps, err := discovery.FilterAppsBySvip(Apps, svip)
    if err != nil {
        return nil, err
    }
    app := apps[rand.Intn(len(apps))]
    return app.Instances[rand.Intn(len(app.Instances))], nil
}

// FilterInstancesByVip 查询指定vip的可用服务实例列表
func (discovery *DiscoveryClient) FilterInstancesByVip(Apps map[string][]*meta.AppInfo, vip string) (instances []*meta.InstanceInfo, err error) {
    if _, err = discovery.isEnabled(); err != nil {
        return nil, err
    }
    Config := discovery.Config
    if *Config.PreferSameZoneEureka && Apps != nil {
        if apps, ok := Apps[Config.Zone]; ok {
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
    if Apps != nil {
        for k, v := range Apps {
            anyMap[k] = v
        }
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

// FilterInstanceByVip 查询指定vip的可用服务实例（随机选择）
func (discovery *DiscoveryClient) FilterInstanceByVip(Apps map[string][]*meta.AppInfo, vip string) (*meta.InstanceInfo, error) {
    instances, err := discovery.FilterInstancesByVip(Apps, vip)
    if err != nil {
        return nil, err
    }
    return instances[rand.Intn(len(instances))], nil
}

// FilterInstancesBySvip 查询指定svip的可用服务实例列表
func (discovery *DiscoveryClient) FilterInstancesBySvip(Apps map[string][]*meta.AppInfo, svip string) (instances []*meta.InstanceInfo, err error) {
    if _, err = discovery.isEnabled(); err != nil {
        return nil, err
    }
    Config := discovery.Config
    if *Config.PreferSameZoneEureka && Apps != nil {
        if apps, ok := Apps[Config.Zone]; ok {
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
    if Apps != nil {
        for k, v := range Apps {
            anyMap[k] = v
        }
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

// FilterInstanceBySvip 查询指定svip的可用服务实例（随机选择）
func (discovery *DiscoveryClient) FilterInstanceBySvip(Apps map[string][]*meta.AppInfo, svip string) (*meta.InstanceInfo, error) {
    instances, err := discovery.FilterInstancesBySvip(Apps, svip)
    if err != nil {
        return nil, err
    }
    return instances[rand.Intn(len(instances))], nil
}
