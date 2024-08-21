module github.com/grafana/plexporter

go 1.20

require (
	github.com/go-kit/log v0.2.1
	github.com/gorilla/websocket v1.5.0
	github.com/jrudio/go-plex-client v0.0.0-20220428052413-e5b4386beb17
	github.com/prometheus/client_golang v1.14.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)

// Replace for fix: https://github.com/jrudio/go-plex-client/pull/56
replace github.com/jrudio/go-plex-client => github.com/jsclayton/go-plex-client v0.0.0-20230428220949-afd78005d7d3
