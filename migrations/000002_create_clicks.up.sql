CREATE TABLE clicks (
    id BIGSERIAL PRIMARY KEY,
    code_id BIGINT NOT NULL REFERENCES codes(id) ON DELETE CASCADE,
    clicked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reference TEXT
);

CREATE INDEX idx_clicks_code_id ON clicks(code_id);