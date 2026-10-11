package meta

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// unitBool returns a pointer to the provided bool (helper for pointer fields).
func unitBool(b bool) *bool { return &b }

// unitFakeHostInfo pins the local host info cache to deterministic values so
// tests never depend on the actual network/host environment.
func unitFakeHostInfo() {
	LocalHostInfo = &HostInfo{Hostname: "unit-host", IpAddress: "10.0.0.1"}
}

func TestUnitMeta_CheckDefaults(t *testing.T) {
	unitFakeHostInfo()
	ast := assert.New(t)

	ec := &EurekaConfig{}
	ast.Nilf(ec.Check(), "Check() should succeed")

	ic, cc := ec.InstanceConfig, ec.ClientConfig
	ast.NotNil(ic)
	ast.NotNil(cc)

	// InstanceConfig defaults
	ast.Equal(DefaultAppName, ic.AppName)
	ast.NotEqual("", ic.InstanceId)
	ast.Equal("unit-host", ic.Hostname)
	ast.Equal(DefaultPreferIpAddress, ic.PreferIpAddress)
	ast.Equal("10.0.0.1", ic.IpAddress)
	ast.Equal(DefaultInstanceEnabledOnIt, ic.InstanceEnabledOnIt)
	ast.NotNil(ic.Metadata)
	ast.Equal(DefaultLeaseRenewalIntervalInSeconds, ic.LeaseRenewalIntervalInSeconds)
	ast.Equal(DefaultLeaseExpirationDurationInSeconds, ic.LeaseExpirationDurationInSeconds)
	ast.Equal(DefaultNonSecurePort, ic.NonSecurePort)
	ast.Equal(DefaultNonSecurePortEnabled, ic.NonSecurePortEnabled)
	ast.Equal(DefaultSecurePort, ic.SecurePort)
	ast.Equal(DefaultSecurePortEnabled, ic.SecurePortEnabled)
	ast.Equal(DefaultAppName, ic.VirtualHostname)
	ast.Equal(DefaultAppName, ic.SecureVirtualHostname)
	ast.Equal("", ic.StatusPageUrl)
	ast.Equal(DefaultStatusPageUrlPath, ic.StatusPageUrlPath)
	ast.Equal("", ic.HomePageUrl)
	ast.Equal(DefaultHomePageUrlPath, ic.HomePageUrlPath)
	ast.Equal("", ic.HealthCheckUrl)
	ast.Equal(DefaultHealthCheckUrlPath, ic.HealthCheckUrlPath)

	// ClientConfig defaults
	ast.Equal("", cc.EurekaServerUsername)
	ast.Equal("", cc.EurekaServerPassword)
	ast.Equal(DefaultEurekaServerReadTimeoutSeconds, cc.EurekaServerReadTimeoutSeconds)
	ast.Equal(DefaultEurekaServerConnectTimeoutSeconds, cc.EurekaServerConnectTimeoutSeconds)
	// nil ClientConfig forces registry/discovery disabled
	ast.Equal(&False, cc.RegistryEnabled)
	ast.Equal(&False, cc.DiscoveryEnabled)
	ast.Equal(DefaultInstanceInfoReplicationIntervalSeconds, cc.InstanceInfoReplicationIntervalSeconds)
	ast.Equal(DefaultInitialInstanceInfoReplicationIntervalSeconds, cc.InitialInstanceInfoReplicationIntervalSeconds)
	ast.Equal(DefaultRegistryFetchIntervalSeconds, cc.RegistryFetchIntervalSeconds)
	ast.Equal(DefaultPreferSameZoneEureka, cc.PreferSameZoneEureka)
	ast.Equal(DefaultRegion, cc.Region)
	ast.Equal(DefaultZone, cc.Zone)
	ast.NotNil(cc.AvailableZones)
	ast.Equal(DefaultZone, cc.AvailableZones[DefaultRegion])
	ast.Equal(DefaultServiceUrl, cc.ServiceUrlOfDefaultZone)
	ast.NotNil(cc.ServiceUrlOfAllZone)
	ast.Equal(DefaultServiceUrl, cc.ServiceUrlOfAllZone[DefaultZone])
	ast.True(ec.checked)
}

func TestUnitMeta_CheckFullyPopulated(t *testing.T) {
	unitFakeHostInfo()
	ast := assert.New(t)

	ec := &EurekaConfig{
		InstanceConfig: &InstanceConfig{
			AppName:                       "  my-app  ",
			InstanceId:                    "  id-1  ",
			Hostname:                      "  host1  ",
			PreferIpAddress:               unitBool(true),
			IpAddress:                     "  1.2.3.4  ",
			InstanceEnabledOnIt:           unitBool(false),
			Metadata:                      map[string]string{"k": "v"},
			LeaseRenewalIntervalInSeconds: 11,
			LeaseExpirationDurationInSeconds: 33,
			NonSecurePort:                 8080,
			NonSecurePortEnabled:          unitBool(false),
			SecurePort:                    8443,
			SecurePortEnabled:             unitBool(true),
			VirtualHostname:               "  vip  ",
			SecureVirtualHostname:         "  svip  ",
			StatusPageUrl:                 "  http://sp  ",
			StatusPageUrlPath:             "  /spp  ",
			HomePageUrl:                   "  http://hp  ",
			HomePageUrlPath:               "  /hpp  ",
			HealthCheckUrl:                "  http://hc  ",
			HealthCheckUrlPath:            "  /hcp  ",
		},
		ClientConfig: &ClientConfig{
			EurekaServerUsername:                          "  user  ",
			EurekaServerPassword:                          "  pass  ",
			EurekaServerReadTimeoutSeconds:                1,
			EurekaServerConnectTimeoutSeconds:             2,
			RegistryEnabled:                               unitBool(false),
			InstanceInfoReplicationIntervalSeconds:        3,
			InitialInstanceInfoReplicationIntervalSeconds: 4,
			DiscoveryEnabled:                              unitBool(false),
			RegistryFetchIntervalSeconds:                  5,
			PreferSameZoneEureka:                          unitBool(false),
			Region:                                        "  r  ",
			Zone:                                          "  z  ",
			AvailableZones:                                map[string]string{"r": "z"},
			ServiceUrlOfDefaultZone:                       "  http://def:1/eureka  ",
			ServiceUrlOfAllZone:                           map[string]string{"z": "  http://z:1/eureka  "},
		},
	}
	ast.Nilf(ec.Check(), "Check() should succeed")

	ic, cc := ec.InstanceConfig, ec.ClientConfig
	ast.Equal("my-app", ic.AppName)
	ast.Equal("id-1", ic.InstanceId)
	ast.Equal("host1", ic.Hostname)
	ast.Equal(unitBool(true), ic.PreferIpAddress)
	ast.Equal("1.2.3.4", ic.IpAddress)
	ast.Equal(unitBool(false), ic.InstanceEnabledOnIt)
	ast.Equal(map[string]string{"k": "v"}, ic.Metadata)
	ast.Equal(11, ic.LeaseRenewalIntervalInSeconds)
	ast.Equal(33, ic.LeaseExpirationDurationInSeconds)
	ast.Equal(8080, ic.NonSecurePort)
	ast.Equal(unitBool(false), ic.NonSecurePortEnabled)
	ast.Equal(8443, ic.SecurePort)
	ast.Equal(unitBool(true), ic.SecurePortEnabled)
	ast.Equal("vip", ic.VirtualHostname)
	ast.Equal("svip", ic.SecureVirtualHostname)
	ast.Equal("http://sp", ic.StatusPageUrl)
	ast.Equal("/spp", ic.StatusPageUrlPath)
	ast.Equal("http://hp", ic.HomePageUrl)
	ast.Equal("/hpp", ic.HomePageUrlPath)
	ast.Equal("http://hc", ic.HealthCheckUrl)
	ast.Equal("/hcp", ic.HealthCheckUrlPath)

	ast.Equal("user", cc.EurekaServerUsername)
	ast.Equal("pass", cc.EurekaServerPassword)
	ast.Equal(1, cc.EurekaServerReadTimeoutSeconds)
	ast.Equal(2, cc.EurekaServerConnectTimeoutSeconds)
	ast.Equal(unitBool(false), cc.RegistryEnabled)
	ast.Equal(3, cc.InstanceInfoReplicationIntervalSeconds)
	ast.Equal(4, cc.InitialInstanceInfoReplicationIntervalSeconds)
	ast.Equal(unitBool(false), cc.DiscoveryEnabled)
	ast.Equal(5, cc.RegistryFetchIntervalSeconds)
	ast.Equal(unitBool(false), cc.PreferSameZoneEureka)
	ast.Equal("r", cc.Region)
	ast.Equal("z", cc.Zone)
	ast.Equal("http://z:1/eureka", cc.ServiceUrlOfAllZone["z"])
	ast.Equal("http://def:1/eureka", cc.ServiceUrlOfDefaultZone)
}

func TestUnitMeta_CheckCached(t *testing.T) {
	ast := assert.New(t)

	// cached success: early return must not run the defaulting logic
	ec := &EurekaConfig{
		InstanceConfig: &InstanceConfig{},
		ClientConfig:   &ClientConfig{},
	}
	ec.checked = true
	ec.checkedError = nil
	ast.Nil(ec.Check())
	ast.Equal("", ec.InstanceConfig.AppName, "cached Check must not default")

	// cached error: returns the cached error untouched
	sentinel := errors.New("cached error")
	ec2 := &EurekaConfig{}
	ec2.checked = true
	ec2.checkedError = sentinel
	ast.Equal(sentinel, ec2.Check())
}

func TestUnitMeta_CheckZoneAppended(t *testing.T) {
	unitFakeHostInfo()
	ast := assert.New(t)

	ec := &EurekaConfig{ClientConfig: &ClientConfig{
		Region:         "r",
		Zone:           "z",
		AvailableZones: map[string]string{"r": "z2"},
	}}
	ast.Nil(ec.Check())
	ast.Equal("z2,z", ec.ClientConfig.AvailableZones["r"])
}

func TestUnitMeta_CheckZoneAlreadyPresent(t *testing.T) {
	unitFakeHostInfo()
	ast := assert.New(t)

	ec := &EurekaConfig{ClientConfig: &ClientConfig{
		Region:         "r",
		Zone:           "z",
		AvailableZones: map[string]string{"r": "a, z "},
	}}
	ast.Nil(ec.Check())
	ast.Equal("a, z ", ec.ClientConfig.AvailableZones["r"])
}

func TestUnitMeta_CheckRegionMissingFromAvailableZones(t *testing.T) {
	unitFakeHostInfo()
	ast := assert.New(t)

	ec := &EurekaConfig{ClientConfig: &ClientConfig{
		Region:         "r",
		Zone:           "z",
		AvailableZones: map[string]string{"other": "x"},
	}}
	ast.Nil(ec.Check())
	ast.Equal("z", ec.ClientConfig.AvailableZones["r"])
}

func TestUnitMeta_CheckDefaultZoneServiceUrl(t *testing.T) {
	unitFakeHostInfo()
	ast := assert.New(t)

	ec := &EurekaConfig{ClientConfig: &ClientConfig{
		Region:                  "r",
		Zone:                    DefaultZone,
		ServiceUrlOfDefaultZone: "  http://custom:9/eureka  ",
	}}
	ast.Nil(ec.Check())
	ast.Equal("http://custom:9/eureka", ec.ClientConfig.ServiceUrlOfAllZone[DefaultZone])
	ast.Equal("http://custom:9/eureka", ec.ClientConfig.ServiceUrlOfDefaultZone)
}

func TestUnitMeta_CheckServiceUrlDefaultAndBlank(t *testing.T) {
	unitFakeHostInfo()
	ast := assert.New(t)

	// missing entry -> DefaultServiceUrl
	ec := &EurekaConfig{ClientConfig: &ClientConfig{Region: "r", Zone: "z"}}
	ast.Nil(ec.Check())
	ast.Equal(DefaultServiceUrl, ec.ClientConfig.ServiceUrlOfAllZone["z"])

	// blank entry -> DefaultServiceUrl
	ec2 := &EurekaConfig{ClientConfig: &ClientConfig{
		Region:              "r",
		Zone:                "z",
		ServiceUrlOfAllZone: map[string]string{"z": "   "},
	}}
	ast.Nil(ec2.Check())
	ast.Equal(DefaultServiceUrl, ec2.ClientConfig.ServiceUrlOfAllZone["z"])
}

func TestUnitMeta_CheckEmptyZoneEntry(t *testing.T) {
	unitFakeHostInfo()
	ast := assert.New(t)

	ec := &EurekaConfig{ClientConfig: &ClientConfig{
		Region:         "r",
		Zone:           "z",
		AvailableZones: map[string]string{"r": "a,"},
	}}
	ast.Nil(ec.Check())
	ast.Equal("a,,z", ec.ClientConfig.AvailableZones["r"])
	ast.Equal(DefaultServiceUrl, ec.ClientConfig.ServiceUrlOfAllZone["a"])
	ast.Equal(DefaultServiceUrl, ec.ClientConfig.ServiceUrlOfAllZone["z"])
}

func TestUnitMeta_GetCurrZoneEurekaServer_Nil(t *testing.T) {
	ast := assert.New(t)
	var ec *EurekaConfig
	server, err := ec.GetCurrZoneEurekaServer()
	ast.Nil(server)
	ast.NotNil(err)
}

func TestUnitMeta_GetCurrZoneEurekaServer(t *testing.T) {
	unitFakeHostInfo()
	ast := assert.New(t)

	ec := &EurekaConfig{ClientConfig: &ClientConfig{
		Region:                           "r",
		Zone:                             "z",
		ServiceUrlOfAllZone:              map[string]string{"z": "http://z:1/eureka"},
		EurekaServerUsername:             "user",
		EurekaServerPassword:             "pass",
		EurekaServerReadTimeoutSeconds:   7,
		EurekaServerConnectTimeoutSeconds: 6,
	}}
	server, err := ec.GetCurrZoneEurekaServer()
	ast.Nil(err)
	ast.NotNil(server)
	ast.Equal("r", server.Region)
	ast.Equal("z", server.Zone)
	ast.Equal("http://z:1/eureka", server.ServiceUrl)
	ast.Equal("user", server.Username)
	ast.Equal("pass", server.Password)
	ast.Equal(7, server.ReadTimeoutSeconds)
	ast.Equal(6, server.ConnectTimeoutSeconds)
}

func TestUnitMeta_GetCurrZoneEurekaServer_CheckError(t *testing.T) {
	ast := assert.New(t)
	sentinel := errors.New("boom")
	ec := &EurekaConfig{}
	ec.checked = true
	ec.checkedError = sentinel
	server, err := ec.GetCurrZoneEurekaServer()
	ast.Nil(server)
	ast.Equal(sentinel, err)
}

func TestUnitMeta_GetAllZoneEurekaServers_Nil(t *testing.T) {
	ast := assert.New(t)
	var ec *EurekaConfig
	servers, err := ec.GetAllZoneEurekaServers()
	ast.Nil(servers)
	ast.NotNil(err)
}

func TestUnitMeta_GetAllZoneEurekaServers(t *testing.T) {
	unitFakeHostInfo()
	ast := assert.New(t)

	ec := &EurekaConfig{ClientConfig: &ClientConfig{
		Region:                           "r",
		Zone:                             "z",
		ServiceUrlOfAllZone:              map[string]string{"z": "http://z:1/eureka", "z2": "http://z2:2/eureka"},
		EurekaServerUsername:             "user",
		EurekaServerPassword:             "pass",
		EurekaServerReadTimeoutSeconds:   7,
		EurekaServerConnectTimeoutSeconds: 6,
	}}
	servers, err := ec.GetAllZoneEurekaServers()
	ast.Nil(err)
	ast.Len(servers, 2)
	ast.Equal("http://z:1/eureka", servers["z"].ServiceUrl)
	ast.Equal("http://z2:2/eureka", servers["z2"].ServiceUrl)
	ast.Equal("r", servers["z"].Region)
	ast.Equal("z", servers["z"].Zone)
	ast.Equal("user", servers["z"].Username)
	ast.Equal("pass", servers["z"].Password)
	ast.Equal(7, servers["z"].ReadTimeoutSeconds)
	ast.Equal(6, servers["z"].ConnectTimeoutSeconds)

	// empty ServiceUrlOfAllZone -> empty map
	ec2 := &EurekaConfig{ClientConfig: &ClientConfig{Region: "r", Zone: "z"}}
	ec2.checked = true
	ec2.checkedError = nil
	// Ensure maps exist for the code path; nil map range is safe.
	ec2.ServiceUrlOfAllZone = map[string]string{}
	servers2, err2 := ec2.GetAllZoneEurekaServers()
	ast.Nil(err2)
	ast.Len(servers2, 0)
}

func TestUnitMeta_GetAllZoneEurekaServers_CheckError(t *testing.T) {
	ast := assert.New(t)
	sentinel := errors.New("boom")
	ec := &EurekaConfig{}
	ec.checked = true
	ec.checkedError = sentinel
	servers, err := ec.GetAllZoneEurekaServers()
	ast.Nil(servers)
	ast.Equal(sentinel, err)
}

func TestUnitMeta_ParseInstanceConfig(t *testing.T) {
	ast := assert.New(t)

	c, err := ParseInstanceConfig([]byte(`{"app-name":"a","non-secure-port":1234}`))
	ast.Nil(err)
	ast.NotNil(c)
	ast.Equal("a", c.AppName)
	ast.Equal(1234, c.NonSecurePort)

	// invalid json: error is returned (the partially built config is also returned)
	_, err = ParseInstanceConfig([]byte(`{not-json`))
	ast.NotNil(err)
}

func TestUnitMeta_ParseClientConfig(t *testing.T) {
	ast := assert.New(t)

	c, err := ParseClientConfig([]byte(`{"region":"r","zone":"z"}`))
	ast.Nil(err)
	ast.NotNil(c)
	ast.Equal("r", c.Region)
	ast.Equal("z", c.Zone)

	_, err = ParseClientConfig([]byte(`{not-json`))
	ast.NotNil(err)
}

func TestUnitMeta_ParseEurekaServer(t *testing.T) {
	ast := assert.New(t)

	s, err := ParseEurekaServer([]byte(`{"region":"r","service-url":"http://x:1/eureka"}`))
	ast.Nil(err)
	ast.NotNil(s)
	ast.Equal("r", s.Region)
	ast.Equal("http://x:1/eureka", s.ServiceUrl)

	_, err = ParseEurekaServer([]byte(`{not-json`))
	ast.NotNil(err)
}

func TestUnitMeta_ParseEurekaConfig(t *testing.T) {
	unitFakeHostInfo()
	ast := assert.New(t)

	ec, err := ParseEurekaConfig([]byte(`{"app-name":"a","region":"r","zone":"z"}`))
	ast.Nil(err)
	ast.NotNil(ec)
	ast.Equal("a", ec.InstanceConfig.AppName)
	ast.Equal("r", ec.ClientConfig.Region)
	ast.True(ec.checked)

	// null fields must not break Check
	ecNull, errNull := ParseEurekaConfig([]byte(`{"prefer-ip-address":null,"available-zones":null,"secure-port":null}`))
	ast.Nil(errNull)
	ast.NotNil(ecNull)
	// null PreferIpAddress gets defaulted by Check()
	ast.Equal(DefaultPreferIpAddress, ecNull.InstanceConfig.PreferIpAddress)

	bad, err := ParseEurekaConfig([]byte(`{not-json`))
	ast.NotNil(err)
	ast.Nil(bad)
}

func TestUnitMeta_GetLocalHostInfo_Cache(t *testing.T) {
	ast := assert.New(t)
	saved := LocalHostInfo
	defer func() { LocalHostInfo = saved }()

	LocalHostInfo = &HostInfo{Hostname: "cached-host", IpAddress: "9.9.9.9"}
	h, err := GetLocalHostInfo()
	ast.Nil(err)
	ast.NotNil(h)
	ast.Equal("cached-host", h.Hostname)
	ast.Equal("9.9.9.9", h.IpAddress)
}

func TestUnitMeta_GetLocalHostInfo_Uncached(t *testing.T) {
	ast := assert.New(t)
	saved := LocalHostInfo
	defer func() { LocalHostInfo = saved }()

	LocalHostInfo = nil
	h, err := GetLocalHostInfo()
	if err != nil {
		t.Fatalf("GetLocalHostInfo() unexpected error: %v", err)
	}
	ast.NotNil(h)
	ast.NotNil(LocalHostInfo)
	ast.NotEqual("", h.Hostname)
}

func TestUnitMeta_GetLocalIpv4Address(t *testing.T) {
	ast := assert.New(t)
	ip, err := GetLocalIpv4Address()
	if err != nil {
		t.Fatalf("GetLocalIpv4Address() unexpected error: %v", err)
	}
	ast.NotEqual("", ip)
}
