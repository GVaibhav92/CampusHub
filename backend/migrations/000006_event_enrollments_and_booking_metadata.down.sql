DROP TABLE IF EXISTS event_enrollments;
DROP TYPE IF EXISTS enrollment_status;
ALTER TABLE bookings DROP COLUMN IF EXISTS course_code;
ALTER TABLE bookings DROP COLUMN IF EXISTS title;
