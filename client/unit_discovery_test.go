package client

import (
    "context"
    "net/http"
    "strings"
    "testing"
    "time"

    "github.com/jiashunx/eureka-client-go/meta"
)

func unitDiscoveryHandler(fail bool) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if fail {
            w.WriteHeader(500)
            return
        }
        w.WriteHeader(200)
        _, _ = w.Write([]byte(unitAppsBody("UNITAPP")))
    }
}

func TestUnitDiscoveryIsEnabled(t *testing.T) {
    c := unitNewClient(t, "")
    d := c.DiscoveryClient()
    if b, err := d.isEnabled(); !b || err != nil {
        t.Fatalf("expected enabled, got %v %v", b, err)
    }
    fa := false
    d.Config.DiscoveryEnabled = &fa
    if b, err := d.isEnabled(); b || err == nil {
        t.Fatalf("expected disabled, got %v %v", b, err)
    }
}

func TestUnitDiscoveryStart(t *testing.T) {
    srv := unitServer(unitDiscoveryHandler(false))
    defer srv.Close()

    c := unitNewClient(t, srv.URL)
    d := c.DiscoveryClient()
    ctx, cancel := context.WithCancel(context.Background())
    if resp := d.start(ctx); resp.Error != nil {
        t.Fatalf("start failed: %v", resp.Error)
    }
    cancel()

    // Discovery disabled still starts but loop skips Discovery0.
    c2 := unitNewClient(t, srv.URL)
    d2 := c2.DiscoveryClient()
    fa := false
    d2.Config.DiscoveryEnabled = &fa
    ctx2, cancel2 := context.WithCancel(context.Background())
    if resp := d2.start(ctx2); resp.Error != nil {
        t.Fatalf("start failed: %v", resp.Error)
    }
    cancel2()
}

func TestUnitDiscoveryLoop(t *testing.T) {
    srv := unitServer(unitDiscoveryHandler(false))
    defer srv.Close()

    c := unitNewClient(t, srv.URL)
    d := c.DiscoveryClient()
    ctx, cancel := context.WithCancel(context.Background())
    done := make(chan struct{})
    go func() {
        d.discovery(ctx)
        close(done)
    }()
    time.Sleep(50 * time.Millisecond)
    cancel()
    select {
    case <-done:
    case <-time.After(3 * time.Second):
        t.Fatalf("discovery loop did not stop")
    }
}

func TestUnitDiscovery0MultiZone(t *testing.T) {
    srv := unitServer(unitDiscoveryHandler(false))
    defer srv.Close()
    failSrv := unitServer(unitDiscoveryHandler(true))
    defer failSrv.Close()

    cfg := &meta.EurekaConfig{
        InstanceConfig: &meta.InstanceConfig{AppName: "a", InstanceId: "i", NonSecurePort: 18080},
        ClientConfig: &meta.ClientConfig{
            Region: "r",
            Zone:   "z1",
            AvailableZones: map[string]string{
                "r": "z1,z2,z3,z4",
            },
            ServiceUrlOfAllZone: map[string]string{
                "z1": srv.URL,
                "z2": srv.URL,
                "z3": failSrv.URL,
                "z4": "http://127.0.0.1:1",
            },
            RegistryFetchIntervalSeconds: 1,
        },
    }
    c, err := NewEurekaClient(cfg)
    if err != nil {
        t.Fatalf("NewEurekaClient: %v", err)
    }
    d := c.DiscoveryClient()
    apps, err := d.Discovery0()
    if err != nil {
        t.Fatalf("Discovery0 failed: %v", err)
    }
    if len(apps["z1"]) != 1 || len(apps["z2"]) != 1 {
        t.Fatalf("expected 1 app per healthy zone, got z1=%d z2=%d", len(apps["z1"]), len(apps["z2"]))
    }
    if len(apps["z3"]) != 0 {
        t.Fatalf("expected empty apps for error zone, got %d", len(apps["z3"]))
    }
    if len(apps["z4"]) != 0 {
        t.Fatalf("expected empty apps for unreachable zone, got %d", len(apps["z4"]))
    }
    if d.Apps == nil {
        t.Fatalf("expected discovery.Apps populated")
    }
}

func TestUnitDiscovery0Recover(t *testing.T) {
    srv := unitServer(unitDiscoveryHandler(false))
    defer srv.Close()

    c := unitNewClient(t, srv.URL)
    d := c.DiscoveryClient()
    d.Logger = &unitPanicOnceLogger{}
    if _, err := d.Discovery0(); err == nil {
        t.Fatalf("expected recover error")
    } else if !strings.Contains(err.Error(), "recover error") {
        t.Fatalf("unexpected error: %v", err)
    }
}

func TestUnitPublicQuery(t *testing.T) {
    srv := unitServer(unitDiscoveryHandler(false))
    defer srv.Close()

    c := unitNewClient(t, srv.URL)
    d := c.DiscoveryClient()

    // with params (logging path) + success.
    ret, err := d.publicQuery("TestQuery", func(params ...any) (any, error) {
        return params[0], nil
    }, "hello")
    if err != nil {
        t.Fatalf("publicQuery failed: %v", err)
    }
    if ret != "hello" {
        t.Fatalf("unexpected ret: %v", ret)
    }

    // without params.
    if _, err := d.publicQuery("TestQueryNoParams", func(params ...any) (any, error) {
        return "x", nil
    }); err != nil {
        t.Fatalf("publicQuery no-params failed: %v", err)
    }

    // isEnabled error path.
    fa := false
    d.Config.DiscoveryEnabled = &fa
    if _, err := d.publicQuery("TestDisabled", func(params ...any) (any, error) { return nil, nil }); err == nil {
        t.Fatalf("expected isEnabled error")
    }

    // Discovery0 error path (panic logger makes Discovery0 fail).
    c2 := unitNewClient(t, srv.URL)
    d2 := c2.DiscoveryClient()
    d2.Logger = &unitPanicOnceLogger{}
    if _, err := d2.publicQuery("TestDiscoveryErr", func(params ...any) (any, error) { return nil, nil }); err == nil {
        t.Fatalf("expected Discovery0 error")
    }

    // params logging panic -> recover path.
    c3 := unitNewClient(t, srv.URL)
    d3 := c3.DiscoveryClient()
    d3.Logger = &unitPanicOnceLogger{}
    if _, err := d3.publicQuery("TestPanic", func(params ...any) (any, error) { return nil, nil }, "arg"); err == nil {
        t.Fatalf("expected recover error")
    } else if !strings.Contains(err.Error(), "recover error") {
        t.Fatalf("unexpected error: %v", err)
    }
}

func TestUnitDiscoveryAccessWrappers(t *testing.T) {
    srv := unitServer(unitDiscoveryHandler(false))
    defer srv.Close()

    c := unitNewClient(t, srv.URL)
    d := c.DiscoveryClient()

    if app, err := d.AccessApp("UNITAPP"); err != nil || app == nil {
        t.Fatalf("AccessApp failed: %v", err)
    }
    if inst, err := d.AccessAppInstance("UNITAPP"); err != nil || inst == nil {
        t.Fatalf("AccessAppInstance failed: %v", err)
    }
    if apps, err := d.AccessAppsByVip("vip1"); err != nil || len(apps) != 1 {
        t.Fatalf("AccessAppsByVip failed: %v", err)
    }
    if inst, err := d.AccessAppInstanceByVip("vip1"); err != nil || inst == nil {
        t.Fatalf("AccessAppInstanceByVip failed: %v", err)
    }
    if apps, err := d.AccessAppsBySvip("svip1"); err != nil || len(apps) != 1 {
        t.Fatalf("AccessAppsBySvip failed: %v", err)
    }
    if inst, err := d.AccessAppInstanceBySvip("svip1"); err != nil || inst == nil {
        t.Fatalf("AccessAppInstanceBySvip failed: %v", err)
    }
    if insts, err := d.AccessInstancesByVip("vip1"); err != nil || len(insts) != 1 {
        t.Fatalf("AccessInstancesByVip failed: %v", err)
    }
    if inst, err := d.AccessInstanceByVip("vip1"); err != nil || inst == nil {
        t.Fatalf("AccessInstanceByVip failed: %v", err)
    }
    if insts, err := d.AccessInstancesBySvip("svip1"); err != nil || len(insts) != 1 {
        t.Fatalf("AccessInstancesBySvip failed: %v", err)
    }
    if inst, err := d.AccessInstanceBySvip("svip1"); err != nil || inst == nil {
        t.Fatalf("AccessInstanceBySvip failed: %v", err)
    }
}

// unitAppsForFilter builds commonly used Apps maps for the Filter* tests.
func unitAppsForFilter() (up, down, other map[string][]*meta.AppInfo) {
    up = map[string][]*meta.AppInfo{
        "defaultZone": {{Name: "UNITAPP", Instances: []*meta.InstanceInfo{unitUpInstance()}}},
    }
    down = map[string][]*meta.AppInfo{
        "defaultZone": {{Name: "UNITAPP", Instances: []*meta.InstanceInfo{{
            AppName: "UNITAPP", InstanceId: "d", Status: meta.StatusDown,
            VipAddress: "vip1", SecureVipAddress: "svip1",
        }}}},
    }
    other = map[string][]*meta.AppInfo{
        "otherZone": {{Name: "UNITAPP", Instances: []*meta.InstanceInfo{unitUpInstance()}}},
    }
    return
}

func TestUnitDiscoveryFilterApp(t *testing.T) {
    c := unitNewClient(t, "")
    d := c.DiscoveryClient()
    tr, fa := true, false
    up, down, other := unitAppsForFilter()

    // same-zone match with available instances.
    d.Config.PreferSameZoneEureka = &tr
    if app, err := d.FilterApp(up, "UNITAPP"); err != nil || app == nil || len(app.Instances) != 1 {
        t.Fatalf("same-zone FilterApp failed: %v %#v", err, app)
    }

    // same-zone app present but no available instances -> falls through -> error.
    if _, err := d.FilterApp(down, "UNITAPP"); err == nil {
        t.Fatalf("expected no available service error")
    }

    // same-zone key missing -> random map path.
    if app, err := d.FilterApp(other, "UNITAPP"); err != nil || app == nil {
        t.Fatalf("random-map FilterApp failed: %v %#v", err, app)
    }

    // PreferSameZone false -> random map path.
    d.Config.PreferSameZoneEureka = &fa
    if app, err := d.FilterApp(up, "UNITAPP"); err != nil || app == nil {
        t.Fatalf("non-prefer FilterApp failed: %v %#v", err, app)
    }

    // empty/nil apps -> error.
    if _, err := d.FilterApp(map[string][]*meta.AppInfo{}, "UNITAPP"); err == nil {
        t.Fatalf("expected error for empty apps")
    }
    if _, err := d.FilterApp(nil, "UNITAPP"); err == nil {
        t.Fatalf("expected error for nil apps")
    }

    // FilterAppInstance success.
    if inst, err := d.FilterAppInstance(up, "UNITAPP"); err != nil || inst == nil {
        t.Fatalf("FilterAppInstance failed: %v", err)
    }
    // FilterAppInstance error propagation.
    if _, err := d.FilterAppInstance(up, "MISSING"); err == nil {
        t.Fatalf("expected FilterAppInstance error")
    }
}

func TestUnitDiscoveryFilterAppsByVip(t *testing.T) {
    c := unitNewClient(t, "")
    d := c.DiscoveryClient()
    tr, fa := true, false
    up, down, _ := unitAppsForFilter()

    d.Config.PreferSameZoneEureka = &tr
    if apps, err := d.FilterAppsByVip(up, "vip1"); err != nil || len(apps) != 1 {
        t.Fatalf("same-zone FilterAppsByVip failed: %v %#v", err, apps)
    }
    // same-zone match but no available instances -> error.
    if _, err := d.FilterAppsByVip(down, "vip1"); err == nil {
        t.Fatalf("expected error for no available instances")
    }
    // vip not found -> error.
    if _, err := d.FilterAppsByVip(up, "nomatch"); err == nil {
        t.Fatalf("expected error for missing vip")
    }
    // PreferSameZone false -> random path.
    d.Config.PreferSameZoneEureka = &fa
    if apps, err := d.FilterAppsByVip(up, "vip1"); err != nil || len(apps) != 1 {
        t.Fatalf("random FilterAppsByVip failed: %v %#v", err, apps)
    }
    // nil apps -> error.
    if _, err := d.FilterAppsByVip(nil, "vip1"); err == nil {
        t.Fatalf("expected error for nil apps")
    }

    // FilterAppInstanceByVip success + error.
    if inst, err := d.FilterAppInstanceByVip(up, "vip1"); err != nil || inst == nil {
        t.Fatalf("FilterAppInstanceByVip failed: %v", err)
    }
    if _, err := d.FilterAppInstanceByVip(up, "nomatch"); err == nil {
        t.Fatalf("expected FilterAppInstanceByVip error")
    }
}

func TestUnitDiscoveryFilterAppsBySvip(t *testing.T) {
    c := unitNewClient(t, "")
    d := c.DiscoveryClient()
    tr, fa := true, false
    up, down, _ := unitAppsForFilter()

    d.Config.PreferSameZoneEureka = &tr
    if apps, err := d.FilterAppsBySvip(up, "svip1"); err != nil || len(apps) != 1 {
        t.Fatalf("same-zone FilterAppsBySvip failed: %v %#v", err, apps)
    }
    if _, err := d.FilterAppsBySvip(down, "svip1"); err == nil {
        t.Fatalf("expected error for no available instances")
    }
    if _, err := d.FilterAppsBySvip(up, "nomatch"); err == nil {
        t.Fatalf("expected error for missing svip")
    }
    d.Config.PreferSameZoneEureka = &fa
    if apps, err := d.FilterAppsBySvip(up, "svip1"); err != nil || len(apps) != 1 {
        t.Fatalf("random FilterAppsBySvip failed: %v %#v", err, apps)
    }
    if _, err := d.FilterAppsBySvip(nil, "svip1"); err == nil {
        t.Fatalf("expected error for nil apps")
    }

    if inst, err := d.FilterAppInstanceBySvip(up, "svip1"); err != nil || inst == nil {
        t.Fatalf("FilterAppInstanceBySvip failed: %v", err)
    }
    if _, err := d.FilterAppInstanceBySvip(up, "nomatch"); err == nil {
        t.Fatalf("expected FilterAppInstanceBySvip error")
    }
}

func TestUnitDiscoveryFilterInstancesByVip(t *testing.T) {
    c := unitNewClient(t, "")
    d := c.DiscoveryClient()
    tr, fa := true, false
    up, down, _ := unitAppsForFilter()

    d.Config.PreferSameZoneEureka = &tr
    if insts, err := d.FilterInstancesByVip(up, "vip1"); err != nil || len(insts) != 1 {
        t.Fatalf("same-zone FilterInstancesByVip failed: %v %#v", err, insts)
    }
    if _, err := d.FilterInstancesByVip(down, "vip1"); err == nil {
        t.Fatalf("expected error for no available instances")
    }
    if _, err := d.FilterInstancesByVip(up, "nomatch"); err == nil {
        t.Fatalf("expected error for missing vip")
    }
    d.Config.PreferSameZoneEureka = &fa
    if insts, err := d.FilterInstancesByVip(up, "vip1"); err != nil || len(insts) != 1 {
        t.Fatalf("random FilterInstancesByVip failed: %v %#v", err, insts)
    }
    if _, err := d.FilterInstancesByVip(nil, "vip1"); err == nil {
        t.Fatalf("expected error for nil apps")
    }

    if inst, err := d.FilterInstanceByVip(up, "vip1"); err != nil || inst == nil {
        t.Fatalf("FilterInstanceByVip failed: %v", err)
    }
    if _, err := d.FilterInstanceByVip(up, "nomatch"); err == nil {
        t.Fatalf("expected FilterInstanceByVip error")
    }
}

func TestUnitDiscoveryFilterInstancesBySvip(t *testing.T) {
    c := unitNewClient(t, "")
    d := c.DiscoveryClient()
    tr, fa := true, false
    up, down, _ := unitAppsForFilter()

    d.Config.PreferSameZoneEureka = &tr
    if insts, err := d.FilterInstancesBySvip(up, "svip1"); err != nil || len(insts) != 1 {
        t.Fatalf("same-zone FilterInstancesBySvip failed: %v %#v", err, insts)
    }
    if _, err := d.FilterInstancesBySvip(down, "svip1"); err == nil {
        t.Fatalf("expected error for no available instances")
    }
    if _, err := d.FilterInstancesBySvip(up, "nomatch"); err == nil {
        t.Fatalf("expected error for missing svip")
    }
    d.Config.PreferSameZoneEureka = &fa
    if insts, err := d.FilterInstancesBySvip(up, "svip1"); err != nil || len(insts) != 1 {
        t.Fatalf("random FilterInstancesBySvip failed: %v %#v", err, insts)
    }
    if _, err := d.FilterInstancesBySvip(nil, "svip1"); err == nil {
        t.Fatalf("expected error for nil apps")
    }

    if inst, err := d.FilterInstanceBySvip(up, "svip1"); err != nil || inst == nil {
        t.Fatalf("FilterInstanceBySvip failed: %v", err)
    }
    if _, err := d.FilterInstanceBySvip(up, "nomatch"); err == nil {
        t.Fatalf("expected FilterInstanceBySvip error")
    }
}

func TestUnitDiscoveryUncovered(t *testing.T) {
    // GetLogger nil branch.
    if (&DiscoveryClient{}).GetLogger() == nil {
        t.Fatalf("expected default logger")
    }

    // Discovery0 with nil Config -> GetAllZoneEurekaServers error branch.
    if _, err := (&DiscoveryClient{}).Discovery0(); err == nil {
        t.Fatalf("expected Discovery0 error for nil config")
    }

    // isEnabled error branches for all Filter* methods.
    c := unitNewClient(t, "")
    d := c.DiscoveryClient()
    fa := false
    d.Config.DiscoveryEnabled = &fa
    up, _, _ := unitAppsForFilter()
    if _, err := d.FilterApp(up, "UNITAPP"); err == nil {
        t.Fatalf("expected FilterApp isEnabled error")
    }
    if _, err := d.FilterAppsByVip(up, "vip1"); err == nil {
        t.Fatalf("expected FilterAppsByVip isEnabled error")
    }
    if _, err := d.FilterAppsBySvip(up, "svip1"); err == nil {
        t.Fatalf("expected FilterAppsBySvip isEnabled error")
    }
    if _, err := d.FilterInstancesByVip(up, "vip1"); err == nil {
        t.Fatalf("expected FilterInstancesByVip isEnabled error")
    }
    if _, err := d.FilterInstancesBySvip(up, "svip1"); err == nil {
        t.Fatalf("expected FilterInstancesBySvip isEnabled error")
    }

    // Access* wrappers error branches (isEnabled disabled).
    if _, err := d.AccessApp("UNITAPP"); err == nil {
        t.Fatalf("expected AccessApp error")
    }
    if _, err := d.AccessAppInstance("UNITAPP"); err == nil {
        t.Fatalf("expected AccessAppInstance error")
    }
    if _, err := d.AccessAppsByVip("vip1"); err == nil {
        t.Fatalf("expected AccessAppsByVip error")
    }
    if _, err := d.AccessAppInstanceByVip("vip1"); err == nil {
        t.Fatalf("expected AccessAppInstanceByVip error")
    }
    if _, err := d.AccessAppsBySvip("svip1"); err == nil {
        t.Fatalf("expected AccessAppsBySvip error")
    }
    if _, err := d.AccessAppInstanceBySvip("svip1"); err == nil {
        t.Fatalf("expected AccessAppInstanceBySvip error")
    }
    if _, err := d.AccessInstancesByVip("vip1"); err == nil {
        t.Fatalf("expected AccessInstancesByVip error")
    }
    if _, err := d.AccessInstanceByVip("vip1"); err == nil {
        t.Fatalf("expected AccessInstanceByVip error")
    }
    if _, err := d.AccessInstancesBySvip("svip1"); err == nil {
        t.Fatalf("expected AccessInstancesBySvip error")
    }
    if _, err := d.AccessInstanceBySvip("svip1"); err == nil {
        t.Fatalf("expected AccessInstanceBySvip error")
    }
}
