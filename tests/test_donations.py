from decimal import Decimal

from app import create_app
from app.models import db
from app.models.api_key import ApiKey
from app.models.donation import Donation
from app.models.organization import Organization
from app.models.user import User


def create_seeded_app(tmp_path):
    app = create_app({
        "TESTING": True,
        "SQLALCHEMY_DATABASE_URI": f"sqlite:///{tmp_path / 'test.db'}",
        "API_RATE_LIMIT_WINDOW": 60,
    })
    with app.app_context():
        db.create_all()
        user = User(telegram_id=123456789, display_name="Sacha")
        db.session.add(user)
        db.session.flush()

        org = Organization(
            slug="clean-water",
            name="Clean Water Trust",
            description="Builds wells",
            category="water",
            status="verified",
            created_by=user.id,
        )
        db.session.add(org)
        db.session.flush()

        key = ApiKey(name="integration", organization_id=org.id)
        key.set_key("secret-key")
        db.session.add(key)
        db.session.commit()
    return app


def test_create_donation_appends_hash_chain(tmp_path):
    app = create_seeded_app(tmp_path)
    client = app.test_client()

    response = client.post(
        "/api/v1/donations",
        headers={"X-API-Key": "secret-key"},
        json={
            "user_telegram_id": 123456789,
            "organization_id": 1,
            "amount": 250.0,
            "currency": "USD",
            "note": "Zakat al-mal 1448",
        },
    )

    assert response.status_code == 201
    payload = response.get_json()
    assert payload["verified"] is True
    assert len(payload["hash"]) == 64

    with app.app_context():
        donation = db.session.execute(db.select(Donation)).scalar_one()
        assert donation.prev_hash
        assert donation.own_hash == payload["hash"]
        assert donation.amount == Decimal("250.00")
