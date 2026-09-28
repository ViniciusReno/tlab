"""Offline independent reference; Python is not an application/test dependency.

Run from any directory to print the reviewed expected values. Never import Go
production pricing/calendar code or regenerate expectations during Go tests.
"""

import csv
import datetime as dt
from decimal import Decimal, ROUND_DOWN, localcontext
import json
from pathlib import Path

HERE = Path(__file__).resolve().parent
CALENDAR = json.loads((HERE.parents[2] / "data/calendar/anbima-2002-2032-v1.json").read_text())
HOLIDAYS = {dt.date.fromisoformat(s) for s in CALENDAR["holidays"]}
ONE_DAY = dt.timedelta(days=1)


def business_days(start, end):
    # Full weeks + the remainder, minus only holidays falling on weekdays.
    weeks, remainder = divmod((end - start).days, 7)
    weekdays = weeks * 5 + sum((start.weekday() + i) % 7 < 5 for i in range(remainder))
    return weekdays - sum(start <= h < end and h.weekday() < 5 for h in HOLIDAYS)


def number(s):
    return Decimal(s.replace(".", "").replace(",", "."))


def reference():
    results = []
    with (HERE / "quote-contexts.csv").open() as f, localcontext() as ctx:
        ctx.prec = 60
        for row in csv.DictReader(f, delimiter=";"):
            date = dt.datetime.strptime(row["Data Base"], "%d/%m/%Y").date()
            maturity = dt.datetime.strptime(row["Data Vencimento"], "%d/%m/%Y").date()
            next_day = date + ONE_DAY
            while next_day.weekday() >= 5 or next_day in HOLIDAYS:
                next_day += ONE_DAY
            for basis, settlement, pu_column, yield_column in [
                ("purchase", next_day, "PU Compra Manha", "Taxa Compra Manha"),
                ("mark_to_market", date, "PU Base Manha", "Taxa Venda Manha"),
                ("early_exit", date if date >= dt.date(2021, 9, 13) else next_day, "PU Venda Manha", "Taxa Venda Manha"),
            ]:
                pu, base_yield = number(row[pu_column]), number(row[yield_column]) / 100
                days = business_days(settlement, maturity)
                # Independently cross-check the closed-form weekday count by enumeration.
                dates = (settlement + i * ONE_DAY for i in range((maturity - settlement).days))
                assert days == sum(d.weekday() < 5 and d not in HOLIDAYS for d in dates)
                years = {}
                for year in range(settlement.year, maturity.year + 1):
                    start = max(settlement, dt.date(year, 1, 1))
                    end = min(maturity, dt.date(year + 1, 1, 1))
                    if start < end:
                        years[str(year)] = business_days(start, end)
                term = Decimal(days) / 252
                theoretical = Decimal(1000) / (1 + base_yield) ** term
                difference = theoretical.quantize(Decimal(".01"), rounding=ROUND_DOWN) - pu
                validated = abs(difference) <= Decimal(".01")
                result = {
                    "bond_id": "prefixado:" + maturity.isoformat(),
                    "quote_date": date.isoformat(), "basis": basis,
                    "settlement_date": settlement.isoformat(), "business_days": days,
                    "days_by_year": years, "calendar_version": CALENDAR["version"],
                    "base_pu": float(pu), "base_yield": float(base_yield),
                    "theoretical_pu": float(theoretical),
                    "truncated_difference": float(difference),
                    "validation_status": "validated" if validated else "calculation_not_validated",
                    "shocks": [],
                }
                if validated:
                    # Hold the implied nominal anchor fixed; independently discount
                    # that anchor at each hypothetical yield using Decimal powers.
                    anchor = pu * (1 + base_yield) ** term
                    for change in (Decimal("-.01"), Decimal(0), Decimal(".01")):
                        hypothetical = base_yield + change
                        scenario = anchor / (1 + hypothetical) ** term
                        result["shocks"].append({"yield": float(hypothetical), "pu": float(scenario), "variation": float(scenario / pu - 1)})
                results.append(result)
    return results


if __name__ == "__main__":
    print(json.dumps(reference(), indent=2))
