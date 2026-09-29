"""Independent 60-digit fixed-indexation IPCA+ reference; never used by Go tests."""
import csv
import datetime as dt
import json
from decimal import Decimal, localcontext
from reference import HERE, ONE_DAY, number

CALENDAR = json.loads((HERE.parents[2] / 'data/calendar/anbima-2002-2050-v1.json').read_text())
HOLIDAYS = {dt.date.fromisoformat(d) for d in CALENDAR['holidays']}


def business_days(start, end):
    weeks, remainder = divmod((end - start).days, 7)
    weekdays = weeks * 5 + sum((start.weekday() + i) % 7 < 5 for i in range(remainder))
    return weekdays - sum(start <= h < end and h.weekday() < 5 for h in HOLIDAYS)


def reference():
    results = []
    with (HERE / 'm3-quotes.csv').open() as f, localcontext() as ctx:
        ctx.prec = 60
        for row in csv.DictReader(f, delimiter=';'):
            if row['Tipo Titulo'] != 'Tesouro IPCA+':
                continue
            date = dt.datetime.strptime(row['Data Base'], '%d/%m/%Y').date()
            maturity = dt.datetime.strptime(row['Data Vencimento'], '%d/%m/%Y').date()
            next_day = date + ONE_DAY
            while next_day.weekday() >= 5 or next_day in HOLIDAYS:
                next_day += ONE_DAY
            for basis, settlement, price_column, rate_column in [
                ('purchase', next_day, 'PU Compra Manha', 'Taxa Compra Manha'),
                ('mark_to_market', date, 'PU Base Manha', 'Taxa Venda Manha'),
                ('early_exit', next_day if date < dt.date(2021, 9, 13) else date, 'PU Venda Manha', 'Taxa Venda Manha'),
            ]:
                days = business_days(settlement, maturity)
                assert days == sum((settlement + i * ONE_DAY).weekday() < 5 and
                                  settlement + i * ONE_DAY not in HOLIDAYS
                                  for i in range((maturity - settlement).days))
                pu, y = number(row[price_column]), number(row[rate_column]) / 100
                term = Decimal(days) / 252
                # This implied indexation anchor comes from the official pair;
                # it is not an independently observed official VNA.
                anchor = pu * (1 + y) ** term
                shocks = []
                for delta in [Decimal('-.01'), Decimal('0'), Decimal('.01')]:
                    price = anchor / (1 + y + delta) ** term
                    shocks.append({'yield': float(y + delta), 'pu': float(price), 'variation': float(price / pu - 1)})
                results.append({'bond_id': 'ipca:' + maturity.isoformat(), 'quote_date': date.isoformat(),
                                'basis': basis, 'settlement_date': settlement.isoformat(), 'business_days': days,
                                'base_pu': float(pu), 'base_yield': float(y),
                                'derived_fixed_indexation_anchor': str(anchor), 'shocks': shocks})
    return results


if __name__ == '__main__':
    print(json.dumps(reference(), indent=2))
