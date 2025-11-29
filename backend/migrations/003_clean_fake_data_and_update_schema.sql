-- Migration: Clean fake data and update team schema
-- Description: Remove all fake/sample data and ensure schema matches admin panel

-- Step 1: Rename leader_id to owner_id if exists
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'teams' AND column_name = 'leader_id'
    ) THEN
        ALTER TABLE teams RENAME COLUMN leader_id TO owner_id;
    END IF;
END $$;

-- Step 2: Add max_members column if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'teams' AND column_name = 'max_members'
    ) THEN
        ALTER TABLE teams ADD COLUMN max_members INTEGER DEFAULT 10;
    END IF;
END $$;

-- Step 3: Add updated_at column if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'teams' AND column_name = 'updated_at'
    ) THEN
        ALTER TABLE teams ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
    END IF;
END $$;

-- Step 4: Add logo_url column if not exists (snake_case)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'teams' AND column_name = 'logo_url'
    ) THEN
        ALTER TABLE teams ADD COLUMN logo_url TEXT;
    END IF;
END $$;

-- Step 5: Remove logoUrl column if exists (camelCase - old format)
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'teams' AND column_name = 'logoUrl'
    ) THEN
        -- Copy data from logoUrl to logo_url if logo_url is empty
        UPDATE teams SET logo_url = "logoUrl" WHERE logo_url IS NULL AND "logoUrl" IS NOT NULL;
        -- Drop old column
        ALTER TABLE teams DROP COLUMN "logoUrl";
    END IF;
END $$;

-- Step 6: Delete all team members (will cascade delete teams if configured)
DELETE FROM team_members;

-- Step 7: Delete all teams (clean slate - only admin panel created teams will remain)
DELETE FROM teams;

-- Step 8: Create trigger to auto-update updated_at
CREATE OR REPLACE FUNCTION update_teams_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop trigger if exists
DROP TRIGGER IF EXISTS update_teams_updated_at ON teams;

-- Create trigger
CREATE TRIGGER update_teams_updated_at
    BEFORE UPDATE ON teams
    FOR EACH ROW
    EXECUTE FUNCTION update_teams_updated_at();

-- Step 9: Verify schema
SELECT 'Migration completed successfully!' as message;

-- Show current teams table structure
SELECT column_name, data_type, column_default
FROM information_schema.columns
WHERE table_name = 'teams'
ORDER BY ordinal_position;
