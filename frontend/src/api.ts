import type { Job, Artifact } from "./types.js";

export const API_BASE = import.meta.env.VITE_API_BASE;

export const WORKER_URL = import.meta.env.VITE_WORKER_URL;

export class JobCreationError extends Error {
  code: string;
  constructor(code: string, message: string){
    super(message);
    this.code = code;
  }
}

// Send a URL to the backend to create a new preview job.
export async function createJob(url: string): Promise<Job> {
  const response = await fetch(`${API_BASE}/jobs`, {
    method: "POST",
    headers: {"Content-Type": "application/json"},
    body: JSON.stringify({ url }),
  });

  if (!response.ok) {
    const errorBody = await response.json();
    throw new JobCreationError(
      errorBody.error_code ?? "unknown_error",
      errorBody.message ?? "Failed to create job");
  }

  return response.json() as Promise<Job>;
}
// Fetches the current status of a job by ID.
export async function getJob(jobId: string): Promise<Job> {
  const response = await fetch(`${API_BASE}/jobs/${jobId}`);
  if (!response.ok){
    throw new Error("Failed to fetch job");
  }
  return response.json() as Promise<Job>;
}

// Fetches the derived artifacts for a job.
export async function getJobArtifacts(jobId: string): Promise<Artifact[]> {
  const response = await fetch(`${API_BASE}/jobs/${jobId}/artifacts`);
  if (!response.ok){
    throw new Error("Failed to fetch artifacts");
  }
  return response.json() as Promise<Artifact[]>
}

// Fires a request to wake the worker if it's asleep (Render free tier). Intentionally not awaited by callers - this just kicks off the wake-up process in the background; errors are ignored since this is best-effort.
export function wakeWorker(): void{
  fetch(WORKER_URL).catch(() => {
    // Ignore failures - if this doesn't work, the job will just wait in the queue a bit longer until something else wakes it.
  });
}