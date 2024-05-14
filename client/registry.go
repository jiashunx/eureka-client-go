package client

import (
    "context"
    "errors"
    "fmt"
    "github.com/jiashunx/eureka-client-go/log"
    "github.com/jiashunx/eureka-client-go/meta"
    "time"
)

// RegistryClient eureka服务注册客户端
type RegistryClient struct {
    HttpClient *HttpClient
    Config     *meta.EurekaConfig
    Logger     log.Logger
    // 是否开启心跳, 仅当集成到 EurekaClient 时有效
    heartbeat bool
    // 心跳失败回调, 仅当集成到 EurekaClient 时有效
    HeartbeatFailFunc func(*CommonResponse)
    // 服务实例状态, 仅当集成到 EurekaClient 时有效
    status meta.InstanceStatus
}

// start 启动eureka服务注册客户端
func (registry *RegistryClient) start(ctx context.Context) (response *CommonResponse) {
    defer func() {
        if rc := recover(); rc != nil {
            response = &CommonResponse{}
            response.Error = errors.New(fmt.Sprintf("RegistryClient.start, recover error: %v", rc))
        }
        if response.Error != nil {
            registry.Logger.Tracef("RegistryClient.start, FAILED >>> error: %v", response.Error)
        }
        if response.Error == nil {
            registry.Logger.Tracef("RegistryClient.start, OK")
        }
    }()
    registry.status = meta.StatusStarting
    if *registry.Config.InstanceEnabledOnIt {
        registry.status = meta.StatusUp
    }
    registry.heartbeat = false
    go registry.beat(ctx)
    if _, err := registry.isEnabled(); err != nil {
        return &CommonResponse{Error: err}
    }
    server, _ := registry.Config.GetCurrZoneEurekaServer()
    instance, err := registry.buildInstanceInfo(registry.status, meta.Added)
    if err != nil {
        return &CommonResponse{Error: err}
    }
    response = registry.HttpClient.Register(server, instance)
    registry.heartbeat = response.Error == nil
    return response
}

// beat 心跳处理
func (registry *RegistryClient) beat(ctx context.Context) {
    ticker := time.NewTicker(time.Duration(registry.Config.LeaseRenewalIntervalInSeconds) * time.Second)
FL:
    for {
        select {
        case <-ctx.Done():
            ticker.Stop()
            break FL
        default:
            go registry.beat0(ctx)
        }
        <-ticker.C
    }
}

// beat 心跳处理
func (registry *RegistryClient) beat0(ctx context.Context) (response *CommonResponse) {
    defer func() {
        if rc := recover(); rc != nil {
            response = &CommonResponse{}
            response.Error = errors.New(fmt.Sprintf("RegistryClient.beat0, recover error: %v", rc))
        }
        if response.Error != nil {
            registry.Logger.Tracef("RegistryClient.beat0, FAILED >>> error: %v", response.Error)
            if registry.HeartbeatFailFunc != nil {
                go registry.HeartbeatFailFunc(response)
            }
        }
        if response.Error != nil {
            registry.Logger.Tracef("RegistryClient.beat0, OK")
        }
    }()
    _, err := registry.isEnabled()
    if err == nil && registry.heartbeat && registry.status == meta.StatusUp {
        return registry.Heartbeat()
    }
    return &CommonResponse{Error: err}
}

// Register 服务注册
func (registry *RegistryClient) Register(status meta.InstanceStatus) *CommonResponse {
    if _, err := registry.isEnabled(); err != nil {
        return &CommonResponse{Error: err}
    }
    server, err := registry.Config.GetCurrZoneEurekaServer()
    if err != nil {
        return &CommonResponse{Error: err}
    }
    instance, err := registry.buildInstanceInfo(status, meta.Added)
    if err != nil {
        return &CommonResponse{Error: err}
    }
    return registry.HttpClient.Register(server, instance)
}

// Heartbeat 心跳
func (registry *RegistryClient) Heartbeat() *CommonResponse {
    if _, err := registry.isEnabled(); err != nil {
        return &CommonResponse{Error: err}
    }
    server, err := registry.Config.GetCurrZoneEurekaServer()
    if err != nil {
        return &CommonResponse{Error: err}
    }
    return registry.HttpClient.Heartbeat(server, registry.Config.AppName, registry.Config.InstanceId)
}

// UnRegister 取消注册服务
func (registry *RegistryClient) UnRegister() (response *CommonResponse) {
    if _, err := registry.isEnabled(); err != nil {
        return &CommonResponse{Error: err}
    }
    server, err := registry.Config.GetCurrZoneEurekaServer()
    if err != nil {
        return &CommonResponse{Error: err}
    }
    response = registry.HttpClient.UnRegister(server, registry.Config.AppName, registry.Config.InstanceId)
    registry.heartbeat = !(response.Error == nil)
    return response
}

// ChangeStatus 变更服务状态
func (registry *RegistryClient) ChangeStatus(status meta.InstanceStatus) (response *CommonResponse) {
    if _, err := registry.isEnabled(); err != nil {
        return &CommonResponse{Error: err}
    }
    server, err := registry.Config.GetCurrZoneEurekaServer()
    if err != nil {
        return &CommonResponse{Error: err}
    }
    switch status {
    case meta.StatusUp, meta.StatusDown, meta.StatusStarting, meta.StatusOutOfService, meta.StatusUnknown:
        response = registry.HttpClient.ChangeStatus(server, registry.Config.AppName, registry.Config.InstanceId, status)
        if response.Error != nil {
            break
        }
        registry.status = status
        registry.heartbeat = status == meta.StatusUp
    default:
        response = &CommonResponse{}
        response.Error = errors.New("status value is invalid: " + string(status))
    }
    return response
}

// ChangeMetadata 变更元数据
func (registry *RegistryClient) ChangeMetadata(metadata map[string]string) (response *CommonResponse) {
    if _, err := registry.isEnabled(); err != nil {
        return &CommonResponse{Error: err}
    }
    server, err := registry.Config.GetCurrZoneEurekaServer()
    if err != nil {
        return &CommonResponse{Error: err}
    }
    response = registry.HttpClient.ModifyMetadata(server, registry.Config.AppName, registry.Config.InstanceId, metadata)
    if response.Error == nil {
        for key, value := range metadata {
            registry.Config.Metadata[key] = value
        }
    }
    return response
}

// isEnabled 服务注册功能是否开启
func (registry *RegistryClient) isEnabled() (bool, error) {
    if !*registry.Config.RegistryEnabled {
        return false, errors.New("eureka client's service registration feature is not enabled")
    }
    return true, nil
}

// buildInstanceInfo 根据配置构造 *meta.InstanceInfo
func (registry *RegistryClient) buildInstanceInfo(status meta.InstanceStatus, action meta.ActionType) (instance *meta.InstanceInfo, err error) {
    Config := registry.Config
    instance = &meta.InstanceInfo{
        InstanceId:                    Config.InstanceId,
        HostName:                      Config.Hostname,
        AppName:                       Config.AppName,
        IpAddr:                        Config.IpAddress,
        Status:                        status,
        OverriddenStatus:              meta.StatusUnknown,
        Port:                          meta.DefaultNonSecurePortWrapper(),
        SecurePort:                    meta.DefaultSecurePortWrapper(),
        CountryId:                     1,
        DataCenterInfo:                meta.DefaultDataCenterInfo(),
        LeaseInfo:                     meta.DefaultLeaseInfo(),
        Metadata:                      make(map[string]string),
        HomePageUrl:                   Config.HomePageUrl,
        StatusPageUrl:                 Config.StatusPageUrl,
        HealthCheckUrl:                Config.HealthCheckUrl,
        VipAddress:                    Config.VirtualHostname,
        SecureVipAddress:              Config.SecureVirtualHostname,
        IsCoordinatingDiscoveryServer: "false",
        LastUpdatedTimestamp:          "",
        LastDirtyTimestamp:            "",
        ActionType:                    action,
        Region:                        Config.Region,
        Zone:                          Config.Zone,
    }
    if *Config.PreferIpAddress {
        instance.HostName = Config.IpAddress
    }
    instance.Port.Port = Config.NonSecurePort
    instance.Port.Enabled = meta.StrFalse
    if *Config.NonSecurePortEnabled {
        instance.Port.Enabled = meta.StrTrue
    }
    instance.SecurePort.Port = Config.SecurePort
    instance.SecurePort.Enabled = meta.StrFalse
    if *Config.SecurePortEnabled {
        instance.SecurePort.Enabled = meta.StrTrue
    }
    instance.LeaseInfo.RenewalIntervalInSecs = Config.LeaseRenewalIntervalInSeconds
    instance.LeaseInfo.DurationInSecs = Config.LeaseExpirationDurationInSeconds
    for k, v := range Config.Metadata {
        instance.Metadata[k] = v
    }
    httpUrl, _ := instance.HttpsServiceUrl()
    httpsUrl, _ := instance.HttpsServiceUrl()
    if instance.HomePageUrl == "" && httpUrl != "" {
        instance.HomePageUrl = httpUrl + Config.HomePageUrlPath
    }
    if instance.HomePageUrl == "" && httpsUrl != "" {
        instance.HomePageUrl = httpsUrl + Config.HomePageUrlPath
    }
    if instance.StatusPageUrl == "" && httpUrl != "" {
        instance.StatusPageUrl = httpUrl + Config.StatusPageUrlPath
    }
    if instance.StatusPageUrl == "" && httpsUrl != "" {
        instance.StatusPageUrl = httpsUrl + Config.StatusPageUrlPath
    }
    if instance.HealthCheckUrl == "" && httpUrl != "" {
        instance.HealthCheckUrl = httpUrl + Config.HealthCheckUrlPath
    }
    if instance.HealthCheckUrl == "" && httpsUrl != "" {
        instance.HealthCheckUrl = httpsUrl + Config.HealthCheckUrlPath
    }
    return instance, nil
}
