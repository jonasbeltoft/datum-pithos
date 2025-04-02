-- Create table: units
CREATE TABLE IF NOT EXISTS units (id INTEGER PRIMARY KEY, name TEXT UNIQUE NOT NULL);

-- Create table: collections
CREATE TABLE IF NOT EXISTS collections (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT
);

-- Create table: samples
CREATE TABLE IF NOT EXISTS samples (
    id INTEGER PRIMARY KEY,
    collection_id INTEGER NOT NULL,
    created_at INTEGER NOT NULL, -- Stores UNIX time at INSERT
    note TEXT,
    FOREIGN KEY (collection_id) REFERENCES collections (id)
);

-- Create table: sample_attributes
CREATE TABLE IF NOT EXISTS sample_attributes (
    id INTEGER PRIMARY KEY,
    collection_id INTEGER NOT NULL,
    unit_id INTEGER, -- Nullable
    name TEXT NOT NULL,
    CONSTRAINT unique_name UNIQUE (collection_id, name),
    FOREIGN KEY (collection_id) REFERENCES collections (id),
    FOREIGN KEY (unit_id) REFERENCES units (id)
);

-- Create table: sample_attribute_values
CREATE TABLE IF NOT EXISTS sample_attribute_values (
    sample_id INTEGER NOT NULL,
    attribute_id INTEGER NOT NULL,
    value TEXT,
    PRIMARY KEY (sample_id, attribute_id),
    FOREIGN KEY (sample_id) REFERENCES samples (id),
    FOREIGN KEY (attribute_id) REFERENCES sample_attributes (id)
);

-- Create table: roles
CREATE TABLE IF NOT EXISTS roles (id INTEGER PRIMARY KEY, name TEXT NOT NULL);

-- Create table: users
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role_id INTEGER, -- Nullable, references roles table
    display_name TEXT,
    access_token TEXT,
    token_expiry_date INTEGER, -- UNIX time in seconds
    FOREIGN KEY (role_id) REFERENCES roles (id)
);

-- Create table: logs
CREATE TABLE IF NOT EXISTS logs (
    id INTEGER PRIMARY KEY,
    created_at INTEGER NOT NULL, -- Stores UNIX time
    instance_user INTEGER NOT NULL, -- References users
    crud_action TEXT NOT NULL, -- Action performed (e.g., CREATE, READ, UPDATE, DELETE)
    value TEXT,
    FOREIGN KEY (instance_user) REFERENCES users (id)
);

-- Initialize roles table only if empty
INSERT INTO
    roles
SELECT
    *
FROM
    (
        VALUES
            (1, 'admin'),
            (2, 'Lab Technician')
    ) source_data
WHERE
    NOT EXISTS (
        SELECT
            NULL
        FROM
            roles
    );