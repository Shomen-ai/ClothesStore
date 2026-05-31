-- Pending email verification for sign-up.
-- One row per email-in-progress; UPSERT on email so re-requests rotate the code
-- rather than piling up rows.
CREATE TABLE pending_registrations (
  email         VARCHAR(255) PRIMARY KEY,
  code          VARCHAR(6)   NOT NULL,
  password_hash TEXT         NOT NULL,
  name          VARCHAR(255) NOT NULL,
  phone         VARCHAR(64)  NOT NULL DEFAULT '',
  attempts      SMALLINT     NOT NULL DEFAULT 0,
  expires_at    TIMESTAMP    NOT NULL,
  resend_after  TIMESTAMP    NOT NULL DEFAULT NOW(),
  created_at    TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_pending_registrations_expires ON pending_registrations(expires_at);
