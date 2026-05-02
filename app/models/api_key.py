from __future__ import annotations

from datetime import datetime, UTC
import hashlib
import secrets

from app.models import db


class ApiKey(db.Model):
    __tablename__ = "api_keys"

    id = db.Column(db.Integer, primary_key=True)
    key_hash = db.Column(db.String(64), unique=True, nullable=False, index=True)
    name = db.Column(db.String(100), nullable=False)
    organization_id = db.Column(db.Integer, db.ForeignKey("organizations.id"), nullable=True)
    is_active = db.Column(db.Boolean, nullable=False, default=True)
    rate_limit = db.Column(db.Integer, nullable=False, default=100)
    created_at = db.Column(db.DateTime, nullable=False, default=lambda: datetime.now(UTC))

    def set_key(self, raw_key: str | None = None) -> str:
        raw_key = raw_key or secrets.token_urlsafe(32)
        self.key_hash = hashlib.sha256(raw_key.encode()).hexdigest()
        return raw_key

    def check_key(self, raw_key: str) -> bool:
        return self.key_hash == hashlib.sha256(raw_key.encode()).hexdigest()

    @classmethod
    def find_active(cls, raw_key: str | None):
        if not raw_key:
            return None
        key_hash = hashlib.sha256(raw_key.encode()).hexdigest()
        return db.session.execute(db.select(cls).filter_by(key_hash=key_hash, is_active=True)).scalar_one_or_none()
