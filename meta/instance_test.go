package meta

import (
    "encoding/json"
    "github.com/stretchr/testify/assert"
    "testing"
)

func TestParseInstanceInfo(t *testing.T) {
    ast := assert.New(t)
    var ii interface{}
    err := json.Unmarshal([]byte(TestInstanceInfo), &ii)
    ast.Nilf(err, "反序列化测试数据失败")
    instance, rc := ParseInstanceInfo(ii.(map[string]interface{}))
    ast.Nilf(rc, "解析InstanceInfo失败: %v", rc)
    ast.Equal("127.0.0.1", instance.HostName)
}

var TestInstanceInfo = `
{
    "instanceId": "127.0.0.1:18080",
    "hostName": "127.0.0.1",
    "app": "SPRINGBOOT278",
    "ipAddr": "127.0.0.1",
    "status": "UP",
    "overriddenStatus": "UNKNOWN",
    "port": {
      "$": 18080,
      "@enabled": "true"
    },
    "securePort": {
      "$": 443,
      "@enabled": "false"
    },
    "countryId": 1,
    "dataCenterInfo": {
      "@class": "com.netflix.appinfo.InstanceInfo$DefaultDataCenterInfo",
      "name": "MyOwn"
    },
    "leaseInfo": {
      "renewalIntervalInSecs": 30,
      "durationInSecs": 90,
      "registrationTimestamp": 1714441080811,
      "lastRenewalTimestamp": 1714441080811,
      "evictionTimestamp": 0,
      "serviceUpTimestamp": 1714437265702
    },
    "metadata": {
      "management.port": "18080",
      "hello": "world"
    },
    "homePageUrl": "http://127.0.0.1:18080/",
    "statusPageUrl": "http://127.0.0.1:18080/actuator/info",
    "healthCheckUrl": "http://127.0.0.1:18080/actuator/health",
    "vipAddress": "springboot278",
    "secureVipAddress": "springboot278",
    "isCoordinatingDiscoveryServer": "false",
    "lastUpdatedTimestamp": "1714441080811",
    "lastDirtyTimestamp": "1714437265630",
    "actionType": "ADDED"
}
`
