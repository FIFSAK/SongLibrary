CREATE TABLE songs
(
    id           SERIAL PRIMARY KEY,
    group_name   TEXT NOT NULL,
    song_name    TEXT NOT NULL,
    release_date DATE NOT NULL,
    text         TEXT NOT NULL,
    link         TEXT NOT NULL,
    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX unique_song ON songs (group_name, song_name);
