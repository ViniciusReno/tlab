CREATE TABLE portfolio_positions (
    id TEXT PRIMARY KEY,
    bond_id TEXT NOT NULL REFERENCES bonds(id),
    purchase_date TEXT,
    quantity REAL NOT NULL CHECK(quantity > 0),
    purchase_pu REAL CHECK(purchase_pu > 0),
    purchase_yield REAL CHECK(purchase_yield > -1),
    invested_brl REAL CHECK(invested_brl > 0),
    note TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
