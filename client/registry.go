package client

import (
    "context"
    "errors"
    "fmt"
    "github.com/jiashunx/eureka-client-go/log"
    "github.com/jiashunx/eureka-client-go/meta"
    "time"
)

// registryClient eureka服务注册客户端
type registryClient struct {
    client    *EurekaClient
    logger    log.Logger
    heartbeat bool                // 是否开启心跳
    status    meta.InstanceStatus // 服务实例状态
}

// start 启动eureka服务注册客户端
func (registry *registryClient) start(ctx context.Context) (response *CommonResponse) {
    defer func() {
        if rc := recover(); rc != nil {
            response = &CommonResponse{}
            response.Error = errors.New(fmt.Sprintf("start, recover error: %v", rc))
        }
        if response.Error != nil {
            registry.logger.Errorf("start, FAILED >>> error: %v", response.Error)
        }
        if response.Error == nil {
            registry.logger.Tracef("start, OK")
        }
    }()
    client := registry.client
    registry.status = meta.StatusStarting
    if *client.config.InstanceEnabledOnIt {
        registry.status = meta.StatusUp
    }
    go registry.beat(ctx)
    if _, err := registry.isEnabled(); err != nil {
        return &CommonResponse{Error: err}
    }
    server, _ := client.config.GetCurrZoneEurekaServer()
    instance, err := registry.buildInstanceInfo(registry.status, meta.Added)
    if err != nil {
        return &CommonResponse{Error: err}
    }
    response = client.HttpClient().Register(server, instance)
    registry.heartbeat = response.Error == nil
    return response
}

// beat 心跳处理
func (registry *registryClient) beat(ctx context.Context) {
    ticker := time.NewTicker(time.Duration(registry.client.config.LeaseRenewalIntervalInSeconds) * time.Second)
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
func (registry *registryClient) beat0(ctx context.Context) (response *CommonResponse) {
    defer func() {
        if rc := recover(); rc != nil {
            response = &CommonResponse{}
            response.Error = errors.New(fmt.Sprintf("beat0, recover error: %v", rc))
        }
        if response.Error != nil {
            registry.logger.Errorf("beat0, FAILED >>> error: %v", response.Error)
        }
        if response.Error != nil {
            registry.logger.Tracef("beat0, OK")
        }
    }()
    client := registry.client
    _, err := registry.isEnabled()
    if err == nil && registry.heartbeat && registry.status == meta.StatusUp {
        var server *meta.EurekaServer
        server, err = client.config.GetCurrZoneEurekaServer()
        if err == nil {
            return client.HttpClient().Heartbeat(server, client.config.AppName, client.config.InstanceId)
        }
    }
    return &CommonResponse{Error: err}
}

// unRegister 取消注册服务
func (registry *registryClient) unRegister() (response *CommonResponse) {
    defer func() {
        if rc := recover(); rc != nil {
            response = &CommonResponse{}
            response.Error = errors.New(fmt.Sprintf("unRegister, recover error: %v", rc))
        }
        if response.Error != nil {
            registry.logger.Errorf("unRegister, FAILED >>> error: %v", response.Error)
        }
        if response.Error == nil {
            registry.logger.Tracef("unRegister, OK")
        }
    }()
    client := registry.client
    server, err := client.config.GetCurrZoneEurekaServer()
    if err != nil {
        return &CommonResponse{Error: err}
    }
    response = client.HttpClient().UnRegister(server, client.config.AppName, client.config.InstanceId)
    registry.heartbeat = !(response.Error == nil)
    return response
}

// changeStatus 变更服务状态
func (registry *registryClient) changeStatus(status meta.InstanceStatus) (response *CommonResponse) {
    defer func() {
        if rc := recover(); rc != nil {
            response = &CommonResponse{}
            response.Error = errors.New(fmt.Sprintf("changeStatus, recover error: %v", rc))
        }
        if response.Error != nil {
            registry.logger.Errorf("changeStatus, FAILED >>> error: %v", response.Error)
        }
        if response.Error == nil {
            registry.logger.Tracef("changeStatus, OK")
        }
    }()
    registry.logger.Tracef("changeStatus, PARAMS >>> status: %v", status)
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
        response = client.HttpClient().ChangeStatus(server, client.config.AppName, client.config.InstanceId, status)
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

// changeMetadata 变更元数据
func (registry *registryClient) changeMetadata(metadata map[string]string) (response *CommonResponse) {
    defer func() {
        if rc := recover(); rc != nil {
            response = &CommonResponse{}
            response.Error = errors.New(fmt.Sprintf("changeMetadata, recover error: %v", rc))
        }
        if response.Error != nil {
            registry.logger.Errorf("changeMetadata, FAILED >>> error: %v", response.Error)
        }
        if response.Error == nil {
            registry.logger.Tracef("changeMetadata, OK")
        }
    }()
    registry.logger.Tracef("changeMetadata, PARAMS >>> metadata: %v", metadata)
    client := registry.client
    if _, err := registry.isEnabled(); err != nil {
        return &CommonResponse{Error: err}
    }
    server, err := client.config.GetCurrZoneEurekaServer()
    if err != nil {
        return &CommonResponse{Error: err}
    }
    response = client.HttpClient().ModifyMetadata(server, client.config.AppName, client.config.InstanceId, metadata)
    if response.Error == nil {
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
func (registry *registryClient) buildInstanceInfo(status meta.InstanceStatus, action meta.ActionType) (instance *meta.InstanceInfo, err error) {
    defer func() {
        if rc := recover(); rc != nil {
            err = errors.New(fmt.Sprintf("buildInstanceInfo, recover error: %v", rc))
        }
        if err != nil {
            registry.logger.Errorf("buildInstanceInfo, FAILED >>> error: %v", err)
        }
        if err == nil {
            registry.logger.Tracef("buildInstanceInfo, OK >>> instance: %v", instance)
        }
    }()
    registry.logger.Tracef("buildInstanceInfo, PARAMS >>> status: %v, action: %v", status, action)
    config := registry.client.config
    instance = &meta.InstanceInfo{
        InstanceId:                    config.InstanceId,
        HostName:                      config.Hostname,
        AppName:                       config.AppName,
        IpAddr:                        config.IpAddress,
        Status:                        status,
        OverriddenStatus:              meta.StatusUnknown,
        Port:                          meta.DefaultNonSecurePortWrapper(),
        SecurePort:                    meta.DefaultSecurePortWrapper(),
        CountryId:                     1,
        DataCenterInfo:                meta.DefaultDataCenterInfo(),
        LeaseInfo:                     meta.DefaultLeaseInfo(),
        Metadata:                      make(map[string]string),
        HomePageUrl:                   config.HomePageUrl,
        StatusPageUrl:                 config.StatusPageUrl,
        HealthCheckUrl:                config.HealthCheckUrl,
        VipAddress:                    config.VirtualHostname,
        SecureVipAddress:              config.SecureVirtualHostname,
        IsCoordinatingDiscoveryServer: "false",
        LastUpdatedTimestamp:          "",
        LastDirtyTimestamp:            "",
        ActionType:                    action,
        Region:                        config.Region,
        Zone:                          config.Zone,
    }
    if *config.PreferIpAddress {
        instance.HostName = config.IpAddress
    }
    instance.Port.Port = config.NonSecurePort
    instance.Port.Enabled = meta.StrFalse
    if *config.NonSecurePortEnabled {
        instance.Port.Enabled = meta.StrTrue
    }
    instance.SecurePort.Port = config.SecurePort
    instance.SecurePort.Enabled = meta.StrFalse
    if *config.SecurePortEnabled {
        instance.SecurePort.Enabled = meta.StrTrue
    }
    instance.LeaseInfo.RenewalIntervalInSecs = config.LeaseRenewalIntervalInSeconds
    instance.LeaseInfo.DurationInSecs = config.LeaseExpirationDurationInSeconds
    for k, v := range config.Metadata {
        instance.Metadata[k] = v
    }
    httpUrl, _ := instance.HttpsServiceUrl()
    httpsUrl, _ := instance.HttpsServiceUrl()
    if instance.HomePageUrl == "" && httpUrl != "" {
        instance.HomePageUrl = httpUrl + config.HomePageUrlPath
    }
    if instance.HomePageUrl == "" && httpsUrl != "" {
        instance.HomePageUrl = httpsUrl + config.HomePageUrlPath
    }
    if instance.StatusPageUrl == "" && httpUrl != "" {
        instance.StatusPageUrl = httpUrl + config.StatusPageUrlPath
    }
    if instance.StatusPageUrl == "" && httpsUrl != "" {
        instance.StatusPageUrl = httpsUrl + config.StatusPageUrlPath
    }
    if instance.HealthCheckUrl == "" && httpUrl != "" {
        instance.HealthCheckUrl = httpUrl + config.HealthCheckUrlPath
    }
    if instance.HealthCheckUrl == "" && httpsUrl != "" {
        instance.HealthCheckUrl = httpsUrl + config.HealthCheckUrlPath
    }
    return instance, nil
}
