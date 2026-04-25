from app.services.hash_chain import GENESIS_HASH, compute_donation_hash, verify_chain


def test_compute_donation_hash_is_deterministic():
    donation = {
        "id": 1,
        "user_id": 11,
        "amount": "100.00",
        "organization_id": 7,
        "timestamp_utc": "2026-04-22T15:00:00+00:00",
    }

    result = compute_donation_hash(donation, GENESIS_HASH)

    assert len(result) == 64
    assert result == compute_donation_hash(donation, GENESIS_HASH)


def test_verify_chain_detects_tampering():
    first = {
        "id": 1,
        "user_id": 11,
        "amount": "100.00",
        "organization_id": 7,
        "timestamp_utc": "2026-04-22T15:00:00+00:00",
    }
    first["prev_hash"] = GENESIS_HASH
    first["own_hash"] = compute_donation_hash(first, GENESIS_HASH)

    second = {
        "id": 2,
        "user_id": 12,
        "amount": "50.00",
        "organization_id": 8,
        "timestamp_utc": "2026-04-23T12:00:00+00:00",
    }
    second["prev_hash"] = first["own_hash"]
    second["own_hash"] = compute_donation_hash(second, first["own_hash"])

    assert verify_chain([first, second])["valid"] is True

    second["amount"] = "75.00"
    verification = verify_chain([first, second])
    assert verification["valid"] is False
    assert verification["broken_at_id"] == 2
