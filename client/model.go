package client

import (
    "github.com/jiashunx/eureka-client-go/meta"
    "net/http"
)

// EurekaRequest 与eureka server通讯请求模型
type EurekaRequest struct {
    ServiceUrl   string
    AuthUsername string
    AuthPassword string
    Method       string
    RequestUrl   string
    RequestUri   string
    Body         string
}

// EurekaResponse 与eureka server通讯响应（包含同批次所有通讯响应）
type EurekaResponse struct {
    UUID         string
    HttpRequest  *http.Request
    HttpResponse *http.Response
    Body         string
    Request      *EurekaRequest
    Error        error
    Responses    []*EurekaResponse
}

// CommonResponse 通用处理接口请求响应
type CommonResponse struct {
    Response   *EurekaResponse
    StatusCode int
    Error      error
}

// InstanceResponse 服务实例查询接口请求响应
type InstanceResponse struct {
    Response   *EurekaResponse
    StatusCode int
    Error      error
    Instance   *meta.InstanceInfo
}

// InstancesResponse 服务实例列表查询接口请求响应
type InstancesResponse struct {
    Response   *EurekaResponse
    StatusCode int
    Error      error
    Instances  []*meta.InstanceInfo
}

// AppsResponse 服务列表查询接口请求响应
type AppsResponse struct {
    Response   *EurekaResponse
    StatusCode int
    Error      error
    Apps       []*meta.AppInfo
}

// EurekaConfigOptions eureka客户端配置冗余信息（可选）
type EurekaConfigOptions struct {
    // 心跳失败回调, 仅当集成到 EurekaClient 时有效
    HeartbeatFailFunc func(*RegistryClient, *CommonResponse)
}
