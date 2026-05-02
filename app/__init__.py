from flask import Flask, render_template

from app.config import Config
from app.models import db
from app.routes import api_v1_bp, admin_bp


def create_app(config_overrides: dict | None = None) -> Flask:
    app = Flask(__name__)
    app.config.from_object(Config)
    if config_overrides:
        app.config.update(config_overrides)

    db.init_app(app)
    with app.app_context():
        db.create_all()

    @app.get("/")
    def dashboard():
        return render_template("dashboard.html", app_name=app.config["APP_NAME"])

    app.register_blueprint(api_v1_bp)
    app.register_blueprint(admin_bp)
    return app
