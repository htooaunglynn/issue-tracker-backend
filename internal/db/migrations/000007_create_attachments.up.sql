CREATE TABLE attachments (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    issue_id    UUID        NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    uploader_id UUID        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    file_name   TEXT        NOT NULL,
    file_path   TEXT        NOT NULL,
    mime_type   TEXT        NOT NULL,
    size_bytes  BIGINT      NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_attachments_issue_id ON attachments (issue_id);
CREATE INDEX idx_attachments_deleted_at ON attachments (deleted_at);
