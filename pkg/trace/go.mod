module github.com/DataDog/datadog-agent/pkg/trace

go 1.21.9

// NOTE: Prefer using simple `require` directives instead of using `replace` if possible.
// See https://github.com/DataDog/datadog-agent/blob/main/docs/dev/gomodreplace.md
// for more details.

// Internal deps fix version
replace github.com/docker/distribution => github.com/docker/distribution v2.8.1+incompatible

require (
	github.com/DataDog/datadog-agent/pkg/obfuscate v0.53.0-rc.2
	github.com/DataDog/datadog-agent/pkg/proto v0.53.0-rc.2
	github.com/DataDog/datadog-agent/pkg/remoteconfig/state v0.53.0-rc.2
	github.com/DataDog/datadog-agent/pkg/util/cgroups v0.53.0-rc.2
	github.com/DataDog/datadog-agent/pkg/util/log v0.53.0-rc.2
	github.com/DataDog/datadog-agent/pkg/util/pointer v0.53.0-rc.2
	github.com/DataDog/datadog-agent/pkg/util/scrubber v0.53.0-rc.2
	github.com/DataDog/datadog-go/v5 v5.5.0
	github.com/DataDog/opentelemetry-mapping-go/pkg/otlp/attributes v0.13.3
	github.com/DataDog/sketches-go v1.4.2
	github.com/Microsoft/go-winio v0.6.1
	github.com/cihub/seelog v0.0.0-20170130134532-f561c5e57575
	github.com/davecgh/go-spew v1.1.1
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.3
	github.com/google/go-cmp v0.6.0
	github.com/google/gofuzz v1.2.0
	github.com/google/uuid v1.6.0
	github.com/shirou/gopsutil/v3 v3.24.1
	github.com/stretchr/testify v1.9.0
	github.com/tinylib/msgp v1.1.8
	github.com/vmihailenco/msgpack/v4 v4.3.12
	go.opentelemetry.io/collector/component v0.97.0
	go.opentelemetry.io/collector/pdata v1.4.0
	go.opentelemetry.io/collector/semconv v0.97.0
	go.opentelemetry.io/otel v1.24.0
	go.opentelemetry.io/otel/metric v1.24.0
	go.uber.org/atomic v1.11.0
	golang.org/x/sys v0.17.0
	golang.org/x/time v0.3.0
	google.golang.org/grpc v1.62.1
	google.golang.org/protobuf v1.33.0
	k8s.io/apimachinery v0.25.5
)

require (
	github.com/DataDog/go-sqllexer v0.0.9 // indirect
	github.com/DataDog/go-tuf v1.1.0-0.5.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/containerd/cgroups/v3 v3.0.2 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-viper/mapstructure/v2 v2.0.0-alpha.1 // indirect
	github.com/godbus/dbus/v5 v5.0.6 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/karrick/godirwalk v1.17.0 // indirect
	github.com/knadh/koanf/maps v0.1.1 // indirect
	github.com/knadh/koanf/providers/confmap v0.1.0 // indirect
	github.com/knadh/koanf/v2 v2.1.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20220913051719-115f729f3c8c // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/opencontainers/runtime-spec v1.1.0-rc.3 // indirect
	github.com/outcaste-io/ristretto v0.2.1 // indirect
	github.com/philhofer/fwd v1.1.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20220216144756-c35f1ee13d7c // indirect
	github.com/prometheus/client_golang v1.19.0 // indirect
	github.com/prometheus/client_model v0.6.0 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/secure-systems-lab/go-securesystemslib v0.7.0 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/vmihailenco/tagparser v0.1.2 // indirect
	github.com/yusufpapurcu/wmi v1.2.3 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.97.0 // indirect
	go.opentelemetry.io/collector/confmap v0.97.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.46.0 // indirect
	go.opentelemetry.io/otel/sdk v1.24.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.24.0 // indirect
	go.opentelemetry.io/otel/trace v1.24.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/tools v0.16.1 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240123012728-ef4313101c80 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/DataDog/datadog-agent => ../../
	github.com/DataDog/datadog-agent/pkg/obfuscate => ../obfuscate
	github.com/DataDog/datadog-agent/pkg/proto => ../proto
	github.com/DataDog/datadog-agent/pkg/remoteconfig/state => ../remoteconfig/state
	github.com/DataDog/datadog-agent/pkg/util/cgroups => ../util/cgroups
	github.com/DataDog/datadog-agent/pkg/util/log => ../util/log
	github.com/DataDog/datadog-agent/pkg/util/pointer => ../util/pointer
	github.com/DataDog/datadog-agent/pkg/util/scrubber => ../util/scrubber
)
