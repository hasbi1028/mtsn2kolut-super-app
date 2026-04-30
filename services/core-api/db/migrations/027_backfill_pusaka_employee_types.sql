UPDATE employees e
SET employment_type = 'pns',
    updated_at = NOW()
WHERE EXISTS (
    SELECT 1
    FROM pusaka_accounts pa
    WHERE pa.employee_id = e.id
)
  AND e.employment_type NOT IN ('pns', 'pppk');
