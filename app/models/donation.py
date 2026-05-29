from __future__ import annotations

from datetime import datetime, UTC
from decimal import Decimal

from app.models import db


class Donation(db.Model):
    __tablename__ = "donations"

    id = db.Column(db.Integer, primary_key=True)
    user_id = db.Column(db.Integer, db.ForeignKey("users.id"), nullable=False, index=True)
    organization_id = db.Column(db.Integer, db.ForeignKey("organizations.id"), nullable=False, index=True)
    amount = db.Column(db.Numeric(10, 2), nullable=False)
    currency = db.Column(db.String(3), nullable=False, default="USD")
    category = db.Column(db.String(50), nullable=False, index=True)
    note = db.Column(db.Text, nullable=True)
    timestamp_utc = db.Column(db.DateTime, nullable=False, default=lambda: datetime.now(UTC))
    timestamp_hijri = db.Column(db.String(20), nullable=False)
    prev_hash = db.Column(db.String(64), nullable=False)
    own_hash = db.Column(db.String(64), nullable=False, unique=True, index=True)

    user = db.relationship("User", back_populates="donations")
    organization = db.relationship("Organization", back_populates="donations")

    def as_chain_payload(self) -> dict:
        timestamp = self.timestamp_utc
        if timestamp.tzinfo is None:
            timestamp = timestamp.replace(tzinfo=UTC)
        return {
            "id": self.id,
            "user_id": self.user_id,
            "amount": f"{Decimal(self.amount):.2f}",
            "organization_id": self.organization_id,
            "timestamp_utc": timestamp.isoformat(),
            "prev_hash": self.prev_hash,
            "own_hash": self.own_hash,
        }
