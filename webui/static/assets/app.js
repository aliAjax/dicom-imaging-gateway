const elements = {
  connection: document.querySelector("#connection"),
  destinationCount: document.querySelector("#destination-count"),
  destinations: document.querySelector("#destinations"),
  instanceCount: document.querySelector("#instance-count"),
  instanceMeta: document.querySelector("#instance-meta"),
  instances: document.querySelector("#instances"),
  jobCount: document.querySelector("#job-count"),
  jobs: document.querySelector("#jobs"),
  notice: document.querySelector("#notice"),
  refresh: document.querySelector("#refresh"),
  serviceStatus: document.querySelector("#service-status"),
};

const escapeHTML = (value) => String(value ?? "")
  .replaceAll("&", "&amp;")
  .replaceAll("<", "&lt;")
  .replaceAll(">", "&gt;")
  .replaceAll('"', "&quot;")
  .replaceAll("'", "&#039;");

async function getJSON(path) {
  const response = await fetch(path, { headers: { Accept: "application/json" } });
  if (!response.ok) throw new Error(`${path} returned HTTP ${response.status}`);
  return response.json();
}

function renderInstances(items) {
  elements.instanceCount.textContent = items.length;
  elements.instanceMeta.textContent = items.length ? `${items.length} loaded` : "Archive is empty";
  elements.instances.innerHTML = items.length ? items.map((item) => `
    <tr>
      <td title="${escapeHTML(item.uid)}">${escapeHTML(item.uid)}</td>
      <td title="${escapeHTML(item.studyUID)}">${escapeHTML(item.studyUID)}</td>
      <td>${escapeHTML(item.status)}</td>
      <td>${escapeHTML(item.version)}</td>
      <td>${escapeHTML(item.createdAt ? new Date(item.createdAt).toLocaleString() : "-")}</td>
    </tr>`).join("") : '<tr><td colspan="5" class="empty">No instances have been received</td></tr>';
}

function renderRows(target, items, fields, emptyText) {
  target.innerHTML = items.length ? items.map((item) => `
    <div class="data-row">
      <div><strong>${escapeHTML(item[fields.title])}</strong><small>${escapeHTML(item[fields.detail])}</small></div>
      <span class="badge">${escapeHTML(item[fields.badge])}</span>
    </div>`).join("") : `<p class="empty">${emptyText}</p>`;
}

async function refresh() {
  elements.refresh.disabled = true;
  elements.notice.hidden = true;
  elements.connection.dataset.state = "loading";
  elements.connection.textContent = "Refreshing";

  try {
    const [ready, instanceData, jobData, destinationData] = await Promise.all([
      getJSON("/readyz"),
      getJSON("/api/v1/instances?limit=50"),
      getJSON("/api/v1/jobs"),
      getJSON("/api/v1/destinations"),
    ]);
    const instances = instanceData.items ?? [];
    const jobs = jobData.items ?? [];
    const destinations = destinationData.items ?? [];

    elements.serviceStatus.textContent = ready.status === "ready" ? "Ready" : "Degraded";
    elements.jobCount.textContent = jobs.length;
    elements.destinationCount.textContent = destinations.length;
    renderInstances(instances);
    renderRows(elements.jobs, jobs, { title: "instanceUID", detail: "destinationID", badge: "status" }, "No route jobs are queued");
    renderRows(elements.destinations, destinations, { title: "name", detail: "endpoint", badge: "enabled" }, "No destinations are configured");
    elements.connection.dataset.state = "ready";
    elements.connection.textContent = "Service connected";
  } catch (error) {
    elements.serviceStatus.textContent = "Unavailable";
    elements.connection.dataset.state = "error";
    elements.connection.textContent = "Service unavailable";
    elements.notice.textContent = error instanceof Error ? error.message : "Unable to load gateway data";
    elements.notice.hidden = false;
  } finally {
    elements.refresh.disabled = false;
  }
}

elements.refresh.addEventListener("click", refresh);
refresh();
