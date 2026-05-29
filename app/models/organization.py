from __future__ import annotations

from datetime import datetime, UTC

from app.models import db


class Organization(db.Model):
    __tablename__ = "organizations"

    id = db.Column(db.Integer, primary_key=True)
    slug = db.Column(db.String(100), unique=True, nullable=False, index=True)
    name = db.Column(db.String(200), nullable=False)
    description = db.Column(db.Text, nullable=False)
    category = db.Column(db.String(50), nullable=False, index=True)
    website = db.Column(db.String(500), nullable=True)
    contact_email = db.Column(db.String(255), nullable=True)
    status = db.Column(db.String(20), nullable=False, default="pending", index=True)
    created_by = db.Column(db.Integer, db.ForeignKey("users.id"), nullable=True)
    created_at = db.Column(db.DateTime, nullable=False, default=lambda: datetime.now(UTC))

    donations = db.relationship("Donation", back_populates="organization")

    def total_received(self) -> float:
        result = db.session.query(db.func.coalesce(db.func.sum(Donation.amount), 0)).filter(
            Donation.organization_id == self.id
        ).scalar()
        return float(result or 0)


from app.models.donation import Donation  # noqa: E402
