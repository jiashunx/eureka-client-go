package client

import (
    "errors"
    "fmt"
    "github.com/jiashunx/eureka-client-go/meta"
    "math/rand"
    "strings"
)

// GetApp 查询服务信息
func GetApp(apps []*meta.AppInfo, appName string) *meta.AppInfo {
    if apps == nil || appName == "" {
        return nil
    }
    for _, app := range apps {
        if strings.ToUpper(app.Name) == strings.ToUpper(appName) {
            return app
        }
    }
    return nil
}

// GetAppInstance 查询服务实例信息
func GetAppInstance(apps []*meta.AppInfo, appName, instanceId string) *meta.InstanceInfo {
    if apps == nil || appName == "" || instanceId == "" {
        return nil
    }
    for _, app := range apps {
        if strings.ToUpper(app.Name) == strings.ToUpper(appName) {
            if app.Instances == nil {
                continue
            }
            for _, instance := range app.Instances {
                if instance.InstanceId == instanceId {
                    return instance
                }
            }
        }
    }
    return nil
}

// GetInstance 查询服务实例信息
func GetInstance(apps []*meta.AppInfo, instanceId string) *meta.InstanceInfo {
    if apps == nil || instanceId == "" {
        return nil
    }
    for _, app := range apps {
        if app.Instances == nil {
            continue
        }
        for _, instance := range app.Instances {
            if instance.InstanceId == instanceId {
                return instance
            }
        }
    }
    return nil
}

// GetAppsByVip 查询有相同vip的服务列表
func GetAppsByVip(apps []*meta.AppInfo, vip string) []*meta.AppInfo {
    vipApps := make([]*meta.AppInfo, 0)
    if apps == nil || vip == "" {
        return vipApps
    }
    for _, app := range apps {
        if app.Instances == nil {
            continue
        }
        instances := make([]*meta.InstanceInfo, 0)
        for _, instance := range app.Instances {
            if instance.VipAddress == vip {
                instances = append(instances, instance)
            }
        }
        if len(instances) > 0 {
            vipApps = append(vipApps, &meta.AppInfo{
                Name:      app.Name,
                Instances: instances,
            })
        }
    }
    return vipApps
}

// GetAppsBySvip 查询有相同svip的服务列表
func GetAppsBySvip(apps []*meta.AppInfo, svip string) []*meta.AppInfo {
    svipApps := make([]*meta.AppInfo, 0)
    if apps == nil || svip == "" {
        return svipApps
    }
    for _, app := range apps {
        if app.Instances == nil {
            continue
        }
        instances := make([]*meta.InstanceInfo, 0)
        for _, instance := range app.Instances {
            if instance.SecureVipAddress == svip {
                instances = append(instances, instance)
            }
        }
        if len(instances) > 0 {
            svipApps = append(svipApps, &meta.AppInfo{
                Name:      app.Name,
                Instances: instances,
            })
        }
    }
    return svipApps
}

// GetInstancesByVip 查询有相同vip的服务实例列表
func GetInstancesByVip(apps []*meta.AppInfo, vip string) []*meta.InstanceInfo {
    instances := make([]*meta.InstanceInfo, 0)
    vipApps := GetAppsByVip(apps, vip)
    for _, app := range vipApps {
        instances = append(instances, app.Instances...)
    }
    return instances
}

// GetInstancesBySvip 查询有相同svip的服务信息列表
func GetInstancesBySvip(apps []*meta.AppInfo, svip string) []*meta.InstanceInfo {
    instances := make([]*meta.InstanceInfo, 0)
    svipApps := GetAppsBySvip(apps, svip)
    for _, app := range svipApps {
        instances = append(instances, app.Instances...)
    }
    return instances
}

// RandomLoopMap 随机遍历map
func RandomLoopMap(m map[string]interface{}, f func(k string, v interface{}) (bool, error)) (err error) {
    defer func() {
        if rc := recover(); rc != nil {
            err = errors.New(fmt.Sprintf("failed to loop map (random), reason: %v", rc))
        }
    }()
    if m == nil || len(m) == 0 || f == nil {
        return nil
    }
    keys := make([]string, 0)
    for k, _ := range m {
        keys = append(keys, k)
    }
    loop := 0
    size := len(keys)
    idx := rand.Intn(size)
    for i := idx; i < size; i++ {
        if i == idx {
            loop++
        }
        if loop > 1 {
            break
        }
        if i == size-1 {
            i = -1
            continue
        }
        k := keys[i]
        t, e := f(k, m[k])
        if e != nil {
            return e
        }
        if t {
            continue
        }
        break
    }
    return nil
}
