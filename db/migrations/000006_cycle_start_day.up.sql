ALTER TABLE users 
ADD COLUMN cycle_start_day INT NOT NULL DEFAULT 1 
CHECK (cycle_start_day >= 1 AND cycle_start_day <= 31);
