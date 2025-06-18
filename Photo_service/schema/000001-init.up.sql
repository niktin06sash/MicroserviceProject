CREATE TABLE usersid (
    userid TEXT PRIMARY KEY,
);

CREATE TABLE photos (
    photoid TEXT PRIMARY KEY,
    userid TEXT REFERENCES usersid(userid) ON DELETE CASCADE,
    url TEXT NOT NULL,
    size BIGINT NOT NULL,
    content_type TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);