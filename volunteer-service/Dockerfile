FROM python:3.11-slim AS builder

WORKDIR /app

COPY requirements.txt .

RUN pip install --no-cache-dir --prefix=/runtime -r requirements.txt

FROM python:3.11-slim

RUN apt-get update && apt-get upgrade -y --no-install-recommends && rm -rf /var/lib/apt/lists/* && \
    groupadd -r appgroup && useradd -r -g appgroup appuser

WORKDIR /app

COPY --from=builder /runtime /usr/local

COPY app.py .

USER appuser

EXPOSE 8083

CMD ["gunicorn", "--bind", "0.0.0.0:8083", "--workers", "2", "--timeout", "60", "app:app"]
