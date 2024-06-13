
### 基于Go实现的Eureka客户端（服务注册、服务发现）

- 实现功能

   - 封装与Eureka Server通讯的Http API: [HttpClient](./client/http.go)

   - 封装服务注册客户端：[RegistryClient](./client/registry.go)

   - 封装服务发现客户端：[DiscoveryClient](./client/discovery.go)

   - 集成以上三类功能的Eureka客户端：[EurekaClient](./client/client.go)，包含服务注册与发现所有功能，开启后自动注册心跳并获取服务信息

- 添加依赖

```shell
go get github.com/jiashunx/eureka-client-go@v1.0.1
```

- 代码样例

   - 参见测试用例：[测试用例](./client/client_test.go)
