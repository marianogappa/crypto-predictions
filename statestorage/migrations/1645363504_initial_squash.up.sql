-- DDL generated by Postico 1.5.6
-- Not all database features are supported. Do not use for backup.

-- Table Definition ----------------------------------------------

CREATE TABLE predictions (
    uuid uuid NOT NULL PRIMARY KEY,
    blob jsonb NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT now(),
    posted_at timestamp without time zone NOT NULL
);

-- Indices -------------------------------------------------------

CREATE INDEX predictions_blob_idx ON predictions USING GIN (blob jsonb_ops);
CREATE INDEX predictions_created_at_idx ON predictions(created_at timestamp_ops);
CREATE INDEX predictions_posted_at_idx ON predictions(posted_at timestamp_ops);
