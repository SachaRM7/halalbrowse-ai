from __future__ import annotations

import os
from pathlib import Path


class Config:
    APP_NAME = "Sadaqah Tracker"
    SECRET_KEY = os.getenv("SECRET_KEY", "dev-secret-key")
    BASE_DIR = Path(__file__).resolve().parents[1]
    SQLALCHEMY_DATABASE_URI = os.getenv("DATABASE_URL", f"sqlite:///{BASE_DIR / 'sadaqah_tracker.db'}")
    SQLALCHEMY_TRACK_MODIFICATIONS = False
    API_RATE_LIMIT_REQUESTS = int(os.getenv("API_RATE_LIMIT_REQUESTS", "100"))
    API_RATE_LIMIT_WINDOW = int(os.getenv("API_RATE_LIMIT_WINDOW", "60"))
    INTERNAL_API_TOKEN = os.getenv("INTERNAL_API_TOKEN", "internal-dev-token")
    DASHBOARD_URL = os.getenv("DASHBOARD_URL", "http://localhost:5000/")
    TELEGRAM_BOT_TOKEN = os.getenv("TELEGRAM_BOT_TOKEN", "")
    TELEGRAM_CHAT_ID = os.getenv("TELEGRAM_CHAT_ID", "")
