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
var clientNotStartedErr = func() error {
    return errors.New("eureka client has not been started")
}

// clientHasBeenStoppedErr 错误:客户端已关闭
var clientHasBeenStoppedErr = func() error {
    return errors.New("eureka client has already been stopped")
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
func (client *EurekaClient) Start() *CommonResponse {
    return client.StartWithCtx(nil)
}

// StartWithCtx 启动eureka客户端并指定 context.Context
func (client *EurekaClient) StartWithCtx(ctx context.Context) (response *CommonResponse) {
    defer func() {
        if rc := recover(); rc != nil {
            response = &CommonResponse{}
            response.Error = errors.New(fmt.Sprintf("StartWithCtx, recover error: %v", rc))
        }
        if response.Error != nil {
            client.logger.Errorf("StartWithCtx, FAILED >>> error: %v", response.Error)
        }
        if response.Error == nil {
            client.logger.Tracef("StartWithCtx, OK")
        }
    }()
    if client.config == nil {
        return &CommonResponse{Error: errors.New("EurekaConfig is nil")}
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
            return &CommonResponse{Error: errors.New("eureka client is still running")}
        }
    }
    client.ctx, client.ctxCancel = context.WithCancel(ctx)
    client.registryClient = &registryClient{client: client, logger: client.logger}
    client.discoveryClient = &discoveryClient{client: client, logger: client.logger}
    subCtx := context.WithValue(client.ctx, eurekaClientUUID, client.UUID)
    if response = client.registryClient.start(subCtx); response.Error != nil {
        client.ctxCancel()
        client.Stop()
        return response
    }
    client.discoveryClient.start(subCtx)
    return &CommonResponse{Error: nil}
}

// Stop 关闭eureka客户端
func (client *EurekaClient) Stop() (response *CommonResponse) {
    defer func() {
        if rc := recover(); rc != nil {
            response = &CommonResponse{}
            response.Error = errors.New(fmt.Sprintf("Stop, recover error: %v", rc))
        }
        if response.Error != nil {
            client.logger.Errorf("Stop, FAILED >>> error: %v", response.Error)
        }
        if response.Error == nil {
            client.logger.Tracef("Stop, OK")
        }
    }()
    if client.ctx != nil {
        select {
        case <-client.ctx.Done():
            return &CommonResponse{Error: clientHasBeenStoppedErr()}
        default:
            response = client.registryClient.unRegister()
            if response.Error == nil {
                client.ctxCancel()
            }
            return response
        }
    }
    return &CommonResponse{Error: clientNotStartedErr()}
}

// ChangeStatus 变更服务状态
func (client *EurekaClient) ChangeStatus(status meta.InstanceStatus) (response *CommonResponse) {
    defer func() {
        if rc := recover(); rc != nil {
            response = &CommonResponse{}
            response.Error = errors.New(fmt.Sprintf("ChangeStatus, recover error: %v", rc))
        }
        if response.Error != nil {
            client.logger.Errorf("ChangeStatus, FAILED >>> error: %v", response.Error)
        }
        if response.Error == nil {
            client.logger.Tracef("ChangeStatus, OK")
        }
    }()
    client.logger.Tracef("ChangeStatus, PARAMS >>> status: %v", status)
    if client.ctx != nil {
        select {
        case <-client.ctx.Done():
            return &CommonResponse{Error: clientHasBeenStoppedErr()}
        default:
            return client.registryClient.changeStatus(status)
        }
    }
    return &CommonResponse{Error: clientNotStartedErr()}
}

// ChangeMetadata 变更元数据
func (client *EurekaClient) ChangeMetadata(metadata map[string]string) (response *CommonResponse) {
    defer func() {
        if rc := recover(); rc != nil {
            response = &CommonResponse{}
            response.Error = errors.New(fmt.Sprintf("ChangeMetadata, recover error: %v", rc))
        }
        if response.Error != nil {
            client.logger.Errorf("ChangeMetadata, FAILED >>> error: %v", response.Error)
        }
        if response.Error == nil {
            client.logger.Tracef("ChangeMetadata, OK")
        }
    }()
    client.logger.Tracef("ChangeMetadata, PARAMS >>> metadata: %v", metadata)
    if client.ctx != nil {
        select {
        case <-client.ctx.Done():
            return &CommonResponse{Error: clientHasBeenStoppedErr()}
        default:
            return client.registryClient.changeMetadata(metadata)
        }
    }
    return &CommonResponse{Error: clientNotStartedErr()}
}

// AccessApp 查询可用服务信息
func (client *EurekaClient) AccessApp(appName string) (app *meta.AppInfo, err error) {
    defer func() {
        if rc := recover(); rc != nil {
            err = errors.New(fmt.Sprintf("AccessApp, recover error: %v", rc))
        }
        if err != nil {
            client.logger.Errorf("AccessApp, FAILED >>> error: %v", err)
        }
        if err == nil {
            client.logger.Tracef("AccessApp, OK >>> app: %v", SummaryApp(app))
        }
    }()
    client.logger.Tracef("AccessApp, PARAMS >>> appName: %v", appName)
    if client.ctx != nil {
        select {
        case <-client.ctx.Done():
            return nil, clientHasBeenStoppedErr()
        default:
            return client.discoveryClient.accessApp(appName)
        }
    }
    return nil, clientNotStartedErr()
}

// AccessAppsByVip 查询指定vip的可用服务列表
func (client *EurekaClient) AccessAppsByVip(vip string) (vipApps []*meta.AppInfo, err error) {
    defer func() {
        if rc := recover(); rc != nil {
            err = errors.New(fmt.Sprintf("AccessAppsByVip, recover error: %v", rc))
        }
        if err != nil {
            client.logger.Errorf("AccessAppsByVip, FAILED >>> error: %v", err)
        }
        if err == nil {
            client.logger.Tracef("AccessAppsByVip, OK >>> vipApps: %v", vipApps)
        }
    }()
    client.logger.Tracef("AccessAppsByVip, PARAMS >>> vip: %v", SummaryApps(vipApps))
    if client.ctx != nil {
        select {
        case <-client.ctx.Done():
            return nil, clientHasBeenStoppedErr()
        default:
            return client.discoveryClient.accessAppsByVip(vip)
        }
    }
    return nil, clientNotStartedErr()
}

// AccessAppsBySvip 查询指定svip的可用服务列表
func (client *EurekaClient) AccessAppsBySvip(svip string) (svipApps []*meta.AppInfo, err error) {
    defer func() {
        if rc := recover(); rc != nil {
            err = errors.New(fmt.Sprintf("AccessAppsBySvip, recover error: %v", rc))
        }
        if err != nil {
            client.logger.Errorf("AccessAppsBySvip, FAILED >>> error: %v", err)
        }
        if err == nil {
            client.logger.Tracef("AccessAppsBySvip, OK >>> svipApps: %v", SummaryApps(svipApps))
        }
    }()
    client.logger.Tracef("AccessAppsBySvip, PARAMS >>> svip: %v", svip)
    if client.ctx != nil {
        select {
        case <-client.ctx.Done():
            return nil, clientHasBeenStoppedErr()
        default:
            return client.discoveryClient.accessAppsBySvip(svip)
        }
    }
    return nil, clientNotStartedErr()
}

// AccessInstancesByVip 查询指定vip的可用服务实例列表
func (client *EurekaClient) AccessInstancesByVip(vip string) (instances []*meta.InstanceInfo, err error) {
    defer func() {
        if rc := recover(); rc != nil {
            err = errors.New(fmt.Sprintf("AccessInstancesByVip, recover error: %v", rc))
        }
        if err != nil {
            client.logger.Errorf("AccessInstancesByVip, FAILED >>> error: %v", err)
        }
        if err == nil {
            client.logger.Tracef("AccessInstancesByVip, OK >>> instances: %v", SummaryInstances(instances))
        }
    }()
    client.logger.Tracef("AccessInstancesByVip, PARAMS >>> vip: %v", vip)
    if client.ctx != nil {
        select {
        case <-client.ctx.Done():
            return nil, clientHasBeenStoppedErr()
        default:
            return client.discoveryClient.accessInstancesByVip(vip)
        }
    }
    return nil, clientNotStartedErr()
}

// AccessInstancesBySvip 查询指定svip的可用服务实例列表
func (client *EurekaClient) AccessInstancesBySvip(svip string) (instances []*meta.InstanceInfo, err error) {
    defer func() {
        if rc := recover(); rc != nil {
            err = errors.New(fmt.Sprintf("AccessInstancesBySvip, recover error: %v", rc))
        }
        if err != nil {
            client.logger.Errorf("AccessInstancesBySvip, FAILED >>> error: %v", err)
        }
        if err == nil {
            client.logger.Tracef("AccessInstancesBySvip, OK >>> instances: %v", SummaryInstances(instances))
        }
    }()
    client.logger.Tracef("AccessInstancesBySvip, PARAMS >>> svip: %v", svip)
    if client.ctx != nil {
        select {
        case <-client.ctx.Done():
            return nil, clientHasBeenStoppedErr()
        default:
            return client.discoveryClient.accessInstancesBySvip(svip)
        }
    }
    return nil, clientNotStartedErr()
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
