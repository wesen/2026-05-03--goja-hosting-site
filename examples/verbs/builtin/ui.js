__package__({
  name: "local-builtin",
  parents: ["examples"],
  short: "Local copies of the built-in goja smoke-test verbs",
});

function renderSampleTable() {
  const ui = require("ui.dsl");
  return ui.render(ui.table.fromRows("sample", [
    { name: "Alice", role: "admin" },
    { name: "Bob", role: "viewer" },
  ]).features(f => f.pagination().sorting()).render({}));
}

__verb__("renderSampleTable", {
  short: "Render a sample HTML table with ui.dsl",
  outputMode: "text",
});
