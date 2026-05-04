-- Migration: 063_cbt_package_duration_bounds

ALTER TABLE cbt_packages
  ADD CONSTRAINT cbt_packages_duration_minutes_bounds
  CHECK (duration_minutes BETWEEN 1 AND 360);
