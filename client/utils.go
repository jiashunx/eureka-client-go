package client

import (
    "errors"
    "fmt"
    "github.com/jiashunx/eureka-client-go/meta"
    "math/rand"
    "strings"
    "time"
)

// FilterApp 查询服务信息
func FilterApp(apps []*meta.AppInfo, appName string) *meta.AppInfo {
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

// FilterAppsByVip 查询有相同vip的服务列表
func FilterAppsByVip(apps []*meta.AppInfo, vip string) []*meta.AppInfo {
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

// FilterAppsBySvip 查询有相同svip的服务列表
func FilterAppsBySvip(apps []*meta.AppInfo, svip string) []*meta.AppInfo {
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

// FilterInstancesByVip 查询有相同vip的服务实例列表
func FilterInstancesByVip(apps []*meta.AppInfo, vip string) []*meta.InstanceInfo {
    instances := make([]*meta.InstanceInfo, 0)
    vipApps := FilterAppsByVip(apps, vip)
    for _, app := range vipApps {
        instances = append(instances, app.Instances...)
    }
    return instances
}

// FilterInstancesBySvip 查询有相同svip的服务信息列表
func FilterInstancesBySvip(apps []*meta.AppInfo, svip string) []*meta.InstanceInfo {
    instances := make([]*meta.InstanceInfo, 0)
    svipApps := FilterAppsBySvip(apps, svip)
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
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    for i := len(keys) - 1; i > 0; i-- {
        j := r.Intn(i + 1)
        keys[i], keys[j] = keys[j], keys[i]
    }
    for _, key := range keys {
        t, e := f(key, m[key])
        if e != nil {
            return e
        }
        if !t {
            break
        }
    }
    return nil
}

// SummaryAppsMap 日志:简化服务信息
func SummaryAppsMap(appsMap map[string][]*meta.AppInfo) map[string][]interface{} {
    summary := make(map[string][]interface{})
    if appsMap != nil {
        for zone, zoneApps := range appsMap {
            summary[zone] = SummaryApps(zoneApps)
        }
    }
    return summary
}

// SummaryApps 日志:简化服务信息
func SummaryApps(apps []*meta.AppInfo) []interface{} {
    summary := make([]interface{}, 0)
    if apps != nil {
        for _, app := range apps {
            summary = append(summary, SummaryApp(app))
        }
    }
    return summary
}

// SummaryApp 日志:简化服务信息
func SummaryApp(app *meta.AppInfo) map[string]interface{} {
    summary := make(map[string]interface{})
    if app != nil {
        summary["app"] = app.Name
        summary["instances"] = SummaryInstances(app.Instances)
    }
    return summary
}

// SummaryInstances 日志:简化服务实例信息
func SummaryInstances(instances []*meta.InstanceInfo) []map[string]string {
    summary := make([]map[string]string, 0)
    if instances != nil {
        for _, instance := range instances {
            summary = append(summary, SummaryInstance(instance))
        }
    }
    return summary
}

// SummaryInstance 日志:简化服务实例信息
func SummaryInstance(instance *meta.InstanceInfo) map[string]string {
    summary := make(map[string]string)
    if instance != nil {
        summary["region"] = instance.Region
        summary["zone"] = instance.Zone
        summary["app"] = instance.AppName
        summary["instanceId"] = instance.InstanceId
        summary["hostName"] = instance.HostName
        summary["ipAddr"] = instance.IpAddr
        summary["vipAddress"] = instance.VipAddress
        summary["secureVipAddress"] = instance.SecureVipAddress
    }
    return summary
}
