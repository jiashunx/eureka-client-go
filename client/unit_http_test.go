package client

import (
    "errors"
    "net/http"
    "strings"
    "testing"

    "github.com/jiashunx/eureka-client-go/meta"
)

func TestUnitHttpDoRequestSuccess(t *testing.T) {
    srv := unitServer(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(200)
        _, _ = w.Write([]byte("ok-body"))
    })
    defer srv.Close()

    hc := &HttpClient{}
    resp := hc.doRequest(200, &meta.EurekaServer{ServiceUrl: srv.URL}, "GET", "/x", nil)
    if resp.Error != nil {
        t.Fatalf("unexpected error: %v", resp.Error)
    }
    if resp.Body != "ok-body" {
        t.Fatalf("unexpected body: %q", resp.Body)
    }
    if resp.HttpResponse == nil || resp.HttpResponse.StatusCode != 200 {
        t.Fatalf("unexpected http response: %#v", resp.HttpResponse)
    }
    if len(resp.Responses) != 1 {
        t.Fatalf("expected 1 response, got %d", len(resp.Responses))
    }
}

func TestUnitHttpDoRequestNilServer(t *testing.T) {
    hc := &HttpClient{}
    resp := hc.doRequest(200, nil, "GET", "/x", nil)
    if resp.Error == nil {
        t.Fatalf("expected recover error for nil server")
    }
}

func TestUnitHttpDoRequestEmptyThenValid(t *testing.T) {
    srv := unitServer(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(200)
    })
    defer srv.Close()

    hc := &HttpClient{}
    resp := hc.doRequest(200, &meta.EurekaServer{ServiceUrl: ",, " + srv.URL}, "GET", "/x", nil)
    if resp.Error != nil {
        t.Fatalf("unexpected error: %v", resp.Error)
    }
}

func TestUnitHttpDoRequestAllEmpty(t *testing.T) {
    hc := &HttpClient{}
    resp := hc.doRequest(200, &meta.EurekaServer{ServiceUrl: " , "}, "GET", "/x", nil)
    if resp.Error == nil || !strings.Contains(resp.Error.Error(), "no eureka server service address available") {
        t.Fatalf("expected no-address error, got %v", resp.Error)
    }
}

func TestUnitHttpDoRequestMalformedUrl(t *testing.T) {
    hc := &HttpClient{}
    resp := hc.doRequest(200, &meta.EurekaServer{ServiceUrl: "http://[::1"}, "GET", "/x", nil)
    if resp.Error == nil {
        t.Fatalf("expected url parse error")
    }
}

func TestUnitHttpDoRequestAuth(t *testing.T) {
    var user, pass string
    var hasAuth bool
    restore := unitInstallDefaultClient(func(r *http.Request) (*http.Response, error) {
        user, pass, hasAuth = r.BasicAuth()
        return unitResponse(r, 200, "ok"), nil
    })
    defer restore()

    hc := &HttpClient{}
    // URL userinfo has the highest priority.
    resp := hc.doRequest(200, &meta.EurekaServer{
        ServiceUrl: "http://u1:p1@example.com/eureka",
        Username:   "u2",
        Password:   "p2",
    }, "GET", "/x", nil)
    if resp.Error != nil {
        t.Fatalf("unexpected error: %v", resp.Error)
    }
    if !hasAuth || user != "u1" || pass != "p1" {
        t.Fatalf("expected userinfo auth u1/p1, got hasAuth=%v user=%q pass=%q", hasAuth, user, pass)
    }
    // Fallback to server credentials.
    _ = hc.doRequest(200, &meta.EurekaServer{
        ServiceUrl: "http://example.com/eureka",
        Username:   "u2",
        Password:   "p2",
    }, "GET", "/x", nil)
    if !hasAuth || user != "u2" || pass != "p2" {
        t.Fatalf("expected server auth u2/p2, got hasAuth=%v user=%q pass=%q", hasAuth, user, pass)
    }
    // No credentials at all.
    hasAuth = true
    _ = hc.doRequest(200, &meta.EurekaServer{ServiceUrl: "http://example.com/eureka"}, "GET", "/x", nil)
    if hasAuth {
        t.Fatalf("expected no auth header")
    }
}

func TestUnitHttpDoRequestNoPort(t *testing.T) {
    var gotURL string
    restore := unitInstallDefaultClient(func(r *http.Request) (*http.Response, error) {
        gotURL = r.URL.String()
        return unitResponse(r, 200, "ok"), nil
    })
    defer restore()

    hc := &HttpClient{}
    // With port.
    _ = hc.doRequest(200, &meta.EurekaServer{ServiceUrl: "http://example.com:8761/eureka"}, "GET", "/x", nil)
    if gotURL != "http://example.com:8761/eureka/x" {
        t.Fatalf("unexpected url with port: %q", gotURL)
    }
    // Without port (URL.Port() == "" branch).
    _ = hc.doRequest(200, &meta.EurekaServer{ServiceUrl: "http://example.com/eureka"}, "GET", "/x", nil)
    if gotURL != "http://example.com/eureka/x" {
        t.Fatalf("unexpected url without port: %q", gotURL)
    }
}

func TestUnitHttpDoRequestTimeout(t *testing.T) {
    srv := unitServer(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(200)
    })
    defer srv.Close()

    hc := &HttpClient{}
    resp := hc.doRequest(200, &meta.EurekaServer{
        ServiceUrl:            srv.URL,
        ReadTimeoutSeconds:    1,
        ConnectTimeoutSeconds: 1,
    }, "GET", "/x", nil)
    if resp.Error != nil {
        t.Fatalf("unexpected error: %v", resp.Error)
    }
}

func TestUnitHttpDoRequestWrongStatus(t *testing.T) {
    srv := unitServer(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(500)
    })
    defer srv.Close()

    hc := &HttpClient{}
    resp := hc.doRequest(200, &meta.EurekaServer{ServiceUrl: srv.URL}, "GET", "/x", nil)
    if resp.Error == nil || !strings.Contains(resp.Error.Error(), "http response code is incorrect") {
        t.Fatalf("expected status error, got %v", resp.Error)
    }
}

func TestUnitHttpDoRequestNewRequestError(t *testing.T) {
    hc := &HttpClient{}
    resp := hc.doRequest(200, &meta.EurekaServer{ServiceUrl: "http://example.com/eureka"}, "GE T", "/x", nil)
    if resp.Error == nil {
        t.Fatalf("expected http.NewRequest error")
    }
}

func TestUnitHttpDoRequestReadBodyError(t *testing.T) {
    restore := unitInstallDefaultClient(func(r *http.Request) (*http.Response, error) {
        return &http.Response{
            StatusCode: 200,
            Body:       unitErrReader{},
            Header:     make(http.Header),
            Request:    r,
        }, nil
    })
    defer restore()

    hc := &HttpClient{}
    resp := hc.doRequest(200, &meta.EurekaServer{ServiceUrl: "http://example.com/eureka"}, "GET", "/x", nil)
    if resp.Error == nil {
        t.Fatalf("expected body read error")
    }
}

func TestUnitHttpDoRequestTransportError(t *testing.T) {
    restore := unitInstallDefaultClient(func(r *http.Request) (*http.Response, error) {
        return nil, errors.New("unit dial failure")
    })
    defer restore()

    hc := &HttpClient{}
    resp := hc.doRequest(200, &meta.EurekaServer{ServiceUrl: "http://example.com/eureka"}, "GET", "/x", nil)
    if resp.Error == nil {
        t.Fatalf("expected transport error")
    }
}

func TestUnitHttpRegister(t *testing.T) {
    inst := unitUpInstance()

    // nil instance
    hc := &HttpClient{}
    if resp := hc.Register(&meta.EurekaServer{}, nil); resp.Error == nil {
        t.Fatalf("expected nil instance error")
    }

    // invalid instance (Check error is hard to trigger, use valid here)
    srv := unitServer(func(w http.ResponseWriter, r *http.Request) {
        if r.Method != "POST" {
            t.Errorf("expected POST, got %s", r.Method)
        }
        if r.URL.Path != "/apps/UNITAPP" {
            t.Errorf("unexpected path: %s", r.URL.Path)
        }
        if r.Header.Get("Content-Type") != "application/json" {
            t.Errorf("unexpected content-type: %s", r.Header.Get("Content-Type"))
        }
        w.WriteHeader(204)
    })
    defer srv.Close()

    resp := hc.Register(&meta.EurekaServer{ServiceUrl: srv.URL}, inst)
    if resp.Error != nil {
        t.Fatalf("unexpected register error: %v", resp.Error)
    }
    if resp.StatusCode != 204 {
        t.Fatalf("expected 204, got %d", resp.StatusCode)
    }
}

func TestUnitHttpRegisterRecover(t *testing.T) {
    // A logger whose Tracef panics on first use triggers Register's recover block.
    hc := &HttpClient{Logger: &unitPanicOnceLogger{}}
    resp := hc.Register(&meta.EurekaServer{}, nil)
    if resp.Error == nil {
        t.Fatalf("expected recover error")
    }
    if !strings.Contains(resp.Error.Error(), "recover error") {
        t.Fatalf("unexpected error: %v", resp.Error)
    }
}

func TestUnitHttpUnRegisterHeartbeat(t *testing.T) {
    var gotMethods []string
    srv := unitServer(func(w http.ResponseWriter, r *http.Request) {
        gotMethods = append(gotMethods, r.Method)
        w.WriteHeader(200)
    })
    defer srv.Close()

    hc := &HttpClient{}
    if resp := hc.UnRegister(&meta.EurekaServer{ServiceUrl: srv.URL}, "APP", "i-1"); resp.Error != nil {
        t.Fatalf("unexpected unregister error: %v", resp.Error)
    }
    if resp := hc.Heartbeat(&meta.EurekaServer{ServiceUrl: srv.URL}, "APP", "i-1"); resp.Error != nil {
        t.Fatalf("unexpected heartbeat error: %v", resp.Error)
    }
    if len(gotMethods) != 2 || gotMethods[0] != "DELETE" || gotMethods[1] != "PUT" {
        t.Fatalf("unexpected methods: %v", gotMethods)
    }
}

func unitQueryHandler(w http.ResponseWriter, r *http.Request) {
    p := r.URL.Path
    switch {
    case p == "/apps", strings.HasPrefix(p, "/vips/"), strings.HasPrefix(p, "/svips/"):
        w.WriteHeader(200)
        _, _ = w.Write([]byte(unitAppsBody("APP")))
    case strings.HasPrefix(p, "/instances/"):
        w.WriteHeader(200)
        _, _ = w.Write([]byte(unitInstanceBody()))
    case strings.HasPrefix(p, "/apps/"):
        if strings.Count(p, "/") == 2 {
            w.WriteHeader(200)
            _, _ = w.Write([]byte(unitInstancesBody()))
        } else {
            w.WriteHeader(200)
            _, _ = w.Write([]byte(unitInstanceBody()))
        }
    default:
        w.WriteHeader(404)
    }
}

func TestUnitHttpQueries(t *testing.T) {
    srv := unitServer(unitQueryHandler)
    defer srv.Close()
    server := &meta.EurekaServer{ServiceUrl: srv.URL}

    hc := &HttpClient{}

    apps := hc.QueryApps(server)
    if apps.Error != nil || len(apps.Apps) != 1 || apps.StatusCode != 200 {
        t.Fatalf("QueryApps failed: %#v", apps)
    }
    instances := hc.QueryApp(server, "APP")
    if instances.Error != nil || len(instances.Instances) != 1 {
        t.Fatalf("QueryApp failed: %#v", instances)
    }
    one := hc.QueryAppInstance(server, "APP", "i-1")
    if one.Error != nil || one.Instance == nil {
        t.Fatalf("QueryAppInstance failed: %#v", one)
    }
    one = hc.QueryInstance(server, "i-1")
    if one.Error != nil || one.Instance == nil {
        t.Fatalf("QueryInstance failed: %#v", one)
    }
    if r := hc.QueryVipApps(server, "vip1"); r.Error != nil || len(r.Apps) != 1 {
        t.Fatalf("QueryVipApps failed: %#v", r)
    }
    if r := hc.QuerySvipApps(server, "svip1"); r.Error != nil || len(r.Apps) != 1 {
        t.Fatalf("QuerySvipApps failed: %#v", r)
    }
}

func TestUnitHttpSimpleWrappers(t *testing.T) {
    srv := unitServer(unitQueryHandler)
    defer srv.Close()
    url := srv.URL

    hc := &HttpClient{}
    if r := hc.SimpleQueryApps(url); r.Error != nil {
        t.Fatalf("SimpleQueryApps: %v", r.Error)
    }
    if r := hc.SimpleQueryApp(url, "APP"); r.Error != nil {
        t.Fatalf("SimpleQueryApp: %v", r.Error)
    }
    if r := hc.SimpleQueryAppInstance(url, "APP", "i-1"); r.Error != nil {
        t.Fatalf("SimpleQueryAppInstance: %v", r.Error)
    }
    if r := hc.SimpleQueryInstance(url, "i-1"); r.Error != nil {
        t.Fatalf("SimpleQueryInstance: %v", r.Error)
    }
    if r := hc.SimpleQueryVipApps(url, "vip1"); r.Error != nil {
        t.Fatalf("SimpleQueryVipApps: %v", r.Error)
    }
    if r := hc.SimpleQuerySvipApps(url, "svip1"); r.Error != nil {
        t.Fatalf("SimpleQuerySvipApps: %v", r.Error)
    }
}

func TestUnitHttpSimpleRegisterUnRegisterHeartbeat(t *testing.T) {
    srv := unitServer(func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case "POST":
            w.WriteHeader(204)
        default:
            w.WriteHeader(200)
        }
    })
    defer srv.Close()

    hc := &HttpClient{}
    if r := hc.SimpleRegister(srv.URL, unitUpInstance()); r.Error != nil || r.StatusCode != 204 {
        t.Fatalf("SimpleRegister failed: %#v", r)
    }
    if r := hc.SimpleUnRegister(srv.URL, "APP", "i-1"); r.Error != nil {
        t.Fatalf("SimpleUnRegister failed: %v", r.Error)
    }
    if r := hc.SimpleHeartbeat(srv.URL, "APP", "i-1"); r.Error != nil {
        t.Fatalf("SimpleHeartbeat failed: %v", r.Error)
    }
}

func TestUnitHttpGetInstancesNilServer(t *testing.T) {
    hc := &HttpClient{}
    if r := hc.QueryApp(nil, "APP"); r.Error == nil {
        t.Fatalf("expected error for nil server")
    }
    if r := hc.QueryInstance(nil, "i-1"); r.Error == nil {
        t.Fatalf("expected error for nil server")
    }
}

func TestUnitHttpChangeStatusModifyMetadata(t *testing.T) {
    var paths []string
    var queries []string
    srv := unitServer(func(w http.ResponseWriter, r *http.Request) {
        paths = append(paths, r.URL.Path)
        queries = append(queries, r.URL.RawQuery)
        w.WriteHeader(200)
    })
    defer srv.Close()
    server := &meta.EurekaServer{ServiceUrl: srv.URL}

    hc := &HttpClient{}
    if r := hc.ChangeStatus(server, "APP", "i-1", meta.StatusOutOfService); r.Error != nil {
        t.Fatalf("ChangeStatus: %v", r.Error)
    }
    if r := hc.SimpleChangeStatus(srv.URL, "APP", "i-1", meta.StatusUp); r.Error != nil {
        t.Fatalf("SimpleChangeStatus: %v", r.Error)
    }
    if r := hc.ModifyMetadata(server, "APP", "i-1", map[string]string{"a": "b"}); r.Error != nil {
        t.Fatalf("ModifyMetadata: %v", r.Error)
    }
    // nil metadata exercises the truncation branch.
    if r := hc.ModifyMetadata(server, "APP", "i-1", nil); r.Error != nil {
        t.Fatalf("ModifyMetadata nil: %v", r.Error)
    }
    if r := hc.SimpleModifyMetadata(srv.URL, "APP", "i-1", map[string]string{"a": "b"}); r.Error != nil {
        t.Fatalf("SimpleModifyMetadata: %v", r.Error)
    }
    if len(paths) == 0 {
        t.Fatalf("no request paths captured")
    }
}

func TestUnitHttpGetAppsErrors(t *testing.T) {
    cases := []struct {
        name string
        body string
    }{
        {"malformed", "not-json"},
        {"array", "[1,2,3]"},
        {"missing-applications", "{}"},
        {"missing-application", `{"applications":{}}`},
        {"bad-app-entry", `{"applications":{"application":[1]}}`},
    }
    for _, c := range cases {
        c := c
        t.Run(c.name, func(t *testing.T) {
            srv := unitServer(func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(200)
                _, _ = w.Write([]byte(c.body))
            })
            defer srv.Close()
            hc := &HttpClient{}
            resp := hc.QueryApps(&meta.EurekaServer{ServiceUrl: srv.URL})
            if resp.Error == nil {
                t.Fatalf("expected error for %s", c.name)
            }
        })
    }

    // server-side error response propagates.
    hc := &HttpClient{}
    if resp := hc.QueryApps(nil); resp.Error == nil {
        t.Fatalf("expected error for nil server")
    }
}

func TestUnitHttpGetInstancesErrors(t *testing.T) {
    cases := []struct {
        name string
        body string
    }{
        {"malformed", "not-json"},
        {"array", "[1,2,3]"},
        {"missing-application", "{}"},
        {"missing-instance", `{"application":{}}`},
        {"bad-instance-entry", `{"application":{"instance":[1]}}`},
    }
    for _, c := range cases {
        c := c
        t.Run(c.name, func(t *testing.T) {
            srv := unitServer(func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(200)
                _, _ = w.Write([]byte(c.body))
            })
            defer srv.Close()
            hc := &HttpClient{}
            resp := hc.QueryApp(&meta.EurekaServer{ServiceUrl: srv.URL}, "APP")
            if resp.Error == nil {
                t.Fatalf("expected error for %s", c.name)
            }
        })
    }
}

func TestUnitHttpGetInstanceErrors(t *testing.T) {
    cases := []struct {
        name string
        body string
    }{
        {"malformed", "not-json"},
        {"array", "[1,2,3]"},
        {"missing-instance", "{}"},
    }
    for _, c := range cases {
        c := c
        t.Run(c.name, func(t *testing.T) {
            srv := unitServer(func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(200)
                _, _ = w.Write([]byte(c.body))
            })
            defer srv.Close()
            hc := &HttpClient{}
            resp := hc.QueryInstance(&meta.EurekaServer{ServiceUrl: srv.URL}, "i-1")
            if resp.Error == nil {
                t.Fatalf("expected error for %s", c.name)
            }
        })
    }
}

func TestUnitHttpCommonHttpError(t *testing.T) {
    hc := &HttpClient{}
    resp := hc.commonHttp(200, nil, "GET", "/x", nil)
    if resp.Error == nil {
        t.Fatalf("expected error")
    }
    if resp.Response == nil || resp.Response.Error == nil {
        t.Fatalf("expected underlying response error")
    }
    if resp.StatusCode != 0 {
        t.Fatalf("expected zero status, got %d", resp.StatusCode)
    }
}
