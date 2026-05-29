from flask import Blueprint, jsonify

admin_bp = Blueprint("admin", __name__, url_prefix="/admin")


@admin_bp.get("/health")
def admin_health():
    return jsonify({"status": "ok", "scope": "admin"})
