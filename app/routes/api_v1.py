from __future__ import annotations

from collections import defaultdict, deque
from datetime import datetime, UTC
from decimal import Decimal, InvalidOperation
import re

from flask import Blueprint, current_app, jsonify, request

from app.models import db
from app.models.api_key import ApiKey
from app.models.donation import Donation
from app.models.halal_business import HalalBusiness
from app.models.organization import Organization
from app.models.user import User
from app.services.hash_chain import GENESIS_HASH, compute_donation_hash, verify_chain
from app.services.hijri_service import current_hijri_date_string, should_send_reminder

api_v1_bp = Blueprint("api_v1", __name__, url_prefix="/api/v1")
_request_log: dict[str, deque[float]] = defaultdict(deque)
CATEGORY_OPTIONS = ["water", "food", "education", "healthcare", "emergency"]


def _json_error(message: str, status_code: int):
    response = jsonify({"error": message})
    response.status_code = status_code
    return response


def _slugify(value: str) -> str:
    value = re.sub(r"[^a-zA-Z0-9]+", "-", value.strip().lower()).strip("-")
    return value or "organization"


def _paginate(query, serializer):
    page = max(int(request.args.get("page", 1)), 1)
    per_page = min(max(int(request.args.get("per_page", 10)), 1), 100)
    pagination = db.paginate(query.order_by(db.text("id ASC")), page=page, per_page=per_page, error_out=False)
    return jsonify({
        "items": [serializer(item) for item in pagination.items],
        "page": page,
        "per_page": per_page,
        "total": pagination.total,
    })


def _authorize(required: bool = True):
    raw_key = request.headers.get("X-API-Key")
    api_key = ApiKey.find_active(raw_key)
    if not api_key and required:
        return None, _json_error("Valid API key required", 401)
    if api_key:
        limited = _check_rate_limit(api_key)
        if limited is not None:
            return None, limited
    return api_key, None


def _check_rate_limit(api_key: ApiKey):
    now = datetime.now(UTC).timestamp()
    key = api_key.key_hash
    window = current_app.config["API_RATE_LIMIT_WINDOW"]
    limit = min(api_key.rate_limit or current_app.config["API_RATE_LIMIT_REQUESTS"], current_app.config["API_RATE_LIMIT_REQUESTS"])
    events = _request_log[key]
    while events and now - events[0] >= window:
        events.popleft()
    if len(events) >= limit:
        return _json_error("Rate limit exceeded", 429)
    events.append(now)
    return None


def _organization_to_dict(org: Organization) -> dict:
    return {
        "id": org.id,
        "slug": org.slug,
        "name": org.name,
        "description": org.description,
        "category": org.category,
        "website": org.website,
        "contact_email": org.contact_email,
        "status": org.status,
        "total_received": org.total_received(),
    }


def _business_to_dict(item: HalalBusiness) -> dict:
    return {
        "id": item.id,
        "name": item.name,
        "address": item.address,
        "city": item.city,
        "country": item.country,
        "category": item.category,
        "halal_status": item.halal_status,
        "cert_authority": item.cert_authority,
        "community_rating": item.rating_avg,
        "rating_count": item.rating_count,
        "status": item.status,
    }


@api_v1_bp.get("/health")
def health():
    return jsonify({"status": "ok"})


@api_v1_bp.get("/categories")
def categories():
    _, error = _authorize(required=False)
    if error is not None:
        return error
    return jsonify({"items": CATEGORY_OPTIONS})


@api_v1_bp.get("/organizations")
def list_organizations():
    category = request.args.get("category")
    status = request.args.get("status", "verified")
    search = request.args.get("q")

    query = db.select(Organization).filter_by(status=status)
    if category:
        query = query.filter(Organization.category == category)
    if search:
        like = f"%{search}%"
        query = query.filter(db.or_(Organization.name.ilike(like), Organization.description.ilike(like)))
    return _paginate(query, _organization_to_dict)


@api_v1_bp.get("/organizations/<string:slug>")
def organization_detail(slug: str):
    org = db.session.execute(db.select(Organization).filter_by(slug=slug)).scalar_one_or_none()
    if not org:
        return _json_error("Organization not found", 404)
    return jsonify(_organization_to_dict(org))


@api_v1_bp.post("/organizations")
def register_organization():
    api_key, error = _authorize(required=True)
    if error is not None:
        return error

    data = request.get_json(silent=True) or {}
    name = (data.get("name") or "").strip()
    description = (data.get("description") or "").strip()
    category = (data.get("category") or "").strip()
    if not name or not description or category not in CATEGORY_OPTIONS:
        return _json_error("name, description and a valid category are required", 400)

    slug = data.get("slug") or _slugify(name)
    if db.session.execute(db.select(Organization).filter_by(slug=slug)).scalar_one_or_none():
        return _json_error("Organization slug already exists", 400)

    org = Organization(
        slug=slug,
        name=name,
        description=description,
        category=category,
        website=data.get("website"),
        contact_email=data.get("contact_email"),
        status="pending",
    )
    if api_key.organization_id:
        org.created_by = db.session.execute(db.select(Organization.created_by).filter_by(id=api_key.organization_id)).scalar_one_or_none()
    db.session.add(org)
    db.session.commit()
    return jsonify({"organization": _organization_to_dict(org)}), 201


@api_v1_bp.post("/donations")
def create_donation():
    _, error = _authorize(required=True)
    if error is not None:
        return error

    data = request.get_json(silent=True) or {}
    telegram_id = data.get("user_telegram_id")
    organization_id = data.get("organization_id")
    try:
        amount = Decimal(str(data.get("amount"))).quantize(Decimal("0.01"))
    except (InvalidOperation, TypeError):
        return _json_error("amount must be a valid decimal", 400)

    if not telegram_id or not organization_id or amount <= 0:
        return _json_error("user_telegram_id, organization_id and positive amount are required", 400)

    organization = db.session.get(Organization, organization_id)
    if not organization or organization.status != "verified":
        return _json_error("Verified organization required", 404)

    user = db.session.execute(db.select(User).filter_by(telegram_id=telegram_id)).scalar_one_or_none()
    if user is None:
        user = User(telegram_id=telegram_id, display_name=f"User {telegram_id}")
        db.session.add(user)
        db.session.flush()

    timestamp_utc = datetime.now(UTC)
    donation = Donation(
        user_id=user.id,
        organization_id=organization.id,
        amount=amount,
        currency=(data.get("currency") or "USD")[:3].upper(),
        category=organization.category,
        note=data.get("note"),
        timestamp_utc=timestamp_utc,
        timestamp_hijri=current_hijri_date_string(timestamp_utc.date()),
        prev_hash=GENESIS_HASH,
        own_hash="",
    )
    db.session.add(donation)
    db.session.flush()

    previous = db.session.execute(
        db.select(Donation).filter(Donation.id < donation.id).order_by(Donation.id.desc())
    ).scalars().first()
    donation.prev_hash = previous.own_hash if previous else GENESIS_HASH
    donation.own_hash = compute_donation_hash(donation.as_chain_payload(), donation.prev_hash)
    db.session.commit()

    verification = verify_chain([item.as_chain_payload() for item in db.session.execute(db.select(Donation).order_by(Donation.id.asc())).scalars()])
    return jsonify({
        "donation_id": donation.id,
        "hash": donation.own_hash,
        "timestamp_utc": donation.timestamp_utc.isoformat(),
        "timestamp_hijri": donation.timestamp_hijri,
        "verified": verification["valid"],
    }), 201


@api_v1_bp.get("/donations/<int:donation_id>/verify")
def verify_donation(donation_id: int):
    donations = db.session.execute(db.select(Donation).filter(Donation.id <= donation_id).order_by(Donation.id.asc())).scalars().all()
    if not donations or donations[-1].id != donation_id:
        return _json_error("Donation not found", 404)
    verification = verify_chain([donation.as_chain_payload() for donation in donations])
    return jsonify(verification)


@api_v1_bp.get("/ledger")
def public_ledger():
    query = db.select(Donation)
    return _paginate(query, lambda item: {
        "id": item.id,
        "amount": float(item.amount),
        "currency": item.currency,
        "category": item.category,
        "organization_id": item.organization_id,
        "timestamp_utc": item.timestamp_utc.isoformat(),
        "timestamp_hijri": item.timestamp_hijri,
        "prev_hash": item.prev_hash,
        "own_hash": item.own_hash,
    })


@api_v1_bp.get("/halal")
def search_halal():
    query = db.select(HalalBusiness).filter(HalalBusiness.status == "active")
    text = request.args.get("q")
    city = request.args.get("city")
    category = request.args.get("category")
    if text:
        like = f"%{text}%"
        query = query.filter(db.or_(HalalBusiness.name.ilike(like), HalalBusiness.address.ilike(like)))
    if city:
        query = query.filter(HalalBusiness.city.ilike(city))
    if category:
        query = query.filter(HalalBusiness.category == category)
    return _paginate(query, _business_to_dict)


@api_v1_bp.post("/halal")
def suggest_halal_business():
    _, error = _authorize(required=True)
    if error is not None:
        return error
    data = request.get_json(silent=True) or {}
    required_fields = ["name", "address", "city", "country", "category"]
    if any(not data.get(field) for field in required_fields):
        return _json_error("name, address, city, country and category are required", 400)
    business = HalalBusiness(
        name=data["name"],
        address=data["address"],
        city=data["city"],
        country=data["country"],
        category=data["category"],
        halal_status=data.get("halal_status", "pending"),
        cert_authority=data.get("cert_authority"),
        status="pending_review",
    )
    db.session.add(business)
    db.session.commit()
    return jsonify({"business": _business_to_dict(business)}), 201


@api_v1_bp.post("/internal/send-reminders")
def send_reminders():
    token = request.headers.get("X-Internal-Token")
    if token != current_app.config["INTERNAL_API_TOKEN"]:
        return _json_error("Forbidden", 403)

    today = datetime.now(UTC).date()
    users = db.session.execute(db.select(User).filter_by(reminder_enabled=True)).scalars().all()
    reminders = []
    for user in users:
        if should_send_reminder(today, user.reminder_lead_days):
            reminders.append({
                "telegram_id": user.telegram_id,
                "message": (
                    f"Assalamu Alaikum {user.display_name}! It's {current_hijri_date_string(today)} — "
                    f"time to calculate and pay your Zakat. Tap to open Sadaqah Tracker: "
                    f"{current_app.config['DASHBOARD_URL']}"
                ),
            })
    return jsonify({"sent": len(reminders), "messages": reminders})
