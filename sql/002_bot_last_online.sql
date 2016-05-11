ALTER TABLE bot
ADD COLUMN last_online_datetime TIMESTAMP WITH TIME ZONE DEFAULT (now()) NOT NULL;
    
CREATE TABLE __bot_last_online_datetime (
  bot_id INTEGER NOT NULL
, last_online_datetime TIMESTAMP WITH TIME ZONE DEFAULT (now()) NOT NULL
);

INSERT INTO __bot_last_online_datetime (
  bot_id
, last_online_datetime
)
SELECT 
  gb.bot_id
, MAX(gm.created_datetime)
FROM move gm
JOIN game_bot gb
  ON gm.game_bot_id = gb.id
GROUP BY gb.bot_id;

UPDATE bot
SET last_online_datetime = __bot_last_online_datetime.last_online_datetime
FROM __bot_last_online_datetime
WHERE bot.id = __bot_last_online_datetime.bot_id;

DROP TABLE __bot_last_online_datetime;