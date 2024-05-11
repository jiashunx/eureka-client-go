package meta

import (
    "errors"
    "fmt"
)

// AppInfo 服务信息
type AppInfo struct {
    Name      string          `json:"name"`
    Instances []*InstanceInfo `json:"instance"`
}

// Copy 复制副本
func (app *AppInfo) Copy() *AppInfo {
    if app == nil {
        return nil
    }
    newApp := &AppInfo{
        Name:      app.Name,
        Instances: nil,
    }
    if app.Instances != nil {
        newApp.Instances = make([]*InstanceInfo, 0)
        for _, instance := range app.Instances {
            newApp.Instances = append(newApp.Instances, instance.Copy())
        }
    }
    return newApp
}

// AvailableInstances 获取可用服务实例
func (app *AppInfo) AvailableInstances() []*InstanceInfo {
    instances := make([]*InstanceInfo, 0)
    if app != nil && app.Instances != nil {
        for _, instance := range app.Instances {
            if instance != nil && instance.Status == StatusUp {
                instances = append(instances, instance.Copy())
            }
        }
    }
    return instances
}

// ParseAppInfo 从map中解析服务实例信息
func ParseAppInfo(m map[string]interface{}) (app *AppInfo, err error) {
    defer func() {
        if rc := recover(); rc != nil {
            app = nil
            err = errors.New(fmt.Sprintf("failed to parse app info, reason: %v", rc))
        }
    }()
    app = &AppInfo{}
    app.Name = m["name"].(string)
    app.Instances = make([]*InstanceInfo, 0)
    for _, v := range m["instance"].([]interface{}) {
        instance, err := ParseInstanceInfo(v.(map[string]interface{}))
        if err != nil {
            return nil, err
        }
        app.Instances = append(app.Instances, instance)
    }
    return app, nil
}
