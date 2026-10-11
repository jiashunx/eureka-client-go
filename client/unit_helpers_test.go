package client

import (
    "errors"
    "fmt"
    "io"
    "net/http"
    "net/http/httptest"
    "strings"
    "sync"
    "testing"

    "github.com/jiashunx/eureka-client-go/log"
    "github.com/jiashunx/eureka-client-go/meta"
)

// ---------------------------------------------------------------------------
// Shared test helpers (all prefixed with "unit" to avoid collisions with the
// existing integration test files).
// ---------------------------------------------------------------------------

// unitInstanceJSON is a valid InstanceInfo payload (status UP, vip/svip set).
const unitInstanceJSON = `{"instanceId":"i-1","hostName":"127.0.0.1","app":"APP","ipAddr":"127.0.0.1","status":"UP","vipAddress":"vip1","secureVipAddress":"svip1","port":{"@enabled":"true","$":18080},"securePort":{"@enabled":"false","$":443}}`

// unitAppsBody builds the JSON body returned by /apps style endpoints.
func unitAppsBody(name string) string {
    return `{"applications":{"application":[{"name":"` + name + `","instance":[` + unitInstanceJSON + `]}]}}`
}

// unitInstancesBody builds the JSON body returned by /apps/{app} endpoints.
func unitInstancesBody() string {
    return `{"application":{"instance":[` + unitInstanceJSON + `]}}`
}

// unitInstanceBody builds the JSON body returned by /apps/{app}/{id} or
// /instances/{id} endpoints.
func unitInstanceBody() string {
    return `{"instance":` + unitInstanceJSON + `}`
}

// unitBaseConfig returns a config suitable for fast unit tests.
func unitBaseConfig(serviceUrl string) *meta.EurekaConfig {
    return &meta.EurekaConfig{
        InstanceConfig: &meta.InstanceConfig{
            AppName:                       "unit-app",
            InstanceId:                    "unit-instance",
            Hostname:                      "127.0.0.1",
            NonSecurePort:                 18080,
            LeaseRenewalIntervalInSeconds: 1,
        },
        ClientConfig: &meta.ClientConfig{
            ServiceUrlOfDefaultZone:      serviceUrl,
            RegistryFetchIntervalSeconds: 1,
        },
    }
}

// unitNewClient creates a checked-in EurekaClient pointed at serviceUrl.
func unitNewClient(t *testing.T, serviceUrl string) *EurekaClient {
    t.Helper()
    c, err := NewEurekaClient(unitBaseConfig(serviceUrl))
    if err != nil {
        t.Fatalf("NewEurekaClient: %v", err)
    }
    return c
}

// unitServer starts an httptest server from handler.
func unitServer(handler http.HandlerFunc) *httptest.Server {
    return httptest.NewServer(handler)
}

// unitUpInstance returns an *meta.InstanceInfo that is UP and has vip/svip set.
func unitUpInstance() *meta.InstanceInfo {
    return &meta.InstanceInfo{
        AppName:          "UNITAPP",
        InstanceId:       "unit-instance",
        HostName:         "127.0.0.1",
        IpAddr:           "127.0.0.1",
        Status:           meta.StatusUp,
        VipAddress:       "vip1",
        SecureVipAddress: "svip1",
        Port:             &meta.PortWrapper{Enabled: meta.StrTrue, Port: 18080},
        SecurePort:       &meta.PortWrapper{Enabled: meta.StrFalse, Port: 443},
    }
}

// unitRoundTripper allows intercepting requests without touching the network.
type unitRoundTripper struct {
    fn func(*http.Request) (*http.Response, error)
}

func (rt *unitRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
    return rt.fn(r)
}

// unitInstallDefaultClient swaps http.DefaultClient for an intercepting one and
// returns a restore function.
func unitInstallDefaultClient(fn func(*http.Request) (*http.Response, error)) func() {
    prev := http.DefaultClient
    http.DefaultClient = &http.Client{Transport: &unitRoundTripper{fn: fn}}
    return func() { http.DefaultClient = prev }
}

// unitResponse builds a canned *http.Response.
func unitResponse(r *http.Request, status int, body string) *http.Response {
    return &http.Response{
        StatusCode: status,
        Status:     fmt.Sprintf("%d", status),
        Body:       io.NopCloser(strings.NewReader(body)),
        Header:     make(http.Header),
        Request:    r,
    }
}

// unitErrReader is an io.ReadCloser whose Read always fails.
type unitErrReader struct{}

func (unitErrReader) Read([]byte) (int, error) { return 0, errors.New("unit read error") }
func (unitErrReader) Close() error             { return nil }

// unitPanicOnceLogger is a log.Logger whose Tracef panics exactly once. It is
// used to reach recover() branches that have no other synchronous trigger.
type unitPanicOnceLogger struct {
    once sync.Once
}

func (l *unitPanicOnceLogger) SetLevel(log.Level)             {}
func (l *unitPanicOnceLogger) Trace(a ...any)                 {}
func (l *unitPanicOnceLogger) Debug(a ...any)                 {}
func (l *unitPanicOnceLogger) Debugf(f string, a ...any)      {}
func (l *unitPanicOnceLogger) Info(a ...any)                  {}
func (l *unitPanicOnceLogger) Infof(f string, a ...any)       {}
func (l *unitPanicOnceLogger) Warn(a ...any)                  {}
func (l *unitPanicOnceLogger) Warnf(f string, a ...any)       {}
func (l *unitPanicOnceLogger) Error(a ...any)                 {}
func (l *unitPanicOnceLogger) Errorf(f string, a ...any)      {}

func (l *unitPanicOnceLogger) Tracef(format string, a ...any) {
    l.once.Do(func() { panic("unit panic logger") })
}
