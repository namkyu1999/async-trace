# Asynchronous tracing

## Reference
- https://github.com/open-telemetry/opentelemetry-specification/issues/740
- https://github.com/honeycombio/buildevents
- https://github.com/jenkinsci/opentelemetry-plugin/tree/main


## Installation
```bash
kubectl create ns observability
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.6.3/cert-manager.yaml
kubectl apply -f https://github.com/jaegertracing/jaeger-operator/releases/download/v1.55.0/jaeger-operator.yaml -n observability
kubectl apply -f ./installation/jaeger.yaml -n observability
kubectl apply -f ./installation/otel-collector.yaml
kubectl apply -f ./installation/application.yaml
```

## Current status
- [x] parent process 가 jaeger에 span을 보내는 것은 확인
- [ ] child process span 안보임
- [ ] parent 와 child 를 env로 연결