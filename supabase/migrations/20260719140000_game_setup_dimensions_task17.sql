ALTER TABLE games ADD COLUMN IF NOT EXISTS map_size TEXT;
ALTER TABLE games ADD COLUMN IF NOT EXISTS climate TEXT;
ALTER TABLE games ADD COLUMN IF NOT EXISTS wetness INTEGER;
ALTER TABLE games ADD CONSTRAINT games_map_size_task17_check CHECK (map_size IS NULL OR map_size IN ('duel', 'small', 'medium', 'large', 'huge'));
ALTER TABLE games ADD CONSTRAINT games_climate_task17_check CHECK (climate IS NULL OR climate IN ('arid', 'cold', 'temperate', 'tropical'));
ALTER TABLE games ADD CONSTRAINT games_wetness_task17_check CHECK (wetness IS NULL OR wetness BETWEEN 0 AND 100);
