globalThis.Pizza = globalThis.Pizza || {};

Pizza.columns = {
  kitchen: [
    ["blocked", "Blocked"],
    ["ready", "Ready"],
    ["working", "Working"],
    ["done", "Done"],
  ],
  delivery: [
    ["waiting", "Waiting to be cooked"],
    ["cooking", "Cooking"],
    ["quality", "Quality check"],
    ["out", "Out for delivery"],
    ["delivered", "Delivered"],
    ["paid", "Paid"],
  ],
};

Pizza.menu = {
  sizes: [
    { id: "small", label: "Small", cents: 1000 },
    { id: "medium", label: "Medium", cents: 1300 },
    { id: "large", label: "Large", cents: 1600 },
  ],
  crusts: [
    { id: "neapolitan", label: "Neapolitan" },
    { id: "thin", label: "Thin" },
    { id: "sourdough", label: "Sourdough" },
    { id: "sicilian", label: "Sicilian" },
  ],
  sauces: [
    { id: "tomato", label: "Tomato" },
    { id: "white", label: "White garlic" },
    { id: "pesto", label: "Pesto" },
    { id: "bbq", label: "BBQ" },
  ],
  toppings: [
    { id: "pepperoni", label: "Pepperoni", cents: 200 },
    { id: "mushrooms", label: "Mushrooms", cents: 100 },
    { id: "olives", label: "Olives", cents: 100 },
    { id: "onions", label: "Red onions", cents: 100 },
    { id: "basil", label: "Fresh basil", cents: 100 },
    { id: "sausage", label: "Sausage", cents: 200 },
    { id: "peppers", label: "Peppers", cents: 100 },
    { id: "extra_cheese", label: "Extra cheese", cents: 200 },
  ],
};

Pizza.taskTemplate = [
  {
    code: "stretch",
    title: "Stretch dough",
    description: "Shape the dough and dust the peel.",
    station: "Dough",
    dependencies: [],
  },
  {
    code: "sauce",
    title: "Sauce base",
    description: "Spread sauce to the edge without flooding the center.",
    station: "Prep",
    dependencies: ["stretch"],
  },
  {
    code: "cheese",
    title: "Add cheese",
    description: "Add mozzarella and balance coverage.",
    station: "Prep",
    dependencies: ["sauce"],
  },
  {
    code: "toppings",
    title: "Add toppings",
    description: "Place toppings evenly so every slice gets something good.",
    station: "Prep",
    dependencies: ["cheese"],
  },
  {
    code: "bake",
    title: "Bake pizza",
    description: "Bake until the crust blisters and the cheese freckles.",
    station: "Oven",
    dependencies: ["toppings"],
  },
  {
    code: "box",
    title: "Slice and box",
    description: "Slice, garnish, box, and hand off to delivery.",
    station: "Expo",
    dependencies: ["bake"],
  },
];

Pizza.util = {
  sessionId(session) {
    if (typeof session === "string") return session;
    return String((session && session.id) || "default");
  },

  money(cents) {
    return "$" + (Number(cents || 0) / 100).toFixed(2);
  },

  orderNumber(id) {
    return "PZ-" + String(id).padStart(4, "0");
  },

  selectedList(value) {
    if (!value) return [];
    if (Array.isArray(value)) return value;
    return [value];
  },

  label(items, id) {
    const item = items.find((x) => x.id === id || x[0] === id);
    return item ? item.label || item[1] : id;
  },

  validColumn(columns, value, fallback) {
    return columns.some(([id]) => id === value) ? value : fallback;
  },

  deliveryRank(status) {
    return Pizza.columns.delivery.findIndex(([id]) => id === status);
  },

  toppingLabels(csv) {
    const ids = String(csv || "")
      .split(",")
      .filter(Boolean);
    if (ids.length === 0) return "plain cheese";
    return ids.map((id) => this.label(Pizza.menu.toppings, id)).join(", ");
  },

  dependencyText(csv) {
    const deps = String(csv || "")
      .split(",")
      .filter(Boolean);
    return deps.length ? deps.join(", ") : "none";
  },

  priceFor(size, toppings) {
    const sizeItem =
      Pizza.menu.sizes.find((x) => x.id === size) || Pizza.menu.sizes[1];
    const toppingTotal = this.selectedList(toppings).reduce((sum, id) => {
      const topping = Pizza.menu.toppings.find((x) => x.id === id);
      return sum + (topping ? topping.cents : 0);
    }, 0);
    return sizeItem.cents + toppingTotal;
  },
};
