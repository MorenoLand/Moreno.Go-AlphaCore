package database

const commonSchema = `
CREATE TABLE IF NOT EXISTS schema_meta (
    key TEXT PRIMARY KEY NOT NULL,
    value TEXT NOT NULL
);
INSERT OR IGNORE INTO schema_meta (key, value) VALUES ('engine', 'sqlite');
INSERT OR IGNORE INTO schema_meta (key, value) VALUES ('schema_version', '1');
`

const authSchema = commonSchema + `
CREATE TABLE IF NOT EXISTS accounts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    ip TEXT NOT NULL DEFAULT '',
    gmlevel INTEGER NOT NULL DEFAULT 1,
    salt TEXT NOT NULL,
    verifier TEXT NOT NULL,
    sessionkey TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS realmlist (
    realm_id INTEGER PRIMARY KEY AUTOINCREMENT,
    realm_name TEXT NOT NULL DEFAULT '',
    proxy_address TEXT NOT NULL DEFAULT '0.0.0.0',
    proxy_port INTEGER NOT NULL DEFAULT 9090,
    realm_address TEXT NOT NULL DEFAULT '0.0.0.0',
    realm_port INTEGER NOT NULL DEFAULT 9100,
    online_player_count INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS applied_updates (
    id TEXT PRIMARY KEY NOT NULL DEFAULT '000000000'
);
INSERT OR IGNORE INTO realmlist (realm_id, realm_name, proxy_address, proxy_port, realm_address, realm_port, online_player_count)
VALUES (1, 'Moreno.AlphaCore', '0.0.0.0', 9090, '0.0.0.0', 9100, 0);
`

const realmSchema = commonSchema + `
CREATE TABLE IF NOT EXISTS characters (
    guid INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id INTEGER NOT NULL DEFAULT 0,
    realm_id INTEGER NOT NULL DEFAULT 1,
    name TEXT NOT NULL DEFAULT '',
    race INTEGER NOT NULL DEFAULT 0,
    "class" INTEGER NOT NULL DEFAULT 0,
    gender INTEGER NOT NULL DEFAULT 0,
    level INTEGER NOT NULL DEFAULT 0,
    xp INTEGER NOT NULL DEFAULT 0,
    money INTEGER NOT NULL DEFAULT 0,
    skin INTEGER NOT NULL DEFAULT 0,
    face INTEGER NOT NULL DEFAULT 0,
    hairstyle INTEGER NOT NULL DEFAULT 0,
    haircolour INTEGER NOT NULL DEFAULT 0,
    facialhair INTEGER NOT NULL DEFAULT 0,
    bankslots INTEGER NOT NULL DEFAULT 0,
    talentpoints INTEGER NOT NULL DEFAULT 0,
    skillpoints INTEGER NOT NULL DEFAULT 0,
    position_x REAL NOT NULL DEFAULT 0,
    position_y REAL NOT NULL DEFAULT 0,
    position_z REAL NOT NULL DEFAULT 0,
    map INTEGER NOT NULL DEFAULT 0,
    orientation REAL NOT NULL DEFAULT 0,
    taximask TEXT,
    explored_areas TEXT,
    online INTEGER NOT NULL DEFAULT 0,
    totaltime INTEGER NOT NULL DEFAULT 0,
    leveltime INTEGER NOT NULL DEFAULT 0,
    extra_flags INTEGER NOT NULL DEFAULT 0,
    zone INTEGER NOT NULL DEFAULT 0,
    taxi_path TEXT,
    drunk INTEGER NOT NULL DEFAULT 0,
    health INTEGER NOT NULL DEFAULT 0,
    power1 INTEGER NOT NULL DEFAULT 0,
    power2 INTEGER NOT NULL DEFAULT 0,
    power3 INTEGER NOT NULL DEFAULT 0,
    power4 INTEGER NOT NULL DEFAULT 0,
    power5 INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS characters_name_idx ON characters (name);
CREATE INDEX IF NOT EXISTS characters_online_idx ON characters (online);
CREATE TABLE IF NOT EXISTS character_addons_settings (
    guid INTEGER PRIMARY KEY NOT NULL,
    flags INTEGER NOT NULL DEFAULT 0,
    settings TEXT,
    updated_at INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (guid) REFERENCES characters (guid) ON DELETE CASCADE ON UPDATE CASCADE
);
`

const worldSchema = commonSchema + `
CREATE TABLE IF NOT EXISTS applied_updates (
    id TEXT PRIMARY KEY NOT NULL DEFAULT '000000000'
);
`

const dbcSchema = commonSchema
