package meta

import (
    "encoding/json"
    "fmt"
    "github.com/stretchr/testify/assert"
    "testing"
)

func TestNewEurekaConfig(t *testing.T) {
    ast := assert.New(t)
    ec := &EurekaConfig{
        InstanceConfig: nil,
        ClientConfig:   nil,
    }
    err := ec.Check()
    ast.Nilf(err, "%v", err)

    ecData, err := json.Marshal(ec)
    ast.Nilf(err, "%v", err)

    nec, err := ParseEurekaConfig(ecData)
    ast.Nilf(err, "%v", err)
    ast.NotNilf(nec, "%v", nec)
    fmt.Printf("nec: %#v\n", nec)

    nic, err := ParseInstanceConfig(ecData)
    ast.Nilf(err, "%v", err)
    ast.NotNilf(nec, "%v", nic)
    fmt.Printf("nic: %#v\n", nic)

    ncc, err := ParseClientConfig(ecData)
    ast.Nilf(err, "%v", err)
    ast.NotNilf(nec, "%v", ncc)
    fmt.Printf("ncc: %#v\n", ncc)
}
