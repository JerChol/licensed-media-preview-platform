const API_BASE = "http://localhost:8080";
// Send a URL to the backend to create a new preview job.
export async function createJob(url) {
    const response = await fetch(`${API_BASE}/jobs`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ url }),
    });
    if (!response.ok) {
        const errorBody = await response.json();
        throw new Error(errorBody.message ?? "Failed to create job");
    }
    return response.json();
}
// Fetches the current status of a job by ID.
export async function getJob(jobId) {
    const response = await fetch(`${API_BASE}/jobs/${jobId}`);
    if (!response.ok) {
        throw new Error("Failed to fetch job");
    }
    return response.json();
}
// Fetches the derived artifacts for a job.
export async function getJobArtifacts(jobId) {
    const response = await fetch(`${API_BASE}/jobs/${jobId}/artifacts`);
    if (!response.ok) {
        throw new Error("Failed to fetch artifacts");
    }
    return response.json();
}
