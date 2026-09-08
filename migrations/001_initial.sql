CREATE TABLE bonds (
    id TEXT PRIMARY KEY,
    kind TEXT NOT NULL,
    name TEXT NOT NULL,
    maturity TEXT NOT NULL
);
CREATE TABLE market_quotes (
    bond_id TEXT NOT NULL REFERENCES bonds(id),
    quote_date TEXT NOT NULL,
    buy_yield REAL,
    sell_yield REAL,
    buy_pu REAL,
    sell_pu REAL,
    base_pu REAL,
    source TEXT NOT NULL,
    PRIMARY KEY (bond_id, quote_date)
);
