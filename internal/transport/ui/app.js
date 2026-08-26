const state = {
  status: document.querySelector("#status-text"),
  dot: document.querySelector("#status-dot"),
  instances: document.querySelector("#instances"),
  instanceCount: document.querySelector("#instance-count"),
  jobCount: document.querySelector("#job-count"),
  destinationCount: document.querySelector("#destination-count"),
};

async function json(path) {
  const response = await fetch(path, { headers: { Accept: "application/json" } });
  if (!response.ok) throw new Error(`${path}: ${response.status}`);
  return response.json();
}

function renderInstances(items) {
  state.instances.replaceChildren();
  if (!items.length) {
    const row = state.instances.insertRow();
    const cell = row.insertCell();
    cell.colSpan = 4;
    cell.className = "empty";
    cell.textContent = "No imaging instances have been ingested.";
    return;
  }
  for (const item of items.slice(0, 20)) {
    const row = state.instances.insertRow();
    for (const value of [item.uid, item.studyUID, item.metadata?.modality, item.status]) {
      row.insertCell().textContent = value || "-";
    }
  }
}

async function refresh() {
  try {
    const [health, instances, jobs, destinations] = await Promise.all([
      json("/readyz"), json("/api/v1/instances?limit=20"), json("/api/v1/jobs"), json("/api/v1/destinations"),
    ]);
    state.status.textContent = health.status === "ready" ? "Gateway ready" : "Gateway starting";
    state.dot.className = "ready";
    state.instanceCount.textContent = instances.items?.length ?? 0;
    state.jobCount.textContent = jobs.items?.length ?? 0;
    state.destinationCount.textContent = destinations.items?.length ?? 0;
    renderInstances(instances.items || []);
  } catch (error) {
    state.status.textContent = "Gateway unavailable";
    state.dot.className = "failed";
    state.instances.innerHTML = `<tr><td colspan="4" class="empty">${error.message}</td></tr>`;
  }
}

document.querySelector("#refresh").addEventListener("click", refresh);
refresh();
