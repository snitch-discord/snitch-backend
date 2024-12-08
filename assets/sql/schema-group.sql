CREATE TABLE users (
    user_id INTEGER PRIMARY KEY
) STRICT;

CREATE TABLE servers (
    server_id INTEGER PRIMARY KEY
) STRICT;

CREATE TABLE reports (
    report_id INTEGER PRIMARY KEY,
    report_text TEXT NOT NULL,
    reporter_id INTEGER NOT NULL REFERENCES users(users_id),
    reported_user_id INTEGER NOT NULL REFERENCES users(users_id),
    origin_server_id INTEGER NOT NULL REFERENCES servers(server_id)
) STRICT;
