package meta

import (
    "encoding/json"
    "errors"
    "fmt"
    "github.com/google/uuid"
    "strconv"
)

// DataCenterInfo 数据中心
type DataCenterInfo struct {
    Class string `json:"@class"`
    Name  string `json:"name"`
}

// Copy 复制副本
func (dc *DataCenterInfo) Copy() *DataCenterInfo {
    if dc == nil {
        return nil
    }
    return &DataCenterInfo{
        Class: dc.Class,
        Name:  dc.Name,
    }
}

// DefaultDataCenterInfo 默认数据中心信息
func DefaultDataCenterInfo() *DataCenterInfo {
    return &DataCenterInfo{
        Class: "com.netflix.appinfo.InstanceInfo$DefaultDataCenterInfo",
        Name:  "MyOwn",
    }
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

// Copy 复制副本
func (lease *LeaseInfo) Copy() *LeaseInfo {
    if lease == nil {
        return nil
    }
    return &LeaseInfo{
        RenewalIntervalInSecs: lease.RenewalIntervalInSecs,
        DurationInSecs:        lease.DurationInSecs,
        RegistrationTimestamp: lease.RegistrationTimestamp,
        LastRenewalTimestamp:  lease.LastRenewalTimestamp,
        EvictionTimestamp:     lease.EvictionTimestamp,
        ServiceUpTimestamp:    lease.ServiceUpTimestamp,
    }
}

// DefaultLeaseInfo 默认服务实例租约信息
func DefaultLeaseInfo() *LeaseInfo {
    return &LeaseInfo{
        RenewalIntervalInSecs: DefaultLeaseRenewalIntervalInSeconds,
        DurationInSecs:        DefaultLeaseExpirationDurationInSeconds,
    }
}

// PortWrapper 端口信息
type PortWrapper struct {
    Enabled string `json:"@enabled"`
    Port    int    `json:"$"`
}

// IsEnabled 端口是否可用
func (wrapper *PortWrapper) IsEnabled() bool {
    return wrapper.Enabled == StrTrue
}

// Copy 复制副本
func (wrapper *PortWrapper) Copy() *PortWrapper {
    if wrapper == nil {
        return nil
    }
    return &PortWrapper{
        Enabled: wrapper.Enabled,
        Port:    wrapper.Port,
    }
}

// DefaultNonSecurePortWrapper 默认http端口信息
func DefaultNonSecurePortWrapper() *PortWrapper {
    return &PortWrapper{
        Enabled: fmt.Sprintf("%t", *DefaultNonSecurePortEnabled),
        Port:    DefaultNonSecurePort,
    }
}

// DefaultSecurePortWrapper 默认https端口信息
func DefaultSecurePortWrapper() *PortWrapper {
    return &PortWrapper{
        Enabled: fmt.Sprintf("%t", *DefaultSecurePortEnabled),
        Port:    DefaultSecurePort,
    }
}

// InstanceStatus 服务实例状态
type InstanceStatus string

const (
    StatusUp           InstanceStatus = "UP"
    StatusDown         InstanceStatus = "DOWN"
    StatusStarting     InstanceStatus = "STARTING"
    StatusOutOfService InstanceStatus = "OUT_OF_SERVICE"
    StatusUnknown      InstanceStatus = "UNKNOWN"
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
    Region                        string            `json:"-"`
    Zone                          string            `json:"-"`
}

// ToJson 对象转json
func (instance *InstanceInfo) ToJson() ([]byte, error) {
    return json.Marshal(instance)
}

// HttpServiceUrl 获取服务实例的http调用地址
func (instance *InstanceInfo) HttpServiceUrl() (string, error) {
    if instance.Port != nil && instance.Port.IsEnabled() {
        return HttpProtocol + instance.HostName + ":" + strconv.Itoa(instance.Port.Port), nil
    }
    return "", errors.New("the non-secure port of the service instance is not enabled")
}

// HttpsServiceUrl 获取服务实例的https调用地址
func (instance *InstanceInfo) HttpsServiceUrl() (string, error) {
    if instance.SecurePort != nil && instance.SecurePort.IsEnabled() {
        return HttpsProtocol + instance.HostName + ":" + strconv.Itoa(instance.SecurePort.Port), nil
    }
    return "", errors.New("the secure port of the service instance is not enabled")
}

// Copy 复制副本
func (instance *InstanceInfo) Copy() *InstanceInfo {
    if instance == nil {
        return nil
    }
    newInstance := &InstanceInfo{
        InstanceId:                    instance.InstanceId,
        HostName:                      instance.HostName,
        AppName:                       instance.AppName,
        IpAddr:                        instance.IpAddr,
        Status:                        instance.Status,
        OverriddenStatus:              instance.OverriddenStatus,
        Port:                          instance.Port.Copy(),
        SecurePort:                    instance.SecurePort.Copy(),
        CountryId:                     instance.CountryId,
        DataCenterInfo:                instance.DataCenterInfo.Copy(),
        LeaseInfo:                     instance.LeaseInfo.Copy(),
        Metadata:                      nil,
        HomePageUrl:                   instance.HomePageUrl,
        StatusPageUrl:                 instance.StatusPageUrl,
        HealthCheckUrl:                instance.HealthCheckUrl,
        VipAddress:                    instance.VipAddress,
        SecureVipAddress:              instance.SecureVipAddress,
        IsCoordinatingDiscoveryServer: instance.IsCoordinatingDiscoveryServer,
        LastUpdatedTimestamp:          instance.LastUpdatedTimestamp,
        LastDirtyTimestamp:            instance.LastDirtyTimestamp,
        ActionType:                    instance.ActionType,
        Region:                        instance.Region,
        Zone:                          instance.Zone,
    }
    if instance.Metadata != nil {
        newInstance.Metadata = make(map[string]string)
        for k, v := range instance.Metadata {
            newInstance.Metadata[k] = v
        }
    }
    return newInstance
}

// ParseInstanceInfo 从json中解析服务实例信息
func ParseInstanceInfo(data []byte) (instance *InstanceInfo, err error) {
    defer func() {
        if rc := recover(); rc != nil {
            instance = nil
            err = errors.New(fmt.Sprintf("ParseInstanceInfo, recover error: %v", rc))
        }
    }()
    instance = &InstanceInfo{}
    err = json.Unmarshal(data, instance)
    if err != nil {
        return nil, err
    }
    return instance, instance.Check()
}

// Check 检查属性
func (instance *InstanceInfo) Check() error {
    hostInfo, err := GetLocalHostInfo()
    if err != nil {
        return err
    }
    if instance.InstanceId == "" {
        instance.InstanceId = uuid.New().String()
    }
    if instance.HostName == "" {
        instance.HostName = hostInfo.Hostname
    }
    if instance.AppName == "" {
        instance.AppName = DefaultAppName
    }
    if instance.IpAddr == "" {
        instance.IpAddr = hostInfo.IpAddress
    }
    if instance.Status == "" {
        instance.Status = StatusStarting
    }
    if instance.OverriddenStatus == "" {
        instance.OverriddenStatus = StatusUnknown
    }
    if instance.SecurePort == nil {
        instance.SecurePort = DefaultSecurePortWrapper()
    }
    if instance.Port == nil {
        instance.Port = DefaultNonSecurePortWrapper()
        if instance.SecurePort.IsEnabled() {
            instance.Port.Enabled = StrFalse
        }
    }
    instance.CountryId = 1
    if instance.DataCenterInfo == nil {
        instance.DataCenterInfo = DefaultDataCenterInfo()
    }
    if instance.LeaseInfo == nil {
        instance.LeaseInfo = DefaultLeaseInfo()
    }
    if instance.Metadata == nil {
        instance.Metadata = make(map[string]string)
    }
    protocol, ipAddr, port := HttpProtocol, instance.IpAddr, instance.Port.Port
    if instance.SecurePort.IsEnabled() {
        protocol, ipAddr, port = HttpsProtocol, instance.IpAddr, instance.SecurePort.Port
    }
    if instance.StatusPageUrl == "" {
        instance.StatusPageUrl = fmt.Sprintf("%s%s:%d%s", protocol, ipAddr, port, DefaultStatusPageUrlPath)
    }
    if instance.HomePageUrl == "" {
        instance.HomePageUrl = fmt.Sprintf("%s%s:%d%s", protocol, ipAddr, port, DefaultHomePageUrlPath)
    }
    if instance.HealthCheckUrl == "" {
        instance.HealthCheckUrl = fmt.Sprintf("%s%s:%d%s", protocol, ipAddr, port, DefaultHealthCheckUrlPath)
    }
    if instance.VipAddress == "" {
        instance.VipAddress = instance.AppName
    }
    if instance.SecureVipAddress == "" {
        instance.SecureVipAddress = instance.AppName
    }
    if instance.IsCoordinatingDiscoveryServer == "" {
        instance.IsCoordinatingDiscoveryServer = StrFalse
    }
    // if instance.LastUpdatedTimestamp == "" {
    //     instance.LastUpdatedTimestamp = fmt.Sprintf("%d", time.Now().UnixMilli())
    // }
    // if instance.LastDirtyTimestamp == "" {
    //     instance.LastDirtyTimestamp = fmt.Sprintf("%d", time.Now().UnixMilli())
    // }
    if instance.ActionType == "" {
        instance.ActionType = Added
    }
    return nil
}
