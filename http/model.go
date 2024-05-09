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
    Request      *EurekaRequest
    Error        error
    Responses    []*EurekaResponse
}

// CommonResponse 通用请求响应模型
type CommonResponse struct {
    Response   *EurekaResponse
    StatusCode int
    Error      error
}

// RegisterResponse 注册新服务请求响应
type RegisterResponse struct {
    *CommonResponse
}

// UnRegisterResponse 取消注册服务请求响应
type UnRegisterResponse struct {
    *CommonResponse
}

// HeartbeatResponse 发送服务心跳请求响应
type HeartbeatResponse struct {
    *CommonResponse
}

// QueryAppsResponse 查询所有服务列表请求响应
type QueryAppsResponse struct {
    *CommonResponse
    Apps []*meta.AppInfo
}

// QueryAppResponse 查询指定appName的服务实例列表请求响应
type QueryAppResponse struct {
    *CommonResponse
    Instances []*meta.InstanceInfo
}

// QueryAppInstanceResponse 查询指定appName&InstanceId服务实例请求响应
type QueryAppInstanceResponse struct {
    *CommonResponse
    Instance *meta.InstanceInfo
}

// QueryInstanceResponse 查询指定InstanceId服务实例请求响应
type QueryInstanceResponse struct {
    *CommonResponse
    Instance *meta.InstanceInfo
}

// ChangeStatusResponse 变更服务状态请求响应
type ChangeStatusResponse struct {
    *CommonResponse
}

// ModifyMetadataResponse 变更元数据请求响应
type ModifyMetadataResponse struct {
    *CommonResponse
}

// QueryVipAppsResponse 查询指定IP下的服务列表请求响应
type QueryVipAppsResponse struct {
    *CommonResponse
    Apps []*meta.AppInfo
}

// QuerySvipAppsResponse 查询指定安全IP下的服务列表请求响应
type QuerySvipAppsResponse struct {
    *CommonResponse
    Apps []*meta.AppInfo
}
