-- Portable schema: runs unchanged on PostgreSQL and SQLite.
-- Timestamps are unix seconds (BIGINT) to avoid dialect differences.
-- Curriculum content (modules, lessons, challenges, ...) is versioned in
-- internal/content/data and loaded at startup; these tables hold learner state.

CREATE TABLE users (
    id            TEXT PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL DEFAULT 'learner',
    created_at    BIGINT NOT NULL
);

CREATE TABLE profiles (
    user_id          TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    display_name     TEXT NOT NULL DEFAULT '',
    bio              TEXT NOT NULL DEFAULT '',
    learning_path    TEXT NOT NULL DEFAULT 'beginner',
    theme            TEXT NOT NULL DEFAULT 'system',
    daily_lessons    INTEGER NOT NULL DEFAULT 1,
    daily_challenges INTEGER NOT NULL DEFAULT 2,
    daily_quizzes    INTEGER NOT NULL DEFAULT 1,
    gamification     INTEGER NOT NULL DEFAULT 1,
    timezone         TEXT NOT NULL DEFAULT 'UTC',
    updated_at       BIGINT NOT NULL
);

CREATE TABLE sessions (
    token_hash TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at BIGINT NOT NULL,
    expires_at BIGINT NOT NULL
);
CREATE INDEX idx_sessions_user ON sessions(user_id);

CREATE TABLE lesson_progress (
    user_id         TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    lesson_id       TEXT NOT NULL,
    exercise_passed INTEGER NOT NULL DEFAULT 0,
    quiz_passed     INTEGER NOT NULL DEFAULT 0,
    completed_at    BIGINT,
    updated_at      BIGINT NOT NULL,
    PRIMARY KEY (user_id, lesson_id)
);

CREATE TABLE challenge_submissions (
    id           TEXT PRIMARY KEY,
    user_id      TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    challenge_id TEXT NOT NULL,
    code         TEXT NOT NULL,
    passed       INTEGER NOT NULL,
    tests_passed INTEGER NOT NULL,
    tests_total  INTEGER NOT NULL,
    created_at   BIGINT NOT NULL
);
CREATE INDEX idx_submissions_user_challenge ON challenge_submissions(user_id, challenge_id, created_at);

CREATE TABLE quiz_attempts (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    lesson_id  TEXT NOT NULL,
    score      INTEGER NOT NULL,
    total      INTEGER NOT NULL,
    passed     INTEGER NOT NULL,
    created_at BIGINT NOT NULL
);
CREATE INDEX idx_quiz_attempts_user_lesson ON quiz_attempts(user_id, lesson_id, created_at);

CREATE TABLE project_progress (
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id TEXT NOT NULL,
    task_id    TEXT NOT NULL,
    done_at    BIGINT NOT NULL,
    PRIMARY KEY (user_id, project_id, task_id)
);

CREATE TABLE user_achievements (
    user_id        TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id TEXT NOT NULL,
    unlocked_at    BIGINT NOT NULL,
    PRIMARY KEY (user_id, achievement_id)
);

CREATE TABLE learning_sessions (
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    day     TEXT NOT NULL,
    seconds INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, day)
);
