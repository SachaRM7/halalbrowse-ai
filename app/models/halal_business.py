from __future__ import annotations

from datetime import datetime, UTC

from app.models import db


class HalalBusiness(db.Model):
    __tablename__ = "halal_businesses"

    id = db.Column(db.Integer, primary_key=True)
    name = db.Column(db.String(200), nullable=False, index=True)
    address = db.Column(db.Text, nullable=False)
    city = db.Column(db.String(100), nullable=False, index=True)
    country = db.Column(db.String(100), nullable=False)
    category = db.Column(db.String(50), nullable=False, index=True)
    halal_status = db.Column(db.String(20), nullable=False, default="pending")
    cert_authority = db.Column(db.String(200), nullable=True)
    rating_avg = db.Column(db.Float, nullable=False, default=0.0)
    rating_count = db.Column(db.Integer, nullable=False, default=0)
    added_by = db.Column(db.Integer, db.ForeignKey("users.id"), nullable=True)
    status = db.Column(db.String(20), nullable=False, default="pending_review", index=True)
    created_at = db.Column(db.DateTime, nullable=False, default=lambda: datetime.now(UTC))
