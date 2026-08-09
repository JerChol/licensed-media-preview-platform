import { createJob, getJob, getJobArtifacts } from "./api.js";
const form = document.querySelector("#job-form");
const urlInput = document.querySelector("#url-input");
const statusDiv = document.querySelector("#status");
const resultDiv = document.querySelector("#result");
form.addEventListener("submit", async (event) => {
    event.preventDefault();
    const url = urlInput.value;
    statusDiv.textContent = "Submitting";
    resultDiv.textContent = "";
    try {
        const job = await createJob(url);
        statusDiv.textContent = `Job Created: ${job.id} (status: ${job.status})`;
        pollJobStatus(job.id);
    }
    catch (err) {
        statusDiv.textContent = `Error: ${err.message}`;
    }
});
async function pollJobStatus(jobId) {
    const job = await getJob(jobId);
    statusDiv.textContent = `Status: ${job.status}`;
    if (job.status === "done") {
        const artifacts = await getJobArtifacts(jobId);
        resultDiv.textContent = JSON.stringify(artifacts, null, 2);
        return;
    }
    // Still queued/processing - check again in a second
    setTimeout(() => pollJobStatus(jobId), 1000);
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
