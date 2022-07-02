ALTER TABLE "prediction_interactions"
  ADD COLUMN "status" text NOT NULL DEFAULT 'POSTED',
  ADD COLUMN "error" text;
