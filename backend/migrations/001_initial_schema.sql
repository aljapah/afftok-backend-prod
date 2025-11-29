-- AffTok Database Schema
-- Initial Migration - All Tables

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- 1. Admin Users Table (for admin panel login)
-- ============================================
CREATE TABLE IF NOT EXISTS admin_users (
    id SERIAL PRIMARY KEY,
    open_id VARCHAR(64) NOT NULL UNIQUE,
    name TEXT,
    email VARCHAR(320),
    login_method VARCHAR(64),
    role VARCHAR(20) DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_signed_in TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- 2. AffTok Users Table (affiliate marketers)
-- ============================================
CREATE TABLE IF NOT EXISTS afftok_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100),
    avatar_url TEXT,
    bio TEXT,
    role VARCHAR(20) DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'suspended')),
    points INTEGER DEFAULT 0,
    level INTEGER DEFAULT 1,
    total_clicks INTEGER DEFAULT 0,
    total_conversions INTEGER DEFAULT 0,
    total_earnings INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- 3. Networks Table (affiliate networks)
-- ============================================
CREATE TABLE IF NOT EXISTS networks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    logo_url TEXT,
    api_url TEXT,
    api_key TEXT,
    postback_url TEXT,
    hmac_secret TEXT,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- 4. Offers Table
-- ============================================
CREATE TABLE IF NOT EXISTS offers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    network_id UUID REFERENCES networks(id) ON DELETE SET NULL,
    external_offer_id VARCHAR(100),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    image_url TEXT,
    logo_url TEXT,
    destination_url TEXT NOT NULL,
    category VARCHAR(50),
    payout INTEGER DEFAULT 0,
    commission INTEGER DEFAULT 0,
    payout_type VARCHAR(20) DEFAULT 'cpa' CHECK (payout_type IN ('cpa', 'cpl', 'cps', 'cpi')),
    rating DECIMAL(3,2) DEFAULT 0.0,
    users_count INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('active', 'inactive', 'pending')),
    total_clicks INTEGER DEFAULT 0,
    total_conversions INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- 5. User Offers Table (junction table)
-- ============================================
CREATE TABLE IF NOT EXISTS user_offers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES afftok_users(id) ON DELETE CASCADE,
    offer_id UUID NOT NULL REFERENCES offers(id) ON DELETE CASCADE,
    affiliate_link TEXT NOT NULL,
    short_link TEXT,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    clicks INTEGER DEFAULT 0,
    conversions INTEGER DEFAULT 0,
    earnings INTEGER DEFAULT 0,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, offer_id)
);

-- ============================================
-- 6. Clicks Table
-- ============================================
CREATE TABLE IF NOT EXISTS clicks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_offer_id UUID NOT NULL REFERENCES user_offers(id) ON DELETE CASCADE,
    ip_address VARCHAR(45),
    user_agent TEXT,
    device VARCHAR(50),
    browser VARCHAR(50),
    os VARCHAR(50),
    country VARCHAR(2),
    city VARCHAR(100),
    referrer TEXT,
    clicked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index for faster queries
CREATE INDEX IF NOT EXISTS idx_clicks_user_offer_id ON clicks(user_offer_id);
CREATE INDEX IF NOT EXISTS idx_clicks_clicked_at ON clicks(clicked_at);

-- ============================================
-- 7. Conversions Table
-- ============================================
CREATE TABLE IF NOT EXISTS conversions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_offer_id UUID NOT NULL REFERENCES user_offers(id) ON DELETE CASCADE,
    click_id UUID REFERENCES clicks(id) ON DELETE SET NULL,
    external_conversion_id VARCHAR(100),
    amount INTEGER DEFAULT 0,
    commission INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    converted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    approved_at TIMESTAMP
);

-- Create index for faster queries
CREATE INDEX IF NOT EXISTS idx_conversions_user_offer_id ON conversions(user_offer_id);
CREATE INDEX IF NOT EXISTS idx_conversions_status ON conversions(status);

-- ============================================
-- 8. Teams Table
-- ============================================
CREATE TABLE IF NOT EXISTS teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    logo_url TEXT,
    leader_id UUID NOT NULL REFERENCES afftok_users(id) ON DELETE CASCADE,
    total_points INTEGER DEFAULT 0,
    member_count INTEGER DEFAULT 1,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- 9. Team Members Table
-- ============================================
CREATE TABLE IF NOT EXISTS team_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES afftok_users(id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'member' CHECK (role IN ('leader', 'member')),
    points INTEGER DEFAULT 0,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(team_id, user_id)
);

-- ============================================
-- 10. Badges Table
-- ============================================
CREATE TABLE IF NOT EXISTS badges (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    icon_url TEXT,
    criteria TEXT NOT NULL,
    points_reward INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- 11. User Badges Table
-- ============================================
CREATE TABLE IF NOT EXISTS user_badges (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES afftok_users(id) ON DELETE CASCADE,
    badge_id UUID NOT NULL REFERENCES badges(id) ON DELETE CASCADE,
    earned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, badge_id)
);

-- ============================================
-- Triggers for updated_at
-- ============================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply trigger to tables with updated_at
CREATE TRIGGER update_admin_users_updated_at BEFORE UPDATE ON admin_users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_afftok_users_updated_at BEFORE UPDATE ON afftok_users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_networks_updated_at BEFORE UPDATE ON networks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_offers_updated_at BEFORE UPDATE ON offers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Initial Data (Optional)
-- ============================================

-- Insert default badges
INSERT INTO badges (name, description, icon_url, criteria, points_reward) VALUES
('First Click', 'Get your first click', 'https://example.com/badges/first-click.png', 'total_clicks >= 1', 10),
('First Conversion', 'Get your first conversion', 'https://example.com/badges/first-conversion.png', 'total_conversions >= 1', 50),
('Rookie Marketer', 'Reach 10 conversions', 'https://example.com/badges/rookie.png', 'total_conversions >= 10', 100),
('Pro Marketer', 'Reach 50 conversions', 'https://example.com/badges/pro.png', 'total_conversions >= 50', 500),
('Expert Marketer', 'Reach 200 conversions', 'https://example.com/badges/expert.png', 'total_conversions >= 200', 2000)
ON CONFLICT DO NOTHING;

-- Success message
SELECT 'Database schema created successfully!' as message;
