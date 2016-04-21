DROP TABLE move;
DROP TABLE game_bot;
DROP TABLE game;
DROP TABLE bot;
DROP TABLE game_type;
DROP TABLE merknera_user;
 

CREATE TABLE merknera_user (
  id               SERIAL PRIMARY KEY
, username         VARCHAR(250) UNIQUE
, token            CHAR(50) UNIQUE
, created_datetime TIMESTAMP WITH TIME ZONE DEFAULT (now())
);

CREATE TABLE game_type (
  id               SERIAL PRIMARY KEY
, mnemonic         VARCHAR(50) UNIQUE
, name             VARCHAR(250)
, created_datetime TIMESTAMP WITH TIME ZONE DEFAULT (now())
);

CREATE TABLE bot (
  id                   SERIAL PRIMARY KEY
, name                 VARCHAR(250)
, version              VARCHAR(100)
, game_type_id         INTEGER REFERENCES game_type (id)
, user_id              INTEGER REFERENCES merknera_user (id)
, rpc_endpoint         VARCHAR(500)
, programming_language VARCHAR(250)
, website              VARCHAR(500) NULL
, status               VARCHAR(20) DEFAULT 'ONLINE'
, created_datetime     TIMESTAMP WITH TIME ZONE DEFAULT (now())
);

CREATE TABLE game (
  id               SERIAL PRIMARY KEY
, game_type_id     INTEGER REFERENCES game_type (id)
, status           VARCHAR(50) DEFAULT 'NOT STARTED'
, created_datetime TIMESTAMP WITH TIME ZONE DEFAULT (now())
);

CREATE TABLE game_bot (
  id               SERIAL PRIMARY KEY
, game_id          INTEGER
, bot_id           INTEGER
, play_sequence    INTEGER
, created_datetime TIMESTAMP WITH TIME ZONE DEFAULT (now())
, UNIQUE (game_id, bot_id)
);

CREATE TABLE move (
  id               SERIAL PRIMARY KEY
, game_bot_id      INTEGER REFERENCES game_bot (id)
, status           VARCHAR(20) DEFAULT 'NOT STARTED'
, created_datetime TIMESTAMP WITH TIME ZONE DEFAULT (now())
);



