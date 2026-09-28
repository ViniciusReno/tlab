"""Independent evidence for the approved settlement amendment, not runtime policy.

Print fixed expectations with 60-digit Decimal arithmetic; no production Go code.
"""
import csv
import datetime as dt
import json
from decimal import Decimal, ROUND_DOWN, localcontext
from reference import HERE, HOLIDAYS, ONE_DAY, business_days, number


def reference():
    results = []
    with (HERE / 'redemption-transition.csv').open() as f, localcontext() as ctx:
        ctx.prec = 60
        for row in csv.DictReader(f, delimiter=';'):
            date = dt.datetime.strptime(row['Data Base'], '%d/%m/%Y').date()
            maturity = dt.datetime.strptime(row['Data Vencimento'], '%d/%m/%Y').date()
            next_day = date + ONE_DAY
            while next_day.weekday() >= 5 or next_day in HOLIDAYS:
                next_day += ONE_DAY
            pu = number(row['PU Venda Manha'])
            y = number(row['Taxa Venda Manha']) / 100
            for convention, settlement in [('D0', date), ('D+1', next_day)]:
                days = business_days(settlement, maturity)
                assert days == sum((settlement + i * ONE_DAY).weekday() < 5 and
                                  settlement + i * ONE_DAY not in HOLIDAYS
                                  for i in range((maturity - settlement).days))
                term = Decimal(days) / 252
                theoretical = Decimal(1000) / (1 + y) ** term
                truncated = theoretical.quantize(Decimal('.01'), rounding=ROUND_DOWN)
                anchor = pu * (1 + y) ** term
                shocks = []
                for delta in [Decimal('-.01'), Decimal('0'), Decimal('.01')]:
                    price = anchor / (1 + y + delta) ** term
                    shocks.append({'yield': float(y + delta), 'pu': float(price), 'variation': float(price / pu - 1)})
                results.append({'bond_id': 'prefixado:' + maturity.isoformat(),
                                'quote_date': date.isoformat(), 'convention': convention,
                                'settlement_date': settlement.isoformat(), 'business_days': days,
                                'official_pu': float(pu), 'base_yield': float(y),
                                'theoretical_truncated': float(truncated),
                                'difference': float(truncated - pu),
                                'matches': abs(truncated - pu) <= Decimal('.01'),
                                'shocks': shocks})
    return results


if __name__ == '__main__':
    print(json.dumps(reference(), indent=2))
