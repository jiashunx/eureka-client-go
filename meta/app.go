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
