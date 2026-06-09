import os
import sys
from unittest.mock import MagicMock, patch

os.environ.setdefault("DATABASE_URL", "postgresql://test:test@localhost/testdb")

with patch("psycopg2.pool.SimpleConnectionPool") as _mock_pool_cls:
    _mock_pool = MagicMock()
    _mock_pool_cls.return_value = _mock_pool
    import app as ngo_app

import pytest
import json


@pytest.fixture
def client():
    ngo_app.app.config["TESTING"] = True
    with ngo_app.app.test_client() as c:
        yield c


@pytest.fixture(autouse=True)
def reset_pool():
    ngo_app.pool = MagicMock()
    yield


def test_health_returns_200(client):
    response = client.get("/health")
    assert response.status_code == 200


def test_health_body(client):
    response = client.get("/health")
    data = response.get_json()
    assert data["status"] == "ok"
    assert data["service"] == "ngo-service"


def test_create_ngo_missing_fields_returns_400(client):
    response = client.post("/ngos", json={"name": "ONG Test"})
    assert response.status_code == 400
    data = response.get_json()
    assert "error" in data


def test_create_ngo_empty_body_returns_400(client):
    response = client.post("/ngos", json={})
    assert response.status_code == 400


def test_create_ngo_success(client):
    mock_conn = MagicMock()
    cursor_ctx = MagicMock()
    cursor_ctx.fetchone.return_value = {
        "id": 1,
        "name": "ONG Teste",
        "email": "ong@teste.com",
        "cause": "Educação",
        "city": "São Paulo",
    }
    mock_conn.cursor.return_value.__enter__ = MagicMock(return_value=cursor_ctx)
    mock_conn.cursor.return_value.__exit__ = MagicMock(return_value=False)
    ngo_app.pool.getconn.return_value = mock_conn

    response = client.post(
        "/ngos",
        json={
            "name": "ONG Teste",
            "email": "ong@teste.com",
            "cause": "Educação",
            "city": "São Paulo",
        },
    )
    assert response.status_code == 201


def test_get_ngos_returns_list(client):
    mock_conn = MagicMock()
    cursor_ctx = MagicMock()
    cursor_ctx.fetchall.return_value = []
    mock_conn.cursor.return_value.__enter__ = MagicMock(return_value=cursor_ctx)
    mock_conn.cursor.return_value.__exit__ = MagicMock(return_value=False)
    ngo_app.pool.getconn.return_value = mock_conn

    response = client.get("/ngos")
    assert response.status_code == 200
    data = response.get_json()
    assert isinstance(data, list)
