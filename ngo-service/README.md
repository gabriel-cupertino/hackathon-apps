# NGO Service — Cadastro e Gestão de ONGs

Microsserviço responsável pelo **cadastro e gestão das ONGs parceiras** da plataforma SolidaryTech. É o ponto de entrada para o onboarding de novas organizações que recebem doações e captam voluntários.

| Item | Valor |
|---|---|
| Linguagem | Python 3 |
| Framework | Flask |
| Banco de dados | PostgreSQL (RDS `ngo-db`) — driver `psycopg2` com connection pool |
| Porta | `8081` |
| Rota no ingress | `/ngo-service` |
| Observabilidade | OpenTelemetry (Distributed Tracing via OTLP) |

## Funcionalidades

- Cadastro de ONGs parceiras (nome, e-mail, causa, cidade)
- Listagem das ONGs cadastradas
- Validação de e-mail único (retorna `409` em duplicidade)
- Pool de conexões PostgreSQL (1–10) para resiliência sob carga
- Health check para probes do Kubernetes

## Endpoints

| Método | Rota | Descrição |
|---|---|---|
| `GET` | `/health` | Liveness/readiness — `{"status":"ok","service":"ngo-service"}` |
| `POST` | `/ngos` | Cadastra uma ONG |
| `GET` | `/ngos` | Lista todas as ONGs (ordem decrescente por id) |

### Exemplo — cadastro

```bash
curl -X POST http://<host>/ngo-service/ngos \
  -H 'Content-Type: application/json' \
  -d '{"name":"Instituto Esperança","email":"contato@esperanca.org","cause":"Educação","city":"São Paulo"}'
# 201 Created → objeto da ONG criada
# 409 Conflict → e-mail já cadastrado
# 400 Bad Request → campos obrigatórios ausentes
```

## Variáveis de ambiente

| Variável | Obrigatória | Descrição |
|---|---|---|
| `DATABASE_URL` | ✅ | DSN do PostgreSQL (ex: `postgresql://user:pass@host:5432/ngodb`) |
| `PORT` | ❌ | Porta HTTP (default `8081`) |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | ❌ | Endpoint OTLP do OTel Collector (habilita tracing) |

## Modelo de dados

Tabela `ngos`: `id`, `name`, `email` (único), `cause`, `city`.

## Execução local

```bash
pip install -r requirements.txt
export DATABASE_URL="postgresql://user:pass@localhost:5432/ngodb"
python app.py            # sobe em http://localhost:8081
```

### Docker

```bash
docker build -t ngo-service .
docker run -p 8081:8081 -e DATABASE_URL="..." ngo-service
```

## Deploy

Imagem publicada no **ECR** pelo pipeline de CI (GitHub Actions com scan de segurança **Trivy**). O deploy no cluster **EKS** é feito via **GitOps (ArgoCD)** a partir do repositório `hackathon-gitops` — nenhum deploy manual.
