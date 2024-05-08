package meta

import (
    "errors"
    "net"
    "os"
    "strings"
)

// InstanceConfig 服务实例配置信息
type InstanceConfig struct {
    // 应用名称，默认：unknown
    AppName string
    // 服务实例ID
    InstanceId string
    // 实例的主机名
    Hostname string
    // 是否优先使用服务实例的IP地址（相较于 Hostname ），默认：false
    PreferIpAddress *bool
    // 实例的 IP Address
    IpAddress string
    // 是否在eureka注册后立即启用实例以获取流量，默认：false（初始注册状态STARTING）
    InstanceEnabledOnIt *bool
    // 实例关联的元数据名称值对，默认：空map
    Metadata map[string]string
    // 客户端发送心跳的时间间隔，默认：30s
    LeaseRenewalIntervalInSeconds int
    // eureka server等待心跳最长时间（超出此时间未接收到心跳则服务实例将不可用，该值应大于 LeaseRenewalIntervalInSeconds ），默认：90s
    LeaseExpirationDurationInSeconds int
    // http通讯端口，默认：80
    NonSecurePort int
    // 是否启用http通信端口，默认：true
    NonSecurePortEnabled *bool
    // https通讯端口，默认：443
    SecurePort int
    // 是否启用https通讯端口，默认：false
    SecurePortEnabled *bool
    // 为此实例定义的虚拟主机名，默认：unknown
    VirtualHostname string
    // 为此实例定义的安全虚拟主机名，默认：unknown
    SecureVirtualHostname string
    // 实例的状态页面绝对URL路径
    StatusPageUrl string
    // 实例的状态页面相对URL路径，默认：/actuator/info
    StatusPageUrlPath string
    // 实例的主页绝对URL路径
    HomePageUrl string
    // 实例的主页相对URL路径，默认：/
    HomePageUrlPath string
    // 实例的健康检查页面绝对URL路径
    HealthCheckUrl string
    // 实例的健康检查页面相对URL路径，默认：/actuator/health
    HealthCheckUrlPath string
}

// ClientConfig 客户端配置信息
type ClientConfig struct {
    // 读取eureka server的超时时间，默认：8s
    EurekaServerReadTimeoutSeconds int
    // 连接eureka server的超时时间，默认：5s
    EurekaServerConnectTimeoutSeconds int
    // 是否开启服务注册，默认：true
    RegisterEnabled *bool
    // 更新实例信息到eureka server的时间间隔，默认：30s
    InstanceInfoReplicationIntervalSeconds int
    // 初始化实例信息到eureka server的时间间隔，默认：30s
    InitialInstanceInfoReplicationIntervalSeconds int
    // 是否开启服务发现，默认：true
    DiscoveryEnabled *bool
    // 从eureka server获取服务注册信息的时间间隔，默认：30s
    RegistryFetchIntervalSeconds int
    // 优先从当前相同zone获取可用服务实例，默认：true
    PreferSameZoneEureka *bool
    // 当前服务实例归属region，默认：default
    Region string
    // 当前region所有zone信息，以逗号分隔，默认：defaultZone
    AvailableZones string
    // 当前服务实例归属zone，若为空则以 AvailableZones 属性逗号分隔的第一个值作为当前服务实例归属zone
    Zone string
    // defaultZone的eureka server服务地址信息，以逗号分隔，默认：http://127.0.0.1:8761/eureka
    ServiceUrlOfDefaultZone string
    // 所有zone的eureka server服务地址信息，若 AvailableZones 中的zone未在当前属性指定eureka server服务地址，则默认：http://127.0.0.1:8761/eureka
    ServiceUrlOfAllZone map[string]string
}

// EurekaConfig eureka客户端配置信息
type EurekaConfig struct {
    *InstanceConfig
    *ClientConfig
}

// NewEurekaConfig 根据 InstanceConfig 及 ClientConfig 构造新的 EurekaConfig 对象
func NewEurekaConfig(ic *InstanceConfig, cc *ClientConfig) (*EurekaConfig, error) {
    hostInfo, err := GetLocalHostInfo()
    if err != nil {
        return nil, err
    }
    // 默认值
    t, f, unknown, Default, defaultZone := true, false, "unknown", "default", "defaultZone"
    // InstanceConfig解析处理
    nic := &InstanceConfig{}
    if ic == nil {
        ic = &InstanceConfig{}
    }
    nic.AppName = strings.TrimSpace(ic.AppName)
    if nic.AppName == "" {
        nic.AppName = unknown
    }
    nic.InstanceId = strings.TrimSpace(ic.InstanceId)
    nic.Hostname = strings.TrimSpace(ic.Hostname)
    if nic.Hostname == "" {
        nic.Hostname = hostInfo.Hostname
    }
    nic.PreferIpAddress = ic.PreferIpAddress
    if nic.PreferIpAddress == nil {
        nic.PreferIpAddress = &f
    }
    nic.IpAddress = strings.TrimSpace(ic.IpAddress)
    if nic.IpAddress == "" {
        nic.IpAddress = hostInfo.IpAddress
    }
    nic.InstanceEnabledOnIt = ic.InstanceEnabledOnIt
    if nic.InstanceEnabledOnIt == nil {
        nic.InstanceEnabledOnIt = &f
    }
    nic.Metadata = ic.Metadata
    if nic.Metadata == nil {
        nic.Metadata = make(map[string]string)
    }
    nic.LeaseRenewalIntervalInSeconds = ic.LeaseRenewalIntervalInSeconds
    if nic.LeaseRenewalIntervalInSeconds <= 0 {
        nic.LeaseRenewalIntervalInSeconds = 30
    }
    nic.LeaseExpirationDurationInSeconds = ic.LeaseExpirationDurationInSeconds
    if nic.LeaseExpirationDurationInSeconds <= 0 {
        nic.LeaseExpirationDurationInSeconds = 90
    }
    nic.NonSecurePort = ic.NonSecurePort
    if nic.NonSecurePort <= 0 {
        nic.NonSecurePort = 80
    }
    nic.NonSecurePortEnabled = ic.NonSecurePortEnabled
    if nic.NonSecurePortEnabled == nil {
        nic.NonSecurePortEnabled = &t
    }
    nic.SecurePort = ic.SecurePort
    if nic.SecurePort <= 0 {
        nic.SecurePort = 443
    }
    nic.SecurePortEnabled = ic.SecurePortEnabled
    if nic.SecurePortEnabled == nil {
        nic.SecurePortEnabled = &f
    }
    nic.VirtualHostname = strings.TrimSpace(ic.VirtualHostname)
    if nic.VirtualHostname == "" {
        nic.VirtualHostname = unknown
    }
    nic.SecureVirtualHostname = strings.TrimSpace(ic.SecureVirtualHostname)
    if nic.SecureVirtualHostname == "" {
        nic.SecureVirtualHostname = unknown
    }
    nic.StatusPageUrl = strings.TrimSpace(ic.StatusPageUrl)
    nic.StatusPageUrlPath = strings.TrimSpace(ic.StatusPageUrlPath)
    if nic.StatusPageUrlPath == "" {
        nic.StatusPageUrlPath = "/actuator/info"
    }
    nic.HomePageUrl = strings.TrimSpace(ic.HomePageUrl)
    nic.HomePageUrlPath = strings.TrimSpace(ic.HomePageUrlPath)
    if nic.HomePageUrlPath == "" {
        nic.HomePageUrlPath = "/"
    }
    nic.HealthCheckUrl = strings.TrimSpace(ic.HealthCheckUrl)
    nic.HealthCheckUrlPath = strings.TrimSpace(ic.HealthCheckUrlPath)
    if nic.HealthCheckUrlPath == "" {
        nic.HealthCheckUrlPath = "/actuator/health"
    }
    // ClientConfig解析处理
    ncc := &ClientConfig{}
    if cc == nil {
        cc = &ClientConfig{}
    }
    ncc.EurekaServerReadTimeoutSeconds = cc.EurekaServerReadTimeoutSeconds
    if ncc.EurekaServerReadTimeoutSeconds <= 0 {
        ncc.EurekaServerReadTimeoutSeconds = 8
    }
    ncc.EurekaServerConnectTimeoutSeconds = cc.EurekaServerConnectTimeoutSeconds
    if ncc.EurekaServerConnectTimeoutSeconds <= 0 {
        ncc.EurekaServerConnectTimeoutSeconds = 5
    }
    ncc.RegisterEnabled = cc.RegisterEnabled
    if ncc.RegisterEnabled == nil {
        ncc.RegisterEnabled = &t
    }
    ncc.InstanceInfoReplicationIntervalSeconds = cc.InstanceInfoReplicationIntervalSeconds
    if ncc.InstanceInfoReplicationIntervalSeconds <= 0 {
        ncc.InstanceInfoReplicationIntervalSeconds = 30
    }
    ncc.InitialInstanceInfoReplicationIntervalSeconds = cc.InitialInstanceInfoReplicationIntervalSeconds
    if ncc.InitialInstanceInfoReplicationIntervalSeconds <= 0 {
        ncc.InitialInstanceInfoReplicationIntervalSeconds = 30
    }
    ncc.DiscoveryEnabled = cc.DiscoveryEnabled
    if ncc.DiscoveryEnabled == nil {
        ncc.DiscoveryEnabled = &t
    }
    ncc.RegistryFetchIntervalSeconds = cc.RegistryFetchIntervalSeconds
    if ncc.RegistryFetchIntervalSeconds <= 0 {
        ncc.RegistryFetchIntervalSeconds = 30
    }
    ncc.PreferSameZoneEureka = cc.PreferSameZoneEureka
    if ncc.PreferSameZoneEureka == nil {
        ncc.PreferSameZoneEureka = &t
    }
    ncc.Region = strings.TrimSpace(cc.Region)
    if ncc.Region == "" {
        ncc.Region = Default
    }
    ncc.AvailableZones = strings.TrimSpace(cc.AvailableZones)
    if ncc.AvailableZones == "" {
        ncc.AvailableZones = defaultZone
    }
    ncc.Zone = strings.TrimSpace(cc.Zone)
    if ncc.Zone == "" {
        ncc.Zone = strings.Split(ncc.AvailableZones, ",")[0]
    }
    ncc.ServiceUrlOfDefaultZone = strings.TrimSpace(cc.ServiceUrlOfDefaultZone)
    if ncc.ServiceUrlOfDefaultZone == "" {
        ncc.ServiceUrlOfDefaultZone = "http://127.0.0.1:8761/eureka"
    }
    ncc.ServiceUrlOfAllZone = cc.ServiceUrlOfAllZone
    if ncc.ServiceUrlOfAllZone == nil {
        ncc.ServiceUrlOfAllZone = make(map[string]string)
    }
    for _, zone := range strings.Split(ncc.AvailableZones, ",") {
        if z, ok := ncc.ServiceUrlOfAllZone[zone]; !ok || strings.TrimSpace(z) == "" {
            ncc.ServiceUrlOfAllZone[zone] = "http://127.0.0.1:8761/eureka"
        }
    }
    return &EurekaConfig{
        nic,
        ncc,
    }, nil
}

// HostInfo 当前主机信息
type HostInfo struct {
    // 主机名
    Hostname string
    // 主机 IP Address
    IpAddress string
}

func GetLocalHostInfo() (*HostInfo, error) {
    hostname, err := os.Hostname()
    if err != nil {
        return nil, errors.New("获取本机hostname失败，原因：" + err.Error())
    }
    ipAddress, err := GetLocalIpv4Address()
    if err != nil {
        return nil, errors.New("获取本机IP失败，原因：" + err.Error())
    }
    return &HostInfo{
        hostname,
        ipAddress,
    }, nil
}

func GetLocalIpv4Address() (string, error) {
    address, err := net.InterfaceAddrs()
    if err != nil {
        return "", err
    }
    for _, addr := range address {
        // 取第一个非lo的网卡IP
        if in, ok := addr.(*net.IPNet); ok && !in.IP.IsLoopback() && in.IP.To4() != nil {
            return in.IP.String(), nil
        }
    }
    return "", errors.New("ipv4 address not found")
}
