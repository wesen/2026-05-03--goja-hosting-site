const ui = require("ui.dsl");
const kanban = require("kanban.dsl");

globalThis.Pizza = globalThis.Pizza || {};

Pizza.views = {
  page(req) {
    const tally = Pizza.store.tally(req.session);

    return ui.page(
      { title: "Pizza Ops" },
      ui.style(Pizza.styles),
      ui.main(
        { class: "page" },
        this.hero(tally),
        this.builderForm(),
        ui.div(
          { class: "boards" },
          this.kitchenSection(req),
          this.deliverySection(req),
        ),
        this.apiLinks(),
      ),
    );
  },

  hero(tally) {
    return ui.header(
      { class: "hero" },
      ui.div(
        { class: "hero-copy" },
        ui.div({ class: "mascot", "aria-hidden": "true" }, "◩"),
        ui.div(
          ui.h1({ class: "brand" }, "Pizza Ops"),
          ui.p(
            { class: "subtitle" },
            "Build a pizza, split it into dependent kitchen tasks, watch the delivery board, and tally paid orders plus tips.",
          ),
        ),
      ),
      ui.div(
        { class: "tally" },
        this.tallyCard("▤", tally.orders, "Orders"),
        this.tallyCard("♞", tally.delivered, "Delivered"),
        this.tallyCard("▣", tally.paid, "Paid"),
        this.tallyCard("▥", Pizza.util.money(tally.revenueCents), "Sales"),
        this.tallyCard("▴", Pizza.util.money(tally.tipCents), "Tips"),
      ),
    );
  },

  tallyCard(icon, value, label) {
    return ui.div(
      { class: "tally-card" },
      ui.span({ class: "tally-icon", "aria-hidden": "true" }, icon),
      ui.div(ui.strong(String(value)), ui.span(label)),
    );
  },

  builderForm() {
    return ui.form(
      { class: "pizza-builder", method: "post", action: "/orders" },
      ui.h2("Build a pizza"),
      ui.div(
        { class: "builder-grid" },
        this.field(
          "Customer",
          ui.input({ name: "customer", placeholder: "Name", required: true }),
        ),
        this.field(
          "Delivery address",
          ui.input({
            name: "address",
            placeholder: "Street or pickup note",
            required: true,
          }),
          "wide",
        ),
        this.field(
          "Size",
          this.select("size", Pizza.menu.sizes, "medium", true),
        ),
        this.field(
          "Crust",
          this.select("crust", Pizza.menu.crusts, "neapolitan"),
        ),
        this.field("Sauce", this.select("sauce", Pizza.menu.sauces, "tomato")),
        this.field(
          "Kitchen note",
          ui.input({
            name: "notes",
            placeholder: "Allergies, gate code, delivery note",
          }),
          "wide",
        ),
        ui.div(
          { class: "topping-picker" },
          ui.strong("Toppings"),
          ui.div(
            { class: "topping-grid" },
            Pizza.menu.toppings.map((topping) =>
              ui.label(
                { class: "topping-option" },
                ui.input({
                  type: "checkbox",
                  name: "toppings",
                  value: topping.id,
                }),
                `${topping.label} +${Pizza.util.money(topping.cents)}`,
              ),
            ),
          ),
        ),
        ui.button(
          { class: "primary-action", type: "submit" },
          "Fire this pizza",
        ),
      ),
    );
  },

  field(label, control, extraClass) {
    return ui.label(
      { class: ["form-field", extraClass || ""].join(" ").trim() },
      label,
      control,
    );
  },

  select(name, items, selected, includePrice) {
    return ui.select(
      { name },
      items.map((item) => {
        const label = includePrice
          ? `${item.label} ${Pizza.util.money(item.cents)}`
          : item.label;
        return ui.option(
          { value: item.id, selected: item.id === selected },
          label,
        );
      }),
    );
  },

  kitchenSection(req) {
    return ui.section(
      { class: "board-panel kitchen-panel" },
      this.boardHeading(
        "Kitchen dependency board",
        "Tasks unlock as prerequisites reach Done.",
      ),
      Pizza.boards.kitchen.render({ session: req.session, query: req.query }),
    );
  },

  deliverySection(req) {
    return ui.section(
      { class: "board-panel delivery-panel" },
      this.boardHeading(
        "Delivery board",
        "Orders advance as the kitchen finishes, then delivery and payment take over.",
      ),
      Pizza.boards.delivery.render({ session: req.session, query: req.query }),
    );
  },

  boardHeading(title, hint) {
    return ui.div(
      { class: "board-heading" },
      ui.h2(title),
      ui.span({ class: "hint" }, hint),
    );
  },

  kitchenCard(task) {
    if (task.kind === "doneGroup") {
      return this.doneKitchenGroup(task);
    }

    return ui.div(
      { class: "task-card" },
      ui.div(
        { class: "card-topline" },
        ui.span({ class: "ticket" }, Pizza.util.orderNumber(task.order_id)),
        ui.span({ class: "chip" }, task.station),
      ),
      ui.h3(task.title),
      ui.p({ class: "description" }, task.description),
      ui.div(
        { class: "chip-row" },
        ui.span({ class: "chip" }, task.customer),
        ui.span({ class: "chip" }, `${task.size} / ${task.crust}`),
        task.deps
          ? ui.span(
              { class: "chip blocked" },
              "needs " + Pizza.util.dependencyText(task.deps),
            )
          : ui.span({ class: "chip ready" }, "no dependencies"),
      ),
    );
  },

  doneKitchenGroup(group) {
    const titles = String(group.done_titles || "")
      .split(",")
      .map((title) => title.trim())
      .filter(Boolean);

    return ui.div(
      { class: "task-card done-group-card" },
      ui.div(
        { class: "card-topline" },
        ui.span({ class: "ticket" }, Pizza.util.orderNumber(group.order_id)),
        ui.span({ class: "chip ready" }, `${group.task_count} done`),
      ),
      ui.h3("Completed kitchen work"),
      ui.ul(
        { class: "done-task-list" },
        titles.map((title) => ui.li(title)),
      ),
      ui.div(
        { class: "chip-row" },
        ui.span({ class: "chip" }, group.customer),
        ui.span({ class: "chip" }, `${group.size} / ${group.crust}`),
        ui.span({ class: "chip ready" }, "ready for handoff"),
      ),
    );
  },

  deliveryCard(order) {
    return ui.div(
      { class: "order-card" },
      ui.div(
        { class: "card-topline" },
        ui.span({ class: "ticket" }, Pizza.util.orderNumber(order.id)),
        ui.span({ class: "price" }, Pizza.util.money(order.total_cents)),
      ),
      ui.h3(order.customer || "Guest"),
      ui.p({ class: "small-text" }, order.address || "Pickup counter"),
      ui.div(
        { class: "chip-row" },
        ui.span(
          { class: "chip" },
          Pizza.util.label(Pizza.menu.sizes, order.size),
        ),
        ui.span(
          { class: "chip" },
          Pizza.util.label(Pizza.menu.crusts, order.crust),
        ),
        ui.span(
          { class: "chip" },
          Pizza.util.label(Pizza.menu.sauces, order.sauce),
        ),
        ui.span({ class: "chip" }, Pizza.util.toppingLabels(order.toppings)),
      ),
      order.notes
        ? ui.p({ class: "small-text" }, "Note: " + order.notes)
        : null,
      Number(order.paid || 0)
        ? ui.p(
            { class: "chip paid-note" },
            "Paid with " + Pizza.util.money(order.tip_cents) + " tip",
          )
        : order.delivery_status === "delivered"
          ? this.paymentForm(order)
          : ui.p(
              { class: "small-text payment-hint" },
              "Payment appears after delivery.",
            ),
    );
  },

  paymentForm(order) {
    return ui.form(
      {
        class: "payment-form",
        method: "post",
        action: "/orders/" + order.id + "/pay",
      },
      ui.input({ name: "tip", placeholder: "Tip", value: "5.00" }),
      ui.button({ type: "submit" }, "Pay"),
    );
  },

  apiLinks() {
    return ui.nav(
      { class: "api-links" },
      ui.a({ href: "/api/tally" }, "tally json"),
      ui.a({ href: "/api/orders" }, "orders json"),
      ui.a({ href: "/api/tasks" }, "tasks json"),
    );
  },
};

Pizza.boards = {
  kitchen: kanban
    .board("pizza-kitchen")
    .title("Kitchen tasks")
    .className("board kitchen-board")
    .columns((columns) =>
      columns
        .column("blocked")
        .title("Blocked")
        .done()
        .column("ready")
        .title("Ready")
        .done()
        .column("working")
        .title("Working")
        .done()
        .column("done")
        .title("Done")
        .terminal(true)
        .done(),
    )
    .data((data) =>
      data
        .cards((ctx) => Pizza.store.listKitchenCards(ctx.session))
        .id((task) => String(task.card_id || task.id))
        .column((task) => task.status)
        .position((task) => Number(task.position || 0))
        .searchText((task) =>
          [
            Pizza.util.orderNumber(task.order_id),
            task.customer,
            task.title,
            task.done_titles,
            task.station,
            task.size,
            Pizza.util.toppingLabels(task.toppings),
            task.deps,
          ]
            .join(" ")
            .toLowerCase(),
        ),
    )
    .features((features) => features.search({ mode: "client" }).dragDrop())
    .render((render) =>
      render
        .card((task) => Pizza.views.kitchenCard(task))
        .emptyColumn(() => ui.div({ class: "empty" }, "No tasks here.")),
    )
    .actions((actions) =>
      actions.cardMoved((event) => Pizza.store.moveKitchenTask(event)),
    )
    .build(),

  delivery: kanban
    .board("pizza-delivery")
    .title("Delivery board")
    .className("board delivery-board")
    .columns((columns) =>
      columns
        .column("waiting")
        .title("Waiting")
        .done()
        .column("cooking")
        .title("Cooking")
        .done()
        .column("quality")
        .title("Quality")
        .done()
        .column("out")
        .title("Out")
        .done()
        .column("delivered")
        .title("Delivered")
        .done()
        .column("paid")
        .title("Paid")
        .terminal(true)
        .done(),
    )
    .data((data) =>
      data
        .cards((ctx) => Pizza.store.listOrders(ctx.session))
        .id((order) => String(order.id))
        .column((order) => order.delivery_status)
        .position((order) => Number(order.position || 0))
        .searchText((order) =>
          [
            Pizza.util.orderNumber(order.id),
            order.customer,
            order.address,
            order.size,
            order.crust,
            order.sauce,
            Pizza.util.toppingLabels(order.toppings),
          ]
            .join(" ")
            .toLowerCase(),
        ),
    )
    .features((features) => features.search({ mode: "client" }).dragDrop())
    .render((render) =>
      render
        .card((order) => Pizza.views.deliveryCard(order))
        .emptyColumn(() => ui.div({ class: "empty" }, "No orders here.")),
    )
    .actions((actions) =>
      actions.cardMoved((event) => Pizza.store.moveDeliveryOrder(event)),
    )
    .build(),
};
