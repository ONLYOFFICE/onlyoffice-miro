CREATE TABLE IF NOT EXISTS authentications (
    team_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    token_type TEXT NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    expires_at INTEGER NOT NULL,
    scope TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (team_id, user_id)
);

CREATE TABLE settings (
    team_id TEXT NOT NULL,
    board_id TEXT NOT NULL,
    address VARCHAR(255) NOT NULL,
    header VARCHAR(255) NOT NULL,
    secret TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (team_id, board_id)
);

CREATE INDEX IF NOT EXISTS idx_authentications_team_id ON authentications(team_id);
CREATE INDEX IF NOT EXISTS idx_authentications_user_id ON authentications(user_id);
CREATE INDEX IF NOT EXISTS idx_settings_team_id ON settings(team_id);
CREATE INDEX IF NOT EXISTS idx_settings_board_id ON settings(board_id);
