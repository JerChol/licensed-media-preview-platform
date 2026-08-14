import "./style.css"; // Ignore any error on this line.
import type { Job, Artifact } from "./types.js";
import { createJob, getJob, getJobArtifacts, JobCreationError, API_BASE } from "./api.js";

const FRIENDLY_ERRORS: Record<string, string> = {
  unknown_source: "That URL isn't from a source we're allowed to preview yet.",
  ambiguous_source: "That URL matches more than one source - try a more specific link.",
  disallowed_file_type: "That file type isn't supported yet - PDFs only for now.",
  bad_input: "That doesn't look like a valid URL",
}

const form = document.querySelector<HTMLFormElement>("#job-form")!;
const urlInput = document.querySelector<HTMLInputElement>("#url-input")!;
const statusDiv = document.querySelector<HTMLDivElement>("#status")!;
const resultDiv = document.querySelector<HTMLDivElement>("#result")!;

form.addEventListener("submit", async (event) => {
  event.preventDefault();

  const url = urlInput.value;
  statusDiv.textContent = "Submitting...";
  resultDiv.textContent = "";

  try {
    const job = await createJob(url);
    statusDiv.textContent =  `Job Created: ${job.id} (status: ${job.status})`;
    pollJobStatus(job.id);
  }catch (err) {
    if (err instanceof JobCreationError) {
      statusDiv.textContent = FRIENDLY_ERRORS[err.code] ?? err.message;
    }else{
      statusDiv.textContent = `Error: ${(err as Error).message}`;
    }
  }
});

async function pollJobStatus(jobId: string) {
  const job = await getJob(jobId);
  statusDiv.textContent = `Status: ${job.status}`;

  if (job.status === "done"){
    const artifacts = await getJobArtifacts(jobId);
    await renderArtifacts(artifacts);
    return;
  }

  if (job.status === "failed") {
    resultDiv.textContent = "Job Failed.";
    return;
  }

  // Still queued/processing - check again in a second
  setTimeout(() => pollJobStatus(jobId), 1000);
}

async function renderArtifacts(artifacts: Artifact[]) {
  resultDiv.innerHTML = "";

  for (const artifact of artifacts) {
    const publicPath = artifact.storage_path.replace(/^data\//, "");
    const fileUrl = `${API_BASE}/artifacts/${artifact.id}`;

    if (artifact.artifact_type === "text_snippet"){
      const response = await fetch(fileUrl);
      if (!response.ok){
        console.error(`Failed to fetch ${fileUrl}: ${response.status}`);
      }
      const text = await response.text();

      const pre = document.createElement("pre");
      pre.textContent = text;
      resultDiv.appendChild(pre);
    }

    if (artifact.artifact_type === "thumbnail"){
      const img = document.createElement("img");
      img.src = fileUrl;
      img.alt = "PDF thumbnail";
      resultDiv.appendChild(img);
    }
  }
}
