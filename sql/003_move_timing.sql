ALTER TABLE move
ADD COLUMN start_datetime TIMESTAMP WITH TIME ZONE;

ALTER TABLE move
ADD COLUMN end_datetime TIMESTAMP WITH TIME ZONE;

DROP TABLE __move_start_end_time;

CREATE TABLE __move_start_end_time (
  move_id INTEGER NOT NULL
, game_bot_id INTEGER NOT NULL
, start_datetime TIMESTAMP WITH TIME ZONE
, end_datetime TIMESTAMP WITH TIME ZONE
);

INSERT INTO __move_start_end_time (
  move_id
, game_bot_id
, start_datetime
, end_datetime
)
SELECT
  m.id
, gb.id
, m.created_datetime
, LEAD(m.created_datetime) OVER (PARTITION BY gb.game_id ORDER BY m.created_datetime)
FROM move m
JOIN game_bot gb
  ON m.game_bot_id = gb.id;
   
UPDATE __move_start_end_time
SET end_datetime = start_datetime + t.calculated_move_duration
FROM (
  SELECT
    gb.id gb_id
  , AVG(COALESCE(mset.end_datetime, mset.start_datetime) - mset.start_datetime) OVER (PARTITION BY gb.id) calculated_move_duration
  FROM __move_start_end_time mset
  JOIN move m
    ON mset.move_id = m.id
  JOIN game_bot gb
    ON m.game_bot_id = gb.id
) t
WHERE game_bot_id = t.gb_id
AND end_datetime IS NULL;

UPDATE move
SET
  start_datetime = mset.start_datetime
, end_datetime = mset.end_datetime
FROM __move_start_end_time mset
WHERE id = mset.move_id;

DROP TABLE __move_start_end_time;