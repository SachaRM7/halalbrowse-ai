from app import create_app
from app.models import db
from app.models.api_key import ApiKey
from app.models.halal_business import HalalBusiness
from app.models.organization import Organization
from app.models.user import User


def seeded_app(tmp_path):
    app = create_app({
        "TESTING": True,
        "SQLALCHEMY_DATABASE_URI": f"sqlite:///{tmp_path / 'api.db'}",
        "API_RATE_LIMIT_WINDOW": 60,
        "API_RATE_LIMIT_REQUESTS": 3,
    })
    with app.app_context():
        db.create_all()
        admin = User(telegram_id=777, display_name="Admin", role="super_admin")
        db.session.add(admin)
        db.session.flush()

        org = Organization(
            slug="water-aid",
            name="Water Aid",
            description="Well building",
            category="water",
            website="https://example.org",
            contact_email="hello@example.org",
            status="verified",
            created_by=admin.id,
        )
        db.session.add(org)

        key = ApiKey(name="admin-key")
        key.set_key("admin-secret")
        db.session.add(key)

        business = HalalBusiness(
            name="Burger Medina",
            address="1 Rue de Paris",
            city="Paris",
            country="France",
            category="restaurant",
            halal_status="certified",
            cert_authority="Paris Halal Council",
            rating_avg=4.5,
            rating_count=2,
            added_by=admin.id,
            status="active",
        )
        db.session.add(business)
        db.session.commit()
    return app


def test_health_and_public_queries(tmp_path):
    app = seeded_app(tmp_path)
    client = app.test_client()

    health = client.get("/api/v1/health")
    assert health.status_code == 200
    assert health.get_json()["status"] == "ok"

    orgs = client.get("/api/v1/organizations?category=water")
    assert orgs.status_code == 200
    assert len(orgs.get_json()["items"]) == 1

    halal = client.get("/api/v1/halal?q=burger&city=Paris")
    assert halal.status_code == 200
    assert len(halal.get_json()["items"]) == 1


def test_organization_registration_requires_api_key_and_rate_limits(tmp_path):
    app = seeded_app(tmp_path)
    client = app.test_client()

    unauthorized = client.post("/api/v1/organizations", json={"name": "No Key", "category": "water"})
    assert unauthorized.status_code == 401

    for idx in range(3):
        response = client.get("/api/v1/categories", headers={"X-API-Key": "admin-secret"})
        assert response.status_code == 200, idx

    limited = client.get("/api/v1/categories", headers={"X-API-Key": "admin-secret"})
    assert limited.status_code == 429
