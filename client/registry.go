package client

import (
    "context"
    "errors"
    "github.com/jiashunx/eureka-client-go/http"
    "github.com/jiashunx/eureka-client-go/meta"
    "time"
)

// RegistryClient eureka服务注册客户端
type RegistryClient struct {
    config    *meta.EurekaConfig
    ctx       context.Context
    heartbeat bool                // 是否开启心跳
    status    meta.InstanceStatus // 服务实例状态
}

// start 启动eureka服务注册客户端
func (client *RegistryClient) start() (response *http.CommonResponse) {
    client.status = meta.StatusStarting
    if *client.config.InstanceEnabledOnIt {
        client.status = meta.StatusUp
    }
    go client.heartBeat()
    if _, err := client.isEnabled(); err != nil {
        return &http.CommonResponse{Error: nil}
    }
    server, _ := client.config.GetCurrZoneEurekaServer()
    response = http.Register(server, client.buildInstanceInfo())
    client.heartbeat = response.Error == nil && response.StatusCode == 204
    return response
}

// heartBeat 心跳处理
func (client *RegistryClient) heartBeat() {
    ticker := time.NewTicker(time.Duration(client.config.LeaseRenewalIntervalInSeconds) * time.Second)
FL:
    for {
        <-ticker.C
        select {
        case <-client.ctx.Done():
            ticker.Stop()
            break FL
        default:
            go client.heartBeat0()
        }
    }
}

// heartBeat 心跳处理
func (client *RegistryClient) heartBeat0() {
    if b, _ := client.isEnabled(); b && client.heartbeat && client.status == meta.StatusUp {
        server, err := client.config.GetCurrZoneEurekaServer()
        if err != nil {
            return
        }
        _ = http.Heartbeat(server, client.config.AppName, client.config.InstanceId)
    }
}

// changeStatus 变更服务状态
func (client *RegistryClient) changeStatus(status meta.InstanceStatus) (response *http.CommonResponse) {
    if _, err := client.isEnabled(); err != nil {
        return &http.CommonResponse{Error: err}
    }
    server, err := client.config.GetCurrZoneEurekaServer()
    if err != nil {
        return &http.CommonResponse{Error: err}
    }
    switch status {
    case meta.StatusUp, meta.StatusDown, meta.StatusStarting, meta.StatusOutOfService, meta.StatusUnknown:
        response = http.ChangeStatus(server, client.config.AppName, client.config.InstanceId, status)
        if response.Error != nil {
            break
        }
        client.status = status
        client.heartbeat = status == meta.StatusUp
    default:
        response = &http.CommonResponse{Error: errors.New("invalid status: " + string(status))}
    }
    return response
}

// changeMetadata 变更元数据
func (client *RegistryClient) changeMetadata(metadata map[string]string) (response *http.CommonResponse) {
    if _, err := client.isEnabled(); err != nil {
        return &http.CommonResponse{Error: err}
    }
    server, err := client.config.GetCurrZoneEurekaServer()
    if err != nil {
        return &http.CommonResponse{Error: err}
    }
    response = http.ModifyMetadata(server, client.config.AppName, client.config.InstanceId, metadata)
    if response.Error == nil && response.StatusCode == 200 {
        for key, value := range metadata {
            client.config.Metadata[key] = value
        }
    }
    return response
}

// isEnabled 服务注册功能是否开启
func (client *RegistryClient) isEnabled() (bool, error) {
    if !*client.config.RegistryEnabled {
        return false, errors.New("service registration feature is not enabled")
    }
    return true, nil
}

// buildInstanceInfo 根据配置构造 *meta.InstanceInfo
func (client *RegistryClient) buildInstanceInfo() *meta.InstanceInfo {
    // TODO 待补充具体逻辑
    return &meta.InstanceInfo{}
}
