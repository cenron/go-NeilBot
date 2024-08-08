CREATE TABLE booty_image (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL,
    mime_type VARCHAR(255) NOT NULL,
    hash VARCHAR(255),
    likes INTEGER NOT NULL DEFAULT 0,
    dislikes INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE booty_message (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    message_id VARCHAR(64) NOT NULL UNIQUE,
    channel_id VARCHAR(64) NOT NULL,
    guild_id VARCHAR(64) NOT NULL,
    image_id INTEGER,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,

    FOREIGN KEY (image_id) REFERENCES booty_image(id)
);