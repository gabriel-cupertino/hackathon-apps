# Volunteer Service — Match entre Voluntários e Campanhas

Microsserviço responsável pelo **cadastro de voluntários e o match com as ONGs/campanhas** da plataforma SolidaryTech. Usa um banco NoSQL (DynamoDB) por ser um workload de leitura/escrita simples e alto volume, sem necessidade de relacionamentos complexos.

| Item | Valor |
|---|---|
| Linguagem | Python 3 |
| Framework | Flask (servido via gunicorn) |
| Banco de dados | AWS DynamoDB (`SolidaryTechVolunteers`) — SDK `boto3` |
| Porta | `8083` |
| Rota no ingress | `/volunteer-service` |
| Observabilidade | OpenTelemetry (Distributed Tracing via OTLP) |

## Funcionalidades

- Cadastro de voluntários (nome, e-mail, ONG de interesse) com `volunteer_id` gerado (UUID)
- Consulta de voluntários por ONG (`ngo_id`)
- Persistência em DynamoDB (acesso via IAM Role do nó — sem credenciais estáticas)
- Health check para probes do Kubernetes

## Endpoints

| Método | Rota | Descrição |
|---|---|---|
| `GET` | `/health` | Liveness/readiness — `{"status":"ok","service":"volunteer-service"}` |
| `POST` | `/volunteers` | Cadastra um voluntário |
| `GET` | `/volunteers/<ngo_id>` | Lista voluntários de uma ONG (Scan com filtro por `ngo_id`) |

### Exemplo — cadastro

```bash
curl -X POST http://<host>/volunteer-service/volunteers \
  -H 'Content-Type: application/json' \
  -d '{"name":"Maria","email":"maria@email.com","ngo_id":1}'
# 201 Created → voluntário com volunteer_id (UUID) e registered_at
# 400 Bad Request → campos obrigatórios ausentes (name, email, ngo_id)
```

```bash
curl http://<host>/volunteer-service/volunteers/1     # voluntários da ONG 1
```

> **Nota de arquitetura:** a consulta usa `Scan` com filtro por `ngo_id` — simplificação didática. Em produção de alto volume, o correto seria um **GSI (Global Secondary Index)** em `ngo_id` para evitar full scan.

## Variáveis de ambiente

| Variável | Obrigatória | Descrição |
|---|---|---|
| `AWS_DYNAMODB_TABLE` | ✅ | Nome da tabela DynamoDB |
| `AWS_REGION` | ❌ | Região AWS (default `us-east-1`) |
| `PORT` | ❌ | Porta HTTP (default `8083`) |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | ❌ | Endpoint OTLP do OTel Collector (habilita tracing) |

## Modelo de dados

Tabela DynamoDB `SolidaryTechVolunteers`: `volunteer_id` (PK, UUID), `name`, `email`, `ngo_id`, `registered_at`.

## Execução local

```bash
pip install -r requirements.txt
export AWS_DYNAMODB_TABLE="SolidaryTechVolunteers"
export AWS_REGION="us-east-1"
python app.py            # sobe em http://localhost:8083
```

### Docker

```bash
docker build -t volunteer-service .
docker run -p 8083:8083 -e AWS_DYNAMODB_TABLE="..." volunteer-service
```

## Deploy

Imagem publicada no **ECR** pelo pipeline de CI (GitHub Actions + scan **Trivy**). Deploy no **EKS** via **GitOps (ArgoCD)** a partir do `hackathon-gitops`. O acesso ao DynamoDB é feito via **IAM Role do nó (LabRole)** através do IMDS — sem chaves estáticas no pod.
