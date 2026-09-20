-- 000003_quote_number_unique.sql
-- Guarantee quote numbers are unique per user (format QT-0001, scoped to each
-- business account).
--
-- Non-destructive: nothing is dropped and no existing unique number is changed.
-- Rows that would violate the new constraint (duplicates that the previous
-- COUNT(*)-based generator could hand out under concurrency, or numbers reused
-- after a deletion) are renumbered to the lowest free number for their user.
-- The earliest-created quote always keeps its number.

-- Keep the deduplication snapshot stable until the constraint is in place.
-- Migrations run in one transaction, so concurrent quote creation waits rather
-- than inserting a new duplicate between the cleanup and ALTER TABLE below.
LOCK TABLE quotes IN ACCESS EXCLUSIVE MODE;

-- 1. Renumber duplicate (user_id, quote_number) pairs.
--    Keep the first row (by created_at, then id) and give every later duplicate
--    the smallest number not already used by that user.
DO $$
DECLARE
    dup       RECORD;
    candidate BIGINT;
BEGIN
    FOR dup IN
        SELECT id, user_id
        FROM (
            SELECT id, user_id, quote_number, created_at,
                   ROW_NUMBER() OVER (
                       PARTITION BY user_id, quote_number
                       ORDER BY created_at, id
                   ) AS rn
            FROM quotes
        ) ranked
        WHERE rn > 1
        ORDER BY user_id, created_at, id
    LOOP
        candidate := 1;
        WHILE EXISTS (
            SELECT 1 FROM quotes
            WHERE user_id = dup.user_id
              AND NULLIF(regexp_replace(quote_number, '[^0-9]', '', 'g'), '')::bigint = candidate
        ) LOOP
            candidate := candidate + 1;
        END LOOP;
        UPDATE quotes
        SET quote_number = 'QT-' || lpad(candidate::text, 4, '0')
        WHERE id = dup.id;
    END LOOP;
END $$;

-- 2. Enforce uniqueness going forward. The application now allocates numbers
--    under a per-user advisory lock, deriving the next number from the highest
--    existing one; this constraint is the database-level backstop.
ALTER TABLE quotes
    ADD CONSTRAINT uq_quotes_user_quote_number UNIQUE (user_id, quote_number);
