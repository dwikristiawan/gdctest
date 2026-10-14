-- 000003_create_task_logs_and_optimal_indexes.up.sql

-- 1. Audit Table task_logs
CREATE TABLE IF NOT EXISTS task_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    previous_assignee_id UUID REFERENCES users(id) ON DELETE SET NULL,
    new_assignee_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    changed_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    action VARCHAR(50) NOT NULL DEFAULT 'ASSIGN',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_task_logs_task_created ON task_logs(task_id, created_at DESC);

-- 2. Partial Index pada task_collaborators untuk role Assignee (role_type = 0)
CREATE INDEX IF NOT EXISTS idx_task_collaborators_assignee ON task_collaborators(task_id, user_id) WHERE role_type = 0;

-- 3. Composite & Functional Query Indexes pada tasks
CREATE INDEX IF NOT EXISTS idx_tasks_team_status_created ON tasks(team_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tasks_title_lower ON tasks(LOWER(title));

-- 4. Reverse Partial Index pada team_members
CREATE INDEX IF NOT EXISTS idx_team_members_user_active ON team_members(user_id) WHERE deleted_at IS NULL;
