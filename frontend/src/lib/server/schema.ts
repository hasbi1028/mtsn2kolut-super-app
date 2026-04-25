import { sqliteTable, text, integer, uniqueIndex, index } from 'drizzle-orm/sqlite-core';
import { sql } from 'drizzle-orm';

const now = sql`CURRENT_TIMESTAMP`;

export const employees = sqliteTable('employees', {
	id:               text('id').primaryKey(),
	nip:              text('nip').notNull().unique(),
	nama:             text('nama').notNull(),
	unit_kerja:       text('unit_kerja').notNull().default(''),
	pusaka_username:  text('pusaka_username').notNull(),
	pusaka_password:  text('pusaka_password').notNull(),
	is_active:        integer('is_active').notNull().default(1),
	created_at:       text('created_at').notNull().default(now),
	updated_at:       text('updated_at').notNull().default(now),
});

export const schedules = sqliteTable('schedules', {
	id:         text('id').primaryKey(),
	label:      text('label').notNull(),
	run_time:   text('run_time').notNull(),
	run_type:   text('run_type', { enum: ['morning', 'afternoon', 'checkin', 'checkout'] }).notNull(),
	is_enabled: integer('is_enabled').notNull().default(1),
	created_at: text('created_at').notNull().default(now),
	updated_at: text('updated_at').notNull().default(now),
});

export const jobs = sqliteTable('jobs', {
	id:            text('id').primaryKey(),
	employee_id:   text('employee_id').notNull().references(() => employees.id),
	run_type:      text('run_type', { enum: ['morning', 'afternoon', 'checkin', 'checkout'] }).notNull(),
	status:        text('status', { enum: ['queued', 'running', 'success', 'failed'] }).notNull(),
	error_message: text('error_message').notNull().default(''),
	claimed_by:    text('claimed_by').notNull().default(''),
	claimed_at:    text('claimed_at'),
	attempts:      integer('attempts').notNull().default(0),
	max_attempts:  integer('max_attempts').notNull().default(3),
	next_retry_at: text('next_retry_at'),
	created_at:    text('created_at').notNull().default(now),
	updated_at:    text('updated_at').notNull().default(now),
}, (t) => [
	index('idx_jobs_status_created_at').on(t.status, t.created_at),
	index('idx_jobs_employee_run_status').on(t.employee_id, t.run_type, t.status),
	index('idx_jobs_next_retry_at').on(t.next_retry_at),
]);

export const attendanceRecords = sqliteTable('attendance_records', {
	id:            text('id').primaryKey(),
	employee_id:   text('employee_id').notNull().references(() => employees.id),
	tanggal:       text('tanggal').notNull(),
	jam_masuk:     text('jam_masuk').notNull().default(''),
	jam_pulang:    text('jam_pulang').notNull().default(''),
	source_job_id: text('source_job_id').notNull().references(() => jobs.id),
	created_at:    text('created_at').notNull().default(now),
	updated_at:    text('updated_at').notNull().default(now),
}, (t) => [
	uniqueIndex('uq_attendance_employee_tanggal').on(t.employee_id, t.tanggal),
]);

export const appSettings = sqliteTable('app_settings', {
	key:        text('key').primaryKey(),
	value:      text('value').notNull().default(''),
	updated_at: text('updated_at').notNull().default(now),
});

// Inferred row types for use across the app
export type Employee        = typeof employees.$inferSelect;
export type NewEmployee     = typeof employees.$inferInsert;
export type Job             = typeof jobs.$inferSelect;
export type NewJob          = typeof jobs.$inferInsert;
export type Schedule        = typeof schedules.$inferSelect;
export type AttendanceRecord= typeof attendanceRecords.$inferSelect;
export type AppSetting      = typeof appSettings.$inferSelect;
export type RunType         = 'morning' | 'afternoon' | 'checkin' | 'checkout';
export type JobStatus       = 'queued' | 'running' | 'success' | 'failed';
