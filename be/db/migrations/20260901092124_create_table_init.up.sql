CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE simulation_runs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    total_tickets INT NOT NULL,
    bot_count INT NOT NULL DEFAULT 0,
    bot_throttle_seconds INT NOT NULL DEFAULT 0,
    max_concurrent INT NOT NULL,
    status simulation_run_status NOT NULL DEFAULT 'running',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE ticket_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    run_id UUID NOT NULL REFERENCES simulation_runs(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    quota INT NOT NULL,
    price NUMERIC(12, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE participants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    run_id UUID NOT NULL REFERENCES simulation_runs(id) ON DELETE CASCADE,
    is_bot BOOLEAN NOT NULL DEFAULT FALSE,
    identifier VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    participant_id UUID NOT NULL REFERENCES participants(id) ON DELETE CASCADE,
    run_id UUID NOT NULL REFERENCES simulation_runs(id) ON DELETE CASCADE,
    status session_status NOT NULL DEFAULT 'queued',
    queued_at TIMESTAMP WITH TIME ZONE,
    ttl_started_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE reservations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    simulation_id UUID NOT NULL REFERENCES simulation_runs(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES ticket_categories(id) ON DELETE RESTRICT,
    status reservation_status NOT NULL DEFAULT 'active',
    reserved_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    released_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reservation_id UUID NOT NULL REFERENCES reservations(id) ON DELETE CASCADE,
    simulation_id UUID NOT NULL REFERENCES simulation_runs(id) ON DELETE CASCADE,
    status payment_status NOT NULL DEFAULT 'pending',
    requested_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_ticket_categories_run_id ON ticket_categories(run_id);
CREATE INDEX idx_participants_run_id ON participants(run_id);
CREATE INDEX idx_sessions_participant_id ON sessions(participant_id);
CREATE INDEX idx_sessions_run_id ON sessions(run_id);
CREATE INDEX idx_sessions_status ON sessions(status);
CREATE INDEX idx_reservations_session_id ON reservations(session_id);
CREATE INDEX idx_reservations_simulation_id ON reservations(simulation_id);
CREATE INDEX idx_reservations_category_id ON reservations(category_id);
CREATE INDEX idx_reservations_status ON reservations(status);
CREATE INDEX idx_payments_reservation_id ON payments(reservation_id);
CREATE INDEX idx_payments_simulation_id ON payments(simulation_id);
CREATE INDEX idx_payments_status ON payments(status);