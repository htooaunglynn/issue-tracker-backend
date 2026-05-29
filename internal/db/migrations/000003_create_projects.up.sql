CREATE TABLE projects (
    id          UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    key         VARCHAR(10)    NOT NULL,
    name        TEXT           NOT NULL,
    description TEXT           NOT NULL DEFAULT '',
    status      project_status NOT NULL DEFAULT 'active',
    owner_id    UUID           NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_projects_key ON projects (key) WHERE deleted_at IS NULL;
CREATE INDEX idx_projects_name ON projects (name);
CREATE INDEX idx_projects_owner_id ON projects (owner_id);
CREATE INDEX idx_projects_deleted_at ON projects (deleted_at);

CREATE TABLE project_members (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID         NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id    UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role       project_role NOT NULL DEFAULT 'reporter',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uq_project_member UNIQUE (project_id, user_id)
);

CREATE INDEX idx_project_members_project_id ON project_members (project_id);
CREATE INDEX idx_project_members_user_id ON project_members (user_id);
