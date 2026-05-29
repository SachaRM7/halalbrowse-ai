from __future__ import annotations

from datetime import datetime, UTC

from app.models import db


class User(db.Model):
    __tablename__ = "users"

    id = db.Column(db.Integer, primary_key=True)
    telegram_id = db.Column(db.BigInteger, unique=True, nullable=False, index=True)
    display_name = db.Column(db.String(100), nullable=False)
    email = db.Column(db.String(255), nullable=True)
    role = db.Column(db.String(20), nullable=False, default="user")
    reminder_enabled = db.Column(db.Boolean, nullable=False, default=True)
    reminder_lead_days = db.Column(db.Integer, nullable=False, default=3)
    created_at = db.Column(db.DateTime, nullable=False, default=lambda: datetime.now(UTC))

    donations = db.relationship("Donation", back_populates="user", cascade="all, delete-orphan")
