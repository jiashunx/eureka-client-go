package http

import (
    "encoding/json"
    "errors"
    "fmt"
    "github.com/jiashunx/eureka-client-go/meta"
    "io/ioutil"
    "math"
    "net/http"
    "net/url"
    "strings"
    "time"
)

// DoRequest 与eureka server通讯处理
func DoRequest(expect int, server *meta.EurekaServer, method string, uri string, payload []byte) *EurekaResponse {
    var responses = make([]*EurekaResponse, 0)
    // 遍历eureka server服务地址，循环发请求直至成功
    for _, serviceUrl := range strings.Split(server.ServiceUrl, ",") {
        request := &EurekaRequest{
            ServiceUrl:   serviceUrl,
            AuthUsername: "",
            AuthPassword: "",
            Method:       method,
            RequestUrl:   "",
            RequestUri:   uri,
            Body:         "",
        }
        response := &EurekaResponse{Request: request}
        if payload != nil {
            request.Body = string(payload)
        }
        URL, err := url.Parse(serviceUrl)
        if err != nil {
            response.Error = err
            responses = append(responses, response)
            continue
        }
        if URL.User != nil && URL.User.String() != "" {
            password, _ := URL.User.Password()
            request.AuthUsername = URL.User.Username()
            request.AuthPassword = password
        } else if server.Username != "" {
            request.AuthUsername = server.Username
            request.AuthPassword = server.Password
        }
        request.RequestUrl = URL.Scheme + "://" + URL.Hostname() + ":" + URL.Port() + URL.Path + strings.TrimSpace(uri)
        if URL.Port() == "" {
            request.RequestUrl = URL.Scheme + "://" + URL.Hostname() + URL.Path + strings.TrimSpace(uri)
        }
        httpRequest, err := http.NewRequest(request.Method, request.RequestUrl, strings.NewReader(request.Body))
        response.HttpRequest = httpRequest
        response.Error = err
        if response.Error != nil {
            responses = append(responses, response)
            continue
        }
        if request.AuthUsername != "" {
            httpRequest.SetBasicAuth(request.AuthUsername, request.AuthPassword)
        }
        httpRequest.Header.Set("Accept", "application/json")
        if request.Body != "" {
            httpRequest.Header.Set("Content-Type", "application/json")
        }
        httpClient := http.DefaultClient
        if server.ReadTimeoutSeconds > 0 || server.ConnectTimeoutSeconds > 0 {
            seconds := time.Duration(int64(math.Max(float64(server.ReadTimeoutSeconds), float64(server.ConnectTimeoutSeconds))))
            httpClient = &http.Client{Timeout: seconds * time.Second}
        }
        httpResponse, err := httpClient.Do(httpRequest)
        response.HttpResponse = httpResponse
        response.Error = err
        responses = append(responses, response)
        if response.Error == nil {
            var body []byte
            body, response.Error = ioutil.ReadAll(httpResponse.Body)
            if response.Error == nil {
                response.Body = string(body)
            }
            _ = httpResponse.Body.Close()
        }
        if response.Error == nil && httpResponse.StatusCode == expect {
            break
        }
    }
    for _, r := range responses {
        r.Responses = responses
    }
    if len(responses) == 0 {
        return &EurekaResponse{
            Request:      nil,
            HttpResponse: nil,
            Error:        errors.New("无可用serviceUrl"),
            Responses:    responses,
        }
    }
    response := responses[len(responses)-1]
    if response.Error == nil && response.HttpResponse.StatusCode != expect {
        response.Error = errors.New(fmt.Sprintf("请求响应码错误, 预期: %d, 实际: %d", expect, response.HttpResponse.StatusCode))
    }
    return response
}

// Register 注册新服务
func Register(server *meta.EurekaServer, instance *meta.InstanceInfo) *RegisterResponse {
    ret := &RegisterResponse{&CommonResponse{}}
    ret.Error = instance.Check()
    if ret.Error != nil {
        return ret
    }
    body := make(map[string]*meta.InstanceInfo)
    body["instance"] = instance
    var payload []byte
    payload, ret.Error = json.Marshal(body)
    if ret.Error != nil {
        return ret
    }
    requestUrl := fmt.Sprintf("/apps/%s", instance.AppName)
    AssignCommonResponse(ret.CommonResponse, DoRequest(204, server, "POST", requestUrl, payload))
    return ret
}

// SimpleRegister 注册新服务
func SimpleRegister(serviceUrl string, instance *meta.InstanceInfo) *RegisterResponse {
    return Register(&meta.EurekaServer{ServiceUrl: serviceUrl}, instance)
}

// UnRegister 取消注册服务
func UnRegister(server *meta.EurekaServer, appName, instanceId string) *UnRegisterResponse {
    ret := &UnRegisterResponse{&CommonResponse{}}
    requestUrl := fmt.Sprintf("/apps/%s/%s", appName, instanceId)
    AssignCommonResponse(ret.CommonResponse, DoRequest(200, server, "DELETE", requestUrl, nil))
    return ret
}

// SimpleUnRegister 取消注册服务
func SimpleUnRegister(serviceUrl, appName, instanceId string) *UnRegisterResponse {
    return UnRegister(&meta.EurekaServer{ServiceUrl: serviceUrl}, appName, instanceId)
}

// Heartbeat 发送服务心跳
func Heartbeat(server *meta.EurekaServer, appName, instanceId string) *HeartbeatResponse {
    ret := &HeartbeatResponse{&CommonResponse{}}
    requestUrl := fmt.Sprintf("/apps/%s/%s", appName, instanceId)
    AssignCommonResponse(ret.CommonResponse, DoRequest(200, server, "PUT", requestUrl, nil))
    return ret
}

// SimpleHeartbeat 发送服务心跳
func SimpleHeartbeat(serviceUrl, appName, instanceId string) *HeartbeatResponse {
    return Heartbeat(&meta.EurekaServer{ServiceUrl: serviceUrl}, appName, instanceId)
}

// QueryApps 查询所有服务列表
func QueryApps(server *meta.EurekaServer) *QueryAppsResponse {
    ret := &QueryAppsResponse{&CommonResponse{}, make([]*meta.AppInfo, 0)}
    AssignCommonResponse(ret.CommonResponse, DoRequest(200, server, "GET", "/apps", nil))
    if ret.Error != nil {
        return ret
    }
    return ret
}

// SimpleQueryApps 查询所有服务列表
func SimpleQueryApps(serviceUrl string) *QueryAppsResponse {
    return QueryApps(&meta.EurekaServer{ServiceUrl: serviceUrl})
}

// QueryApp 查询指定appName的服务实例列表
func QueryApp(server *meta.EurekaServer, appName string) *QueryAppResponse {
    return nil
}

// SimpleQueryApp 查询指定appName的服务实例列表
func SimpleQueryApp(serviceUrl, appName string) *QueryAppResponse {
    return QueryApp(&meta.EurekaServer{ServiceUrl: serviceUrl}, appName)
}

// QueryAppInstance 查询指定appName&InstanceId服务实例
func QueryAppInstance(server *meta.EurekaServer, appName, instanceId string) *QueryAppInstanceResponse {
    return nil
}

// SimpleQueryAppInstance 查询指定appName&InstanceId服务实例
func SimpleQueryAppInstance(serviceUrl, appName, instanceId string) *QueryAppInstanceResponse {
    return QueryAppInstance(&meta.EurekaServer{ServiceUrl: serviceUrl}, appName, instanceId)
}

// QueryInstance 查询指定InstanceId服务实例
func QueryInstance(server *meta.EurekaServer, instanceId string) *QueryInstanceResponse {
    return nil
}

// SimpleQueryInstance 查询指定InstanceId服务实例
func SimpleQueryInstance(serviceUrl, instanceId string) *QueryInstanceResponse {
    return QueryInstance(&meta.EurekaServer{ServiceUrl: serviceUrl}, instanceId)
}

// ChangeStatus 变更服务状态
func ChangeStatus(server *meta.EurekaServer, appName, instanceId string, status meta.InstanceStatus) *ChangeStatusResponse {
    ret := &ChangeStatusResponse{&CommonResponse{}}
    requestUrl := fmt.Sprintf("/apps/%s/%s/status?value=%s", appName, instanceId, string(status))
    AssignCommonResponse(ret.CommonResponse, DoRequest(200, server, "PUT", requestUrl, nil))
    return ret
}

// SimpleChangeStatus 变更服务状态
func SimpleChangeStatus(serviceUrl, appName, instanceId string, status meta.InstanceStatus) *ChangeStatusResponse {
    return ChangeStatus(&meta.EurekaServer{ServiceUrl: serviceUrl}, appName, instanceId, status)
}

// ModifyMetadata 变更元数据
func ModifyMetadata(server *meta.EurekaServer, appName, instanceId, key, value string) *ModifyMetadataResponse {
    ret := &ModifyMetadataResponse{&CommonResponse{}}
    requestUrl := fmt.Sprintf("/apps/%s/%s/metadata?%s=%s", appName, instanceId, key, value)
    AssignCommonResponse(ret.CommonResponse, DoRequest(200, server, "PUT", requestUrl, nil))
    return ret
}

// SimpleModifyMetadata 变更元数据
func SimpleModifyMetadata(serviceUrl, appName, instanceId, key, value string) *ModifyMetadataResponse {
    return ModifyMetadata(&meta.EurekaServer{ServiceUrl: serviceUrl}, appName, instanceId, key, value)
}

// QueryVipApps 查询指定IP下的服务列表
func QueryVipApps(server *meta.EurekaServer, vipAddress string) *QueryVipAppsResponse {
    return nil
}

// SimpleQueryVipApps 查询指定IP下的服务列表
func SimpleQueryVipApps(serviceUrl, vipAddress string) *QueryVipAppsResponse {
    return QueryVipApps(&meta.EurekaServer{ServiceUrl: serviceUrl}, vipAddress)
}

// QuerySvipApps 查询指定安全IP下的服务列表
func QuerySvipApps(server *meta.EurekaServer, svipAddress string) *QuerySvipAppsResponse {
    return nil
}

// SimpleQuerySvipApps 查询指定安全IP下的服务列表
func SimpleQuerySvipApps(serviceUrl, svipAddress string) *QuerySvipAppsResponse {
    return QuerySvipApps(&meta.EurekaServer{ServiceUrl: serviceUrl}, svipAddress)
}

// AssignCommonResponse 针对 *CommonResponse 进行赋值处理
func AssignCommonResponse(ret *CommonResponse, response *EurekaResponse) *CommonResponse {
    ret.Response = response
    if ret.Response != nil {
        ret.Error = ret.Response.Error
        if ret.Response.HttpResponse != nil {
            ret.StatusCode = ret.Response.HttpResponse.StatusCode
        }
    }
    return ret
}
