DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS events CASCADE;

CREATE EXTENSION IF NOT EXISTS CITEXT;

CREATE TABLE users
(
    id               BIGSERIAL PRIMARY KEY NOT NULL UNIQUE,
    mail_user_id     VARCHAR(128)          NOT NULL UNIQUE CHECK ( mail_user_id <> '' ),

    telegram_user_id BIGINT                NOT NULL UNIQUE

--     nickname         VARCHAR(256)          NOT NULL,
--     fullname         VARCHAR(512)          NOT NULL,
);

CREATE TABLE events
(
    id          BIGSERIAL PRIMARY KEY NOT NULL UNIQUE,
    user_id     BIGINT                NOT NULL,
    name        CITEXT,
    description CITEXT,

    FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
);
