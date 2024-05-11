package client

import (
    "context"
    "errors"
    "fmt"
    "github.com/jiashunx/eureka-client-go/meta"
)

// clientNotStartedErr 错误:客户端未启动
var clientNotStartedErr = func(format string, a ...any) error {
    if format == "" {
        return errors.New("eureka client has not been started")
    }
    return errors.New(fmt.Sprintf(format, a...) + ", reason: eureka client has not been started")
}

// clientHasBeenStoppedErr 错误:客户端已关闭
var clientHasBeenStoppedErr = func(format string, a ...any) error {
    if format == "" {
        return errors.New("eureka client has already been stopped")
    }
    return errors.New(fmt.Sprintf(format, a...) + ", reason: eureka client has already been stopped")
}

// EurekaClient eureka客户端模型
type EurekaClient struct {
    config          *meta.EurekaConfig
    ctx             context.Context
    ctxCancel       context.CancelFunc
    registryClient  *registryClient
    discoveryClient *discoveryClient
    HttpClient      *HttpClient
}

// Start 启动eureka客户端
func (client *EurekaClient) Start() error {
    return client.StartWithCtx(nil)
}

// StartWithCtx 启动eureka客户端并指定 context.Context
func (client *EurekaClient) StartWithCtx(ctx context.Context) error {
    if client.config == nil {
        return errors.New("EurekaConfig is nil")
    }
    if ctx == nil {
        ctx = context.Background()
    }
    if client.ctx != nil {
        select {
        case <-client.ctx.Done():
            break
        default:
            return errors.New("failed to start eureka client, reason: eureka client is still running")
        }
    }
    client.ctx, client.ctxCancel = context.WithCancel(ctx)
    client.registryClient = &registryClient{client: client}
    client.discoveryClient = &discoveryClient{client: client}
    if response := client.registryClient.start(); response.Error != nil {
        client.ctxCancel()
        client.Stop()
        return response.Error
    }
    client.discoveryClient.start()
    return nil
}

// Stop 关闭eureka客户端
func (client *EurekaClient) Stop() *CommonResponse {
    if client.ctx != nil {
        select {
        case <-client.ctx.Done():
            return &CommonResponse{Error: clientHasBeenStoppedErr("failed to stop eureka client")}
        default:
            response := client.registryClient.unRegister()
            if response.Error == nil {
                client.ctxCancel()
            }
            return response
        }
    }
    return &CommonResponse{Error: clientNotStartedErr("failed to stop eureka client")}
}

// ChangeStatus 变更服务状态
func (client *EurekaClient) ChangeStatus(status meta.InstanceStatus) *CommonResponse {
    if client.ctx != nil {
        select {
        case <-client.ctx.Done():
            return &CommonResponse{Error: clientHasBeenStoppedErr("failed to change service instance's status, status: %v", status)}
        default:
            return client.registryClient.changeStatus(status)
        }
    }
    return &CommonResponse{Error: clientNotStartedErr("failed to change service instance's status, status: %v", status)}
}

// ChangeMetadata 变更元数据
func (client *EurekaClient) ChangeMetadata(metadata map[string]string) *CommonResponse {
    if client.ctx != nil {
        select {
        case <-client.ctx.Done():
            return &CommonResponse{Error: clientHasBeenStoppedErr("failed to change service instance's metadata, metadata: %v", metadata)}
        default:
            return client.registryClient.changeMetadata(metadata)
        }
    }
    return &CommonResponse{Error: clientNotStartedErr("failed to change service instance's metadata, metadata: %v", metadata)}
}

// GetApp 查询服务信息
func (client *EurekaClient) GetApp(appName string) (*meta.AppInfo, error) {
    if client.ctx != nil {
        select {
        case <-client.ctx.Done():
            return nil, clientHasBeenStoppedErr("failed to query the service, appName: %s", appName)
        default:
            return client.discoveryClient.getApp(appName)
        }
    }
    return nil, clientNotStartedErr("failed to query the service, appName: %s", appName)
}

// GetAppInstance 查询服务实例信息
func (client *EurekaClient) GetAppInstance(appName, instanceId string) (*meta.InstanceInfo, error) {
    if client.ctx != nil {
        select {
        case <-client.ctx.Done():
            return nil, clientHasBeenStoppedErr("failed to query the service instance, appName: %s, instanceId: %s", appName, instanceId)
        default:
            return client.discoveryClient.getAppInstance(appName, instanceId)
        }
    }
    return nil, clientNotStartedErr("failed to query the service instance, appName: %s, instanceId: %s", appName, instanceId)
}

// GetInstance 查询服务实例信息
func (client *EurekaClient) GetInstance(instanceId string) (*meta.InstanceInfo, error) {
    if client.ctx != nil {
        select {
        case <-client.ctx.Done():
            return nil, clientHasBeenStoppedErr("failed to query the service instance, instanceId: %s", instanceId)
        default:
            return client.discoveryClient.getInstance(instanceId)
        }
    }
    return nil, clientNotStartedErr("failed to query the service instance, instanceId: %s", instanceId)
}

// GetAppsByVip 查询指定vip的服务列表
func (client *EurekaClient) GetAppsByVip(vip string) ([]*meta.AppInfo, error) {
    if client.ctx != nil {
        select {
        case <-client.ctx.Done():
            return nil, clientHasBeenStoppedErr("failed to query the service, vip: %s", vip)
        default:
            return client.discoveryClient.getAppsByVip(vip)
        }
    }
    return nil, clientNotStartedErr("failed to query the service, vip: %s", vip)
}

// GetAppsBySvip 查询指定svip的服务列表
func (client *EurekaClient) GetAppsBySvip(svip string) ([]*meta.AppInfo, error) {
    if client.ctx != nil {
        select {
        case <-client.ctx.Done():
            return nil, clientHasBeenStoppedErr("failed to query the service, svip: %s", svip)
        default:
            return client.discoveryClient.getAppsBySvip(svip)
        }
    }
    return nil, clientNotStartedErr("failed to query the service, svip: %s", svip)
}

// GetInstancesByVip 查询指定vip的服务实例列表
func (client *EurekaClient) GetInstancesByVip(vip string) ([]*meta.InstanceInfo, error) {
    if client.ctx != nil {
        select {
        case <-client.ctx.Done():
            return nil, clientHasBeenStoppedErr("failed to query the service instance, vip: %s", vip)
        default:
            return client.discoveryClient.getInstancesByVip(vip)
        }
    }
    return nil, clientNotStartedErr("failed to query the service instance, vip: %s", vip)
}

// GetInstancesBySvip 查询指定svip的服务实例列表
func (client *EurekaClient) GetInstancesBySvip(svip string) ([]*meta.InstanceInfo, error) {
    if client.ctx != nil {
        select {
        case <-client.ctx.Done():
            return nil, clientHasBeenStoppedErr("failed to query the service instance, svip: %s", svip)
        default:
            return client.discoveryClient.getInstancesBySvip(svip)
        }
    }
    return nil, clientNotStartedErr("failed to query the service instance, svip: %s", svip)
}

// NewEurekaClient 根据 *meta.EurekaConfig 创建eureka客户端
func NewEurekaClient(config *meta.EurekaConfig) (*EurekaClient, error) {
    if config == nil {
        return nil, errors.New("EurekaConfig is nil")
    }
    eurekaConfig := &meta.EurekaConfig{
        InstanceConfig: config.InstanceConfig,
        ClientConfig:   config.ClientConfig,
    }
    if err := eurekaConfig.Check(); err != nil {
        return nil, err
    }
    httpClient := &HttpClient{}
    return &EurekaClient{
        config:          eurekaConfig,
        ctx:             nil,
        ctxCancel:       nil,
        registryClient:  nil,
        discoveryClient: nil,
        HttpClient:      httpClient,
    }, nil
}
