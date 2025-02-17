# Benchmarker

```
.
├── cmd
│   ├── bench
│   │   ├── main.go
│   │   ├── root.go
│   │   ├── run.go
│   │   └── supervise.go
│   └── lightbench
│       └── main.go
├── deploy
├── internal
│   ├── aws
│   ├── logger
│   └── signal
├── scenario
│   ├── action
│   ├── api
│   ├── fixture
│   ├── model
│   ├── validate
│   ├── scenario.go
│   ├── scenario_xxx.go
│   ├──   :
│   └── scenario_zzz.go
├── README.md
├── logging_policy.md
├── Dockerfile
├── go.mod
├── go.sum
├── bench.go
└── option.go
```


## Tasks
[![xc compatible](https://xcfile.dev/badge.svg)](https://xcfile.dev)
### Bench
Run benchmark

Inputs: TARGET_HOST
env: TARGET_HOST="localhost:8080"
```
go run ./cmd/bench/... run $TARGET_HOST
```

### LightBench
Run light-weight benchmark

Inputs: TARGET_HOST
env: TARGET_HOST="localhost:8080"
```
go run ./cmd/lightbench/... $TARGET_HOST
```
