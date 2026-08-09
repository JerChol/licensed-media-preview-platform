// Mirrors the Go `JobStatus` type - a string, but restricted to specific known values.
export type JobStatus = "queued" | "processing" | "done" | "failed";

// Mirrors the Go `Job` struct
export interface Job {
  id: string;
  source_url: string;
  resolved_url: string;
  resolved_source_id: string;
  status: JobStatus;
  progress: number;
  error_message?: string;
  created_at: string;
  updated_at: string;
}

//Mirrors the Go `Artifact` struct
export interface Artifact {
  id: string;
  job_id: string;
  artifact_type: string;
  storage_path: string;
  created_at: string;
}