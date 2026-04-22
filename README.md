# Sadaqah Tracker

Sadaqah Tracker is a Flask + SQLite application for transparent Zakat and Sadaqah logging. It includes a tamper-evident SHA-256 donation ledger, verified organizations, halal business search, and a Hijri-aware reminder endpoint.

## Project layout

- `app/` Flask application package
- `scripts/sadaqah-reminder.sh` cron-friendly reminder trigger
- `tests/` pytest coverage for core behaviors

## Setup

```bash
cd /root/.hermes/agents/coder/code/20260422
python3 -m venv .venv
source .venv/bin/activate
pip install -r app/requirements.txt
cp .env.example .env
python run.py
```

App default URL: `http://localhost:5000`

## Run tests

```bash
cd /root/.hermes/agents/coder/code/20260422
source .venv/bin/activate
pytest -q tests
```

## API summary

- `GET /api/v1/health`
- `GET /api/v1/categories`
- `GET /api/v1/organizations?category=water`
- `GET /api/v1/organizations/<slug>`
- `POST /api/v1/organizations` with `X-API-Key`
- `POST /api/v1/donations` with `X-API-Key`
- `GET /api/v1/donations/<id>/verify`
- `GET /api/v1/ledger`
- `GET /api/v1/halal?q=burger&city=Paris`
- `POST /api/v1/halal` with `X-API-Key`
- `POST /api/v1/internal/send-reminders` with `X-Internal-Token`

## Seed an API key

Example Python shell snippet:

```python
from app import create_app
from app.models import db
from app.models.api_key import ApiKey

app = create_app()
with app.app_context():
    key = ApiKey(name="local-dev")
    raw = key.set_key()
    db.session.add(key)
    db.session.commit()
    print(raw)
```

## Reminder cron

Example crontab entry:

```cron
0 9 * * * /root/.hermes/agents/coder/code/20260422/scripts/sadaqah-reminder.sh >> /tmp/sadaqah-reminder.log 2>&1
```

## Contribution guide

1. Create a branch.
2. Add or update tests first.
3. Run `pytest -q tests`.
4. Open a pull request with a short change summary.

## License

MIT
