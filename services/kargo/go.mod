module github.com/totomz/template-burrito/services/kargo

go 1.19

replace github.com/totomz/template-burrito/common/httpserver => ../../common/httpserver

require github.com/totomz/template-burrito/common/httpserver v0.0.0-00010101000000-000000000000

replace github.com/totomz/template-burrito/common/burrito-common => ../../common/burrito-common

require github.com/totomz/template-burrito/common/burrito-common v0.0.0-00010101000000-000000000000

require (
	github.com/form3tech-oss/jwt-go v3.2.5+incompatible
	go.opencensus.io v0.24.0
)

require (
	contrib.go.opencensus.io/exporter/prometheus v0.4.2 // indirect
	contrib.go.opencensus.io/exporter/zipkin v0.1.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/openzipkin/zipkin-go v0.2.2 // indirect
	github.com/prometheus/client_golang v1.13.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/prometheus/statsd_exporter v0.22.7 // indirect
	golang.org/x/sys v0.0.0-20220708085239-5a0f0661e09d // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
