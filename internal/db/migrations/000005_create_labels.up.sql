CREATE TABLE labels (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name       TEXT        NOT NULL,
    color      VARCHAR(7)  NOT NULL DEFAULT '#6B7280',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Label name must be unique within a project
CREATE UNIQUE INDEX idx_labels_project_name
    ON labels (project_id, name)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_labels_project_id ON labels (project_id);
CREATE INDEX idx_labels_deleted_at ON labels (deleted_at);

CREATE TABLE issue_labels (
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    label_id UUID NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
    PRIMARY KEY (issue_id, label_id)
);

CREATE INDEX idx_issue_labels_label_id ON issue_labels (label_id);
