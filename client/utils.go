package client

import "github.com/jiashunx/eureka-client-go/meta"

// GetApp 查询服务信息
func GetApp(apps []*meta.AppInfo, appName string) *meta.AppInfo {
    return nil
}

// GetAppInstance 查询服务实例信息
func GetAppInstance(apps []*meta.AppInfo, appName, instanceId string) *meta.InstanceInfo {
    return nil
}

// GetInstance 查询服务实例信息
func GetInstance(apps []*meta.AppInfo, instanceId string) *meta.InstanceInfo {
    return nil
}

// GetAppsByVip 查询有相同vip的服务信息列表
func GetAppsByVip(apps []*meta.AppInfo, appName string) []*meta.AppInfo {
    return nil
}

// GetAppsBySvip 查询有相同svip的服务信息列表
func GetAppsBySvip(apps []*meta.AppInfo, appName string) []*meta.AppInfo {
    return nil
}
