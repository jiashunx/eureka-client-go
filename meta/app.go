package meta

import (
    "encoding/json"
    "errors"
    "fmt"
)

// AppInfo 服务信息
type AppInfo struct {
    Name      string          `json:"name"`
    Instances []*InstanceInfo `json:"instance"`
    Region    string          `json:"-"`
    Zone      string          `json:"-"`
}

// Copy 复制副本
func (app *AppInfo) Copy() *AppInfo {
    if app == nil {
        return nil
    }
    newApp := &AppInfo{
        Name:      app.Name,
        Instances: nil,
        Region:    app.Region,
        Zone:      app.Zone,
    }
    if app.Instances != nil {
        newApp.Instances = make([]*InstanceInfo, 0)
        for _, instance := range app.Instances {
            newApp.Instances = append(newApp.Instances, instance.Copy())
        }
    }
    return newApp
}

// CopyWithInstances 复制副本（同时更新服务实例列表）
func (app *AppInfo) CopyWithInstances(instances []*InstanceInfo) *AppInfo {
    if app == nil {
        return nil
    }
    return &AppInfo{
        Name:      app.Name,
        Instances: instances,
        Region:    app.Region,
        Zone:      app.Zone,
    }
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

// ParseAppInfo 从json中解析服务信息
func ParseAppInfo(data []byte) (app *AppInfo, err error) {
    defer func() {
        if rc := recover(); rc != nil {
            err = errors.New(fmt.Sprintf("failed to parse app info, recover error: %v", rc))
        }
    }()
    app = &AppInfo{}
    err = json.Unmarshal(data, app)
    if err != nil {
        return nil, err
    }
    if app.Instances == nil {
        app.Instances = make([]*InstanceInfo, 0)
    }
    return app, nil
}
