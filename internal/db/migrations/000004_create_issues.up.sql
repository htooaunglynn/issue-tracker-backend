CREATE TABLE issues (
    id          UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID           NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    number      INTEGER        NOT NULL,
    title       TEXT           NOT NULL,
    description TEXT           NOT NULL DEFAULT '',
    type        issue_type     NOT NULL DEFAULT 'task',
    priority    issue_priority NOT NULL DEFAULT 'medium',
    status      issue_status   NOT NULL DEFAULT 'open',
    reporter_id UUID           NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    assignee_id UUID           REFERENCES users(id) ON DELETE SET NULL,
    due_date    TIMESTAMPTZ,
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

-- Composite unique: issue number is unique within a project
CREATE UNIQUE INDEX idx_issues_project_number
    ON issues (project_id, number)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_issues_project_id ON issues (project_id);
CREATE INDEX idx_issues_title ON issues (title);
CREATE INDEX idx_issues_priority ON issues (priority);
CREATE INDEX idx_issues_status ON issues (status);
CREATE INDEX idx_issues_reporter_id ON issues (reporter_id);
CREATE INDEX idx_issues_assignee_id ON issues (assignee_id);
CREATE INDEX idx_issues_deleted_at ON issues (deleted_at);
