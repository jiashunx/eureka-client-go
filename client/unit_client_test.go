package client

import (
    "context"
    "strings"
    "testing"

    "github.com/jiashunx/eureka-client-go/log"
    "github.com/jiashunx/eureka-client-go/meta"
)

func TestUnitClientNewEurekaClient(t *testing.T) {
    // nil config -> recover into error.
    if _, err := NewEurekaClient(nil); err == nil {
        t.Fatalf("expected error for nil config")
    }

    // valid config.
    c, err := NewEurekaClient(unitBaseConfig(""))
    if err != nil {
        t.Fatalf("NewEurekaClient: %v", err)
    }
    if c == nil || c.UUID == "" {
        t.Fatalf("expected non-nil client with UUID")
    }
    if c.RegistryClient() == nil || c.DiscoveryClient() == nil || c.HttpClient() == nil {
        t.Fatalf("expected non-nil sub clients")
    }

    // options nil.
    if _, err := NewEurekaClientWithOptions(unitBaseConfig(""), nil); err != nil {
        t.Fatalf("NewEurekaClientWithOptions nil options: %v", err)
    }

    // options with HeartbeatFunc.
    called := false
    c2, err := NewEurekaClientWithOptions(unitBaseConfig(""), &EurekaConfigOptions{
        HeartbeatFunc: func(*CommonResponse) { called = true },
    })
    if err != nil {
        t.Fatalf("NewEurekaClientWithOptions: %v", err)
    }
    if c2.RegistryClient().HeartbeatFunc == nil {
        t.Fatalf("expected HeartbeatFunc wired")
    }
    c2.RegistryClient().HeartbeatFunc(&CommonResponse{})
    if !called {
        t.Fatalf("expected HeartbeatFunc call")
    }
}

func TestUnitClientSetGetLogger(t *testing.T) {
    c, err := NewEurekaClient(unitBaseConfig(""))
    if err != nil {
        t.Fatalf("NewEurekaClient: %v", err)
    }
    // default logger.
    if c.GetLogger() == nil {
        t.Fatalf("expected default logger")
    }
    // set nil -> error.
    if err := c.SetLogger(nil); err == nil {
        t.Fatalf("expected error for nil logger")
    }
    // set non-nil.
    logger := log.DefaultLogger()
    if err := c.SetLogger(logger); err != nil {
        t.Fatalf("SetLogger failed: %v", err)
    }
    if c.HttpClient().Logger != logger {
        t.Fatalf("expected logger propagated to HttpClient")
    }

    // default GetLogger on zero-value client.
    if (&EurekaClient{}).GetLogger() == nil {
        t.Fatalf("expected default logger for zero value")
    }
}

func TestUnitClientStartNilConfig(t *testing.T) {
    resp := (&EurekaClient{}).Start()
    if resp.Error == nil || !strings.Contains(resp.Error.Error(), "EurekaConfig is nil") {
        t.Fatalf("unexpected error: %v", resp.Error)
    }
}

func TestUnitClientStartStop(t *testing.T) {
    srv := unitServer(unitRegistryHandler(204, 200, 200))
    defer srv.Close()

    c := unitNewClient(t, srv.URL)
    resp := c.Start()
    if resp.Error != nil {
        t.Fatalf("Start failed: %v", resp.Error)
    }
    // already running.
    if r := c.Start(); r.Error == nil || !strings.Contains(r.Error.Error(), "still running") {
        t.Fatalf("expected still running error, got %v", r.Error)
    }
    // stop.
    if r := c.Stop(); r.Error != nil {
        t.Fatalf("Stop failed: %v", r.Error)
    }
    // restart after stop (covers the ctx.Done branch in StartWithCtx).
    if r := c.Start(); r.Error != nil {
        t.Fatalf("restart failed: %v", r.Error)
    }
    if r := c.Stop(); r.Error != nil {
        t.Fatalf("Stop after restart failed: %v", r.Error)
    }

    // Stop on a fresh (never started) client.
    c2 := unitNewClient(t, srv.URL)
    if r := c2.Stop(); r.Error == nil {
        t.Fatalf("expected Stop error for non-started client")
    }
}

func TestUnitClientStartWithCtx(t *testing.T) {
    srv := unitServer(unitRegistryHandler(204, 200, 200))
    defer srv.Close()

    c := unitNewClient(t, srv.URL)
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    resp := c.StartWithCtx(ctx)
    if resp.Error != nil {
        t.Fatalf("StartWithCtx failed: %v", resp.Error)
    }
    if r := c.Stop(); r.Error != nil {
        t.Fatalf("Stop failed: %v", r.Error)
    }
}

func TestUnitClientStartRecover(t *testing.T) {
    c := unitNewClient(t, "")
    c.registryClient = nil
    resp := c.StartWithCtx(context.Background())
    if resp.Error == nil || !strings.Contains(resp.Error.Error(), "recover error") {
        t.Fatalf("expected recover error, got %v", resp.Error)
    }
}

func TestUnitClientStartRegistryFailure(t *testing.T) {
    srv := unitServer(unitRegistryHandler(500, 200, 200))
    defer srv.Close()

    c := unitNewClient(t, srv.URL)
    resp := c.Start()
    if resp.Error == nil {
        t.Fatalf("expected start failure")
    }
}

func TestUnitClientStopError(t *testing.T) {
    srv := unitServer(unitRegistryHandler(204, 500, 200))
    defer srv.Close()

    c := unitNewClient(t, srv.URL)
    if resp := c.Start(); resp.Error != nil {
        t.Fatalf("Start failed: %v", resp.Error)
    }
    if resp := c.Stop(); resp.Error == nil {
        t.Fatalf("expected Stop error")
    }
    // ctx was not cancelled because UnRegister failed; force stop to clean up.
    c.ForceStop()
}

func TestUnitClientForceStop(t *testing.T) {
    // nil ctx -> only logs.
    (&EurekaClient{}).ForceStop()

    srv := unitServer(unitRegistryHandler(204, 200, 200))
    defer srv.Close()

    c := unitNewClient(t, srv.URL)
    if resp := c.Start(); resp.Error != nil {
        t.Fatalf("Start failed: %v", resp.Error)
    }
    c.ForceStop() // ctx active -> unregister + cancel
    c.ForceStop() // ctx done -> no-op
}

func TestUnitClientChangeStatusMetadata(t *testing.T) {
    srv := unitServer(unitRegistryHandler(204, 200, 200))
    defer srv.Close()

    c := unitNewClient(t, srv.URL)
    // error paths: client not started -> exec error propagated by wrappers.
    if resp := c.ChangeStatus(meta.StatusUp); resp.Error == nil {
        t.Fatalf("expected ChangeStatus error for non-started client")
    }
    if resp := c.ChangeMetadata(map[string]string{"k": "v"}); resp.Error == nil {
        t.Fatalf("expected ChangeMetadata error for non-started client")
    }
    if resp := c.Start(); resp.Error != nil {
        t.Fatalf("Start failed: %v", resp.Error)
    }
    if resp := c.ChangeStatus(meta.StatusUp); resp.Error != nil {
        t.Fatalf("ChangeStatus failed: %v", resp.Error)
    }
    c.discoveryClient.Config.Metadata = map[string]string{}
    if resp := c.ChangeMetadata(map[string]string{"k": "v"}); resp.Error != nil {
        t.Fatalf("ChangeMetadata failed: %v", resp.Error)
    }
    if r := c.Stop(); r.Error != nil {
        t.Fatalf("Stop failed: %v", r.Error)
    }
}

func TestUnitClientAccessWrappers(t *testing.T) {
    c := unitNewClient(t, "")
    c.ctx = context.Background()
    c.discoveryClient.Apps = map[string][]*meta.AppInfo{
        "defaultZone": {{Name: "UNITAPP", Instances: []*meta.InstanceInfo{unitUpInstance()}}},
    }

    if app, err := c.AccessApp("UNITAPP"); err != nil || app == nil {
        t.Fatalf("AccessApp failed: %v", err)
    }
    if apps, err := c.AccessAppsByVip("vip1"); err != nil || len(apps) != 1 {
        t.Fatalf("AccessAppsByVip failed: %v", err)
    }
    if inst, err := c.AccessAppInstanceByVip("vip1"); err != nil || inst == nil {
        t.Fatalf("AccessAppInstanceByVip failed: %v", err)
    }
    if apps, err := c.AccessAppsBySvip("svip1"); err != nil || len(apps) != 1 {
        t.Fatalf("AccessAppsBySvip failed: %v", err)
    }
    if inst, err := c.AccessAppInstanceBySvip("svip1"); err != nil || inst == nil {
        t.Fatalf("AccessAppInstanceBySvip failed: %v", err)
    }
    if insts, err := c.AccessInstancesByVip("vip1"); err != nil || len(insts) != 1 {
        t.Fatalf("AccessInstancesByVip failed: %v", err)
    }
    if inst, err := c.AccessInstanceByVip("vip1"); err != nil || inst == nil {
        t.Fatalf("AccessInstanceByVip failed: %v", err)
    }
    if insts, err := c.AccessInstancesBySvip("svip1"); err != nil || len(insts) != 1 {
        t.Fatalf("AccessInstancesBySvip failed: %v", err)
    }
    if inst, err := c.AccessInstanceBySvip("svip1"); err != nil || inst == nil {
        t.Fatalf("AccessInstanceBySvip failed: %v", err)
    }

    // error path: empty apps -> no available service.
    c.discoveryClient.Apps = map[string][]*meta.AppInfo{}
    if _, err := c.AccessApp("UNITAPP"); err == nil {
        t.Fatalf("expected AccessApp error")
    }
    if _, err := c.AccessAppsByVip("vip1"); err == nil {
        t.Fatalf("expected AccessAppsByVip error")
    }
    if _, err := c.AccessAppInstanceByVip("vip1"); err == nil {
        t.Fatalf("expected AccessAppInstanceByVip error")
    }
    if _, err := c.AccessAppsBySvip("svip1"); err == nil {
        t.Fatalf("expected AccessAppsBySvip error")
    }
    if _, err := c.AccessAppInstanceBySvip("svip1"); err == nil {
        t.Fatalf("expected AccessAppInstanceBySvip error")
    }
    if _, err := c.AccessInstancesByVip("vip1"); err == nil {
        t.Fatalf("expected AccessInstancesByVip error")
    }
    if _, err := c.AccessInstanceByVip("vip1"); err == nil {
        t.Fatalf("expected AccessInstanceByVip error")
    }
    if _, err := c.AccessInstancesBySvip("svip1"); err == nil {
        t.Fatalf("expected AccessInstancesBySvip error")
    }
    if _, err := c.AccessInstanceBySvip("svip1"); err == nil {
        t.Fatalf("expected AccessInstanceBySvip error")
    }
}

func TestUnitClientExec(t *testing.T) {
    c := unitNewClient(t, "")

    // ctx nil -> not started error.
    if _, err := c.exec("X", func(params ...any) (any, error) { return nil, nil }); err == nil {
        t.Fatalf("expected not started error")
    }

    // ctx done -> stopped error.
    ctx, cancel := context.WithCancel(context.Background())
    cancel()
    c.ctx = ctx
    if _, err := c.exec("X", func(params ...any) (any, error) { return nil, nil }); err == nil {
        t.Fatalf("expected stopped error")
    }

    // ctx active with params (logging branch).
    c.ctx = context.Background()
    ret, err := c.exec("X", func(params ...any) (any, error) { return params[0], nil }, "value")
    if err != nil || ret != "value" {
        t.Fatalf("exec unexpected result: %v %v", ret, err)
    }

    // panicking closure -> recover.
    if _, err := c.exec("X", func(params ...any) (any, error) { panic("boom") }); err == nil {
        t.Fatalf("expected recover error")
    } else if !strings.Contains(err.Error(), "recover error") {
        t.Fatalf("unexpected error: %v", err)
    }
}
