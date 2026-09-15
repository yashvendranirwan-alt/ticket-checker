const state = {
  token: localStorage.getItem("ticket_token"),
  username: localStorage.getItem("ticket_username"),
  tickets: [],
  filter: "all",
  registerMode: false,
};

const $ = (selector) => document.querySelector(selector);
const authView = $("#auth-view");
const appView = $("#app-view");
const authForm = $("#auth-form");
const authMessage = $("#auth-message");
const ticketDialog = $("#ticket-dialog");

function setMessage(element, message = "") {
  element.textContent = message;
}

function showApp() {
  authView.classList.add("hidden");
  appView.classList.remove("hidden");
  $("#user-label").textContent = `@${state.username}`;
  loadTickets();
}

function showAuth() {
  state.token = null;
  state.username = null;
  localStorage.removeItem("ticket_token");
  localStorage.removeItem("ticket_username");
  appView.classList.add("hidden");
  authView.classList.remove("hidden");
  authForm.reset();
}

async function api(path, options = {}) {
  const headers = { "Content-Type": "application/json", ...(options.headers || {}) };
  if (state.token) headers.Authorization = `Bearer ${state.token}`;
  const response = await fetch(path, { ...options, headers });
  let body = {};
  try { body = await response.json(); } catch (_) { /* Empty response. */ }
  if (response.status === 401 && path.startsWith("/tickets")) {
    showAuth();
    throw new Error("Your session has expired. Please sign in again.");
  }
  if (!response.ok) throw new Error(body.error || "Something went wrong.");
  return body;
}

function renderAuthMode() {
  const register = state.registerMode;
  $("#auth-eyebrow").textContent = register ? "First time here" : "Welcome back";
  $("#auth-title").textContent = register ? "Create your desk" : "Sign in to your desk";
  $("#auth-subtitle").textContent = register ? "Start with a name and a password." : "Pick up where you left off.";
  $("#auth-submit-label").textContent = register ? "Create account" : "Sign in";
  $("#auth-toggle").textContent = register ? "Already have an account? Sign in" : "Need an account? Create one";
  $("#password").autocomplete = register ? "new-password" : "current-password";
  setMessage(authMessage);
}

async function handleAuth(event) {
  event.preventDefault();
  setMessage(authMessage);
  const form = new FormData(authForm);
  const username = form.get("username").trim();
  const password = form.get("password");
  const endpoint = state.registerMode ? "/auth/register" : "/auth/login";
  const button = authForm.querySelector("button[type=submit]");
  button.disabled = true;
  try {
    if (state.registerMode) await api(endpoint, { method: "POST", body: JSON.stringify({ username, password }) });
    const result = state.registerMode
      ? await api("/auth/login", { method: "POST", body: JSON.stringify({ username, password }) })
      : await api(endpoint, { method: "POST", body: JSON.stringify({ username, password }) });
    state.token = result.token;
    state.username = username;
    localStorage.setItem("ticket_token", state.token);
    localStorage.setItem("ticket_username", username);
    showApp();
  } catch (error) {
    setMessage(authMessage, error.message);
  } finally {
    button.disabled = false;
  }
}

function formatStatus(status) {
  return status === "in_progress" ? "In progress" : status[0].toUpperCase() + status.slice(1);
}

function nextStatus(status) {
  return status === "open" ? "in_progress" : status === "in_progress" ? "closed" : null;
}

function renderTickets() {
  const list = $("#ticket-list");
  const empty = $("#empty-state");
  const visible = state.filter === "all" ? state.tickets : state.tickets.filter((ticket) => ticket.status === state.filter);
  $("#stat-all").textContent = state.tickets.length;
  $("#stat-progress").textContent = state.tickets.filter((ticket) => ticket.status === "in_progress").length;
  $("#stat-closed").textContent = state.tickets.filter((ticket) => ticket.status === "closed").length;
  list.innerHTML = visible.map((ticket) => {
    const next = nextStatus(ticket.status);
    return `<article class="ticket-row"><span class="ticket-id">#${String(ticket.id).padStart(3, "0")}</span><span class="ticket-title">${escapeHtml(ticket.title)}</span><span class="ticket-date">${formatStatus(ticket.status)}</span><span class="status-badge ${ticket.status}">${formatStatus(ticket.status)}</span>${next ? `<button class="status-action" data-ticket-id="${ticket.id}" type="button">Move to ${formatStatus(next)} -></button>` : ""}</article>`;
  }).join("");
  empty.classList.toggle("hidden", visible.length > 0);
  list.classList.toggle("hidden", visible.length === 0);
}

function escapeHtml(value) {
  return value.replace(/[&<>'"]/g, (character) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "'": "&#39;", '"': "&quot;" })[character]);
}

async function loadTickets() {
  try {
    state.tickets = await api("/tickets");
    renderTickets();
  } catch (error) {
    setMessage($("#app-message"), error.message);
  }
}

async function updateStatus(id, status) {
  try {
    await api(`/tickets/${id}/status`, { method: "PATCH", body: JSON.stringify({ status }) });
    await loadTickets();
  } catch (error) {
    setMessage($("#app-message"), error.message);
  }
}

async function createTicket(event) {
  event.preventDefault();
  const titleInput = $("#ticket-title");
  const button = event.target.querySelector("button[type=submit]");
  button.disabled = true;
  setMessage($("#ticket-message"));
  try {
    await api("/tickets", { method: "POST", body: JSON.stringify({ title: titleInput.value.trim() }) });
    ticketDialog.close();
    titleInput.value = "";
    await loadTickets();
  } catch (error) {
    setMessage($("#ticket-message"), error.message);
  } finally {
    button.disabled = false;
  }
}

authForm.addEventListener("submit", handleAuth);
$("#auth-toggle").addEventListener("click", () => { state.registerMode = !state.registerMode; renderAuthMode(); });
$("#logout-button").addEventListener("click", showAuth);
$("#new-ticket-button").addEventListener("click", () => ticketDialog.showModal());
$("#empty-create-button").addEventListener("click", () => ticketDialog.showModal());
$("#close-dialog").addEventListener("click", () => ticketDialog.close());
$("#ticket-form").addEventListener("submit", createTicket);
$("#ticket-list").addEventListener("click", (event) => {
  const button = event.target.closest("[data-ticket-id]");
  if (!button) return;
  const ticket = state.tickets.find((item) => item.id === Number(button.dataset.ticketId));
  if (ticket) updateStatus(ticket.id, nextStatus(ticket.status));
});

document.querySelectorAll(".filter").forEach((button) => button.addEventListener("click", () => {
  state.filter = button.dataset.filter;
  document.querySelectorAll(".filter").forEach((item) => item.classList.toggle("active", item === button));
  renderTickets();
}));

renderAuthMode();
if (state.token && state.username) showApp();
