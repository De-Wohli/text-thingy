CREATE TABLE accounts (
    id UUID PRIMARY KEY,
    display_name TEXT NOT NULL,
    honor INT NOT NULL DEFAULT 0,
    gold INT NOT NULL DEFAULT 50,
    active_character_id UUID,
    coord_x INT NOT NULL DEFAULT 10,
    coord_y INT NOT NULL DEFAULT 7,
    party_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE characters (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    race_id TEXT NOT NULL,
    class_id TEXT NOT NULL,
    level INT NOT NULL DEFAULT 1,
    status TEXT NOT NULL DEFAULT 'IDLE',
    hp_current INT NOT NULL,
    hp_max INT NOT NULL,
    ability_str INT NOT NULL,
    ability_dex INT NOT NULL,
    ability_con INT NOT NULL,
    ability_int INT NOT NULL,
    ability_wis INT NOT NULL,
    ability_cha INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_characters_account_id ON characters(account_id);

ALTER TABLE accounts
    ADD CONSTRAINT fk_active_character
    FOREIGN KEY (active_character_id) REFERENCES characters(id) ON DELETE SET NULL;

CREATE TABLE honor_log (
    id BIGSERIAL PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    delta INT NOT NULL,
    typology TEXT NOT NULL,
    reason TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE dungeons (
    id UUID PRIMARY KEY,
    party_id UUID NOT NULL,
    grid JSONB NOT NULL,
    rooms JSONB NOT NULL,
    encounters JSONB NOT NULL,
    resolved BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at TIMESTAMPTZ
);
