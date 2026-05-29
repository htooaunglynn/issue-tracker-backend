CREATE TABLE issue_activities (
    id         UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    issue_id   UUID            NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    actor_id   UUID            NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    action     activity_action NOT NULL,
    field      TEXT,
    old_value  TEXT,
    new_value  TEXT,
    created_at TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_issue_activities_issue_id ON issue_activities (issue_id);
CREATE INDEX idx_issue_activities_actor_id ON issue_activities (actor_id);

CREATE TABLE notifications (
    id          UUID              PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID              NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type        notification_type NOT NULL,
    title       TEXT              NOT NULL,
    message     TEXT              NOT NULL,
    entity_type TEXT              NOT NULL DEFAULT '',
    entity_id   UUID,
    is_read     BOOLEAN           NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ       NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ       NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_notifications_user_id ON notifications (user_id);
CREATE INDEX idx_notifications_is_read ON notifications (is_read) WHERE deleted_at IS NULL;
