package client

import (
    "context"
    "net/http"
    "strings"
    "testing"
    "time"

    "github.com/jiashunx/eureka-client-go/meta"
)

func unitRegistryHandler(registerStatus, deleteStatus, putStatus int) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case "POST":
            w.WriteHeader(registerStatus)
        case "DELETE":
            w.WriteHeader(deleteStatus)
        case "PUT":
            w.WriteHeader(putStatus)
        case "GET":
            w.WriteHeader(200)
            _, _ = w.Write([]byte(unitAppsBody("UNITAPP")))
        default:
            w.WriteHeader(404)
        }
    }
}

func TestUnitRegistryIsEnabled(t *testing.T) {
    c := unitNewClient(t, "")
    rc := c.RegistryClient()
    if b, err := rc.isEnabled(); !b || err != nil {
        t.Fatalf("expected enabled, got %v %v", b, err)
    }
    fa := false
    rc.Config.RegistryEnabled = &fa
    if b, err := rc.isEnabled(); b || err == nil {
        t.Fatalf("expected disabled, got %v %v", b, err)
    }
}

func TestUnitRegistryStart(t *testing.T) {
    srv := unitServer(unitRegistryHandler(204, 200, 200))
    defer srv.Close()

    // InstanceEnabledOnIt false -> STARTING.
    c := unitNewClient(t, srv.URL)
    rc := c.RegistryClient()
    fa := false
    rc.Config.InstanceEnabledOnIt = &fa
    ctx, cancel := context.WithCancel(context.Background())
    resp := rc.start(ctx)
    cancel()
    if resp.Error != nil {
        t.Fatalf("start failed: %v", resp.Error)
    }
    if rc.status != meta.StatusStarting {
        t.Fatalf("expected STARTING, got %s", rc.status)
    }
    if !rc.heartbeat {
        t.Fatalf("expected heartbeat enabled after success")
    }
    if rc.GetLogger() == nil {
        t.Fatalf("expected logger")
    }

    // InstanceEnabledOnIt true -> UP.
    c2 := unitNewClient(t, srv.URL)
    rc2 := c2.RegistryClient()
    tr := true
    rc2.Config.InstanceEnabledOnIt = &tr
    ctx2, cancel2 := context.WithCancel(context.Background())
    resp2 := rc2.start(ctx2)
    cancel2()
    if resp2.Error != nil {
        t.Fatalf("start failed: %v", resp2.Error)
    }
    if rc2.status != meta.StatusUp {
        t.Fatalf("expected UP, got %s", rc2.status)
    }
}

func TestUnitRegistryStartFailure(t *testing.T) {
    srv := unitServer(unitRegistryHandler(500, 200, 200))
    defer srv.Close()

    c := unitNewClient(t, srv.URL)
    rc := c.RegistryClient()
    ctx, cancel := context.WithCancel(context.Background())
    resp := rc.start(ctx)
    cancel()
    if resp.Error == nil {
        t.Fatalf("expected register failure")
    }
    if rc.heartbeat {
        t.Fatalf("expected heartbeat disabled after failure")
    }
}

func TestUnitRegistryStartRecover(t *testing.T) {
    rc := &RegistryClient{}
    resp := rc.start(context.Background())
    if resp.Error == nil {
        t.Fatalf("expected recover error")
    }
    if !strings.Contains(resp.Error.Error(), "recover error") {
        t.Fatalf("unexpected error: %v", resp.Error)
    }
}

func TestUnitRegistryBeat(t *testing.T) {
    srv := unitServer(unitRegistryHandler(204, 200, 200))
    defer srv.Close()

    c := unitNewClient(t, srv.URL)
    rc := c.RegistryClient()
    ctx, cancel := context.WithCancel(context.Background())
    done := make(chan struct{})
    go func() {
        rc.beat(ctx)
        close(done)
    }()
    time.Sleep(50 * time.Millisecond)
    cancel()
    select {
    case <-done:
    case <-time.After(3 * time.Second):
        t.Fatalf("beat loop did not stop")
    }
}

func TestUnitRegistryBeat0(t *testing.T) {
    srv := unitServer(unitRegistryHandler(204, 200, 200))
    defer srv.Close()

    // heartbeat false -> no Heartbeat(), HeartbeatFunc invoked.
    c := unitNewClient(t, srv.URL)
    rc := c.RegistryClient()
    called := make(chan *CommonResponse, 1)
    rc.HeartbeatFunc = func(r *CommonResponse) { called <- r }
    rc.heartbeat = false
    if resp := rc.beat0(context.Background()); resp.Error != nil {
        t.Fatalf("beat0 unexpected error: %v", resp.Error)
    }
    select {
    case <-called:
    case <-time.After(time.Second):
        t.Fatalf("HeartbeatFunc not called")
    }

    // isEnabled error path.
    c2 := unitNewClient(t, srv.URL)
    rc2 := c2.RegistryClient()
    fa := false
    rc2.Config.RegistryEnabled = &fa
    rc2.heartbeat = false
    if resp := rc2.beat0(context.Background()); resp.Error == nil {
        t.Fatalf("expected isEnabled error")
    }

    // heartbeat true & status UP -> Heartbeat().
    c3 := unitNewClient(t, srv.URL)
    rc3 := c3.RegistryClient()
    rc3.heartbeat = true
    rc3.status = meta.StatusUp
    if resp := rc3.beat0(context.Background()); resp.Error != nil {
        t.Fatalf("beat0 heartbeat unexpected error: %v", resp.Error)
    }

    // heartbeat true but status != UP -> no Heartbeat() call (nil error).
    c4 := unitNewClient(t, srv.URL)
    rc4 := c4.RegistryClient()
    rc4.heartbeat = true
    rc4.status = meta.StatusDown
    if resp := rc4.beat0(context.Background()); resp.Error != nil {
        t.Fatalf("beat0 unexpected error: %v", resp.Error)
    }
}

func TestUnitRegistryRegisterHeartbeatUnRegister(t *testing.T) {
    srv := unitServer(unitRegistryHandler(204, 200, 200))
    defer srv.Close()

    c := unitNewClient(t, srv.URL)
    rc := c.RegistryClient()
    if r := rc.Register(meta.StatusUp); r.Error != nil {
        t.Fatalf("Register failed: %v", r.Error)
    }
    if r := rc.Heartbeat(); r.Error != nil {
        t.Fatalf("Heartbeat failed: %v", r.Error)
    }
    if r := rc.UnRegister(); r.Error != nil {
        t.Fatalf("UnRegister failed: %v", r.Error)
    }
    if rc.heartbeat {
        t.Fatalf("heartbeat should be false after successful unregister")
    }

    // UnRegister failure path.
    srvFail := unitServer(unitRegistryHandler(204, 500, 200))
    defer srvFail.Close()
    c2 := unitNewClient(t, srvFail.URL)
    rc2 := c2.RegistryClient()
    if r := rc2.UnRegister(); r.Error == nil {
        t.Fatalf("expected unregister failure")
    }
    if !rc2.heartbeat {
        t.Fatalf("heartbeat should be true after failed unregister")
    }
}

func TestUnitRegistryChangeStatus(t *testing.T) {
    srv := unitServer(unitRegistryHandler(204, 200, 200))
    defer srv.Close()

    c := unitNewClient(t, srv.URL)
    rc := c.RegistryClient()
    statuses := []meta.InstanceStatus{
        meta.StatusUp, meta.StatusDown, meta.StatusStarting, meta.StatusOutOfService, meta.StatusUnknown,
    }
    for _, st := range statuses {
        if r := rc.ChangeStatus(st); r.Error != nil {
            t.Fatalf("ChangeStatus(%s) failed: %v", st, r.Error)
        }
        if rc.status != st {
            t.Fatalf("expected status %s, got %s", st, rc.status)
        }
        if st == meta.StatusUp {
            if !rc.heartbeat {
                t.Fatalf("expected heartbeat true for UP")
            }
        } else if rc.heartbeat {
            t.Fatalf("expected heartbeat false for %s", st)
        }
    }

    // invalid status -> default branch.
    if r := rc.ChangeStatus("INVALID"); r.Error == nil {
        t.Fatalf("expected invalid status error")
    }

    // remote failure -> error propagated, status unchanged.
    srvFail := unitServer(unitRegistryHandler(204, 200, 500))
    defer srvFail.Close()
    c2 := unitNewClient(t, srvFail.URL)
    rc2 := c2.RegistryClient()
    rc2.status = meta.StatusDown
    if r := rc2.ChangeStatus(meta.StatusUp); r.Error == nil {
        t.Fatalf("expected change status failure")
    }
    if rc2.status != meta.StatusDown {
        t.Fatalf("status should not change on failure, got %s", rc2.status)
    }
}

func TestUnitRegistryChangeMetadata(t *testing.T) {
    srv := unitServer(unitRegistryHandler(204, 200, 200))
    defer srv.Close()

    c := unitNewClient(t, srv.URL)
    rc := c.RegistryClient()
    rc.Config.Metadata = map[string]string{}
    if r := rc.ChangeMetadata(map[string]string{"hello": "world"}); r.Error != nil {
        t.Fatalf("ChangeMetadata failed: %v", r.Error)
    }
    if rc.Config.Metadata["hello"] != "world" {
        t.Fatalf("metadata not updated: %#v", rc.Config.Metadata)
    }

    // failure -> metadata not updated.
    srvFail := unitServer(unitRegistryHandler(204, 200, 500))
    defer srvFail.Close()
    c2 := unitNewClient(t, srvFail.URL)
    rc2 := c2.RegistryClient()
    rc2.Config.Metadata = map[string]string{}
    if r := rc2.ChangeMetadata(map[string]string{"hello": "world"}); r.Error == nil {
        t.Fatalf("expected ChangeMetadata failure")
    }
    if len(rc2.Config.Metadata) != 0 {
        t.Fatalf("metadata should not be updated on failure: %#v", rc2.Config.Metadata)
    }
}

func TestUnitRegistryBuildInstanceInfo(t *testing.T) {
    c := unitNewClient(t, "")
    rc := c.RegistryClient()
    tr, fa := true, false

    // http enabled, prefer IP, empty metadata, URL fallbacks via http.
    rc.Config.PreferIpAddress = &tr
    rc.Config.NonSecurePortEnabled = &tr
    rc.Config.SecurePortEnabled = &fa
    rc.Config.Metadata = map[string]string{}
    inst, err := rc.buildInstanceInfo(meta.StatusUp, meta.Added)
    if err != nil {
        t.Fatalf("buildInstanceInfo: %v", err)
    }
    if inst.HostName != rc.Config.IpAddress {
        t.Fatalf("expected HostName = IpAddress, got %q", inst.HostName)
    }
    if !inst.Port.IsEnabled() || inst.SecurePort.IsEnabled() {
        t.Fatalf("unexpected port enabled flags: %v/%v", inst.Port.Enabled, inst.SecurePort.Enabled)
    }
    if !strings.HasPrefix(inst.HomePageUrl, "http://") {
        t.Fatalf("expected http HomePageUrl, got %q", inst.HomePageUrl)
    }
    if !strings.HasPrefix(inst.StatusPageUrl, "http://") || !strings.HasPrefix(inst.HealthCheckUrl, "http://") {
        t.Fatalf("expected http status/health urls, got %q %q", inst.StatusPageUrl, inst.HealthCheckUrl)
    }

    // prefer hostname (false branch).
    rc.Config.PreferIpAddress = &fa
    inst2, _ := rc.buildInstanceInfo(meta.StatusUp, meta.Added)
    if inst2.HostName != rc.Config.Hostname {
        t.Fatalf("expected HostName = config hostname, got %q", inst2.HostName)
    }

    // https only -> https fallbacks.
    rc.Config.NonSecurePortEnabled = &fa
    rc.Config.SecurePortEnabled = &tr
    rc.Config.HomePageUrl = ""
    rc.Config.StatusPageUrl = ""
    rc.Config.HealthCheckUrl = ""
    inst3, _ := rc.buildInstanceInfo(meta.StatusUp, meta.Added)
    if !strings.HasPrefix(inst3.HomePageUrl, "https://") || !strings.HasPrefix(inst3.StatusPageUrl, "https://") || !strings.HasPrefix(inst3.HealthCheckUrl, "https://") {
        t.Fatalf("expected https urls, got %q %q %q", inst3.HomePageUrl, inst3.StatusPageUrl, inst3.HealthCheckUrl)
    }

    // both ports disabled -> no fallback.
    rc.Config.NonSecurePortEnabled = &fa
    rc.Config.SecurePortEnabled = &fa
    rc.Config.HomePageUrl = ""
    rc.Config.StatusPageUrl = ""
    rc.Config.HealthCheckUrl = ""
    inst4, _ := rc.buildInstanceInfo(meta.StatusUp, meta.Added)
    if inst4.HomePageUrl != "" || inst4.StatusPageUrl != "" || inst4.HealthCheckUrl != "" {
        t.Fatalf("expected empty urls, got %q %q %q", inst4.HomePageUrl, inst4.StatusPageUrl, inst4.HealthCheckUrl)
    }

    // pre-set URLs -> no fallback (condition false).
    rc.Config.NonSecurePortEnabled = &tr
    rc.Config.HomePageUrl = "http://custom/"
    rc.Config.StatusPageUrl = "http://custom/status"
    rc.Config.HealthCheckUrl = "http://custom/health"
    inst5, _ := rc.buildInstanceInfo(meta.StatusUp, meta.Added)
    if inst5.HomePageUrl != "http://custom/" || inst5.StatusPageUrl != "http://custom/status" || inst5.HealthCheckUrl != "http://custom/health" {
        t.Fatalf("expected preset urls, got %q %q %q", inst5.HomePageUrl, inst5.StatusPageUrl, inst5.HealthCheckUrl)
    }

    // metadata copied.
    rc.Config.Metadata = map[string]string{"k": "v"}
    inst6, _ := rc.buildInstanceInfo(meta.StatusUp, meta.Added)
    if inst6.Metadata["k"] != "v" {
        t.Fatalf("expected metadata copied, got %#v", inst6.Metadata)
    }
}

func TestUnitRegistryUncovered(t *testing.T) {
    // start isEnabled error branch (beat goroutine still launched then stopped).
    c := unitNewClient(t, "")
    rc := c.RegistryClient()
    fa := false
    rc.Config.RegistryEnabled = &fa
    ctx, cancel := context.WithCancel(context.Background())
    resp := rc.start(ctx)
    cancel()
    if resp.Error == nil {
        t.Fatalf("expected start isEnabled error")
    }

    // beat0 recover branch: nil Config panics inside isEnabled.
    rc2 := &RegistryClient{}
    if resp := rc2.beat0(context.Background()); resp.Error == nil {
        t.Fatalf("expected beat0 recover error")
    }

    // isEnabled error branches for the public registry methods.
    c3 := unitNewClient(t, "")
    rc3 := c3.RegistryClient()
    rc3.Config.RegistryEnabled = &fa
    if r := rc3.Register(meta.StatusUp); r.Error == nil {
        t.Fatalf("expected Register isEnabled error")
    }
    if r := rc3.Heartbeat(); r.Error == nil {
        t.Fatalf("expected Heartbeat isEnabled error")
    }
    if r := rc3.UnRegister(); r.Error == nil {
        t.Fatalf("expected UnRegister isEnabled error")
    }
    if r := rc3.ChangeStatus(meta.StatusUp); r.Error == nil {
        t.Fatalf("expected ChangeStatus isEnabled error")
    }
    if r := rc3.ChangeMetadata(map[string]string{"k": "v"}); r.Error == nil {
        t.Fatalf("expected ChangeMetadata isEnabled error")
    }
}
