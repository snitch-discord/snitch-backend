CREATE TABLE groups (
    group_id INTEGER PRIMARY KEY,
    group_name TEXT NOT NULL
);

CREATE TABLE servers (
    server_id INTEGER NOT NULL,
    output_channel INTEGER NOT NULL,
    group_id INTEGER NOT NULL REFERENCES groups(group_id),
    permission_level INTEGER NOT NULL,
    PRIMARY KEY (server_id, group_id)
);
