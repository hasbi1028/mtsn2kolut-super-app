export type RunType = 'morning' | 'afternoon' | 'checkin' | 'checkout';

export interface ClaimedJob {
  id: string;
  employee_id: string;
  run_type: RunType;
  attempts: number;
  max_attempts: number;
  pusaka_username: string;
  pusaka_password: string;
}

export interface AttendanceRecord {
  tanggal: string;
  jam_masuk: string;
  jam_pulang: string;
}

export interface GeoCoords {
  latitude: number;
  longitude: number;
}

export interface RuntimeConfig {
  maxConcurrent: number;
  headless: boolean;
}

export interface ConsumerState {
  id: number;
  consumerId: string;
  stopRequested: boolean;
  activeJobIds: Set<string>;
  activeJobControllers: Map<string, AbortController>;
}
