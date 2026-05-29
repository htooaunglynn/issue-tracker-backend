CREATE TABLE comments (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    issue_id   UUID        NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    author_id  UUID        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    body       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_comments_issue_id ON comments (issue_id);
CREATE INDEX idx_comments_author_id ON comments (author_id);
CREATE INDEX idx_comments_deleted_at ON comments (deleted_at);

-- comment_mentions: tracks @mentions inside comments
CREATE TABLE comment_mentions (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    comment_id  UUID        NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_comment_mention UNIQUE (comment_id, user_id)
);

CREATE INDEX idx_comment_mentions_user_id ON comment_mentions (user_id);
