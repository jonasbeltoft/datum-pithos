-- Create table: units
CREATE TABLE IF NOT EXISTS units (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    CONSTRAINT unique_name UNIQUE (name)
);

-- Create table: collections
CREATE TABLE IF NOT EXISTS collections (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    CONSTRAINT unique_name UNIQUE (name)
);

-- Create table: samples
CREATE TABLE IF NOT EXISTS samples (
    id INTEGER PRIMARY KEY,
    collection_id INTEGER NOT NULL,
    created_at INTEGER NOT NULL, -- Stores UNIX time at INSERT
    note TEXT,
    CONSTRAINT fk_collection FOREIGN KEY (collection_id) REFERENCES collections (id) ON DELETE CASCADE
);

-- Create table: sample_attributes
CREATE TABLE IF NOT EXISTS sample_attributes (
    id INTEGER PRIMARY KEY,
    collection_id INTEGER NOT NULL,
    unit_id INTEGER, -- Nullable
    name TEXT NOT NULL,
    CONSTRAINT unique_name UNIQUE (collection_id, name),
    CONSTRAINT fk_collection FOREIGN KEY (collection_id) REFERENCES collections (id) ON DELETE CASCADE,
    CONSTRAINT fk_unit FOREIGN KEY (unit_id) REFERENCES units (id)
);

-- Create table: sample_attribute_values
CREATE TABLE IF NOT EXISTS sample_attribute_values (
    sample_id INTEGER NOT NULL,
    attribute_id INTEGER NOT NULL,
    value TEXT,
    PRIMARY KEY (sample_id, attribute_id),
    CONSTRAINT fk_sample FOREIGN KEY (sample_id) REFERENCES samples (id) ON DELETE CASCADE,
    CONSTRAINT fk_attribute FOREIGN KEY (attribute_id) REFERENCES sample_attributes (id) ON DELETE CASCADE
);

-- Create table: roles
CREATE TABLE IF NOT EXISTS roles (id INTEGER PRIMARY KEY, name TEXT NOT NULL);

-- Create table: users
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    username TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role_id INTEGER, -- Nullable, references roles table
    display_name TEXT,
    access_token TEXT,
    token_expiry_date INTEGER, -- UNIX time in seconds
    CONSTRAINT unique_username UNIQUE (username),
    CONSTRAINT fk_role FOREIGN KEY (role_id) REFERENCES roles (id)
);

-- Create table: logs
CREATE TABLE IF NOT EXISTS logs (
    id INTEGER PRIMARY KEY,
    created_at INTEGER NOT NULL, -- Stores UNIX time
    instance_user INTEGER, -- References users
    crud_action TEXT NOT NULL, -- Action performed
    request_url TEXT,
    request_body TEXT,
    response_code INTEGER
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