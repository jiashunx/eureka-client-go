package http

import "github.com/jiashunx/eureka-client-go/meta"

// 与eureka server通讯的接口处理

// Register 注册新服务
func Register(server *meta.EurekaServer, instance *meta.InstanceInfo) (int, error) {
    return 0, nil
}

// SimpleRegister 注册新服务
func SimpleRegister(serviceUrl string, instance *meta.InstanceInfo) (int, error) {
    return Register(&meta.EurekaServer{ServiceUrl: serviceUrl}, instance)
}

// UnRegister 取消注册服务
func UnRegister(server *meta.EurekaServer, appName, instanceId string) (int, error) {
    return 0, nil
}

// SimpleUnRegister 取消注册服务
func SimpleUnRegister(serviceUrl, appName, instanceId string) (int, error) {
    return UnRegister(&meta.EurekaServer{ServiceUrl: serviceUrl}, appName, instanceId)
}

// Heartbeat 发送服务心跳
func Heartbeat(server *meta.EurekaServer, appName, instanceId string) (int, error) {
    return 0, nil
}

// SimpleHeartbeat 发送服务心跳
func SimpleHeartbeat(serviceUrl, appName, instanceId string) (int, error) {
    return Heartbeat(&meta.EurekaServer{ServiceUrl: serviceUrl}, appName, instanceId)
}

// QueryApps 查询所有服务
func QueryApps(server *meta.EurekaServer) ([]*meta.AppInfo, error) {
    return nil, nil
}

// SimpleQueryApps 查询所有服务
func SimpleQueryApps(serviceUrl string) ([]*meta.AppInfo, error) {
    return QueryApps(&meta.EurekaServer{ServiceUrl: serviceUrl})
}

// QueryApp 查询指定appName的服务列表
func QueryApp(server *meta.EurekaServer, appName string) ([]*meta.InstanceInfo, error) {
    return nil, nil
}

// SimpleQueryApp 查询指定appName的服务列表
func SimpleQueryApp(serviceUrl, appName string) ([]*meta.InstanceInfo, error) {
    return QueryApp(&meta.EurekaServer{ServiceUrl: serviceUrl}, appName)
}

// QueryAppInstance 查询指定appName&InstanceId
func QueryAppInstance(server *meta.EurekaServer, appName, instanceId string) (*meta.InstanceInfo, error) {
    return nil, nil
}

// SimpleQueryAppInstance 查询指定appName&InstanceId
func SimpleQueryAppInstance(serviceUrl, appName, instanceId string) (*meta.InstanceInfo, error) {
    return QueryAppInstance(&meta.EurekaServer{ServiceUrl: serviceUrl}, appName, instanceId)
}

// QueryInstance 查询指定InstanceId服务列表
func QueryInstance(server *meta.EurekaServer, instanceId string) (*meta.InstanceInfo, error) {
    return nil, nil
}

// SimpleQueryInstance 查询指定InstanceId服务列表
func SimpleQueryInstance(serviceUrl, instanceId string) (*meta.InstanceInfo, error) {
    return QueryInstance(&meta.EurekaServer{ServiceUrl: serviceUrl}, instanceId)
}

// ChangeStatus 变更服务状态
func ChangeStatus(server *meta.EurekaServer, appName, instanceId string, status meta.InstanceStatus) (int, error) {
    return 0, nil
}

// SimpleChangeStatus 变更服务状态
func SimpleChangeStatus(serviceUrl, appName, instanceId string, status meta.InstanceStatus) (int, error) {
    return ChangeStatus(&meta.EurekaServer{ServiceUrl: serviceUrl}, appName, instanceId, status)
}

// ModifyMetadata 变更元数据
func ModifyMetadata(server *meta.EurekaServer, appName, instanceId, key, value string) (int, error) {
    return 0, nil
}

// SimpleModifyMetadata 变更元数据
func SimpleModifyMetadata(serviceUrl, appName, instanceId, key, value string) (int, error) {
    return ModifyMetadata(&meta.EurekaServer{ServiceUrl: serviceUrl}, appName, instanceId, key, value)
}

// QueryVipApps 查询指定IP下的服务列表
func QueryVipApps(server *meta.EurekaServer, vipAddress string) ([]*meta.AppInfo, error) {
    return nil, nil
}

// SimpleQueryVipApps 查询指定IP下的服务列表
func SimpleQueryVipApps(serviceUrl, vipAddress string) ([]*meta.AppInfo, error) {
    return QueryVipApps(&meta.EurekaServer{ServiceUrl: serviceUrl}, vipAddress)
}

// QuerySvipApps 查询指定安全IP下的服务列表
func QuerySvipApps(server *meta.EurekaServer, svipAddress string) ([]*meta.AppInfo, error) {
    return nil, nil
}

// SimpleQuerySvipApps 查询指定安全IP下的服务列表
func SimpleQuerySvipApps(serviceUrl, svipAddress string) ([]*meta.AppInfo, error) {
    return QuerySvipApps(&meta.EurekaServer{ServiceUrl: serviceUrl}, svipAddress)
}
