module nats-llm/llm-example

go 1.24.1

require (
	github.com/hofer/nats-llm v0.0.0
	github.com/nats-io/nats.go v1.45.0
	github.com/ollama/ollama v0.11.9
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
)

replace github.com/hofer/nats-llm v0.0.0 => ../
