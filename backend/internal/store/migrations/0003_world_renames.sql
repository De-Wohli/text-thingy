-- The world location IDs changed in the VTT rework:
--   town_square   → the_town      (hub renamed for clarity)
--   mine_entrance → west_fields   (mine entrance was removed; dungeon
--                                  trigger moved to the western fields)
-- Any accounts stuck at one of the old IDs gets moved to the_town.

UPDATE accounts
SET location_id = 'the_town'
WHERE location_id IN ('town_square', 'mine_entrance');

-- Update the column default so new accounts land at the_town.
ALTER TABLE accounts
    ALTER COLUMN location_id SET DEFAULT 'the_town';
