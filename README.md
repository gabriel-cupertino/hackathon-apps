# hackathon-apps — Microsserviços SolidaryTech

Monorepo com os **3 microsserviços** da plataforma SolidaryTech (Hackathon Fase 5) e seus pipelines de **CI/CD com DevSecOps**. Cada serviço usa uma stack diferente, simulando um ambiente corporativo distribuído.

## Serviços

| Serviço | Stack | Banco | Porta | Papel |
|---|---|---|---|---|
| [ngo-service](ngo-service/README.md) | Python / Flask | RDS PostgreSQL | 8081 | Cadastro e gestão de ONGs |
| [donation-service](donation-service/README.md) ★ | Go | RDS PostgreSQL + SQS | 8082 | **Hot Path** — processamento de doações |
| [volunteer-service](volunteer-service/README.md) | Python / Flask | DynamoDB | 8083 | Match de voluntários e campanhas |

> Cada serviço tem seu próprio **README** com endpoints, variáveis de ambiente, modelo de dados e observabilidade — veja os links acima.

## Observabilidade

Todos instrumentados com **OpenTelemetry** (Distributed Tracing → OTel Collector → Datadog APM). O `donation-service` (Hot Path) expõe ainda **métricas Prometheus** em `/metrics`, base do dashboard SRE e dos alertas de SLO.

## CI/CD + DevSecOps

Um pipeline por serviço em [`.github/workflows/`](.github/workflows/) (GitHub Actions), com os estágios:

1. **Build & Unit Test**
2. **Lint / Static Analysis** (golangci-lint no Go)
3. **Security Scan (DevSecOps):** **Trivy** (SCA — filesystem/deps) + **SAST** (gosec no Go / bandit no Python)
4. **Docker Build & Push → ECR**
5. **GitOps Update:** atualiza a tag da imagem em [`hackathon-gitops`](https://github.com/gabriel-cupertino/hackathon-gitops)

Isso fecha o ciclo: `push → CI (build/scan) → imagem no ECR → commit no GitOps → ArgoCD sincroniza no EKS`.

## Ecossistema (repositórios relacionados)

| Repo | Função |
|---|---|
| **hackathon-apps** (este) | Código dos serviços + CI/CD |
| [hackathon-iac](https://github.com/gabriel-cupertino/hackathon-iac) | Terraform — EKS, RDS, ECR, ArgoCD, observabilidade, DR |
| [hackathon-gitops](https://github.com/gabriel-cupertino/hackathon-gitops) | Manifestos K8s sincronizados pelo ArgoCD |

## Rodando um serviço localmente

Cada serviço tem instruções no seu README. Resumo:

```bash
cd donation-service && go run .          # Go  → :8082
cd ngo-service && python app.py          # Flask → :8081
cd volunteer-service && python app.py    # Flask → :8083
```
