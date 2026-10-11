package client

import (
    "errors"
    "testing"

    "github.com/jiashunx/eureka-client-go/meta"
)

func TestUnitUtilsFilterApp(t *testing.T) {
    if got := FilterApp(nil, "x"); got != nil {
        t.Fatalf("nil apps expected nil, got %v", got)
    }
    if got := FilterApp([]*meta.AppInfo{{Name: "A"}}, ""); got != nil {
        t.Fatalf("empty appName expected nil, got %v", got)
    }
    apps := []*meta.AppInfo{{Name: "Eureka-App"}}
    if got := FilterApp(apps, "eureka-app"); got == nil || got.Name != "Eureka-App" {
        t.Fatalf("case-insensitive match failed, got %v", got)
    }
    if got := FilterApp(apps, "missing"); got != nil {
        t.Fatalf("no match expected nil, got %v", got)
    }
}

func TestUnitUtilsFilterAppsByVip(t *testing.T) {
    if got := FilterAppsByVip(nil, "vip1"); len(got) != 0 {
        t.Fatalf("nil apps expected empty, got %d", len(got))
    }
    if got := FilterAppsByVip([]*meta.AppInfo{{Name: "A"}}, ""); len(got) != 0 {
        t.Fatalf("empty vip expected empty, got %d", len(got))
    }
    apps := []*meta.AppInfo{
        {Name: "nil-instances", Instances: nil},
        {Name: "match", Instances: []*meta.InstanceInfo{{VipAddress: "vip1"}, {VipAddress: "other"}}},
        {Name: "no-match", Instances: []*meta.InstanceInfo{{VipAddress: "other"}}},
    }
    got := FilterAppsByVip(apps, "vip1")
    if len(got) != 1 {
        t.Fatalf("expected 1 vip app, got %d", len(got))
    }
    if got[0].Name != "match" || len(got[0].Instances) != 1 {
        t.Fatalf("unexpected vip app: %#v", got[0])
    }
}

func TestUnitUtilsFilterAppsBySvip(t *testing.T) {
    if got := FilterAppsBySvip(nil, "svip1"); len(got) != 0 {
        t.Fatalf("nil apps expected empty, got %d", len(got))
    }
    if got := FilterAppsBySvip([]*meta.AppInfo{{Name: "A"}}, ""); len(got) != 0 {
        t.Fatalf("empty svip expected empty, got %d", len(got))
    }
    apps := []*meta.AppInfo{
        {Name: "nil-instances", Instances: nil},
        {Name: "match", Instances: []*meta.InstanceInfo{{SecureVipAddress: "svip1"}, {SecureVipAddress: "other"}}},
        {Name: "no-match", Instances: []*meta.InstanceInfo{{SecureVipAddress: "other"}}},
    }
    got := FilterAppsBySvip(apps, "svip1")
    if len(got) != 1 || got[0].Name != "match" || len(got[0].Instances) != 1 {
        t.Fatalf("unexpected svip apps: %#v", got)
    }
}

func TestUnitUtilsFilterInstancesByVip(t *testing.T) {
    if got := FilterInstancesByVip(nil, "vip1"); len(got) != 0 {
        t.Fatalf("nil apps expected empty, got %d", len(got))
    }
    apps := []*meta.AppInfo{
        {Name: "a", Instances: []*meta.InstanceInfo{{VipAddress: "vip1"}}},
        {Name: "b", Instances: []*meta.InstanceInfo{{VipAddress: "vip2"}}},
    }
    got := FilterInstancesByVip(apps, "vip1")
    if len(got) != 1 {
        t.Fatalf("expected 1 instance, got %d", len(got))
    }
}

func TestUnitUtilsFilterInstancesBySvip(t *testing.T) {
    if got := FilterInstancesBySvip(nil, "svip1"); len(got) != 0 {
        t.Fatalf("nil apps expected empty, got %d", len(got))
    }
    apps := []*meta.AppInfo{
        {Name: "a", Instances: []*meta.InstanceInfo{{SecureVipAddress: "svip1"}}},
        {Name: "b", Instances: []*meta.InstanceInfo{{SecureVipAddress: "svip2"}}},
    }
    got := FilterInstancesBySvip(apps, "svip1")
    if len(got) != 1 {
        t.Fatalf("expected 1 instance, got %d", len(got))
    }
}

func TestUnitUtilsRandomLoopMap(t *testing.T) {
    // nil map, nil f -> no-op
    if err := RandomLoopMap(nil, nil); err != nil {
        t.Fatalf("nil map expected nil error, got %v", err)
    }
    // empty map
    if err := RandomLoopMap(map[string]interface{}{}, func(string, interface{}) (bool, error) {
        return true, nil
    }); err != nil {
        t.Fatalf("empty map expected nil error, got %v", err)
    }
    // nil f
    if err := RandomLoopMap(map[string]interface{}{"a": 1}, nil); err != nil {
        t.Fatalf("nil f expected nil error, got %v", err)
    }
    // entry returns false -> loop breaks, still nil error
    if err := RandomLoopMap(map[string]interface{}{"a": 1, "b": 2}, func(k string, v interface{}) (bool, error) {
        return false, nil
    }); err != nil {
        t.Fatalf("false return expected nil error, got %v", err)
    }
    // entry returns error -> propagated
    boom := errors.New("boom")
    if err := RandomLoopMap(map[string]interface{}{"a": 1}, func(k string, v interface{}) (bool, error) {
        return true, boom
    }); err != boom {
        t.Fatalf("expected boom error, got %v", err)
    }
    // panicking f -> recovered into error
    if err := RandomLoopMap(map[string]interface{}{"a": 1}, func(k string, v interface{}) (bool, error) {
        panic("panic in f")
    }); err == nil {
        t.Fatalf("expected recover error, got nil")
    }
}

func TestUnitUtilsSummary(t *testing.T) {
    if got := SummaryAppsMap(nil); len(got) != 0 {
        t.Fatalf("nil appsMap expected empty, got %d", len(got))
    }
    if got := SummaryAppsMap(map[string][]*meta.AppInfo{"z": {{Name: "A"}}}); len(got) != 1 {
        t.Fatalf("expected 1 summary zone, got %d", len(got))
    }
    if got := SummaryApps(nil); len(got) != 0 {
        t.Fatalf("nil apps expected empty, got %d", len(got))
    }
    if got := SummaryApps([]*meta.AppInfo{nil, {Name: "A"}}); len(got) != 2 {
        t.Fatalf("expected 2 summaries, got %d", len(got))
    }
    if got := SummaryApp(nil); len(got) != 0 {
        t.Fatalf("nil app expected empty map, got %d", len(got))
    }
    if got := SummaryApp(&meta.AppInfo{Name: "A"}); got["app"] != "A" {
        t.Fatalf("unexpected summary app: %#v", got)
    }
    if got := SummaryInstances(nil); len(got) != 0 {
        t.Fatalf("nil instances expected empty, got %d", len(got))
    }
    if got := SummaryInstances([]*meta.InstanceInfo{nil, {InstanceId: "i"}}); len(got) != 2 {
        t.Fatalf("expected 2 instance summaries, got %d", len(got))
    }
    if got := SummaryInstance(nil); len(got) != 0 {
        t.Fatalf("nil instance expected empty map, got %d", len(got))
    }
    if got := SummaryInstance(&meta.InstanceInfo{InstanceId: "i", Region: "r", Zone: "z"}); got["instanceId"] != "i" {
        t.Fatalf("unexpected instance summary: %#v", got)
    }
}
