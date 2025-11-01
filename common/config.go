package common

import (
	_ "embed"
	"os"
	"time"

	"github.com/goccy/go-yaml"
)

var LocalConfig = &Config{
	System: System{
		User:                  "",
		Password:              "",
		SignKey:               "",
		Addr:                  "0.0.0.0:80",
		URLPrefix:             "/",
		DataDir:               "./data",
		DSN:                   "",
		Cert:                  "",
		Key:                   "",
		ReduceMemoryUsage:     false,
		ProxyHeader:           "",
		MaxBatchPushCount:     -1,
		MaxAPNSClientCount:    1,
		MaxDeviceKeyArrLength: 10,
		Concurrency:           256 * 1024,
		ReadTimeout:           3 * time.Second,
		WriteTimeout:          3 * time.Second,
		IdleTimeout:           10 * time.Second,
		ProxyDownload:         false,
		Debug:                 false,
		Version:               "",
		BuildDate:             "",
		CommitID:              "",
		ICPInfo:               "",
		Voice:                 false,
		Auths:                 []string{},
	},
	Apple: Apple{
		ApnsPrivateKey: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBHkwdwIBAQQgvjopbchDpzJNojnc
o7ErdZQFZM7Qxho6m61gqZuGVRigCgYIKoZIzj0DAQehRANCAAQ8ReU0fBNg+sA+
ZdDf3w+8FRQxFBKSD/Opt7n3tmtnmnl9Vrtw/nUXX4ldasxA2gErXR4YbEL9Z+uJ
REJP/5bp
-----END PRIVATE KEY-----`,
		Topic:   "me.uuneo.Meoworld",
		KeyID:   "BNY5GUGV38",
		TeamID:  "FUWV6U942Q",
		Develop: false,
	},
}

func SetDefaultVersionOrCommID(version, buildDate, commID string) {
	if len(version) > 0 {
		LocalConfig.System.Version = version
	} else {
		LocalConfig.System.Version = "1.0.0"
	}
	if len(commID) > 0 {
		LocalConfig.System.CommitID = commID
	} else {
		LocalConfig.System.CommitID = "f7efb70"
	}
	if len(buildDate) > 0 {
		LocalConfig.System.BuildDate = buildDate
	} else {
		LocalConfig.System.BuildDate = "2025-01-01 09:20:33"
	}
}

// SynchronousFieldFile Prevent problems with the fields
func SynchronousFieldFile() {
	data, err := yaml.Marshal(LocalConfig)
	if err != nil {
		panic(err)
	}

	header := `# ============================================
# Meoworld Server Configuration
# Generated automatically. Do not edit manually.
# Modify values carefully, then restart the service.
# ============================================

`
	// 拼接注释头 + YAML 内容
	finalData := append([]byte(header), data...)

	// 输出到文件
	if err = os.WriteFile("config.yaml", finalData, 0644); err != nil {
		panic(err)
	}
}
