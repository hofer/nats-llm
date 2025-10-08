module nats-llm/llm-example

go 1.24.2

toolchain go1.24.8

require (
	github.com/hofer/nats-llm v0.0.0
	github.com/nats-io/nats.go v1.46.0
	github.com/ollama/ollama v0.12.3
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
)

replace github.com/hofer/nats-llm v0.0.0 => ../
