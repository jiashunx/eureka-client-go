package http

import (
    "encoding/json"
    "errors"
    "fmt"
    "github.com/jiashunx/eureka-client-go/meta"
    "math"
    "net/http"
    "net/url"
    "strings"
    "time"
)

// DoRequest 与eureka server通讯处理
func DoRequest(server *meta.EurekaServer, expect int, method string, uri string, payload []byte) (*http.Response, error) {
    var request *http.Request
    var response *http.Response
    var err error
    // 遍历eureka server服务地址，循环发请求直至成功
    for _, su := range strings.Split(server.ServiceUrl, ",") {
        method = strings.TrimSpace(method)
        var URL *url.URL
        URL, err = url.Parse(su)
        if err != nil {
            continue
        }
        nu := URL.Scheme + "://" + URL.Hostname() + ":" + URL.Port() + URL.Path + strings.TrimSpace(uri)
        if URL.Port() == "" {
            nu = URL.Scheme + "://" + URL.Hostname() + URL.Path + strings.TrimSpace(uri)
        }
        body := ""
        if payload != nil {
            body = strings.TrimSpace(string(payload))
        }
        request, err = http.NewRequest(method, nu, strings.NewReader(body))
        if err != nil {
            continue
        }
        if URL.User != nil && URL.User.String() != "" {
            password, _ := URL.User.Password()
            request.SetBasicAuth(URL.User.Username(), password)
        } else if server.Username != "" && server.Password != "" {
            request.SetBasicAuth(server.Username, server.Password)
        }
        request.Header.Set("Accept", "application/json")
        if body != "" {
            request.Header.Set("Content-Type", "application/json")
        }
        client := http.DefaultClient
        if server.ReadTimeoutSeconds > 0 || server.ConnectTimeoutSeconds > 0 {
            seconds := time.Duration(int64(math.Max(float64(server.ReadTimeoutSeconds), float64(server.ConnectTimeoutSeconds))))
            client = &http.Client{Timeout: seconds * time.Second}
        }
        response, err = client.Do(request)
        if err != nil {
            continue
        }
        if response.StatusCode == expect {
            return response, nil
        }
    }
    if err != nil {
        return nil, err
    }
    if response != nil {
        return response, nil
    }
    return nil, errors.New("无可用serviceUrl")
}

// Register 注册新服务
func Register(server *meta.EurekaServer, instance *meta.InstanceInfo) (int, error) {
    err := instance.Check()
    if err != nil {
        return 0, err
    }
    body := make(map[string]*meta.InstanceInfo)
    body["instance"] = instance
    payload, err := json.Marshal(body)
    if err != nil {
        return 0, err
    }
    response, err := DoRequest(server, 204, "POST", fmt.Sprintf("/apps/%s", instance.AppName), payload)
    if err != nil {
        return 0, err
    }
    return response.StatusCode, nil
}

// SimpleRegister 注册新服务
func SimpleRegister(serviceUrl string, instance *meta.InstanceInfo) (int, error) {
    return Register(&meta.EurekaServer{ServiceUrl: serviceUrl}, instance)
}

// UnRegister 取消注册服务
func UnRegister(server *meta.EurekaServer, appName, instanceId string) (int, error) {
    response, err := DoRequest(server, 200, "DELETE", fmt.Sprintf("/apps/%s/%s", appName, instanceId), nil)
    if err != nil {
        return 0, err
    }
    return response.StatusCode, nil
}

// SimpleUnRegister 取消注册服务
func SimpleUnRegister(serviceUrl, appName, instanceId string) (int, error) {
    return UnRegister(&meta.EurekaServer{ServiceUrl: serviceUrl}, appName, instanceId)
}

// Heartbeat 发送服务心跳
func Heartbeat(server *meta.EurekaServer, appName, instanceId string) (int, error) {
    response, err := DoRequest(server, 200, "PUT", fmt.Sprintf("/apps/%s/%s", appName, instanceId), nil)
    if err != nil {
        return 0, err
    }
    return response.StatusCode, nil
}

// SimpleHeartbeat 发送服务心跳
func SimpleHeartbeat(serviceUrl, appName, instanceId string) (int, error) {
    return Heartbeat(&meta.EurekaServer{ServiceUrl: serviceUrl}, appName, instanceId)
}

// QueryApps 查询所有服务
func QueryApps(server *meta.EurekaServer) ([]*meta.AppInfo, error) {
    return nil, nil
}

// SimpleQueryApps 查询所有服务
func SimpleQueryApps(serviceUrl string) ([]*meta.AppInfo, error) {
    return QueryApps(&meta.EurekaServer{ServiceUrl: serviceUrl})
}

// QueryApp 查询指定appName的服务列表
func QueryApp(server *meta.EurekaServer, appName string) ([]*meta.InstanceInfo, error) {
    return nil, nil
}

// SimpleQueryApp 查询指定appName的服务列表
func SimpleQueryApp(serviceUrl, appName string) ([]*meta.InstanceInfo, error) {
    return QueryApp(&meta.EurekaServer{ServiceUrl: serviceUrl}, appName)
}

// QueryAppInstance 查询指定appName&InstanceId
func QueryAppInstance(server *meta.EurekaServer, appName, instanceId string) (*meta.InstanceInfo, error) {
    return nil, nil
}

// SimpleQueryAppInstance 查询指定appName&InstanceId
func SimpleQueryAppInstance(serviceUrl, appName, instanceId string) (*meta.InstanceInfo, error) {
    return QueryAppInstance(&meta.EurekaServer{ServiceUrl: serviceUrl}, appName, instanceId)
}

// QueryInstance 查询指定InstanceId服务列表
func QueryInstance(server *meta.EurekaServer, instanceId string) (*meta.InstanceInfo, error) {
    return nil, nil
}

// SimpleQueryInstance 查询指定InstanceId服务列表
func SimpleQueryInstance(serviceUrl, instanceId string) (*meta.InstanceInfo, error) {
    return QueryInstance(&meta.EurekaServer{ServiceUrl: serviceUrl}, instanceId)
}

// ChangeStatus 变更服务状态
func ChangeStatus(server *meta.EurekaServer, appName, instanceId string, status meta.InstanceStatus) (int, error) {
    return 0, nil
}

// SimpleChangeStatus 变更服务状态
func SimpleChangeStatus(serviceUrl, appName, instanceId string, status meta.InstanceStatus) (int, error) {
    return ChangeStatus(&meta.EurekaServer{ServiceUrl: serviceUrl}, appName, instanceId, status)
}

// ModifyMetadata 变更元数据
func ModifyMetadata(server *meta.EurekaServer, appName, instanceId, key, value string) (int, error) {
    return 0, nil
}

// SimpleModifyMetadata 变更元数据
func SimpleModifyMetadata(serviceUrl, appName, instanceId, key, value string) (int, error) {
    return ModifyMetadata(&meta.EurekaServer{ServiceUrl: serviceUrl}, appName, instanceId, key, value)
}

// QueryVipApps 查询指定IP下的服务列表
func QueryVipApps(server *meta.EurekaServer, vipAddress string) ([]*meta.AppInfo, error) {
    return nil, nil
}

// SimpleQueryVipApps 查询指定IP下的服务列表
func SimpleQueryVipApps(serviceUrl, vipAddress string) ([]*meta.AppInfo, error) {
    return QueryVipApps(&meta.EurekaServer{ServiceUrl: serviceUrl}, vipAddress)
}

// QuerySvipApps 查询指定安全IP下的服务列表
func QuerySvipApps(server *meta.EurekaServer, svipAddress string) ([]*meta.AppInfo, error) {
    return nil, nil
}

// SimpleQuerySvipApps 查询指定安全IP下的服务列表
func SimpleQuerySvipApps(serviceUrl, svipAddress string) ([]*meta.AppInfo, error) {
    return QuerySvipApps(&meta.EurekaServer{ServiceUrl: serviceUrl}, svipAddress)
}
