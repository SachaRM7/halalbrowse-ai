from __future__ import annotations

from calendar import monthrange
from datetime import date

try:
    from hijridate import Gregorian
except ImportError:  # pragma: no cover - fallback for environments without deps
    Gregorian = None


def current_hijri_date_string(day: date | None = None) -> str:
    day = day or date.today()
    if Gregorian is not None:
        hijri = Gregorian(day.year, day.month, day.day).to_hijri()
        return f"{hijri.year:04d}-{hijri.month:02d}-{hijri.day:02d}"
    return day.isoformat()


def should_send_reminder(day: date, lead_days: int) -> bool:
    hijri_str = current_hijri_date_string(day)
    try:
        _, _, hijri_day = [int(part) for part in hijri_str.split("-")]
    except ValueError:
        hijri_day = day.day
    if hijri_day == 1:
        return True
    last_day = monthrange(day.year, day.month)[1]
    return (last_day - day.day) <= max(lead_days, 0)
