from flask_sqlalchemy import SQLAlchemy

db = SQLAlchemy()

from app.models.api_key import ApiKey  # noqa: E402,F401
from app.models.donation import Donation  # noqa: E402,F401
from app.models.halal_business import HalalBusiness  # noqa: E402,F401
from app.models.organization import Organization  # noqa: E402,F401
from app.models.user import User  # noqa: E402,F401
