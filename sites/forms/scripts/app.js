const db = require("database");
const express = require("express");
const ui = require("ui.dsl");

const app = express.app();

const fieldTypes = ["text", "textarea", "email", "number", "date", "select", "checkbox"];

function migrate() {
  db.exec(`CREATE TABLE IF NOT EXISTS forms (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    slug TEXT NOT NULL UNIQUE,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
  )`);
  db.exec(`CREATE TABLE IF NOT EXISTS form_fields (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    form_id INTEGER NOT NULL,
    label TEXT NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT 'text',
    placeholder TEXT NOT NULL DEFAULT '',
    options TEXT NOT NULL DEFAULT '',
    required INTEGER NOT NULL DEFAULT 0,
    position INTEGER NOT NULL DEFAULT 0
  )`);
  db.exec(`CREATE TABLE IF NOT EXISTS form_responses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    form_id INTEGER NOT NULL,
    payload TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
  )`);
}

function slugify(value) {
  const slug = String(value || "form")
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 48);
  return slug || "form";
}

function fieldName(label) {
  return slugify(label).replace(/-/g, "_") || "field";
}

function uniqueSlug(base) {
  let slug = slugify(base);
  let suffix = 2;
  while (db.query("SELECT id FROM forms WHERE slug = ?", slug).length > 0) {
    slug = `${slugify(base)}-${suffix++}`;
  }
  return slug;
}

function nextPosition(formId) {
  const rows = db.query("SELECT COALESCE(MAX(position),0) AS max_position FROM form_fields WHERE form_id = ?", Number(formId));
  return Number(rows[0]?.max_position || 0) + 10;
}

function getForm(idOrSlug) {
  const numeric = Number(idOrSlug);
  if (Number.isFinite(numeric) && String(idOrSlug).match(/^\d+$/)) {
    return db.query("SELECT * FROM forms WHERE id = ?", numeric)[0];
  }
  return db.query("SELECT * FROM forms WHERE slug = ?", String(idOrSlug || ""))[0];
}

function listForms() {
  return db.query(`SELECT f.*, COUNT(r.id) AS response_count
    FROM forms f
    LEFT JOIN form_responses r ON r.form_id = f.id
    GROUP BY f.id
    ORDER BY f.id DESC`);
}

function listFields(formId) {
  return db.query("SELECT * FROM form_fields WHERE form_id = ? ORDER BY position,id", Number(formId));
}

function listResponses(formId) {
  return db.query("SELECT * FROM form_responses WHERE form_id = ? ORDER BY id DESC", Number(formId));
}

function seedIfEmpty() {
  const rows = db.query("SELECT COUNT(*) AS count FROM forms");
  if (Number(rows[0]?.count || 0) > 0) return;
  const slug = uniqueSlug("Website intake");
  db.exec("INSERT INTO forms(title,description,slug) VALUES (?, ?, ?)", "Website intake", "Capture enough context to generate a first-pass project brief.", slug);
  const form = getForm(slug);
  [
    ["Project name", "project_name", "text", "Customer portal", "", 1],
    ["Contact email", "contact_email", "email", "you@example.com", "", 1],
    ["Launch date", "launch_date", "date", "", "", 0],
    ["Project type", "project_type", "select", "", "Marketing site\nInternal tool\nCustomer app\nAutomation", 1],
    ["Must-have features", "must_have_features", "textarea", "Authentication, dashboard, exports...", "", 0],
    ["Needs follow-up", "needs_follow_up", "checkbox", "", "", 0],
  ].forEach((field, index) => {
    db.exec(
      "INSERT INTO form_fields(form_id,label,name,type,placeholder,options,required,position) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
      form.id,
      field[0],
      field[1],
      field[2],
      field[3],
      field[4],
      field[5],
      (index + 1) * 10,
    );
  });
}

function stylesheet() {
  return `
  :root{--ink:#111827;--muted:#6b7280;--line:#1f2937;--paper:#fff7ed;--panel:#ffffff;--accent:#fb923c;--accent2:#14b8a6;font-family:Inter,ui-sans-serif,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif}*{box-sizing:border-box}body{margin:0;color:var(--ink);background:radial-gradient(circle at top left,#ffedd5,#f8fafc 42%,#ccfbf1)}a{color:#0f766e;font-weight:800;text-decoration:none}.shell{max-width:1180px;margin:0 auto;padding:32px}.hero{display:grid;grid-template-columns:1.4fr .8fr;gap:18px;align-items:stretch;margin-bottom:18px}.panel,.card,.builder,.preview{background:rgba(255,255,255,.88);border:3px solid var(--line);border-radius:22px;box-shadow:8px 10px 0 rgba(15,23,42,.18);padding:18px}.brand{font-size:52px;line-height:.95;letter-spacing:-3px;margin:0}.muted{color:var(--muted)}.grid{display:grid;grid-template-columns:1fr 1fr;gap:18px}.forms{display:grid;gap:12px}.form-row{display:flex;justify-content:space-between;gap:12px;align-items:center;background:white;border:2px solid var(--line);border-radius:16px;padding:13px}.actions{display:flex;gap:8px;flex-wrap:wrap}.button,button{border:2px solid var(--line);border-radius:999px;background:#fed7aa;color:#111827;padding:9px 13px;font-weight:900;cursor:pointer;font:inherit}.button.secondary,button.secondary{background:#ccfbf1}.button.dark{background:#111827;color:white}input,textarea,select{width:100%;border:2px solid var(--line);border-radius:12px;padding:10px 12px;background:white;font:inherit}textarea{min-height:96px}.stack{display:grid;gap:10px}.inline{display:grid;grid-template-columns:1fr 1fr;gap:10px}.field{display:grid;gap:5px}.field label{font-weight:900}.field small{color:var(--muted)}.builder{margin-top:18px}.preview{margin-top:18px}.field-list{display:grid;gap:10px}.field-pill{border:2px dashed #0f766e;border-radius:14px;padding:10px;background:#f0fdfa}.response{white-space:pre-wrap;background:#0f172a;color:#e2e8f0;border-radius:14px;padding:12px;overflow:auto}.topbar{display:flex;justify-content:space-between;gap:12px;align-items:center;margin-bottom:16px}@media(max-width:860px){.hero,.grid,.inline{grid-template-columns:1fr}.brand{font-size:40px}.shell{padding:18px}}`;
}

function layout(title, children) {
  return ui.page(
    { title },
    ui.style(stylesheet()),
    ui.main({ class: "shell" }, children),
  );
}

function createFormPanel() {
  return ui.form(
    { class: "panel stack", method: "post", action: "/forms" },
    ui.h2("Create a form"),
    ui.div({ class: "field" }, ui.label("Title"), ui.input({ name: "title", placeholder: "Conference registration", required: true })),
    ui.div({ class: "field" }, ui.label("Description"), ui.textarea({ name: "description", placeholder: "What should respondents know before filling this out?" })),
    ui.button({ type: "submit" }, "Create form"),
  );
}

function homePage() {
  const forms = listForms();
  return layout(
    "Form Studio",
    ui.fragment(
      ui.section(
        { class: "hero" },
        ui.div({ class: "panel" }, ui.h1({ class: "brand" }, "Form Studio"), ui.p({ class: "muted" }, "Generate small hosted forms, add fields, preview the respondent experience, and collect JSON responses in SQLite."), ui.div({ class: "actions" }, ui.a({ class: "button dark", href: "/api/forms" }, "Forms API"))),
        createFormPanel(),
      ),
      ui.section(
        { class: "panel" },
        ui.h2("Your forms"),
        forms.length === 0
          ? ui.p({ class: "muted" }, "No forms yet. Create one above.")
          : ui.div(
              { class: "forms" },
              forms.map((form) =>
                ui.div(
                  { class: "form-row" },
                  ui.div(ui.strong(form.title), ui.div({ class: "muted" }, `/${form.slug} · ${form.response_count || 0} responses`)),
                  ui.div({ class: "actions" }, ui.a({ class: "button", href: `/forms/${form.id}` }, "Builder"), ui.a({ class: "button secondary", href: `/f/${form.slug}` }, "Open")),
                ),
              ),
            ),
      ),
    ),
  );
}

function fieldEditor(form) {
  return ui.form(
    { class: "builder stack", method: "post", action: `/forms/${form.id}/fields` },
    ui.h2("Add a field"),
    ui.div(
      { class: "inline" },
      ui.div({ class: "field" }, ui.label("Label"), ui.input({ name: "label", placeholder: "Company size", required: true })),
      ui.div(
        { class: "field" },
        ui.label("Type"),
        ui.select({ name: "type" }, fieldTypes.map((type) => ui.option({ value: type }, type))),
      ),
    ),
    ui.div(
      { class: "inline" },
      ui.div({ class: "field" }, ui.label("Placeholder"), ui.input({ name: "placeholder", placeholder: "1-50 employees" })),
      ui.div({ class: "field" }, ui.label("Options"), ui.input({ name: "options", placeholder: "One option per line or comma separated" }), ui.small("Used by select fields.")),
    ),
    ui.label(ui.input({ type: "checkbox", name: "required", value: "1" }), " Required"),
    ui.button({ type: "submit" }, "Add field"),
  );
}

function renderInput(field, value) {
  const common = { name: field.name, id: field.name, placeholder: field.placeholder || "", required: Number(field.required || 0) ? true : undefined };
  if (field.type === "textarea") return ui.textarea(common, value || "");
  if (field.type === "select") {
    const opts = String(field.options || "").split(/[\n,]+/).map((s) => s.trim()).filter(Boolean);
    return ui.select(common, opts.map((opt) => ui.option({ value: opt }, opt)));
  }
  if (field.type === "checkbox") return ui.label(ui.input({ type: "checkbox", name: field.name, value: "yes" }), " Yes");
  return ui.input({ ...common, type: field.type || "text", value: value || "" });
}

function publicForm(form, fields) {
  return ui.form(
    { class: "preview stack", method: "post", action: `/f/${form.slug}/responses` },
    ui.h2(form.title),
    form.description ? ui.p({ class: "muted" }, form.description) : null,
    fields.length === 0 ? ui.p({ class: "muted" }, "This form has no fields yet.") : null,
    fields.map((field) => ui.div({ class: "field" }, ui.label({ for: field.name }, field.label + (Number(field.required || 0) ? " *" : "")), renderInput(field))),
    fields.length ? ui.button({ type: "submit" }, "Submit response") : null,
  );
}

function builderPage(form) {
  const fields = listFields(form.id);
  const responses = listResponses(form.id);
  return layout(
    `${form.title} builder`,
    ui.fragment(
      ui.div({ class: "topbar" }, ui.a({ href: "/" }, "← Forms"), ui.div({ class: "actions" }, ui.a({ class: "button secondary", href: `/f/${form.slug}` }, "Open public form"), ui.a({ class: "button", href: `/api/forms/${form.id}` }, "JSON"))),
      ui.section({ class: "panel" }, ui.h1(form.title), ui.p({ class: "muted" }, form.description || "No description yet."), ui.p(ui.strong("Public URL: "), ui.a({ href: `/f/${form.slug}` }, `/f/${form.slug}`))),
      ui.div(
        { class: "grid" },
        ui.div(fieldEditor(form), ui.div({ class: "panel builder" }, ui.h2("Fields"), fields.length ? ui.div({ class: "field-list" }, fields.map((field) => ui.div({ class: "field-pill" }, ui.strong(field.label), ui.div({ class: "muted" }, `${field.name} · ${field.type}${Number(field.required || 0) ? " · required" : ""}`)))) : ui.p({ class: "muted" }, "No fields yet."))),
        ui.div(publicForm(form, fields)),
      ),
      ui.section({ class: "panel builder" }, ui.h2("Responses"), responses.length ? ui.div({ class: "stack" }, responses.map((response) => ui.div(ui.strong(response.created_at), ui.pre({ class: "response" }, JSON.stringify(JSON.parse(response.payload), null, 2))))) : ui.p({ class: "muted" }, "No responses yet.")),
    ),
  );
}

function submittedPage(form) {
  return layout("Response saved", ui.section({ class: "panel" }, ui.h1("Thanks!"), ui.p("Your response was saved."), ui.div({ class: "actions" }, ui.a({ class: "button", href: `/f/${form.slug}` }, "Submit another"), ui.a({ class: "button secondary", href: `/forms/${form.id}` }, "Back to builder"))));
}

migrate();
seedIfEmpty();

app.get("/", (req, res) => res.html(homePage()));
app.get("/favicon.ico", (req, res) => res.status(204).end());
app.get("/forms/:id", (req, res) => {
  const form = getForm(req.params.id);
  if (!form) return res.status(404).send("form not found");
  return res.html(builderPage(form));
});
app.post("/forms", (req, res) => {
  const body = req.body || {};
  const title = String(body.title || "").trim();
  if (!title) return res.status(400).send("title is required");
  const slug = uniqueSlug(title);
  db.exec("INSERT INTO forms(title,description,slug) VALUES (?, ?, ?)", title, String(body.description || ""), slug);
  const form = getForm(slug);
  return res.redirect(`/forms/${form.id}`);
});
app.post("/forms/:id/fields", (req, res) => {
  const form = getForm(req.params.id);
  if (!form) return res.status(404).send("form not found");
  const body = req.body || {};
  const label = String(body.label || "").trim();
  if (!label) return res.status(400).send("label is required");
  const type = fieldTypes.includes(String(body.type || "")) ? String(body.type) : "text";
  db.exec(
    "INSERT INTO form_fields(form_id,label,name,type,placeholder,options,required,position) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
    form.id,
    label,
    fieldName(label),
    type,
    String(body.placeholder || ""),
    String(body.options || ""),
    body.required ? 1 : 0,
    nextPosition(form.id),
  );
  return res.redirect(`/forms/${form.id}`);
});
app.get("/f/:slug", (req, res) => {
  const form = getForm(req.params.slug);
  if (!form) return res.status(404).send("form not found");
  return res.html(layout(form.title, publicForm(form, listFields(form.id))));
});
app.post("/f/:slug/responses", (req, res) => {
  const form = getForm(req.params.slug);
  if (!form) return res.status(404).send("form not found");
  const fields = listFields(form.id);
  const body = req.body || {};
  const payload = {};
  fields.forEach((field) => {
    payload[field.name] = field.type === "checkbox" ? (body[field.name] ? true : false) : String(body[field.name] || "");
  });
  db.exec("INSERT INTO form_responses(form_id,payload) VALUES (?, ?)", form.id, JSON.stringify(payload));
  return res.html(submittedPage(form));
});
app.get("/api/forms", (req, res) => res.json(listForms()));
app.get("/api/forms/:id", (req, res) => {
  const form = getForm(req.params.id);
  if (!form) return res.status(404).json({ error: "form not found" });
  return res.json({ form, fields: listFields(form.id), responses: listResponses(form.id).map((row) => ({ ...row, payload: JSON.parse(row.payload) })) });
});
