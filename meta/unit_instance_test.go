package meta

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var unitInstanceInfoJson = `
{
    "instanceId": "unit-1",
    "hostName": "unit-host",
    "app": "UNITAPP",
    "ipAddr": "10.0.0.1",
    "status": "UP",
    "overriddenStatus": "UNKNOWN",
    "port": { "$": 18080, "@enabled": "true" },
    "securePort": { "$": 443, "@enabled": "false" },
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
    "metadata": { "hello": "world" },
    "homePageUrl": "http://unit-host:18080/",
    "statusPageUrl": "http://unit-host:18080/actuator/info",
    "healthCheckUrl": "http://unit-host:18080/actuator/health",
    "vipAddress": "unitapp",
    "secureVipAddress": "unitapp",
    "isCoordinatingDiscoveryServer": "false",
    "actionType": "ADDED"
}
`

func TestUnitMeta_DataCenterInfo_Copy(t *testing.T) {
	ast := assert.New(t)

	var dc *DataCenterInfo
	ast.Nil(dc.Copy())

	orig := &DataCenterInfo{Class: "c", Name: "n"}
	cp := orig.Copy()
	ast.NotNil(cp)
	ast.Equal("c", cp.Class)
	ast.Equal("n", cp.Name)
	cp.Class = "changed"
	ast.Equal("c", orig.Class, "copy must be independent")
}

func TestUnitMeta_DefaultDataCenterInfo(t *testing.T) {
	ast := assert.New(t)
	dc := DefaultDataCenterInfo()
	ast.NotNil(dc)
	ast.Equal("com.netflix.appinfo.InstanceInfo$DefaultDataCenterInfo", dc.Class)
	ast.Equal("MyOwn", dc.Name)
}

func TestUnitMeta_LeaseInfo_Copy(t *testing.T) {
	ast := assert.New(t)

	var lease *LeaseInfo
	ast.Nil(lease.Copy())

	orig := &LeaseInfo{
		RenewalIntervalInSecs: 1,
		DurationInSecs:        2,
		RegistrationTimestamp: 3,
		LastRenewalTimestamp:  4,
		EvictionTimestamp:     5,
		ServiceUpTimestamp:    6,
	}
	cp := orig.Copy()
	ast.NotNil(cp)
	ast.Equal(1, cp.RenewalIntervalInSecs)
	ast.Equal(2, cp.DurationInSecs)
	ast.Equal(int64(3), cp.RegistrationTimestamp)
	ast.Equal(int64(4), cp.LastRenewalTimestamp)
	ast.Equal(int64(5), cp.EvictionTimestamp)
	ast.Equal(int64(6), cp.ServiceUpTimestamp)
	cp.RenewalIntervalInSecs = 99
	ast.Equal(1, orig.RenewalIntervalInSecs, "copy must be independent")
}

func TestUnitMeta_DefaultLeaseInfo(t *testing.T) {
	ast := assert.New(t)
	lease := DefaultLeaseInfo()
	ast.NotNil(lease)
	ast.Equal(DefaultLeaseRenewalIntervalInSeconds, lease.RenewalIntervalInSecs)
	ast.Equal(DefaultLeaseExpirationDurationInSeconds, lease.DurationInSecs)
}

func TestUnitMeta_PortWrapper_IsEnabled(t *testing.T) {
	ast := assert.New(t)
	ast.True((&PortWrapper{Enabled: StrTrue}).IsEnabled())
	ast.False((&PortWrapper{Enabled: StrFalse}).IsEnabled())
	ast.False((&PortWrapper{Enabled: ""}).IsEnabled())
}

func TestUnitMeta_PortWrapper_Copy(t *testing.T) {
	ast := assert.New(t)

	var w *PortWrapper
	ast.Nil(w.Copy())

	orig := &PortWrapper{Enabled: StrTrue, Port: 8080}
	cp := orig.Copy()
	ast.NotNil(cp)
	ast.Equal(StrTrue, cp.Enabled)
	ast.Equal(8080, cp.Port)
	cp.Port = 1
	ast.Equal(8080, orig.Port, "copy must be independent")
}

func TestUnitMeta_DefaultPortWrappers(t *testing.T) {
	ast := assert.New(t)

	nonSecure := DefaultNonSecurePortWrapper()
	ast.Equal(fmt.Sprintf("%t", *DefaultNonSecurePortEnabled), nonSecure.Enabled)
	ast.Equal(DefaultNonSecurePort, nonSecure.Port)

	secure := DefaultSecurePortWrapper()
	ast.Equal(fmt.Sprintf("%t", *DefaultSecurePortEnabled), secure.Enabled)
	ast.Equal(DefaultSecurePort, secure.Port)
}

func TestUnitMeta_InstanceInfo_HttpServiceUrl(t *testing.T) {
	ast := assert.New(t)

	instance := &InstanceInfo{HostName: "h", Port: &PortWrapper{Enabled: StrTrue, Port: 8080}}
	url, err := instance.HttpServiceUrl()
	ast.Nil(err)
	ast.Equal("http://h:8080", url)

	instance.Port = &PortWrapper{Enabled: StrFalse, Port: 8080}
	url, err = instance.HttpServiceUrl()
	ast.NotNil(err)
	ast.Equal("", url)

	instance.Port = nil
	url, err = instance.HttpServiceUrl()
	ast.NotNil(err)
	ast.Equal("", url)
}

func TestUnitMeta_InstanceInfo_HttpsServiceUrl(t *testing.T) {
	ast := assert.New(t)

	instance := &InstanceInfo{HostName: "h", SecurePort: &PortWrapper{Enabled: StrTrue, Port: 8443}}
	url, err := instance.HttpsServiceUrl()
	ast.Nil(err)
	ast.Equal("https://h:8443", url)

	instance.SecurePort = &PortWrapper{Enabled: StrFalse, Port: 8443}
	url, err = instance.HttpsServiceUrl()
	ast.NotNil(err)
	ast.Equal("", url)

	instance.SecurePort = nil
	url, err = instance.HttpsServiceUrl()
	ast.NotNil(err)
	ast.Equal("", url)
}

func TestUnitMeta_InstanceInfo_Copy(t *testing.T) {
	ast := assert.New(t)

	var nilInstance *InstanceInfo
	ast.Nil(nilInstance.Copy())

	// nil sub-wrappers and nil metadata
	instance := &InstanceInfo{
		InstanceId: "id",
		HostName:   "h",
		AppName:    "a",
		IpAddr:     "1.2.3.4",
		Status:     StatusUp,
	}
	cp := instance.Copy()
	ast.NotNil(cp)
	ast.Equal("id", cp.InstanceId)
	ast.Equal("h", cp.HostName)
	ast.Nil(cp.Port)
	ast.Nil(cp.SecurePort)
	ast.Nil(cp.DataCenterInfo)
	ast.Nil(cp.LeaseInfo)
	ast.Nil(cp.Metadata)

	// full copy with metadata and independent nested values
	instance.Port = &PortWrapper{Enabled: StrTrue, Port: 80}
	instance.SecurePort = &PortWrapper{Enabled: StrFalse, Port: 443}
	instance.DataCenterInfo = &DataCenterInfo{Class: "c", Name: "n"}
	instance.LeaseInfo = &LeaseInfo{RenewalIntervalInSecs: 5}
	instance.Metadata = map[string]string{"k": "v"}
	cp2 := instance.Copy()
	ast.NotNil(cp2.Metadata)
	ast.Equal("v", cp2.Metadata["k"])
	cp2.Metadata["k"] = "changed"
	ast.Equal("v", instance.Metadata["k"], "metadata copy must be independent")
	ast.Equal(80, cp2.Port.Port)
	cp2.Port.Port = 1
	ast.Equal(80, instance.Port.Port, "port copy must be independent")
}

func TestUnitMeta_InstanceInfo_Check_AllBlank(t *testing.T) {
	unitFakeHostInfo()
	ast := assert.New(t)

	instance := &InstanceInfo{}
	ast.Nil(instance.Check())

	ast.NotEqual("", instance.InstanceId)
	ast.Equal("unit-host", instance.HostName)
	ast.Equal(DefaultAppName, instance.AppName)
	ast.Equal("10.0.0.1", instance.IpAddr)
	ast.Equal(StatusStarting, instance.Status)
	ast.Equal(StatusUnknown, instance.OverriddenStatus)
	ast.NotNil(instance.SecurePort)
	ast.False(instance.SecurePort.IsEnabled())
	ast.NotNil(instance.Port)
	ast.True(instance.Port.IsEnabled())
	ast.Equal(1, instance.CountryId)
	ast.NotNil(instance.DataCenterInfo)
	ast.NotNil(instance.LeaseInfo)
	ast.NotNil(instance.Metadata)
	ast.True(strings.HasPrefix(instance.StatusPageUrl, HttpProtocol))
	ast.True(strings.HasPrefix(instance.HomePageUrl, HttpProtocol))
	ast.True(strings.HasPrefix(instance.HealthCheckUrl, HttpProtocol))
	ast.Equal(DefaultAppName, instance.VipAddress)
	ast.Equal(DefaultAppName, instance.SecureVipAddress)
	ast.Equal(StrFalse, instance.IsCoordinatingDiscoveryServer)
	ast.Equal(Added, instance.ActionType)
}

func TestUnitMeta_InstanceInfo_Check_AllSet(t *testing.T) {
	unitFakeHostInfo()
	ast := assert.New(t)

	instance := &InstanceInfo{
		InstanceId:                    "id",
		HostName:                      "h",
		AppName:                       "a",
		IpAddr:                        "1.2.3.4",
		Status:                        StatusDown,
		OverriddenStatus:              StatusOutOfService,
		Port:                          &PortWrapper{Enabled: StrTrue, Port: 8080},
		SecurePort:                    &PortWrapper{Enabled: StrFalse, Port: 8443},
		DataCenterInfo:                &DataCenterInfo{Class: "c", Name: "n"},
		LeaseInfo:                     &LeaseInfo{RenewalIntervalInSecs: 5},
		Metadata:                      map[string]string{"k": "v"},
		HomePageUrl:                   "home",
		StatusPageUrl:                 "status",
		HealthCheckUrl:                "health",
		VipAddress:                    "vip",
		SecureVipAddress:              "svip",
		IsCoordinatingDiscoveryServer: StrTrue,
		ActionType:                    Modified,
	}
	ast.Nil(instance.Check())

	ast.Equal("id", instance.InstanceId)
	ast.Equal("h", instance.HostName)
	ast.Equal(StatusDown, instance.Status)
	ast.Equal(StatusOutOfService, instance.OverriddenStatus)
	ast.Equal("home", instance.HomePageUrl)
	ast.Equal("status", instance.StatusPageUrl)
	ast.Equal("health", instance.HealthCheckUrl)
	ast.Equal("vip", instance.VipAddress)
	ast.Equal("svip", instance.SecureVipAddress)
	ast.Equal(StrTrue, instance.IsCoordinatingDiscoveryServer)
	ast.Equal(Modified, instance.ActionType)
}

func TestUnitMeta_InstanceInfo_Check_SecurePortEnablesDefaults(t *testing.T) {
	unitFakeHostInfo()
	ast := assert.New(t)

	// Secure port enabled, no non-secure port: default non-secure port must be disabled.
	instance := &InstanceInfo{
		SecurePort: &PortWrapper{Enabled: StrTrue, Port: 443},
	}
	ast.Nil(instance.Check())
	ast.True(instance.SecurePort.IsEnabled())
	ast.NotNil(instance.Port)
	ast.Equal(StrFalse, instance.Port.Enabled, "default non-secure port must be disabled")
	ast.True(strings.HasPrefix(instance.StatusPageUrl, HttpsProtocol))
}

func TestUnitMeta_InstanceInfo_Check_NonSecurePortOnly(t *testing.T) {
	unitFakeHostInfo()
	ast := assert.New(t)

	// Only non-secure port set -> nil secure port gets the default (disabled).
	instance := &InstanceInfo{
		Port: &PortWrapper{Enabled: StrTrue, Port: 8080},
	}
	ast.Nil(instance.Check())
	ast.NotNil(instance.SecurePort)
	ast.False(instance.SecurePort.IsEnabled())
	ast.Equal(443, instance.SecurePort.Port)
	ast.True(strings.HasPrefix(instance.StatusPageUrl, HttpProtocol))
}

func TestUnitMeta_InstanceInfo_ToJson(t *testing.T) {
	ast := assert.New(t)

	instance := &InstanceInfo{InstanceId: "id", HostName: "h", AppName: "a"}
	data, err := instance.ToJson()
	ast.Nil(err)
	ast.Contains(string(data), `"instanceId":"id"`)

	var back InstanceInfo
	ast.Nil(json.Unmarshal(data, &back))
	ast.Equal("id", back.InstanceId)
}

func TestUnitMeta_ParseInstanceInfo(t *testing.T) {
	unitFakeHostInfo()
	ast := assert.New(t)

	instance, err := ParseInstanceInfo([]byte(unitInstanceInfoJson))
	ast.Nil(err)
	ast.NotNil(instance)
	ast.Equal("unit-1", instance.InstanceId)
	ast.Equal(StatusUp, instance.Status)
	ast.Equal(Added, instance.ActionType)

	bad, err := ParseInstanceInfo([]byte(`{not-json`))
	ast.NotNil(err)
	ast.Nil(bad)
}
