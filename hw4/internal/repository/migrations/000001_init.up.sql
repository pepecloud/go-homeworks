CREATE TABLE IF NOT EXISTS orders (
    id INTEGER PRIMARY KEY,
    status BOOLEAN NOT NULL,
    amount INTEGER NOT NULL CHECK (amount > 0),
    source TEXT NOT NULL DEFAULT 'api',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS transactions (
    id INTEGER PRIMARY KEY,
    amount INTEGER NOT NULL CHECK (amount > 0),
    date TEXT NOT NULL,
    source TEXT NOT NULL DEFAULT 'api',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS change_history (
    id BIGSERIAL PRIMARY KEY,
    entity_type TEXT NOT NULL,
    action TEXT NOT NULL,
    entity_id INTEGER NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_change_history_entity ON change_history (entity_type, entity_id);
