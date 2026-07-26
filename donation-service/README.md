# Donation Service — Processamento de Doações (Hot Path) ★

Microsserviço **crítico** da plataforma SolidaryTech: processa as doações em tempo real e publica eventos assíncronos para notificação. É o **Hot Path** — tem SLO formal, instrumentação de rastreamento distribuído de ponta a ponta e é o foco da estratégia de DR.

| Item | Valor |
|---|---|
| Linguagem | Go |
| Banco de dados | PostgreSQL (RDS `donation-db`) — driver `pgx` instrumentado com `otelsql` |
| Mensageria | AWS SQS (evento de notificação por doação) |
| Porta | `8082` |
| Rota no ingress | `/donation-service` |
| Observabilidade | Métricas Prometheus (`/metrics`) + Distributed Tracing (OpenTelemetry) |

## Funcionalidades

- Registro de doações (persistência transacional no PostgreSQL, status `APPROVED` simulando gateway)
- Publicação assíncrona de evento de notificação no **SQS** a cada doação (sem bloquear a resposta)
- Listagem de doações
- **Métricas Prometheus** expostas em `/metrics` (base do dashboard SRE e dos alertas de SLO)
- **Distributed Tracing** completo: span HTTP raiz (`otelhttp`) → span do banco (`otelsql`) → span produtor do SQS
- Health check para probes do Kubernetes

## SLOs (Golden Metrics)

| SLI | SLO |
|---|---|
| Latência: 99% das requisições < 300ms | 99,9% |
| Taxa de erro: < 0,1% de respostas 5xx | 99,9% |

Métricas expostas: `donation_http_requests_total` e `donation_http_request_duration_seconds`.

## Endpoints

| Método | Rota | Descrição |
|---|---|---|
| `GET` | `/health` | Liveness/readiness — `{"status":"ok","service":"donation-service"}` |
| `POST` | `/donations` | Registra uma doação (persiste + publica evento SQS) |
| `GET` | `/donations` | Lista todas as doações (ordem decrescente por id) |
| `GET` | `/metrics` | Métricas no formato Prometheus |

### Exemplo — nova doação

```bash
curl -X POST http://<host>/donation-service/donations \
  -H 'Content-Type: application/json' \
  -d '{"ngo_id":1,"amount":50.00,"donor_name":"João"}'
# 201 Created → doação com id, status "APPROVED" e created_at
# 400 Bad Request → payload inválido
```

## Variáveis de ambiente

| Variável | Obrigatória | Descrição |
|---|---|---|
| `DATABASE_URL` | ✅ | DSN do PostgreSQL |
| `PORT` | ❌ | Porta HTTP (default `8082`) |
| `AWS_SQS_URL` | ❌ | URL da fila SQS (habilita publicação de eventos) |
| `AWS_REGION` | ❌ | Região AWS para o cliente SQS |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | ❌ | Endpoint gRPC do OTel Collector (habilita tracing) |

## Modelo de dados

Tabela `donations`: `id`, `ngo_id`, `amount` (NUMERIC), `donor_name`, `status`, `created_at`.

## Execução local

```bash
go mod download
export DATABASE_URL="postgres://user:pass@localhost:5432/donationdb"
go run .                 # sobe em http://localhost:8082
```

### Docker

```bash
docker build -t donation-service .
docker run -p 8082:8082 -e DATABASE_URL="..." donation-service
```

## Deploy

Imagem publicada no **ECR** pelo pipeline de CI (GitHub Actions + scan **Trivy**). Deploy no **EKS** via **GitOps (ArgoCD)** a partir do `hackathon-gitops`. Por ser o Hot Path, roda com `minReplicas: 2` (HPA) e é o serviço replicado no ambiente de **Disaster Recovery** (read replica cross-region).
