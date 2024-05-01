package meta

import (
    "encoding/json"
)

// DataCenterInfo 数据中心
type DataCenterInfo struct {
    Class string `json:"@class"`
    Name  string `json:"name"`
}

// ParseDataCenterInfo 从map中解析数据中心信息
func ParseDataCenterInfo(m map[string]interface{}) (dc *DataCenterInfo, err interface{}) {
    defer func() {
        if rc := recover(); rc != nil {
            dc = nil
            err = rc
        }
    }()
    dc = &DataCenterInfo{}
    dc.Class = m["@class"].(string)
    dc.Name = m["name"].(string)
    return dc, nil
}

// DefaultDataCenterInfo 默认数据中心信息
var DefaultDataCenterInfo = &DataCenterInfo{
    Class: "com.netflix.appinfo.InstanceInfo$DefaultDataCenterInfo",
    Name:  "MyOwn",
}

// LeaseInfo 服务实例租约信息
type LeaseInfo struct {
    RenewalIntervalInSecs int   `json:"renewalIntervalInSecs"`
    DurationInSecs        int   `json:"durationInSecs"`
    RegistrationTimestamp int64 `json:"registrationTimestamp"`
    LastRenewalTimestamp  int64 `json:"lastRenewalTimestamp"`
    EvictionTimestamp     int64 `json:"evictionTimestamp"`
    ServiceUpTimestamp    int64 `json:"serviceUpTimestamp"`
}

// ParseLeaseInfo 从map中解析服务实例租约信息
func ParseLeaseInfo(m map[string]interface{}) (lease *LeaseInfo, err interface{}) {
    defer func() {
        if rc := recover(); rc != nil {
            lease = nil
            err = rc
        }
    }()
    lease = &LeaseInfo{}
    lease.RenewalIntervalInSecs = int(m["renewalIntervalInSecs"].(float64))
    lease.DurationInSecs = int(m["durationInSecs"].(float64))
    lease.RegistrationTimestamp = int64(m["registrationTimestamp"].(float64))
    lease.LastRenewalTimestamp = int64(m["lastRenewalTimestamp"].(float64))
    lease.EvictionTimestamp = int64(m["evictionTimestamp"].(float64))
    lease.ServiceUpTimestamp = int64(m["serviceUpTimestamp"].(float64))
    return lease, nil
}

// PortWrapper 端口信息
type PortWrapper struct {
    Enabled string `json:"@enabled"`
    Port    int    `json:"$"`
}

// ParsePortWrapper 从map中解析端口信息
func ParsePortWrapper(m map[string]interface{}) (wrapper *PortWrapper, err interface{}) {
    defer func() {
        if rc := recover(); rc != nil {
            wrapper = nil
            err = rc
        }
    }()
    wrapper = &PortWrapper{}
    wrapper.Enabled = m["@enabled"].(string)
    wrapper.Port = int(m["$"].(float64))
    return wrapper, nil
}

// InstanceStatus 服务实例状态
type InstanceStatus string

const (
    Up           InstanceStatus = "UP"
    Down         InstanceStatus = "DOWN"
    Starting     InstanceStatus = "STARTING"
    OutOfService InstanceStatus = "OUT_OF_SERVICE"
    Unknown      InstanceStatus = "UNKNOWN"
)

// PortType 端口类型
type PortType string

const (
    Secure   PortType = "SECURE"
    UnSecure PortType = "UNSECURE"
)

const (
    HttpProtocol  = "http://"
    HttpsProtocol = "https://"
)

// ActionType 实例操作类型
type ActionType string

const (
    Added    ActionType = "ADDED"
    Modified ActionType = "MODIFIED"
    Deleted  ActionType = "DELETED"
)

// InstanceInfo 服务实例信息
type InstanceInfo struct {
    InstanceId                    string            `json:"instanceId"`
    HostName                      string            `json:"hostName"`
    AppName                       string            `json:"app"`
    IpAddr                        string            `json:"ipAddr"`
    Status                        InstanceStatus    `json:"status"`
    OverriddenStatus              InstanceStatus    `json:"overriddenStatus"`
    Port                          *PortWrapper      `json:"port"`
    SecurePort                    *PortWrapper      `json:"securePort"`
    CountryId                     int               `json:"countryId"`
    DataCenterInfo                *DataCenterInfo   `json:"dataCenterInfo"`
    LeaseInfo                     *LeaseInfo        `json:"leaseInfo"`
    Metadata                      map[string]string `json:"metadata"`
    HomePageUrl                   string            `json:"homePageUrl"`
    StatusPageUrl                 string            `json:"statusPageUrl"`
    HealthCheckUrl                string            `json:"healthCheckUrl"`
    VipAddress                    string            `json:"vipAddress"`
    SecureVipAddress              string            `json:"secureVipAddress"`
    IsCoordinatingDiscoveryServer string            `json:"isCoordinatingDiscoveryServer"`
    LastUpdatedTimestamp          string            `json:"lastUpdatedTimestamp"`
    LastDirtyTimestamp            string            `json:"lastDirtyTimestamp"`
    ActionType                    ActionType        `json:"actionType"`
}

// ToJson 对象转json
func (instance *InstanceInfo) ToJson() ([]byte, error) {
    return json.Marshal(instance)
}

// ParseInstanceInfo 从map中解析服务实例信息
func ParseInstanceInfo(m map[string]interface{}) (instance *InstanceInfo, err interface{}) {
    defer func() {
        if rc := recover(); rc != nil {
            instance = nil
            err = rc
        }
    }()
    instance = &InstanceInfo{}
    instance.InstanceId = m["instanceId"].(string)
    instance.HostName = m["hostName"].(string)
    instance.AppName = m["app"].(string)
    instance.IpAddr = m["ipAddr"].(string)
    instance.Status = InstanceStatus(m["status"].(string))
    instance.OverriddenStatus = InstanceStatus(m["overriddenStatus"].(string))
    instance.Port, err = ParsePortWrapper(m["port"].(map[string]interface{}))
    if err != nil {
        return nil, err
    }
    instance.SecurePort, err = ParsePortWrapper(m["securePort"].(map[string]interface{}))
    if err != nil {
        return nil, err
    }
    instance.CountryId = int(m["countryId"].(float64))
    instance.DataCenterInfo, err = ParseDataCenterInfo(m["dataCenterInfo"].(map[string]interface{}))
    if err != nil {
        return nil, err
    }
    instance.LeaseInfo, err = ParseLeaseInfo(m["leaseInfo"].(map[string]interface{}))
    if err != nil {
        return nil, err
    }
    instance.Metadata = make(map[string]string)
    for k, v := range m["metadata"].(map[string]interface{}) {
        instance.Metadata[k] = v.(string)
    }
    instance.HomePageUrl = m["homePageUrl"].(string)
    instance.StatusPageUrl = m["statusPageUrl"].(string)
    instance.HealthCheckUrl = m["healthCheckUrl"].(string)
    instance.VipAddress = m["vipAddress"].(string)
    instance.SecureVipAddress = m["secureVipAddress"].(string)
    instance.IsCoordinatingDiscoveryServer = m["isCoordinatingDiscoveryServer"].(string)
    instance.LastUpdatedTimestamp = m["lastUpdatedTimestamp"].(string)
    instance.LastDirtyTimestamp = m["lastDirtyTimestamp"].(string)
    instance.ActionType = ActionType(m["actionType"].(string))
    return instance, nil
}
