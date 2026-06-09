import os
import pytest
from unittest.mock import MagicMock, patch

os.environ.setdefault("AWS_DYNAMODB_TABLE", "TestVolunteers")
os.environ.setdefault("AWS_REGION", "us-east-1")

with patch("boto3.resource") as _mock_boto3:
    _mock_table = MagicMock()
    _mock_dynamodb = MagicMock()
    _mock_dynamodb.Table.return_value = _mock_table
    _mock_boto3.return_value = _mock_dynamodb
    import app as volunteer_app


@pytest.fixture
def client():
    volunteer_app.app.config["TESTING"] = True
    with volunteer_app.app.test_client() as c:
        yield c


@pytest.fixture(autouse=True)
def reset_table():
    volunteer_app.table = MagicMock()
    yield


def test_health_returns_200(client):
    response = client.get("/health")
    assert response.status_code == 200


def test_health_body(client):
    response = client.get("/health")
    data = response.get_json()
    assert data["status"] == "ok"
    assert data["service"] == "volunteer-service"


def test_register_volunteer_missing_fields_returns_400(client):
    response = client.post("/volunteers", json={"name": "João"})
    assert response.status_code == 400
    data = response.get_json()
    assert "error" in data


def test_register_volunteer_empty_body_returns_400(client):
    response = client.post("/volunteers", json={})
    assert response.status_code == 400


def test_register_volunteer_success(client):
    volunteer_app.table.put_item.return_value = {}

    response = client.post(
        "/volunteers",
        json={"name": "João Silva", "email": "joao@test.com", "ngo_id": 1},
    )
    assert response.status_code == 201
    data = response.get_json()
    assert data["name"] == "João Silva"
    assert data["email"] == "joao@test.com"
    assert "volunteer_id" in data


def test_get_volunteers_by_ngo(client):
    volunteer_app.table.scan.return_value = {
        "Items": [
            {
                "volunteer_id": "abc-123",
                "name": "João",
                "email": "joao@test.com",
                "ngo_id": 1,
            }
        ]
    }

    response = client.get("/volunteers/1")
    assert response.status_code == 200
    data = response.get_json()
    assert isinstance(data, list)
    assert len(data) == 1


def test_get_volunteers_empty_result(client):
    volunteer_app.table.scan.return_value = {"Items": []}

    response = client.get("/volunteers/99")
    assert response.status_code == 200
    data = response.get_json()
    assert data == []
