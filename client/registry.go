package client

import (
    "context"
    "errors"
    "github.com/jiashunx/eureka-client-go/meta"
    "time"
)

// registryClient eureka服务注册客户端
type registryClient struct {
    client    *EurekaClient
    heartbeat bool                // 是否开启心跳
    status    meta.InstanceStatus // 服务实例状态
}

// start 启动eureka服务注册客户端
func (registry *registryClient) start() (response *CommonResponse) {
    client := registry.client
    registry.status = meta.StatusStarting
    if *client.config.InstanceEnabledOnIt {
        registry.status = meta.StatusUp
    }
    go registry.heartBeat(registry.client.ctx)
    if _, err := registry.isEnabled(); err != nil {
        return &CommonResponse{Error: nil}
    }
    server, _ := client.config.GetCurrZoneEurekaServer()
    instance, err := registry.buildInstanceInfo(registry.status, meta.Added)
    if err != nil {
        return &CommonResponse{Error: errors.New("failed to create service instance, reason: " + err.Error())}
    }
    response = client.HttpClient.Register(server, instance)
    registry.heartbeat = response.Error == nil
    return response
}

// heartBeat 心跳处理
func (registry *registryClient) heartBeat(ctx context.Context) {
    ticker := time.NewTicker(time.Duration(registry.client.config.LeaseRenewalIntervalInSeconds) * time.Second)
FL:
    for {
        <-ticker.C
        select {
        case <-ctx.Done():
            ticker.Stop()
            break FL
        default:
            go registry.heartBeat0()
        }
    }
}

// heartBeat 心跳处理
func (registry *registryClient) heartBeat0() {
    client := registry.client
    if b, _ := registry.isEnabled(); b && registry.heartbeat && registry.status == meta.StatusUp {
        server, err := client.config.GetCurrZoneEurekaServer()
        if err != nil {
            return
        }
        _ = client.HttpClient.Heartbeat(server, client.config.AppName, client.config.InstanceId)
    }
}

// unRegister 取消注册服务
func (registry *registryClient) unRegister() *CommonResponse {
    client := registry.client
    server, err := client.config.GetCurrZoneEurekaServer()
    if err != nil {
        return &CommonResponse{Error: err}
    }
    response := client.HttpClient.UnRegister(server, client.config.AppName, client.config.InstanceId)
    registry.heartbeat = !(response.Error == nil)
    return response
}

// changeStatus 变更服务状态
func (registry *registryClient) changeStatus(status meta.InstanceStatus) (response *CommonResponse) {
    client := registry.client
    if _, err := registry.isEnabled(); err != nil {
        return &CommonResponse{Error: err}
    }
    server, err := client.config.GetCurrZoneEurekaServer()
    if err != nil {
        return &CommonResponse{Error: err}
    }
    switch status {
    case meta.StatusUp, meta.StatusDown, meta.StatusStarting, meta.StatusOutOfService, meta.StatusUnknown:
        response = client.HttpClient.ChangeStatus(server, client.config.AppName, client.config.InstanceId, status)
        if response.Error != nil {
            break
        }
        registry.status = status
        registry.heartbeat = status == meta.StatusUp
    default:
        response = &CommonResponse{Error: errors.New("failed to change service instance's status, reason: status is invalid: " + string(status))}
    }
    return response
}

// changeMetadata 变更元数据
func (registry *registryClient) changeMetadata(metadata map[string]string) (response *CommonResponse) {
    client := registry.client
    if _, err := registry.isEnabled(); err != nil {
        return &CommonResponse{Error: err}
    }
    server, err := client.config.GetCurrZoneEurekaServer()
    if err != nil {
        return &CommonResponse{Error: err}
    }
    response = client.HttpClient.ModifyMetadata(server, client.config.AppName, client.config.InstanceId, metadata)
    if response.Error == nil && response.StatusCode == 200 {
        for key, value := range metadata {
            client.config.Metadata[key] = value
        }
    }
    return response
}

// isEnabled 服务注册功能是否开启
func (registry *registryClient) isEnabled() (bool, error) {
    client := registry.client
    if !*client.config.RegistryEnabled {
        return false, errors.New("eureka client's service registration feature is not enabled")
    }
    return true, nil
}

// buildInstanceInfo 根据配置构造 *meta.InstanceInfo
func (registry *registryClient) buildInstanceInfo(status meta.InstanceStatus, action meta.ActionType) (*meta.InstanceInfo, error) {
    client := registry.client
    // TODO 待补充具体逻辑
    return &meta.InstanceInfo{
        AppName:    client.config.AppName,
        InstanceId: client.config.InstanceId,
    }, nil
}
