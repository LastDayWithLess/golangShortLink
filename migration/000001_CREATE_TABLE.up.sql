
CREATE TABLE IF NOT EXISTS links (
    id SERIAL PRIMARY KEY,
    url TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS short_links (
    id SERIAL PRIMARY KEY,
    id_url INTEGER NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    short_url VARCHAR(6) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    accessed_at TIMESTAMP DEFAULT NULL, 
    accessed_count INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_short_links_short_url ON short_links(short_url);
CREATE INDEX IF NOT EXISTS idx_short_links_id_url ON short_links(id_url);