package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// tracer para spans manuais (ex.: chamada ao SQS). Usa o TracerProvider global
// configurado em initTracer(); se o tracing estiver desativado, vira no-op.
var tracer = otel.Tracer("donation-service")

type Donation struct {
	ID        int       `json:"id"`
	NgoID     int       `json:"ngo_id"`
	Amount    float64   `json:"amount"`
	DonorName string    `json:"donor_name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type App struct {
	DB          *sql.DB
	SqsSvc      *sqs.SQS
	SqsQueueURL string
}

func main() {
	_ = godotenv.Load()

	shutdown := initTracer()
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			log.Printf("Erro ao encerrar OTel tracer: %v", err)
		}
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL é obrigatória")
	}

	db, err := otelsql.Open("pgx", dbURL, otelsql.WithAttributes(semconv.DBSystemPostgreSQL))
	if err != nil || db.Ping() != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}
	log.Println("Conectado ao PostgreSQL (donation-service).")

	var sqsSvc *sqs.SQS
	queueURL := os.Getenv("AWS_SQS_URL")
	region := os.Getenv("AWS_REGION")
	if queueURL != "" && region != "" {
		sess, _ := session.NewSession(&aws.Config{Region: aws.String(region)})
		sqsSvc = sqs.New(sess)
		log.Println("Integração com AWS SQS ativada.")
	}

	app := &App{DB: db, SqsSvc: sqsSvc, SqsQueueURL: queueURL}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", app.HealthHandler)
	mux.HandleFunc("/donations", app.DonationHandler)
	mux.Handle("/metrics", metricsHandler())

	handler := prometheusMiddleware(otelhttp.NewHandler(mux, "donation-service"))
	log.Printf("donation-service rodando na porta %s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

func (a *App) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(`{"status":"ok","service":"donation-service"}`)); err != nil {
		log.Printf("Erro ao escrever resposta health: %v", err)
	}
}

func (a *App) DonationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodPost {
		var d Donation
		if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
			http.Error(w, `{"error":"Payload inválido"}`, http.StatusBadRequest)
			return
		}

		d.Status = "APPROVED" // Simulação de gateway de pagamento
		err := a.DB.QueryRowContext(r.Context(),
			"INSERT INTO donations (ngo_id, amount, donor_name, status) VALUES ($1, $2, $3, $4) RETURNING id, created_at",
			d.NgoID, d.Amount, d.DonorName, d.Status,
		).Scan(&d.ID, &d.CreatedAt)

		if err != nil {
			log.Printf("Erro ao salvar doação: %v", err)
			http.Error(w, `{"error":"Erro interno"}`, http.StatusInternalServerError)
			return
		}

		if a.SqsSvc != nil {
			// WithoutCancel preserva o span do request (para o trace) mas evita que o
			// contexto seja cancelado quando o handler retorna antes da goroutine terminar.
			go a.sendNotificationEvent(context.WithoutCancel(r.Context()), d)
		}

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(d); err != nil {
			log.Printf("Erro ao codificar resposta: %v", err)
		}
		return
	}

	if r.Method == http.MethodGet {
		rows, err := a.DB.QueryContext(r.Context(), "SELECT id, ngo_id, amount, donor_name, status, created_at FROM donations ORDER BY id DESC")
		if err != nil {
			http.Error(w, `{"error":"Erro interno"}`, http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		donations := []Donation{}
		for rows.Next() {
			var d Donation
			if err := rows.Scan(&d.ID, &d.NgoID, &d.Amount, &d.DonorName, &d.Status, &d.CreatedAt); err != nil {
				log.Printf("Erro ao ler linha: %v", err)
				continue
			}
			donations = append(donations, d)
		}

		if err := json.NewEncoder(w).Encode(donations); err != nil {
			log.Printf("Erro ao codificar resposta: %v", err)
		}
		return
	}

	http.Error(w, `{"error":"Método não permitido"}`, http.StatusMethodNotAllowed)
}

func (a *App) sendNotificationEvent(ctx context.Context, d Donation) {
	ctx, span := tracer.Start(ctx, "SQS SendMessage",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			attribute.String("messaging.system", "aws_sqs"),
			attribute.String("messaging.destination.name", a.SqsQueueURL),
		),
	)
	defer span.End()

	body, _ := json.Marshal(d)
	_, err := a.SqsSvc.SendMessageWithContext(ctx, &sqs.SendMessageInput{
		MessageBody: aws.String(string(body)),
		QueueUrl:    aws.String(a.SqsQueueURL),
	})
	if err != nil {
		span.RecordError(err)
		log.Printf("Falha ao despachar evento SQS: %v", err)
	}
}