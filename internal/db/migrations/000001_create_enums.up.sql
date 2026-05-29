-- Global user roles
CREATE TYPE global_role AS ENUM ('admin', 'manager', 'developer', 'reporter');

-- Project-level roles
CREATE TYPE project_role AS ENUM ('admin', 'manager', 'developer', 'reporter');

-- Project statuses
CREATE TYPE project_status AS ENUM ('active', 'archived');

-- Issue types
CREATE TYPE issue_type AS ENUM ('bug', 'feature', 'task', 'improvement');

-- Issue priorities
CREATE TYPE issue_priority AS ENUM ('low', 'medium', 'high', 'critical');

-- Issue statuses
CREATE TYPE issue_status AS ENUM ('open', 'in_progress', 'resolved', 'closed', 'reopened');

-- Activity actions
CREATE TYPE activity_action AS ENUM (
    'created', 'updated', 'status_changed', 'assigned', 'commented'
);

-- Notification types
CREATE TYPE notification_type AS ENUM (
    'assignment', 'comment', 'status_change', 'mention'
);
