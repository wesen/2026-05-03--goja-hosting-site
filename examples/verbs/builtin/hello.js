__package__({
  name: "local-builtin",
  parents: ["examples"],
  short: "Local copies of the built-in goja smoke-test verbs",
});

function hello(name) {
  return { greeting: "hello " + (name || "world") };
}

__verb__("hello", {
  short: "Return a greeting from the built-in verb repository",
  fields: {
    name: { type: "string", default: "world", help: "Name to greet" },
  },
});
