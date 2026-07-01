-- The overworld moved from a tile grid (coord_x/coord_y) to a small
-- location graph (see backend/internal/world). A player's position is now
-- just which location they're standing in.
ALTER TABLE accounts DROP COLUMN coord_x;
ALTER TABLE accounts DROP COLUMN coord_y;
ALTER TABLE accounts ADD COLUMN location_id TEXT NOT NULL DEFAULT 'town_square';

-- Parties were referenced (accounts.party_id) since the very first
-- migration but never had a backing table or real formation flow.
CREATE TABLE parties (
    id UUID PRIMARY KEY,
    leader_account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE accounts
    ADD CONSTRAINT fk_party
    FOREIGN KEY (party_id) REFERENCES parties(id) ON DELETE SET NULL;

-- The dungeon grid was only ever rendered as ASCII; the room-card redesign
-- replaced that rendering and the grid column has been dead weight since.
ALTER TABLE dungeons DROP COLUMN grid;
