package meta

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var unitAppInfoJson = `
{
    "name": "UNITAPP",
    "instance": [
        {
            "instanceId": "unit-1",
            "hostName": "unit-host",
            "app": "UNITAPP",
            "ipAddr": "10.0.0.1",
            "status": "UP",
            "overriddenStatus": "UNKNOWN",
            "port": { "$": 18080, "@enabled": "true" },
            "securePort": { "$": 443, "@enabled": "false" },
            "actionType": "ADDED"
        }
    ]
}
`

func TestUnitMeta_AppInfo_Copy(t *testing.T) {
	ast := assert.New(t)

	var nilApp *AppInfo
	ast.Nil(nilApp.Copy())

	// Instances nil
	app := &AppInfo{Name: "a", Region: "r", Zone: "z"}
	cp := app.Copy()
	ast.NotNil(cp)
	ast.Equal("a", cp.Name)
	ast.Equal("r", cp.Region)
	ast.Equal("z", cp.Zone)
	ast.Nil(cp.Instances)

	// Instances set (including a nil entry)
	app.Instances = []*InstanceInfo{
		{InstanceId: "id", HostName: "h", Status: StatusUp},
		nil,
	}
	cp2 := app.Copy()
	ast.NotNil(cp2.Instances)
	ast.Len(cp2.Instances, 2)
	ast.Equal("id", cp2.Instances[0].InstanceId)
	ast.Nil(cp2.Instances[1])
	cp2.Instances[0].InstanceId = "changed"
	ast.Equal("id", app.Instances[0].InstanceId, "instances copy must be independent")
}

func TestUnitMeta_AppInfo_CopyWithInstances(t *testing.T) {
	ast := assert.New(t)

	var nilApp *AppInfo
	ast.Nil(nilApp.CopyWithInstances(nil))

	app := &AppInfo{Name: "a", Region: "r", Zone: "z"}
	instances := []*InstanceInfo{{InstanceId: "id", Status: StatusUp}}
	cp := app.CopyWithInstances(instances)
	ast.NotNil(cp)
	ast.Equal("a", cp.Name)
	ast.Equal("r", cp.Region)
	ast.Equal("z", cp.Zone)
	ast.Equal(instances, cp.Instances)
}

func TestUnitMeta_AppInfo_AvailableInstances(t *testing.T) {
	ast := assert.New(t)

	// nil app
	var nilApp *AppInfo
	got := nilApp.AvailableInstances()
	ast.NotNil(got)
	ast.Len(got, 0)

	// Instances nil
	app := &AppInfo{Name: "a"}
	got = app.AvailableInstances()
	ast.NotNil(got)
	ast.Len(got, 0)

	// mixed statuses incl. nil entry
	app.Instances = []*InstanceInfo{
		nil,
		{InstanceId: "down", Status: StatusDown},
		{InstanceId: "up", HostName: "h", Status: StatusUp},
		{InstanceId: "starting", Status: StatusStarting},
	}
	got = app.AvailableInstances()
	ast.Len(got, 1)
	ast.Equal("up", got[0].InstanceId)
	got[0].InstanceId = "changed"
	ast.Equal("up", app.Instances[2].InstanceId, "available instances must be copies")
}

func TestUnitMeta_ParseAppInfo(t *testing.T) {
	unitFakeHostInfo()
	ast := assert.New(t)

	app, err := ParseAppInfo([]byte(unitAppInfoJson))
	ast.Nil(err)
	ast.NotNil(app)
	ast.Equal("UNITAPP", app.Name)
	ast.Len(app.Instances, 1)
	ast.Equal("unit-1", app.Instances[0].InstanceId)

	// missing "instance" -> nil slice replaced with empty slice
	noInstance, err := ParseAppInfo([]byte(`{"name":"EMPTY"}`))
	ast.Nil(err)
	ast.NotNil(noInstance)
	ast.NotNil(noInstance.Instances)
	ast.Len(noInstance.Instances, 0)

	bad, err := ParseAppInfo([]byte(`{not-json`))
	ast.NotNil(err)
	ast.Nil(bad)
}
