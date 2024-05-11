package client

import (
    "context"
    "errors"
    "github.com/jiashunx/eureka-client-go/meta"
)

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
            return &CommonResponse{Error: errors.New("failed to stop eureka client, reason: eureka client has already been stopped")}
        default:
            response := client.registryClient.unRegister()
            if response.Error == nil {
                client.ctxCancel()
            }
            return response
        }
    }
    return &CommonResponse{Error: errors.New("failed to stop eureka client, reason: eureka client has not been started")}
}

// ChangeStatus 变更服务状态
func (client *EurekaClient) ChangeStatus(status meta.InstanceStatus) *CommonResponse {
    if client.ctx != nil {
        select {
        case <-client.ctx.Done():
            return &CommonResponse{Error: errors.New("failed to change service instance's status, reason: eureka client has already been stopped")}
        default:
            return client.registryClient.changeStatus(status)
        }
    }
    return &CommonResponse{Error: errors.New("failed to change service instance's status, reason: eureka client has not been started")}
}

// ChangeMetadata 变更元数据
func (client *EurekaClient) ChangeMetadata(metadata map[string]string) *CommonResponse {
    if client.ctx != nil {
        select {
        case <-client.ctx.Done():
            return &CommonResponse{Error: errors.New("failed to change service instance's metadata, reason: eureka client has already been stopped")}
        default:
            return client.registryClient.changeMetadata(metadata)
        }
    }
    return &CommonResponse{Error: errors.New("failed to change service instance's metadata, reason: eureka client has not been started")}
}

// EnabledRegistry 开启/关闭服务注册功能
func (client *EurekaClient) EnabledRegistry(enabled bool) *EurekaClient {
    client.config.RegistryEnabled = &enabled
    return client
}

// EnableDiscovery 开启/关闭服务发现功能
func (client *EurekaClient) EnableDiscovery(enabled bool) *EurekaClient {
    client.config.DiscoveryEnabled = &enabled
    return client
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
