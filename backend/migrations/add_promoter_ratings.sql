CREATE TABLE IF NOT EXISTS promoter_ratings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    promoter_id UUID NOT NULL REFERENCES afftok_users(id) ON DELETE CASCADE,
    visitor_ip VARCHAR(45),
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_promoter_visitor UNIQUE(promoter_id, visitor_ip)
);

CREATE INDEX idx_promoter_ratings_promoter_id ON promoter_ratings(promoter_id);
CREATE INDEX idx_promoter_ratings_created_at ON promoter_ratings(created_at);
