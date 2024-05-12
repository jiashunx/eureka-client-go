package client

import (
    "encoding/json"
    "errors"
    "fmt"
    "github.com/google/uuid"
    "github.com/jiashunx/eureka-client-go/log"
    "github.com/jiashunx/eureka-client-go/meta"
    "io/ioutil"
    "math"
    "net/http"
    "net/url"
    "strings"
    "time"
)

// HttpClient eureka客户端与服务端进行http通讯的客户端模型
type HttpClient struct {
    logger log.Logger
}

// doRequest 与eureka server通讯处理
func (client *HttpClient) doRequest(expect int, server *meta.EurekaServer, method string, uri string, payload []byte) (ret *EurekaResponse) {
    var responses = make([]*EurekaResponse, 0)
    defer func() {
        if rc := recover(); rc != nil {
            responses = append(responses, &EurekaResponse{
                UUID:         strings.ReplaceAll(uuid.New().String(), "-", ""),
                Request:      nil,
                HttpResponse: nil,
                Error:        errors.New(fmt.Sprintf("doRequest, recover error: %v", rc)),
                Responses:    nil,
            })
        }
        for _, r := range responses {
            r.Responses = responses
        }
        ret = responses[len(responses)-1]
        if ret.Error != nil {
            client.logger.Errorf("doRequest, FAILED >>> error: %v", ret.Error)
        }
        if ret.Error == nil {
            client.logger.Tracef("doRequest, OK >>> body: %v", ret.Body)
        }
    }()
    if server == nil {
        panic(errors.New("EurekaServer is nil"))
    }
    client.logger.Tracef("doRequest, PARAMS >>> expect: %d, method: %s, uri: %s, server: %#v", expect, method, uri, server)
    // 遍历eureka server服务地址，循环发请求直至成功
    for idx, serviceUrl := range strings.Split(server.ServiceUrl, ",") {
        serviceUrl = strings.TrimSpace(serviceUrl)
        if serviceUrl == "" {
            continue
        }
        request := &EurekaRequest{
            ServiceUrl:   serviceUrl,
            AuthUsername: "",
            AuthPassword: "",
            Method:       method,
            RequestUrl:   "",
            RequestUri:   uri,
            Body:         "",
        }
        response := &EurekaResponse{
            UUID:    strings.ReplaceAll(uuid.New().String(), "-", ""),
            Request: request,
        }
        if payload != nil {
            request.Body = string(payload)
        }
        URL, err := url.Parse(serviceUrl)
        if err != nil {
            response.Error = err
            responses = append(responses, response)
            client.logger.Tracef("doRequest, failed to parse serviceUrl >> idx: %d, serviceUrl: %s, error: %v", idx, serviceUrl, err)
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
        client.logger.Tracef("doRequest, create request object >>> idx: %d, method: %s, requestUrl: %s, body: %s", idx, method, request.RequestUrl, request.Body)
        httpRequest, err := http.NewRequest(request.Method, request.RequestUrl, strings.NewReader(request.Body))
        response.HttpRequest = httpRequest
        response.Error = err
        if response.Error != nil {
            responses = append(responses, response)
            client.logger.Tracef("doRequest, failed to create request object >>> idx: %d, error: %v", idx, err)
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
        if response.Error == nil {
            if httpResponse.StatusCode == expect {
                break
            }
            response.Error = errors.New(fmt.Sprintf("the http response code is incorrect, expect: %d, actual: %d", expect, response.HttpResponse.StatusCode))
        }
        if response.Error != nil {
            client.logger.Tracef("doRequest, request failed >>> idx: %d, error: %v", idx, err)
        }
    }
    if len(responses) == 0 {
        panic(errors.New("no eureka server service address available"))
    }
    return nil
}

// Register 注册新服务
func (client *HttpClient) Register(server *meta.EurekaServer, instance *meta.InstanceInfo) (ret *CommonResponse) {
    defer func() {
        if rc := recover(); rc != nil {
            ret = &CommonResponse{}
            ret.Error = errors.New(fmt.Sprintf("Register, recover error: %v", rc))
        }
        if ret.Error != nil {
            client.logger.Errorf("Register, FAILED >>> error: %v", ret.Error)
        }
        if ret.Error == nil {
            client.logger.Tracef("Register, OK")
        }
    }()
    client.logger.Tracef("Register, PARAMS >>> server: %v, instance: %v", server, instance)
    if instance == nil {
        return &CommonResponse{Error: errors.New("InstanceInfo is nil")}
    }
    err := instance.Check()
    if err != nil {
        return &CommonResponse{Error: err}
    }
    body := make(map[string]*meta.InstanceInfo)
    body["instance"] = instance
    payload, err := json.Marshal(body)
    if err != nil {
        return &CommonResponse{Error: err}
    }
    requestUrl := fmt.Sprintf("/apps/%s", instance.AppName)
    return client.commonHttp(204, server, "POST", requestUrl, payload)
}

// SimpleRegister 注册新服务
func (client *HttpClient) SimpleRegister(serviceUrl string, instance *meta.InstanceInfo) *CommonResponse {
    return client.Register(&meta.EurekaServer{ServiceUrl: serviceUrl}, instance)
}

// UnRegister 取消注册服务
func (client *HttpClient) UnRegister(server *meta.EurekaServer, appName, instanceId string) *CommonResponse {
    requestUrl := fmt.Sprintf("/apps/%s/%s", appName, instanceId)
    return client.commonHttp(200, server, "DELETE", requestUrl, nil)
}

// SimpleUnRegister 取消注册服务
func (client *HttpClient) SimpleUnRegister(serviceUrl, appName, instanceId string) *CommonResponse {
    return client.UnRegister(&meta.EurekaServer{ServiceUrl: serviceUrl}, appName, instanceId)
}

// Heartbeat 发送服务心跳
func (client *HttpClient) Heartbeat(server *meta.EurekaServer, appName, instanceId string) *CommonResponse {
    requestUrl := fmt.Sprintf("/apps/%s/%s", appName, instanceId)
    return client.commonHttp(200, server, "PUT", requestUrl, nil)
}

// SimpleHeartbeat 发送服务心跳
func (client *HttpClient) SimpleHeartbeat(serviceUrl, appName, instanceId string) *CommonResponse {
    return client.Heartbeat(&meta.EurekaServer{ServiceUrl: serviceUrl}, appName, instanceId)
}

// QueryApps 查询所有服务列表
func (client *HttpClient) QueryApps(server *meta.EurekaServer) *AppsResponse {
    return client.getApps(server, "/apps")
}

// SimpleQueryApps 查询所有服务列表
func (client *HttpClient) SimpleQueryApps(serviceUrl string) *AppsResponse {
    return client.QueryApps(&meta.EurekaServer{ServiceUrl: serviceUrl})
}

// QueryApp 查询指定appName的服务实例列表
func (client *HttpClient) QueryApp(server *meta.EurekaServer, appName string) *InstancesResponse {
    return client.getInstances(server, fmt.Sprintf("/apps/%s", appName))
}

// SimpleQueryApp 查询指定appName的服务实例列表
func (client *HttpClient) SimpleQueryApp(serviceUrl, appName string) *InstancesResponse {
    return client.QueryApp(&meta.EurekaServer{ServiceUrl: serviceUrl}, appName)
}

// QueryAppInstance 查询指定appName&InstanceId服务实例
func (client *HttpClient) QueryAppInstance(server *meta.EurekaServer, appName, instanceId string) *InstanceResponse {
    return client.getInstance(server, fmt.Sprintf("/apps/%s/%s", appName, instanceId))
}

// SimpleQueryAppInstance 查询指定appName&InstanceId服务实例
func (client *HttpClient) SimpleQueryAppInstance(serviceUrl, appName, instanceId string) *InstanceResponse {
    return client.QueryAppInstance(&meta.EurekaServer{ServiceUrl: serviceUrl}, appName, instanceId)
}

// QueryInstance 查询指定InstanceId服务实例
func (client *HttpClient) QueryInstance(server *meta.EurekaServer, instanceId string) *InstanceResponse {
    return client.getInstance(server, fmt.Sprintf("/instances/%s", instanceId))
}

// SimpleQueryInstance 查询指定InstanceId服务实例
func (client *HttpClient) SimpleQueryInstance(serviceUrl, instanceId string) *InstanceResponse {
    return client.QueryInstance(&meta.EurekaServer{ServiceUrl: serviceUrl}, instanceId)
}

// ChangeStatus 变更服务状态
func (client *HttpClient) ChangeStatus(server *meta.EurekaServer, appName, instanceId string, status meta.InstanceStatus) *CommonResponse {
    requestUrl := fmt.Sprintf("/apps/%s/%s/status?value=%s", appName, instanceId, string(status))
    return client.commonHttp(200, server, "PUT", requestUrl, nil)
}

// SimpleChangeStatus 变更服务状态
func (client *HttpClient) SimpleChangeStatus(serviceUrl, appName, instanceId string, status meta.InstanceStatus) *CommonResponse {
    return client.ChangeStatus(&meta.EurekaServer{ServiceUrl: serviceUrl}, appName, instanceId, status)
}

// ModifyMetadata 变更元数据
func (client *HttpClient) ModifyMetadata(server *meta.EurekaServer, appName, instanceId string, metadata map[string]string) *CommonResponse {
    requestUrl := fmt.Sprintf("/apps/%s/%s/metadata?", appName, instanceId)
    if metadata != nil {
        for k, v := range metadata {
            requestUrl = requestUrl + k + "=" + v + "&"
        }
    }
    requestUrl = requestUrl[0:(len(requestUrl) - 2)]
    return client.commonHttp(200, server, "PUT", requestUrl, nil)
}

// SimpleModifyMetadata 变更元数据
func (client *HttpClient) SimpleModifyMetadata(serviceUrl, appName, instanceId string, metadata map[string]string) *CommonResponse {
    return client.ModifyMetadata(&meta.EurekaServer{ServiceUrl: serviceUrl}, appName, instanceId, metadata)
}

// QueryVipApps 查询指定虚拟主机名下的服务列表
func (client *HttpClient) QueryVipApps(server *meta.EurekaServer, vipAddress string) *AppsResponse {
    return client.getApps(server, fmt.Sprintf("/vips/%s", vipAddress))
}

// SimpleQueryVipApps 查询指定虚拟主机名下的服务列表
func (client *HttpClient) SimpleQueryVipApps(serviceUrl, vipAddress string) *AppsResponse {
    return client.QueryVipApps(&meta.EurekaServer{ServiceUrl: serviceUrl}, vipAddress)
}

// QuerySvipApps 查询指定安全虚拟主机名下的服务列表
func (client *HttpClient) QuerySvipApps(server *meta.EurekaServer, svipAddress string) *AppsResponse {
    return client.getApps(server, fmt.Sprintf("/svips/%s", svipAddress))
}

// SimpleQuerySvipApps 查询指定安全虚拟主机名下的服务列表
func (client *HttpClient) SimpleQuerySvipApps(serviceUrl, svipAddress string) *AppsResponse {
    return client.QuerySvipApps(&meta.EurekaServer{ServiceUrl: serviceUrl}, svipAddress)
}

// commonHttp 与eureka server通讯公共方法
func (client *HttpClient) commonHttp(expect int, server *meta.EurekaServer, method string, url string, payload []byte) *CommonResponse {
    ret := &CommonResponse{}
    ret.Response = client.doRequest(expect, server, method, url, payload)
    if ret.Response.Error != nil {
        ret.Error = ret.Response.Error
    }
    if ret.Response.HttpResponse != nil {
        ret.StatusCode = ret.Response.HttpResponse.StatusCode
    }
    return ret
}

// getApps 查询服务列表
func (client *HttpClient) getApps(server *meta.EurekaServer, uri string) (ret *AppsResponse) {
    ret = &AppsResponse{Apps: make([]*meta.AppInfo, 0)}
    defer func() {
        if rc := recover(); rc != nil {
            ret.Error = errors.New(fmt.Sprintf("getApps, recover error: %v", rc))
        }
        if ret.Error != nil {
            client.logger.Errorf("getApps, FAILED >>> error: %v", ret.Error)
        }
        if ret.Error == nil {
            client.logger.Tracef("getApps, OK >>> ret: %v", SummaryApps(ret.Apps))
        }
    }()
    client.logger.Tracef("getApps, PARAMS >>> server: %v, uri: %s", server, uri)
    ret.Response = client.doRequest(200, server, "GET", uri, nil)
    if ret.Response.Error != nil {
        ret.Error = ret.Response.Error
    }
    if ret.Response.HttpResponse != nil {
        ret.StatusCode = ret.Response.HttpResponse.StatusCode
    }
    if ret.Error != nil {
        return ret
    }
    var ii interface{}
    ret.Error = json.Unmarshal([]byte(ret.Response.Body), &ii)
    if ret.Error != nil {
        return ret
    }
    ij := ii.(map[string]interface{})["applications"]
    if ij == nil {
        ret.Error = errors.New("the query yielded no results: 'applications'")
        return ret
    }
    ik := ij.(map[string]interface{})["application"]
    if ik == nil {
        ret.Error = errors.New("the query yielded no results: 'applications'.'application'")
        return ret
    }
    for _, m := range ik.([]interface{}) {
        var data []byte
        data, ret.Error = json.Marshal(m)
        if ret.Error != nil {
            return ret
        }
        var app *meta.AppInfo
        app, ret.Error = meta.ParseAppInfo(data)
        if ret.Error != nil {
            return ret
        }
        ret.Apps = append(ret.Apps, app)
    }
    return ret
}

// getInstances 查询服务实例列表
func (client *HttpClient) getInstances(server *meta.EurekaServer, uri string) (ret *InstancesResponse) {
    ret = &InstancesResponse{Instances: make([]*meta.InstanceInfo, 0)}
    defer func() {
        if rc := recover(); rc != nil {
            ret.Error = errors.New(fmt.Sprintf("getInstances, recover error: %v", rc))
        }
        if ret.Error != nil {
            client.logger.Errorf("getInstances, FAILED >>> error: %v", ret.Error)
        }
        if ret.Error == nil {
            client.logger.Tracef("getInstances, OK >>> ret: %v", SummaryInstances(ret.Instances))
        }
    }()
    client.logger.Tracef("getInstances, PARAMS >>> server: %v, uri: %s", server, uri)
    ret.Response = client.doRequest(200, server, "GET", uri, nil)
    if ret.Response.Error != nil {
        ret.Error = ret.Response.Error
    }
    if ret.Response.HttpResponse != nil {
        ret.StatusCode = ret.Response.HttpResponse.StatusCode
    }
    if ret.Error != nil {
        return ret
    }
    var ii interface{}
    ret.Error = json.Unmarshal([]byte(ret.Response.Body), &ii)
    if ret.Error != nil {
        return ret
    }
    ij := ii.(map[string]interface{})["application"]
    if ij == nil {
        ret.Error = errors.New("the query yielded no results: 'application'")
        return ret
    }
    ik := ij.(map[string]interface{})["instance"]
    if ik == nil {
        ret.Error = errors.New("the query yielded no results: 'application'.'instance'")
        return ret
    }
    for _, m := range ik.([]interface{}) {
        var data []byte
        data, ret.Error = json.Marshal(m)
        if ret.Error != nil {
            return ret
        }
        var instance *meta.InstanceInfo
        instance, ret.Error = meta.ParseInstanceInfo(data)
        if ret.Error != nil {
            return ret
        }
        ret.Instances = append(ret.Instances, instance)
    }
    return ret
}

// getInstance 查询服务实例
func (client *HttpClient) getInstance(server *meta.EurekaServer, uri string) (ret *InstanceResponse) {
    ret = &InstanceResponse{}
    defer func() {
        if rc := recover(); rc != nil {
            ret.Error = errors.New(fmt.Sprintf("getInstance, recover error: %v", rc))
        }
        if ret.Error != nil {
            client.logger.Errorf("getInstance, FAILED >>> error: %v", ret.Error)
        }
        if ret.Error == nil {
            client.logger.Tracef("getInstance, OK >>> ret: %v", SummaryInstance(ret.Instance))
        }
    }()
    client.logger.Tracef("getInstance, PARAMS >>> server: %v, uri: %s", server, uri)
    ret.Response = client.doRequest(200, server, "GET", uri, nil)
    if ret.Response.Error != nil {
        ret.Error = ret.Response.Error
    }
    if ret.Response.HttpResponse != nil {
        ret.StatusCode = ret.Response.HttpResponse.StatusCode
    }
    if ret.Error != nil {
        return ret
    }
    var ii interface{}
    ret.Error = json.Unmarshal([]byte(ret.Response.Body), &ii)
    if ret.Error != nil {
        return ret
    }
    ij := ii.(map[string]interface{})["instance"]
    if ij == nil {
        ret.Error = errors.New("the query yielded no results: 'instance'")
        return ret
    }
    var data []byte
    data, ret.Error = json.Marshal(ij)
    if ret.Error != nil {
        return ret
    }
    ret.Instance, ret.Error = meta.ParseInstanceInfo(data)
    return ret
}
