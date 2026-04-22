from __future__ import annotations

import hashlib
from decimal import Decimal

GENESIS_HASH = hashlib.sha256("SadaqahTracker Genesis Block 2026".encode()).hexdigest()


def compute_donation_hash(donation: dict, prev_hash: str) -> str:
    payload = (
        f"{donation['id']}"
        f"{donation['user_id']}"
        f"{Decimal(str(donation['amount'])):.2f}"
        f"{donation['organization_id']}"
        f"{donation['timestamp_utc']}"
        f"{prev_hash}"
    )
    return hashlib.sha256(payload.encode()).hexdigest()


def verify_chain(donations: list[dict]) -> dict:
    prev_hash = GENESIS_HASH
    for donation in donations:
        expected = compute_donation_hash(donation, prev_hash)
        if donation.get("prev_hash") != prev_hash or expected != donation.get("own_hash"):
            return {"valid": False, "broken_at_id": donation.get("id")}
        prev_hash = donation["own_hash"]
    return {"valid": True, "broken_at_id": None}
