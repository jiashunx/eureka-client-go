
### 基于Go实现的Eureka客户端（服务注册、服务发现）

轻量级、无框架依赖的 Eureka 客户端 Go 实现，提供服务注册、心跳续约、服务发现、实例状态与元数据变更等能力。

- 功能特性

   - 封装与 Eureka Server 通讯的 Http API：[HttpClient](./client/http.go)

   - 封装服务注册客户端（注册、心跳、取消注册、状态/元数据变更）：[RegistryClient](./client/registry.go)

   - 封装服务发现客户端（按 app / vip / svip 查询，支持同 zone 优先与随机获取）：[DiscoveryClient](./client/discovery.go)

   - 集成以上能力的 Eureka 客户端：[EurekaClient](./client/client.go)，开启后自动注册并发送心跳、定期拉取服务信息

   - 配置模型与默认值：[meta](./meta/config.go)

   - 日志接口及默认控制台实现：[log](./log/log.go)

- 目录结构

```text
eureka-client-go
├── client/          # Eureka 客户端核心实现（Http / Registry / Discovery / EurekaClient）
├── meta/            # 配置模型、服务实例模型及 JSON 解析、默认值处理
├── log/             # 日志接口与默认实现
└── http/            # IDEA HTTP Client 请求样例
```

- 添加依赖

```shell
go get github.com/jiashunx/eureka-client-go@v1.0.3
```

- 快速开始

```go
package main

import (
    "fmt"

    "github.com/jiashunx/eureka-client-go/client"
    "github.com/jiashunx/eureka-client-go/log"
    "github.com/jiashunx/eureka-client-go/meta"
)

func main() {
    // 创建 Eureka 客户端
    eurekaClient, err := client.NewEurekaClient(&meta.EurekaConfig{
        InstanceConfig: &meta.InstanceConfig{
            AppName:               "demo-app",
            InstanceId:            "127.0.0.1:8080",
            IpAddress:             "127.0.0.1",
            PreferIpAddress:       &meta.True,
            NonSecurePort:         8080,
            VirtualHostname:       "demo-app",
            SecureVirtualHostname: "secure-demo-app",
        },
        ClientConfig: &meta.ClientConfig{
            ServiceUrlOfDefaultZone: "http://127.0.0.1:8761/eureka",
            RegistryFetchIntervalSeconds: 10,
            LeaseRenewalIntervalInSeconds: 10,
        },
    })
    if err != nil {
        panic(err)
    }

    // 调整日志级别（可选）
    eurekaClient.GetLogger().SetLevel(log.InfoLevel)

    // 启动客户端：自动注册服务、发送心跳、拉取服务列表
    if response := eurekaClient.Start(); response.Error != nil {
        panic(response.Error)
    }
    defer eurekaClient.Stop()

    // 服务发现：查询可用服务实例
    app, err := eurekaClient.AccessApp("demo-app")
    if err != nil {
        panic(err)
    }
    fmt.Printf("available app: %v\n", app.Name)

    // 按 vip 随机获取一个可用实例
    instance, err := eurekaClient.AccessAppInstanceByVip("demo-app")
    if err != nil {
        panic(err)
    }
    fmt.Printf("instance: %s:%d\n", instance.IpAddr, instance.Port.Port)
}
```

- EurekaClient 方法一览

   | 方法 | 说明 |
   | --- | --- |
   | `NewEurekaClient(config)` | 根据配置创建客户端 |
   | `NewEurekaClientWithOptions(config, options)` | 创建客户端并可指定心跳回调等选项 |
   | `Start()` / `StartWithCtx(ctx)` | 启动客户端（注册 + 心跳 + 服务发现） |
   | `Stop()` | 优雅关闭（先取消注册，成功后再关闭） |
   | `ForceStop()` | 强制关闭 |
   | `ChangeStatus(status)` | 变更服务实例状态 |
   | `ChangeMetadata(metadata)` | 变更服务实例元数据 |
   | `AccessApp(appName)` | 查询指定应用可用服务 |
   | `AccessAppsByVip(vip)` / `AccessAppsBySvip(svip)` | 查询指定 vip / svip 的可用服务列表 |
   | `AccessAppInstanceByVip(vip)` / `AccessAppInstanceBySvip(svip)` | 查询指定 vip / svip 的可用实例（随机） |
   | `AccessInstancesByVip(vip)` / `AccessInstancesBySvip(svip)` | 查询指定 vip / svip 的可用实例列表 |
   | `AccessInstanceByVip(vip)` / `AccessInstanceBySvip(svip)` | 查询指定 vip / svip 的可用实例（随机） |
   | `RegistryClient()` / `DiscoveryClient()` / `HttpClient()` | 获取底层客户端 |
   | `SetLogger(logger)` / `GetLogger()` | 设置 / 获取日志对象 |

- 手动使用底层客户端

```go
// 手动服务注册
registryClient := eurekaClient.RegistryClient()
_ = registryClient.Register(meta.StatusUp)
_ = registryClient.Heartbeat()
_ = registryClient.ChangeStatus(meta.StatusDown)
_ = registryClient.ChangeMetadata(map[string]string{"region": "cn"})
_ = registryClient.UnRegister()

// 手动服务发现
discoveryClient := eurekaClient.DiscoveryClient()
_, _ = discoveryClient.AccessApp("demo-app")
_, _ = discoveryClient.AccessAppsByVip("demo-app")

// 直接使用 Http API
httpClient := eurekaClient.HttpClient()
_ = httpClient.SimpleQueryApps("http://127.0.0.1:8761/eureka")
_ = httpClient.SimpleRegister("http://127.0.0.1:8761/eureka", instanceInfo)
```

- 配置说明

   - InstanceConfig（服务实例配置）

   | 配置项 | 默认值 | 说明 |
   | --- | --- | --- |
   | `AppName` | `unknown` | 应用名称 |
   | `InstanceId` | 随机 uuid | 服务实例 ID |
   | `Hostname` | 本机 hostname | 实例主机名 |
   | `PreferIpAddress` | `false` | 是否优先使用 IP 地址替代 Hostname |
   | `IpAddress` | 本机 IPv4 | 实例 IP 地址 |
   | `InstanceEnabledOnIt` | `true` | 注册后是否立即可用（状态直接为 UP） |
   | `Metadata` | 空 map | 实例元数据 |
   | `LeaseRenewalIntervalInSeconds` | `30` | 心跳间隔（秒） |
   | `LeaseExpirationDurationInSeconds` | `90` | 租约过期时间（秒） |
   | `NonSecurePort` | `80` | http 端口 |
   | `NonSecurePortEnabled` | `true` | 是否启用 http 端口 |
   | `SecurePort` | `443` | https 端口 |
   | `SecurePortEnabled` | `false` | 是否启用 https 端口 |
   | `VirtualHostname` | 同 AppName | 虚拟主机名（vip） |
   | `SecureVirtualHostname` | 同 AppName | 安全虚拟主机名（svip） |
   | `StatusPageUrl` / `StatusPageUrlPath` | 空 / `/actuator/info` | 状态页地址及相对路径 |
   | `HomePageUrl` / `HomePageUrlPath` | 空 / `/` | 主页地址及相对路径 |
   | `HealthCheckUrl` / `HealthCheckUrlPath` | 空 / `/actuator/health` | 健康检查地址及相对路径 |

   - ClientConfig（客户端配置）

   | 配置项 | 默认值 | 说明 |
   | --- | --- | --- |
   | `EurekaServerUsername` / `EurekaServerPassword` | 空 | Eureka Server BasicAuth 账号密码 |
   | `EurekaServerReadTimeoutSeconds` | `8` | 读取超时（秒） |
   | `EurekaServerConnectTimeoutSeconds` | `5` | 连接超时（秒） |
   | `RegistryEnabled` | `true` | 是否开启服务注册 |
   | `DiscoveryEnabled` | `true` | 是否开启服务发现 |
   | `InstanceInfoReplicationIntervalSeconds` | `30` | 实例信息复制间隔（秒） |
   | `InitialInstanceInfoReplicationIntervalSeconds` | `30` | 初始实例信息复制间隔（秒） |
   | `RegistryFetchIntervalSeconds` | `30` | 拉取注册信息间隔（秒） |
   | `PreferSameZoneEureka` | `true` | 是否优先使用同 zone 的实例 |
   | `Region` | `default` | 归属 region |
   | `Zone` | `defaultZone` | 归属 zone |
   | `AvailableZones` | 空 map | region 与 zone 的映射 |
   | `ServiceUrlOfDefaultZone` | `http://127.0.0.1:8761/eureka` | defaultZone 的服务地址（逗号分隔） |
   | `ServiceUrlOfAllZone` | 空 map | 各 zone 的服务地址 |

- 测试

   - 单元测试（无需依赖外部环境）

```shell
go test -run '^TestUnit' ./... -cover
```

   - 集成测试：`client/client_test.go`、`client/http_test.go` 需要本地或可访问的 Eureka Server（默认 `http://127.0.0.1:20000/eureka`），按需运行。

   - 单元测试覆盖情况参见：[unit-test.md](./unit-test.md)

- 许可证

   本项目基于 [MIT License](./LICENSE) 开源。
