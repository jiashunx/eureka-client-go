
### 单元测试说明（eureka-client-go）

本文档记录本工程单元测试的覆盖情况、运行方式与未覆盖分支分析。

- 测试范围与约定

   - 本次补充的单元测试全部为**新增文件**，未改动任何源代码，也未改动工程内已有的测试文件（`client/client_test.go`、`client/http_test.go`、`meta/*_test.go`、`log/log_test.go`）。
   - 新增单元测试函数统一以 `TestUnit` 开头，可通过 `-run '^TestUnit'` 独立运行，从而与需要真实 Eureka Server 的集成测试完全隔离。
   - 为在不改动源码的前提下测试网络交互逻辑，统一使用 `net/http/httptest` 构造本地测试服务，因此单元测试**无需任何外部依赖**，且执行迅速。

- 运行方式

```shell
# 运行全部单元测试并统计覆盖率（推荐）
go test ./... -run '^TestUnit' -count=1 -covermode=count -coverprofile=coverage.out

# 查看覆盖率明细
go tool cover -func=coverage.out

# 生成 HTML 覆盖率报告
go tool cover -html=coverage.out
```

> 注：工程内 `client/client_test.go`、`client/http_test.go` 为集成测试，依赖真实 Eureka Server 且有较长等待，不在本单元测试范围内。

- 覆盖率汇总

| 包 | 语句覆盖率 | 说明 |
| --- | --- | --- |
| `github.com/jiashunx/eureka-client-go/log` | **100.0%** | 全部函数 100% |
| `github.com/jiashunx/eureka-client-go/client` | **98.1%** | 核心逻辑（utils/discovery）100% |
| `github.com/jiashunx/eureka-client-go/meta` | **95.6%** | 模型/默认值逻辑基本全覆盖 |
| **合计** | **97.6%** | 未覆盖行均为环境依赖或不可达分支 |

- 各包函数级覆盖（仅列出未达 100% 的函数）

   - `client`（98.1%）

   | 函数 | 覆盖率 | 未覆盖原因 |
   | --- | --- | --- |
   | `EurekaClient.StartWithCtx` | 87.1% | `discoveryClient.start` 恒返回 `Error=nil`，其错误分支不可达 |
   | `NewEurekaClientWithOptions` | 92.3% | `newConfig.Check()` 仅在获取本机信息失败时出错，不可构造 |
   | `HttpClient.Register` | 90.5% | `instance.Check()` 出错与 `json.Marshal` 出错分支（见下） |
   | `HttpClient.getApps` / `getInstances` / `getInstance` | 96.7%~97.4% | 对已 `Unmarshal` 结果的 `json.Marshal` 错误分支不可达 |
   | `RegistryClient.start` | 95.5% | `buildInstanceInfo` 恒返回 `nil` 错误 |
   | `RegistryClient.Register` | 77.8% | `isEnabled`/`GetCurrZoneEurekaServer`/`buildInstanceInfo` 错误分支不可达 |
   | `RegistryClient.Heartbeat` | 83.3% | `isEnabled`/`GetCurrZoneEurekaServer` 错误分支不可达 |
   | `RegistryClient.UnRegister` | 87.5% | 同上 |
   | `RegistryClient.ChangeStatus` / `ChangeMetadata` | 92.9% / 90.0% | `isEnabled`/`GetCurrZoneEurekaServer` 错误分支不可达 |

   - `meta`（95.6%）

   | 函数 | 覆盖率 | 未覆盖原因 |
   | --- | --- | --- |
   | `ParseInstanceConfig` / `ParseClientConfig` / `ParseEurekaServer` / `ParseEurekaConfig` / `ParseInstanceInfo` / `ParseAppInfo` | 77.8%~90.0% | 各解析函数的 `recover` 分支：`json.Unmarshal` 对畸形输入返回 error 而非 panic，不可达 |
   | `EurekaConfig.Check` | 99.3% | `GetLocalHostInfo` 出错分支不可达 |
   | `GetLocalHostInfo` | 77.8% | `os.Hostname()` / `GetLocalIpv4Address` 失败分支依赖运行环境，不可构造 |
   | `GetLocalIpv4Address` | 71.4% | `net.InterfaceAddrs()` 出错与 “ipv4 address not found” 依赖运行环境，不可构造 |
   | `InstanceInfo.Check` | 97.8% | `GetLocalHostInfo` 出错分支不可达 |

- 未覆盖分支归类

   以下分支在本工程与运行环境下**不可达**（编写用例无法触发），故未强行覆盖：

   1. **`recover` 兜底分支**：`meta` 各 `Parse*` 函数、`log` 的 `print/printf`（后者已通过 `defaultLog=nil` 触发覆盖）。
   2. **本机信息获取失败分支**：`os.Hostname()`、`net.InterfaceAddrs()`、无 IPv4 网卡等，依赖宿主环境。
   3. **恒为 nil 的错误分支**：`DiscoveryClient.start`、`RegistryClient.buildInstanceInfo` 永远返回 `Error=nil`。
   4. **对具体结构体的 `json.Marshal` 错误分支**：`map[string]*meta.InstanceInfo` 等类型不可能序列化失败。

- 新增测试文件清单

   | 文件 | 覆盖目标 |
   | --- | --- |
   | `log/log_unit_test.go` | `log.go` 全部日志级别方法、`SetLevel`、`print/printf`（含 recover）、默认日志对象 |
   | `meta/unit_config_test.go` | `EurekaConfig.Check` 默认值/缓存、`GetCurrZoneEurekaServer`、`GetAllZoneEurekaServers`、各 `Parse*`、`GetLocalHostInfo`/`GetLocalIpv4Address` |
   | `meta/unit_instance_test.go` | `InstanceInfo` 及各子结构的 `Copy`/`Check`/`Http(s)ServiceUrl`/`ToJson`/`ParseInstanceInfo` |
   | `meta/unit_app_test.go` | `AppInfo.Copy`/`CopyWithInstances`/`AvailableInstances`/`ParseAppInfo` |
   | `client/unit_helpers_test.go` | 测试公共辅助（httptest 服务、配置、JSON 构造等） |
   | `client/unit_utils_test.go` | `FilterApp`、`FilterAppsByVip/Svip`、`FilterInstancesByVip/Svip`、`RandomLoopMap`、`Summary*` |
   | `client/unit_http_test.go` | `HttpClient` 全部请求方法、`doRequest` 各错误分支、`getApps/getInstances/getInstance` 解析分支 |
   | `client/unit_registry_test.go` | `RegistryClient` 注册/心跳/取消注册/状态/元数据、`start/beat/beat0`、`buildInstanceInfo` |
   | `client/unit_discovery_test.go` | `DiscoveryClient` 服务发现、`Discovery0` 多 zone、`publicQuery`、全部 `Filter*`/`Access*` |
   | `client/unit_client_test.go` | `EurekaClient` 生命周期、`exec`、日志设置、全部 `Access*`、构造方法 |

- 测试中发现但未修复的源码问题（仅记录，未改动源码）

   1. `HttpClient.ModifyMetadata`（`client/http.go`）拼接 URL 后无条件截掉末尾 2 个字符，导致生成的请求 URL 非预期（无 metadata 时得到 `.../metadat`，有 metadata 时丢失最后一个值的末位字符）。
   2. `RegistryClient.beat0`（`client/registry.go`）中 `if response.Error != nil` 条件重复，成功日志永远不会输出，失败时反而打印 “OK”。
   3. `EurekaConfig.Check`（`meta/config.go`）中 `NonSecurePortEnabled`/`SecurePortEnabled` 的 “端口 > 0 则自动启用” 赋值随即被覆盖，属无效逻辑。
   4. 多处配置指针（如 `RegistryEnabled`、`InstanceEnabledOnIt`、`PreferIpAddress`）在未调用 `Check()` 归一化的情况下被直接解引用，手工构造配置时存在 panic 风险。
   5. `log` 包的 `print/printf` 会吞掉所有 panic，配置异常时日志会静默丢失；`LevelNames`/`DefaultLoggerImpl` 为可变的导出全局变量，存在并发读写风险。

- 验证结论

   - `go build ./...`、`go vet ./...` 均通过。
   - `go test ./... -run '^TestUnit'` 全部通过。
   - 工程内既有集成测试与源代码均保持原样，未做任何修改。
