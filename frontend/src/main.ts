import type { Job, Artifact } from "./types.js";
import { createJob, getJob, getJobArtifacts } from "./api.js";

const form = document.querySelector<HTMLFormElement>("#job-form")!;
const urlInput = document.querySelector<HTMLInputElement>("#url-input")!;
const statusDiv = document.querySelector<HTMLDivElement>("#status")!;
const resultDiv = document.querySelector<HTMLDivElement>("#result")!;

form.addEventListener("submit", async (event) => {
  event.preventDefault();

  const url = urlInput.value;
  statusDiv.textContent = "Submitting";
  resultDiv.textContent = "";

  try {
    const job = await createJob(url);
    statusDiv.textContent =  `Job Created: ${job.id} (status: ${job.status})`;
    pollJobStatus(job.id);
  }catch (err) {
    statusDiv.textContent = `Error: ${(err as Error).message}`;
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
    if (artifact.artifact_type === "text_snippet"){
      const publicPath = artifact.storage_path.replace(/^data\//, "");
      const response = await fetch(`http://localhost:8080/${publicPath}`);
      const text = await response.text();

      const pre = document.createElement("pre");
      pre.textContent = text;
      resultDiv.appendChild(pre);
    }
  }
}

/* ________________TESTING CODE__________________

async function run() {
  const job =  await createJob("https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf");
  console.log("Created Job: ", job);

  //Wait a moment for the worker to process it.
  await new Promise((resolve) => setTimeout(resolve, 2000));

  const updatedJob = await getJob(job.id);
  console.log("Job status: ", updatedJob.status);

  const artifacts = await getJobArtifacts(job.id);
  console.log("Artifacts:", artifacts);
}

run()

*/