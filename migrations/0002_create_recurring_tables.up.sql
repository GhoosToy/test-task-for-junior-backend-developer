CREATE TABLE IF NOT EXISTS task_templates (
                                              id BIGSERIAL PRIMARY KEY,
                                              title VARCHAR(255) NOT NULL,
    description TEXT,
    recurrence JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    );

CREATE TABLE IF NOT EXISTS task_instances (
                                              id BIGSERIAL PRIMARY KEY,
                                              template_id BIGINT NOT NULL REFERENCES task_templates(id) ON DELETE CASCADE,
    due_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'new',
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(template_id, due_date)
    );

CREATE INDEX idx_task_instances_template_id ON task_instances(template_id);
CREATE INDEX idx_task_instances_due_date ON task_instances(due_date);
CREATE INDEX idx_task_instances_status ON task_instances(status);