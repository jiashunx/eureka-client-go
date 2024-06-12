package client

import (
    "context"
    "errors"
    "fmt"
    "github.com/google/uuid"
    "github.com/jiashunx/eureka-client-go/log"
    "github.com/jiashunx/eureka-client-go/meta"
    "strconv"
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
    httpClient      *HttpClient
    registryClient  *RegistryClient
    discoveryClient *DiscoveryClient
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
            response.Error = errors.New(fmt.Sprintf("EurekaClient.StartWithCtx, recover error: %v", rc))
        }
        if response.Error != nil {
            client.logger.Errorf("EurekaClient.StartWithCtx, FAILED >>> error: %v", response.Error)
        }
        if response.Error == nil {
            client.logger.Tracef("EurekaClient.StartWithCtx, OK")
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
    subCtx := context.WithValue(client.ctx, eurekaClientUUID, client.UUID)
    if response = client.registryClient.start(subCtx); response.Error != nil {
        client.logger.Errorf("EurekaClient.StartWithCtx, failed to start registry client, try to stop eureka client")
        client.Stop()
        client.ctxCancel()
        return response
    }
    if response = client.discoveryClient.start(subCtx); response.Error != nil {
        client.logger.Errorf("EurekaClient.StartWithCtx, failed to start discovery client, try to stop eureka client")
        client.Stop()
        client.ctxCancel()
        return response
    }
    return &CommonResponse{Error: nil}
}

// Stop 关闭eureka客户端（方法执行成功后才关闭）
func (client *EurekaClient) Stop() *CommonResponse {
    ret, err := client.exec("Stop", func(params ...any) (any, error) {
        response := client.registryClient.UnRegister()
        if response.Error == nil {
            client.ctxCancel()
            return response, nil
        }
        return nil, response.Error
    })
    if err != nil {
        return &CommonResponse{Error: err}
    }
    return ret.(*CommonResponse)
}

// ForceStop 强行关闭eureka客户端
func (client *EurekaClient) ForceStop() {
    if client.ctx != nil {
        select {
        case <-client.ctx.Done():
            break
        default:
            defer client.ctxCancel()
            client.logger.Tracef("EurekaClient.ForceStop, try to stop eureka client")
            response := client.registryClient.UnRegister()
            if response.Error != nil {
                client.logger.Tracef("EurekaClient.ForceStop, failed to unRegister, error: %v", response.Error)
            }
        }
    }
    client.logger.Tracef("EurekaClient.ForceStop, OK")
}

// ChangeStatus 变更服务状态
func (client *EurekaClient) ChangeStatus(status meta.InstanceStatus) *CommonResponse {
    ret, err := client.exec("ChangeStatus", func(params ...any) (any, error) {
        return client.registryClient.ChangeStatus(params[0].(meta.InstanceStatus)), nil
    }, status)
    if err != nil {
        return &CommonResponse{Error: err}
    }
    return ret.(*CommonResponse)
}

// ChangeMetadata 变更元数据
func (client *EurekaClient) ChangeMetadata(metadata map[string]string) *CommonResponse {
    ret, err := client.exec("ChangeMetadata", func(params ...any) (any, error) {
        return client.registryClient.ChangeMetadata(params[0].(map[string]string)), nil
    }, metadata)
    if err != nil {
        return &CommonResponse{Error: err}
    }
    return ret.(*CommonResponse)
}

// AccessApp 查询可用服务信息
func (client *EurekaClient) AccessApp(appName string) (*meta.AppInfo, error) {
    ret, err := client.exec("AccessApp", func(params ...any) (any, error) {
        return client.discoveryClient.FilterApp(client.discoveryClient.Apps, params[0].(string))
    }, appName)
    if err != nil {
        return nil, err
    }
    return ret.(*meta.AppInfo), nil
}

// AccessAppsByVip 查询指定vip的可用服务列表
func (client *EurekaClient) AccessAppsByVip(vip string) ([]*meta.AppInfo, error) {
    ret, err := client.exec("AccessAppsByVip", func(params ...any) (any, error) {
        return client.discoveryClient.FilterAppsByVip(client.discoveryClient.Apps, params[0].(string))
    }, vip)
    if err != nil {
        return nil, err
    }
    return ret.([]*meta.AppInfo), nil
}

// AccessAppInstanceByVip 查询指定vip的可用服务实例（随机选择）
func (client *EurekaClient) AccessAppInstanceByVip(vip string) (*meta.InstanceInfo, error) {
    ret, err := client.exec("AccessAppInstanceByVip", func(params ...any) (any, error) {
        return client.discoveryClient.FilterAppInstanceByVip(client.discoveryClient.Apps, params[0].(string))
    }, vip)
    if err != nil {
        return nil, err
    }
    return ret.(*meta.InstanceInfo), nil
}

// AccessAppsBySvip 查询指定svip的可用服务列表
func (client *EurekaClient) AccessAppsBySvip(svip string) ([]*meta.AppInfo, error) {
    ret, err := client.exec("AccessAppsBySvip", func(params ...any) (any, error) {
        return client.discoveryClient.FilterAppsBySvip(client.discoveryClient.Apps, params[0].(string))
    }, svip)
    if err != nil {
        return nil, err
    }
    return ret.([]*meta.AppInfo), nil
}

// AccessAppInstanceBySvip 查询指定svip的可用服务实例列表（随机选择）
func (client *EurekaClient) AccessAppInstanceBySvip(svip string) (*meta.InstanceInfo, error) {
    ret, err := client.exec("AccessAppInstanceBySvip", func(params ...any) (any, error) {
        return client.discoveryClient.FilterAppInstanceBySvip(client.discoveryClient.Apps, params[0].(string))
    }, svip)
    if err != nil {
        return nil, err
    }
    return ret.(*meta.InstanceInfo), nil
}

// AccessInstancesByVip 查询指定vip的可用服务实例列表
func (client *EurekaClient) AccessInstancesByVip(vip string) ([]*meta.InstanceInfo, error) {
    ret, err := client.exec("AccessInstancesByVip", func(params ...any) (any, error) {
        return client.discoveryClient.FilterInstancesByVip(client.discoveryClient.Apps, params[0].(string))
    }, vip)
    if err != nil {
        return nil, err
    }
    return ret.([]*meta.InstanceInfo), nil
}

// AccessInstanceByVip 查询指定vip的可用服务实例（随机选择）
func (client *EurekaClient) AccessInstanceByVip(vip string) (*meta.InstanceInfo, error) {
    ret, err := client.exec("AccessInstanceByVip", func(params ...any) (any, error) {
        return client.discoveryClient.FilterInstanceByVip(client.discoveryClient.Apps, params[0].(string))
    }, vip)
    if err != nil {
        return nil, err
    }
    return ret.(*meta.InstanceInfo), nil
}

// AccessInstancesBySvip 查询指定svip的可用服务实例列表
func (client *EurekaClient) AccessInstancesBySvip(svip string) ([]*meta.InstanceInfo, error) {
    ret, err := client.exec("AccessInstancesBySvip", func(params ...any) (any, error) {
        return client.discoveryClient.FilterInstancesBySvip(client.discoveryClient.Apps, params[0].(string))
    }, svip)
    if err != nil {
        return nil, err
    }
    return ret.([]*meta.InstanceInfo), nil
}

// AccessInstanceBySvip 查询指定svip的可用服务实例列表（随机选择）
func (client *EurekaClient) AccessInstanceBySvip(svip string) (*meta.InstanceInfo, error) {
    ret, err := client.exec("AccessInstanceBySvip", func(params ...any) (any, error) {
        return client.discoveryClient.FilterInstanceBySvip(client.discoveryClient.Apps, params[0].(string))
    }, svip)
    if err != nil {
        return nil, err
    }
    return ret.(*meta.InstanceInfo), nil
}

// exec 处理并返回（同步检查当前客户端运行状态状态）
func (client *EurekaClient) exec(name string, r func(params ...any) (any, error), params ...any) (ret any, err error) {
    defer func() {
        if rc := recover(); rc != nil {
            err = errors.New(fmt.Sprintf("EurekaClient.%s, recover error: %v", name, rc))
        }
        if err != nil {
            client.logger.Errorf("EurekaClient.%s, FAILED >>> error: %v", name, err)
        }
        if err == nil {
            client.logger.Tracef("EurekaClient.%s, OK >>> ret: %v", name, ret)
        }
    }()
    if len(params) > 0 {
        sp := make([]any, 0)
        sp = append(sp, name)
        sl := make([]string, 0)
        for idx, param := range params {
            sl = append(sl, "arg"+strconv.Itoa(idx)+": %v")
            sp = append(sp, param)
        }
        client.logger.Tracef("EurekaClient.%s, PARAMS >>> "+strings.Join(sl, ", "), sp...)
    }
    if client.ctx != nil {
        select {
        case <-client.ctx.Done():
            return nil, clientHasBeenStoppedErr()
        default:
            return r(params...)
        }
    }
    return nil, clientNotStartedErr()
}

// RegistryClient 获取与eureka通讯的 *RegistryClient
func (client *EurekaClient) RegistryClient() *RegistryClient {
    return client.registryClient
}

// DiscoveryClient 获取与eureka通讯的 *DiscoveryClient
func (client *EurekaClient) DiscoveryClient() *DiscoveryClient {
    return client.discoveryClient
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
    client.httpClient.Logger = logger
    client.registryClient.Logger = logger
    client.discoveryClient.Logger = logger
    return nil
}

// GetLogger 获取客户端日志对象
func (client *EurekaClient) GetLogger() log.Logger {
    return client.logger
}

// NewEurekaClient 根据 *meta.EurekaConfig 创建eureka客户端
func NewEurekaClient(config *meta.EurekaConfig) (client *EurekaClient, err error) {
    return NewEurekaClientWithOptions(config, nil)
}

// NewEurekaClientWithOptions 根据 *meta.EurekaConfig 及 *EurekaConfigOptions 创建eureka客户端
func NewEurekaClientWithOptions(config *meta.EurekaConfig, options *EurekaConfigOptions) (client *EurekaClient, err error) {
    defer func() {
        if rc := recover(); rc != nil {
            err = errors.New(fmt.Sprintf("NewEurekaClient, recover error: %v", rc))
        }
    }()
    if config == nil {
        panic("EurekaConfig is nil")
    }
    if options == nil {
        options = &EurekaConfigOptions{}
    }
    newConfig := &meta.EurekaConfig{
        InstanceConfig: config.InstanceConfig,
        ClientConfig:   config.ClientConfig,
    }
    if err = newConfig.Check(); err != nil {
        return nil, err
    }
    logger := log.DefaultLogger()
    httpClient := &HttpClient{Logger: logger}
    return &EurekaClient{
        UUID:       strings.ReplaceAll(uuid.New().String(), "-", ""),
        config:     newConfig,
        rootCtx:    nil,
        ctx:        nil,
        ctxCancel:  nil,
        httpClient: httpClient,
        registryClient: &RegistryClient{
            HttpClient:    httpClient,
            Config:        newConfig,
            Logger:        logger,
            HeartbeatFunc: options.HeartbeatFunc,
        },
        discoveryClient: &DiscoveryClient{
            HttpClient: httpClient,
            Config:     newConfig,
            Logger:     logger,
            Apps:       make(map[string][]*meta.AppInfo),
        },
        logger: logger,
    }, nil
}
