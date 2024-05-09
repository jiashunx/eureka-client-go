package http

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
    HttpRequest  *http.Request
    HttpResponse *http.Response
    Body         string
    Request      *EurekaRequest
    Error        error
    Responses    []*EurekaResponse
}

// RegisterResponse 注册新服务请求响应
type RegisterResponse struct {
    Response   *EurekaResponse
    StatusCode int
    Error      error
}

// UnRegisterResponse 取消注册服务请求响应
type UnRegisterResponse struct {
    Response   *EurekaResponse
    StatusCode int
    Error      error
}

// HeartbeatResponse 发送服务心跳请求响应
type HeartbeatResponse struct {
    Response   *EurekaResponse
    StatusCode int
    Error      error
}

// QueryAppsResponse 查询所有服务列表请求响应
type QueryAppsResponse struct {
    Response   *EurekaResponse
    StatusCode int
    Error      error
    Apps       []*meta.AppInfo
}

// QueryAppResponse 查询指定appName的服务实例列表请求响应
type QueryAppResponse struct {
    Response   *EurekaResponse
    StatusCode int
    Error      error
    Instances  []*meta.InstanceInfo
}

// QueryAppInstanceResponse 查询指定appName&InstanceId服务实例请求响应
type QueryAppInstanceResponse struct {
    Response   *EurekaResponse
    StatusCode int
    Error      error
    Instance   *meta.InstanceInfo
}

// QueryInstanceResponse 查询指定InstanceId服务实例请求响应
type QueryInstanceResponse struct {
    Response   *EurekaResponse
    StatusCode int
    Error      error
    Instance   *meta.InstanceInfo
}

// ChangeStatusResponse 变更服务状态请求响应
type ChangeStatusResponse struct {
    Response   *EurekaResponse
    StatusCode int
    Error      error
}

// ModifyMetadataResponse 变更元数据请求响应
type ModifyMetadataResponse struct {
    Response   *EurekaResponse
    StatusCode int
    Error      error
}

// QueryVipAppsResponse 查询指定IP下的服务列表请求响应
type QueryVipAppsResponse struct {
    Response   *EurekaResponse
    StatusCode int
    Error      error
    Apps       []*meta.AppInfo
}

// QuerySvipAppsResponse 查询指定安全IP下的服务列表请求响应
type QuerySvipAppsResponse struct {
    Response   *EurekaResponse
    StatusCode int
    Error      error
    Apps       []*meta.AppInfo
}
