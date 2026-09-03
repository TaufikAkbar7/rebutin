CREATE TYPE simulation_run_status AS ENUM ('running', 'ended');
CREATE TYPE session_status AS ENUM ('queued', 'active', 'expired', 'completed');
CREATE TYPE reservation_status AS ENUM ('active', 'released', 'completed');
CREATE TYPE payment_status AS ENUM ('pending', 'success', 'failed', 'refunded');