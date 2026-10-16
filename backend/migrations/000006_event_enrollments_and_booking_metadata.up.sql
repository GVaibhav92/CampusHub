-- Add title and course_code to bookings
ALTER TABLE bookings
  ADD COLUMN title VARCHAR(150),
  ADD COLUMN course_code VARCHAR(30);

-- Backfill any existing records
UPDATE bookings SET title = 'Academic Session' WHERE title IS NULL;

-- Enforce NOT NULL without a permanent default on title
ALTER TABLE bookings
  ALTER COLUMN title SET NOT NULL;

-- Create enrollment status enum and table
CREATE TYPE enrollment_status AS ENUM ('ENROLLED', 'CANCELLED_BY_STUDENT', 'CANCELLED_BY_FACULTY');

CREATE TABLE event_enrollments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    booking_id UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status enrollment_status NOT NULL DEFAULT 'ENROLLED',
    enrolled_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT unique_student_booking UNIQUE (tenant_id, booking_id, student_id)
);

CREATE INDEX idx_enrollments_student ON event_enrollments(tenant_id, student_id, status);
