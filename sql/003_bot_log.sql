CREATE TABLE bot_log (
  id               SERIAL PRIMARY KEY NOT NULL
, bot_id           INTEGER REFERENCES bot (id) NOT NULL
, message          VARCHAR(1000)
, created_datetime TIMESTAMP WITH TIME ZONE DEFAULT (now()) NOT NULL
);

CREATE INDEX ON bot_log (bot_id);