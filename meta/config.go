package meta

import (
    "errors"
    "github.com/google/uuid"
    "net"
    "os"
    "strings"
)

var (
    True                                                 = true
    False                                                = false
    StrTrue                                              = "true"
    StrFalse                                             = "false"
    DefaultAppName                                       = "unknown"
    DefaultPreferIpAddress                               = &False
    DefaultInstanceEnabledOnIt                           = &False
    DefaultLeaseRenewalIntervalInSeconds                 = 30
    DefaultLeaseExpirationDurationInSeconds              = 90
    DefaultNonSecurePort                                 = 80
    DefaultNonSecurePortEnabled                          = &True
    DefaultSecurePort                                    = 443
    DefaultSecurePortEnabled                             = &False
    DefaultVirtualHostname                               = "unknown"
    DefaultSecureVirtualHostname                         = "unknown"
    DefaultStatusPageUrlPath                             = "/actuator/info"
    DefaultHomePageUrlPath                               = "/"
    DefaultHealthCheckUrlPath                            = "/actuator/health"
    DefaultEurekaServerReadTimeoutSeconds                = 8
    DefaultEurekaServerConnectTimeoutSeconds             = 5
    DefaultRegisterEnabled                               = &True
    DefaultInstanceInfoReplicationIntervalSeconds        = 30
    DefaultInitialInstanceInfoReplicationIntervalSeconds = 30
    DefaultDiscoveryEnabled                              = &True
    DefaultRegistryFetchIntervalSeconds                  = 30
    DefaultPreferSameZoneEureka                          = &True
    DefaultRegion                                        = "default"
    DefaultZone                                          = "defaultZone"
    DefaultServiceUrl                                    = "http://127.0.0.1:8761/eureka"
)

// InstanceConfig 服务实例配置信息
type InstanceConfig struct {
    // 应用名称, 默认: DefaultAppName
    AppName string
    // 服务实例ID, 默认为生成的uuid
    InstanceId string
    // 实例的主机名, 默认为本机hostname
    Hostname string
    // 是否优先使用服务实例的IP地址(相较于 Hostname), 默认: DefaultPreferIpAddress
    PreferIpAddress *bool
    // 实例的 IP Address, 默认为本机IP
    IpAddress string
    // 是否在eureka注册后立即启用实例以获取流量, 默认: DefaultInstanceEnabledOnIt
    InstanceEnabledOnIt *bool
    // 实例关联的元数据名称值对, 默认为空map
    Metadata map[string]string
    // 客户端发送心跳的时间间隔, 默认: DefaultLeaseRenewalIntervalInSeconds
    LeaseRenewalIntervalInSeconds int
    // eureka server等待心跳最长时间(超出此时间未接收到心跳则服务实例将不可用, 该值应大于 LeaseRenewalIntervalInSeconds), 默认: DefaultLeaseExpirationDurationInSeconds
    LeaseExpirationDurationInSeconds int
    // http通讯端口, 默认: DefaultNonSecurePort
    NonSecurePort int
    // 是否启用http通信端口, 默认: DefaultNonSecurePortEnabled
    NonSecurePortEnabled *bool
    // https通讯端口, 默认: DefaultSecurePort
    SecurePort int
    // 是否启用https通讯端口, 默认: DefaultSecurePortEnabled
    SecurePortEnabled *bool
    // 为此实例定义的虚拟主机名, 默认: DefaultVirtualHostname
    VirtualHostname string
    // 为此实例定义的安全虚拟主机名, 默认: DefaultSecureVirtualHostname
    SecureVirtualHostname string
    // 实例的状态页面绝对URL路径, 默认为空
    StatusPageUrl string
    // 实例的状态页面相对URL路径, 默认: DefaultStatusPageUrlPath
    StatusPageUrlPath string
    // 实例的主页绝对URL路径, 默认为空
    HomePageUrl string
    // 实例的主页相对URL路径, 默认: DefaultHomePageUrlPath
    HomePageUrlPath string
    // 实例的健康检查页面绝对URL路径, 默认为空
    HealthCheckUrl string
    // 实例的健康检查页面相对URL路径, 默认: DefaultHealthCheckUrlPath
    HealthCheckUrlPath string
}

// ClientConfig 客户端配置信息
type ClientConfig struct {
    // eureka server BasicAuth用户名, 默认为空
    EurekaServerUsername string
    // eureka server BasicAuth密码, 默认为空
    EurekaServerPassword string
    // 读取eureka server的超时时间, 默认: DefaultEurekaServerReadTimeoutSeconds
    EurekaServerReadTimeoutSeconds int
    // 连接eureka server的超时时间, 默认: DefaultEurekaServerConnectTimeoutSeconds
    EurekaServerConnectTimeoutSeconds int
    // 是否开启服务注册, 默认: DefaultRegisterEnabled
    RegisterEnabled *bool
    // 更新实例信息到eureka server的时间间隔, 默认: DefaultInstanceInfoReplicationIntervalSeconds
    InstanceInfoReplicationIntervalSeconds int
    // 初始化实例信息到eureka server的时间间隔, 默认: DefaultInitialInstanceInfoReplicationIntervalSeconds
    InitialInstanceInfoReplicationIntervalSeconds int
    // 是否开启服务发现, 默认: DefaultDiscoveryEnabled
    DiscoveryEnabled *bool
    // 从eureka server获取服务注册信息的时间间隔, 默认: DefaultRegistryFetchIntervalSeconds
    RegistryFetchIntervalSeconds int
    // 优先从当前相同zone获取可用服务实例, 默认: DefaultPreferSameZoneEureka
    PreferSameZoneEureka *bool
    // 当前服务实例归属region, 默认: DefaultRegion
    Region string
    // 当前region所有zone信息, 以逗号分隔, 默认: DefaultZone
    AvailableZones string
    // 当前服务实例归属zone, 若为空则以 AvailableZones 属性逗号分隔的第一个值作为当前服务实例归属zone, 默认: DefaultZone
    Zone string
    // defaultZone的eureka server服务地址信息, 以逗号分隔, 默认: DefaultServiceUrl
    ServiceUrlOfDefaultZone string
    // 所有zone的eureka server服务地址信息, 若 AvailableZones 中的zone未在当前属性指定eureka server服务地址, 默认: DefaultServiceUrl
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
    // InstanceConfig解析处理
    nic := &InstanceConfig{}
    if ic == nil {
        ic = &InstanceConfig{}
    }
    nic.AppName = strings.TrimSpace(ic.AppName)
    if nic.AppName == "" {
        nic.AppName = DefaultAppName
    }
    nic.InstanceId = strings.TrimSpace(ic.InstanceId)
    if nic.InstanceId == "" {
        nic.InstanceId = uuid.New().String()
    }
    nic.Hostname = strings.TrimSpace(ic.Hostname)
    if nic.Hostname == "" {
        nic.Hostname = hostInfo.Hostname
    }
    nic.PreferIpAddress = ic.PreferIpAddress
    if nic.PreferIpAddress == nil {
        nic.PreferIpAddress = DefaultPreferIpAddress
    }
    nic.IpAddress = strings.TrimSpace(ic.IpAddress)
    if nic.IpAddress == "" {
        nic.IpAddress = hostInfo.IpAddress
    }
    nic.InstanceEnabledOnIt = ic.InstanceEnabledOnIt
    if nic.InstanceEnabledOnIt == nil {
        nic.InstanceEnabledOnIt = DefaultInstanceEnabledOnIt
    }
    nic.Metadata = ic.Metadata
    if nic.Metadata == nil {
        nic.Metadata = make(map[string]string)
    }
    nic.LeaseRenewalIntervalInSeconds = ic.LeaseRenewalIntervalInSeconds
    if nic.LeaseRenewalIntervalInSeconds <= 0 {
        nic.LeaseRenewalIntervalInSeconds = DefaultLeaseRenewalIntervalInSeconds
    }
    nic.LeaseExpirationDurationInSeconds = ic.LeaseExpirationDurationInSeconds
    if nic.LeaseExpirationDurationInSeconds <= 0 {
        nic.LeaseExpirationDurationInSeconds = DefaultLeaseExpirationDurationInSeconds
    }
    nic.NonSecurePort = ic.NonSecurePort
    if nic.NonSecurePort <= 0 {
        nic.NonSecurePort = DefaultNonSecurePort
    }
    nic.NonSecurePortEnabled = ic.NonSecurePortEnabled
    if nic.NonSecurePortEnabled == nil {
        nic.NonSecurePortEnabled = DefaultNonSecurePortEnabled
    }
    nic.SecurePort = ic.SecurePort
    if nic.SecurePort <= 0 {
        nic.SecurePort = DefaultSecurePort
    }
    nic.SecurePortEnabled = ic.SecurePortEnabled
    if nic.SecurePortEnabled == nil {
        nic.SecurePortEnabled = DefaultSecurePortEnabled
    }
    nic.VirtualHostname = strings.TrimSpace(ic.VirtualHostname)
    if nic.VirtualHostname == "" {
        nic.VirtualHostname = DefaultVirtualHostname
    }
    nic.SecureVirtualHostname = strings.TrimSpace(ic.SecureVirtualHostname)
    if nic.SecureVirtualHostname == "" {
        nic.SecureVirtualHostname = DefaultSecureVirtualHostname
    }
    nic.StatusPageUrl = strings.TrimSpace(ic.StatusPageUrl)
    nic.StatusPageUrlPath = strings.TrimSpace(ic.StatusPageUrlPath)
    if nic.StatusPageUrlPath == "" {
        nic.StatusPageUrlPath = DefaultStatusPageUrlPath
    }
    nic.HomePageUrl = strings.TrimSpace(ic.HomePageUrl)
    nic.HomePageUrlPath = strings.TrimSpace(ic.HomePageUrlPath)
    if nic.HomePageUrlPath == "" {
        nic.HomePageUrlPath = DefaultHomePageUrlPath
    }
    nic.HealthCheckUrl = strings.TrimSpace(ic.HealthCheckUrl)
    nic.HealthCheckUrlPath = strings.TrimSpace(ic.HealthCheckUrlPath)
    if nic.HealthCheckUrlPath == "" {
        nic.HealthCheckUrlPath = DefaultHealthCheckUrlPath
    }
    // ClientConfig解析处理
    ncc := &ClientConfig{}
    if cc == nil {
        cc = &ClientConfig{}
    }
    ncc.EurekaServerUsername = strings.TrimSpace(cc.EurekaServerUsername)
    ncc.EurekaServerPassword = strings.TrimSpace(cc.EurekaServerPassword)
    ncc.EurekaServerReadTimeoutSeconds = cc.EurekaServerReadTimeoutSeconds
    if ncc.EurekaServerReadTimeoutSeconds <= 0 {
        ncc.EurekaServerReadTimeoutSeconds = DefaultEurekaServerReadTimeoutSeconds
    }
    ncc.EurekaServerConnectTimeoutSeconds = cc.EurekaServerConnectTimeoutSeconds
    if ncc.EurekaServerConnectTimeoutSeconds <= 0 {
        ncc.EurekaServerConnectTimeoutSeconds = DefaultEurekaServerConnectTimeoutSeconds
    }
    ncc.RegisterEnabled = cc.RegisterEnabled
    if ncc.RegisterEnabled == nil {
        ncc.RegisterEnabled = DefaultRegisterEnabled
    }
    ncc.InstanceInfoReplicationIntervalSeconds = cc.InstanceInfoReplicationIntervalSeconds
    if ncc.InstanceInfoReplicationIntervalSeconds <= 0 {
        ncc.InstanceInfoReplicationIntervalSeconds = DefaultInstanceInfoReplicationIntervalSeconds
    }
    ncc.InitialInstanceInfoReplicationIntervalSeconds = cc.InitialInstanceInfoReplicationIntervalSeconds
    if ncc.InitialInstanceInfoReplicationIntervalSeconds <= 0 {
        ncc.InitialInstanceInfoReplicationIntervalSeconds = DefaultInitialInstanceInfoReplicationIntervalSeconds
    }
    ncc.DiscoveryEnabled = cc.DiscoveryEnabled
    if ncc.DiscoveryEnabled == nil {
        ncc.DiscoveryEnabled = DefaultDiscoveryEnabled
    }
    ncc.RegistryFetchIntervalSeconds = cc.RegistryFetchIntervalSeconds
    if ncc.RegistryFetchIntervalSeconds <= 0 {
        ncc.RegistryFetchIntervalSeconds = DefaultRegistryFetchIntervalSeconds
    }
    ncc.PreferSameZoneEureka = cc.PreferSameZoneEureka
    if ncc.PreferSameZoneEureka == nil {
        ncc.PreferSameZoneEureka = DefaultPreferSameZoneEureka
    }
    ncc.Region = strings.TrimSpace(cc.Region)
    if ncc.Region == "" {
        ncc.Region = DefaultRegion
    }
    ncc.AvailableZones = strings.TrimSpace(cc.AvailableZones)
    if ncc.AvailableZones == "" {
        ncc.AvailableZones = DefaultZone
    }
    ncc.Zone = strings.TrimSpace(cc.Zone)
    if ncc.Zone == "" {
        ncc.Zone = strings.Split(ncc.AvailableZones, ",")[0]
    }
    ncc.ServiceUrlOfDefaultZone = strings.TrimSpace(cc.ServiceUrlOfDefaultZone)
    if ncc.ServiceUrlOfDefaultZone == "" {
        ncc.ServiceUrlOfDefaultZone = DefaultServiceUrl
    }
    ncc.ServiceUrlOfAllZone = cc.ServiceUrlOfAllZone
    if ncc.ServiceUrlOfAllZone == nil {
        ncc.ServiceUrlOfAllZone = make(map[string]string)
    }
    for _, zone := range strings.Split(ncc.AvailableZones, ",") {
        if z, ok := ncc.ServiceUrlOfAllZone[zone]; !ok || strings.TrimSpace(z) == "" {
            ncc.ServiceUrlOfAllZone[zone] = DefaultServiceUrl
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

// LocalHostInfo 本机信息缓存
var LocalHostInfo *HostInfo

// GetLocalHostInfo 获取本机信息
func GetLocalHostInfo() (*HostInfo, error) {
    if LocalHostInfo == nil {
        hostname, err := os.Hostname()
        if err != nil {
            return nil, errors.New("获取本机hostname失败, 原因: " + err.Error())
        }
        ipAddress, err := GetLocalIpv4Address()
        if err != nil {
            return nil, errors.New("获取本机IP失败, 原因: " + err.Error())
        }
        LocalHostInfo = &HostInfo{
            hostname,
            ipAddress,
        }
    }
    return LocalHostInfo, nil
}

// GetLocalIpv4Address 获取本机IP(ipv4)
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

// EurekaServer eureka server 连接信息
type EurekaServer struct {
    // 服务地址
    ServiceUrl string
    // BasicAuth用户名
    Username string
    // BasicAuth密码
    Password string
    // 读超时秒数
    ReadTimeoutSeconds int
    // 连接超时秒数
    ConnectTimeoutSeconds int
}
