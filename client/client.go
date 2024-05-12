package client

import (
    "context"
    "errors"
    "fmt"
    "github.com/google/uuid"
    "github.com/jiashunx/eureka-client-go/log"
    "github.com/jiashunx/eureka-client-go/meta"
    "strings"
)

// eurekaClientUUID context中存储的客户端uuid属性名称
var eurekaClientUUID = "EurekaClientUUID"

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
    UUID            string
    config          *meta.EurekaConfig
    rootCtx         context.Context
    ctx             context.Context
    ctxCancel       context.CancelFunc
    registryClient  *registryClient
    discoveryClient *discoveryClient
    httpClient      *HttpClient
    logger          log.Logger
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
    if client.rootCtx == nil {
        client.rootCtx = context.TODO()
    }
    if ctx == nil {
        ctx = client.rootCtx
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
    client.registryClient = &registryClient{client: client, logger: client.logger}
    client.discoveryClient = &discoveryClient{client: client, logger: client.logger}
    subCtx := context.WithValue(client.ctx, eurekaClientUUID, client.UUID)
    if response := client.registryClient.start(subCtx); response.Error != nil {
        client.ctxCancel()
        client.Stop()
        return response.Error
    }
    client.discoveryClient.start(subCtx)
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

// AccessApp 查询可用服务信息
func (client *EurekaClient) AccessApp(appName string) (*meta.AppInfo, error) {
    if client.ctx != nil {
        select {
        case <-client.ctx.Done():
            return nil, clientHasBeenStoppedErr("failed to query available service, appName: %s", appName)
        default:
            return client.discoveryClient.accessApp(appName)
        }
    }
    return nil, clientNotStartedErr("failed to query available service, appName: %s", appName)
}

// AccessAppsByVip 查询指定vip的可用服务列表
func (client *EurekaClient) AccessAppsByVip(vip string) ([]*meta.AppInfo, error) {
    if client.ctx != nil {
        select {
        case <-client.ctx.Done():
            return nil, clientHasBeenStoppedErr("failed to query available service, vip: %s", vip)
        default:
            return client.discoveryClient.accessAppsByVip(vip)
        }
    }
    return nil, clientNotStartedErr("failed to query available service, vip: %s", vip)
}

// AccessAppsBySvip 查询指定svip的可用服务列表
func (client *EurekaClient) AccessAppsBySvip(svip string) ([]*meta.AppInfo, error) {
    if client.ctx != nil {
        select {
        case <-client.ctx.Done():
            return nil, clientHasBeenStoppedErr("failed to query available service, svip: %s", svip)
        default:
            return client.discoveryClient.accessAppsBySvip(svip)
        }
    }
    return nil, clientNotStartedErr("failed to query available service, svip: %s", svip)
}

// AccessInstancesByVip 查询指定vip的可用服务实例列表
func (client *EurekaClient) AccessInstancesByVip(vip string) ([]*meta.InstanceInfo, error) {
    if client.ctx != nil {
        select {
        case <-client.ctx.Done():
            return nil, clientHasBeenStoppedErr("failed to query available service instance, vip: %s", vip)
        default:
            return client.discoveryClient.accessInstancesByVip(vip)
        }
    }
    return nil, clientNotStartedErr("failed to query available service instance, vip: %s", vip)
}

// AccessInstancesBySvip 查询指定svip的可用服务实例列表
func (client *EurekaClient) AccessInstancesBySvip(svip string) ([]*meta.InstanceInfo, error) {
    if client.ctx != nil {
        select {
        case <-client.ctx.Done():
            return nil, clientHasBeenStoppedErr("failed to query available service instance, svip: %s", svip)
        default:
            return client.discoveryClient.accessInstancesBySvip(svip)
        }
    }
    return nil, clientNotStartedErr("failed to query available service instance, svip: %s", svip)
}

// HttpClient 获取与eureka通讯的 *HttpClient
func (client *EurekaClient) HttpClient() *HttpClient {
    return client.httpClient
}

// SetLogger 设置客户端日志对象
func (client *EurekaClient) SetLogger(logger log.Logger) error {
    if logger == nil {
        return errors.New("log.Logger is nil")
    }
    client.logger = logger
    if client.registryClient != nil {
        client.registryClient.logger = logger
    }
    if client.discoveryClient != nil {
        client.discoveryClient.logger = logger
    }
    client.httpClient.logger = logger
    return nil
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
    logger := log.DefaultLogger()
    httpClient := &HttpClient{logger: logger}
    return &EurekaClient{
        UUID:            strings.ReplaceAll(uuid.New().String(), "-", ""),
        config:          eurekaConfig,
        rootCtx:         nil,
        ctx:             nil,
        ctxCancel:       nil,
        registryClient:  nil,
        discoveryClient: nil,
        httpClient:      httpClient,
        logger:          logger,
    }, nil
}
