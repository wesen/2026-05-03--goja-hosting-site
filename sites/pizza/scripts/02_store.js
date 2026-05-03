const db = require("database");

globalThis.Pizza = globalThis.Pizza || {};

Pizza.store = {
  migrate() {
    db.exec(`CREATE TABLE IF NOT EXISTS pizza_orders (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      session_id TEXT NOT NULL DEFAULT 'default',
      customer TEXT NOT NULL DEFAULT '',
      address TEXT NOT NULL DEFAULT '',
      size TEXT NOT NULL DEFAULT 'medium',
      crust TEXT NOT NULL DEFAULT 'neapolitan',
      sauce TEXT NOT NULL DEFAULT 'tomato',
      toppings TEXT NOT NULL DEFAULT '',
      notes TEXT NOT NULL DEFAULT '',
      delivery_status TEXT NOT NULL DEFAULT 'waiting',
      total_cents INTEGER NOT NULL DEFAULT 0,
      tip_cents INTEGER NOT NULL DEFAULT 0,
      paid INTEGER NOT NULL DEFAULT 0,
      position INTEGER NOT NULL DEFAULT 0,
      created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
    )`);

    db.exec(`CREATE TABLE IF NOT EXISTS pizza_tasks (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      session_id TEXT NOT NULL DEFAULT 'default',
      order_id INTEGER NOT NULL,
      code TEXT NOT NULL,
      title TEXT NOT NULL,
      description TEXT NOT NULL DEFAULT '',
      station TEXT NOT NULL DEFAULT '',
      deps TEXT NOT NULL DEFAULT '',
      status TEXT NOT NULL DEFAULT 'blocked',
      position INTEGER NOT NULL DEFAULT 0,
      updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
    )`);
  },

  seedIfEmpty(session) {
    const sid = Pizza.util.sessionId(session);
    const row = db.query(
      "SELECT COUNT(*) AS count FROM pizza_orders WHERE session_id = ?",
      sid,
    )[0];

    if (Number((row && row.count) || 0) > 0) return;

    const first = this.createOrder(sid, {
      customer: "Mara",
      address: "14 Basil Lane",
      size: "large",
      crust: "sourdough",
      sauce: "tomato",
      toppings: ["pepperoni", "mushrooms", "extra_cheese"],
      notes: "Ring twice.",
    });

    this.createOrder(sid, {
      customer: "Ken",
      address: "8 Arcade Ave",
      size: "medium",
      crust: "thin",
      sauce: "white",
      toppings: ["olives", "basil", "peppers"],
      notes: "No-contact delivery.",
    });

    // Make the seed data demonstrate dependency unlocking immediately.
    ["stretch", "sauce", "cheese"].forEach((code) => {
      db.exec(
        "UPDATE pizza_tasks SET status = 'done' WHERE session_id = ? AND order_id = ? AND code = ?",
        sid,
        first.id,
        code,
      );
    });
    this.refreshOrder(sid, first.id);
  },

  createOrder(session, body) {
    const sid = Pizza.util.sessionId(session);
    const toppings = Pizza.util
      .selectedList(body.toppings || body.topping)
      .filter(Boolean);
    const size = String(body.size || "medium");
    const totalCents = Pizza.util.priceFor(size, toppings);

    db.exec(
      `INSERT INTO pizza_orders(
        session_id, customer, address, size, crust, sauce, toppings, notes,
        delivery_status, total_cents, position
      ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'waiting', ?, ?)`,
      sid,
      String(body.customer || "Guest"),
      String(body.address || "Pickup counter"),
      size,
      String(body.crust || "neapolitan"),
      String(body.sauce || "tomato"),
      toppings.join(","),
      String(body.notes || ""),
      totalCents,
      this.nextOrderPosition(sid, "waiting"),
    );

    const order = db.query(
      "SELECT * FROM pizza_orders WHERE session_id = ? ORDER BY id DESC LIMIT 1",
      sid,
    )[0];

    Pizza.taskTemplate.forEach((task, index) => {
      db.exec(
        `INSERT INTO pizza_tasks(
          session_id, order_id, code, title, description, station, deps, status, position
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
        sid,
        order.id,
        task.code,
        task.title,
        task.description,
        task.station,
        task.dependencies.join(","),
        task.dependencies.length === 0 ? "ready" : "blocked",
        (index + 1) * 10,
      );
    });

    return order;
  },

  listOrders(session) {
    const sid = Pizza.util.sessionId(session);
    this.seedIfEmpty(sid);

    return db.query(
      "SELECT * FROM pizza_orders WHERE session_id = ? ORDER BY position, id",
      sid,
    );
  },

  listTasks(session) {
    const sid = Pizza.util.sessionId(session);
    this.seedIfEmpty(sid);

    this.listOrders(sid).forEach((order) => this.refreshOrder(sid, order.id));

    return db.query(
      `SELECT
        task.*,
        orders.customer,
        orders.size,
        orders.crust,
        orders.sauce,
        orders.toppings,
        orders.notes,
        orders.delivery_status,
        orders.total_cents
      FROM pizza_tasks task
      JOIN pizza_orders orders
        ON orders.id = task.order_id
       AND orders.session_id = task.session_id
      WHERE task.session_id = ?
      ORDER BY task.position, task.id`,
      sid,
    );
  },

  listKitchenCards(session) {
    const rows = this.listTasks(session);
    const cards = [];
    const doneByOrder = {};

    rows.forEach((task) => {
      if (task.status !== "done") {
        cards.push(task);
        return;
      }

      const key = String(task.order_id);
      if (!doneByOrder[key]) {
        doneByOrder[key] = {
          card_id: "done-order-" + task.order_id,
          kind: "doneGroup",
          order_id: task.order_id,
          customer: task.customer,
          size: task.size,
          crust: task.crust,
          sauce: task.sauce,
          toppings: task.toppings,
          status: "done",
          station: "Complete",
          position: Number(task.position || 0),
          tasks: [],
        };
      }

      doneByOrder[key].position = Math.min(
        Number(doneByOrder[key].position || 0),
        Number(task.position || 0),
      );
      doneByOrder[key].tasks.push(task);
    });

    Object.keys(doneByOrder).forEach((key) => {
      const group = doneByOrder[key];
      group.task_count = group.tasks.length;
      group.done_titles = group.tasks.map((task) => task.title).join(", ");
      cards.push(group);
    });

    return cards;
  },

  moveKitchenTask(event) {
    const sid = Pizza.util.sessionId(event.session);

    if (String(event.cardId || "").startsWith("done-order-")) {
      return {
        ok: false,
        error:
          "Done groups are summaries. Drag individual tasks before they are complete.",
      };
    }

    const taskId = Number(event.cardId);
    const toStatus = Pizza.util.validColumn(
      Pizza.columns.kitchen,
      event.to && event.to.columnId,
      "blocked",
    );
    const toIndex = Number((event.to && event.to.index) || 0);

    const task = db.query(
      "SELECT * FROM pizza_tasks WHERE session_id = ? AND id = ?",
      sid,
      taskId,
    )[0];

    if (!task) {
      return { ok: false, error: "Task not found" };
    }

    if (
      ["ready", "working", "done"].includes(toStatus) &&
      !this.dependenciesAreDone(task)
    ) {
      return {
        ok: false,
        error: "Dependencies first: " + Pizza.util.dependencyText(task.deps),
      };
    }

    if (task.status === "blocked" && toStatus === "done") {
      return {
        ok: false,
        error: "Blocked tasks must become Ready or Working before Done.",
      };
    }

    const fromStatus = task.status;
    const destination = db.query(
      `SELECT * FROM pizza_tasks
       WHERE session_id = ? AND status = ? AND id != ?
       ORDER BY position, id`,
      sid,
      toStatus,
      taskId,
    );

    destination.splice(
      Math.max(0, Math.min(toIndex, destination.length)),
      0,
      task,
    );

    db.exec(
      "UPDATE pizza_tasks SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE session_id = ? AND id = ?",
      toStatus,
      sid,
      taskId,
    );

    destination.forEach((row, index) => {
      db.exec(
        "UPDATE pizza_tasks SET position = ? WHERE session_id = ? AND id = ?",
        (index + 1) * 10,
        sid,
        row.id,
      );
    });

    if (fromStatus !== toStatus) this.normalizeTasks(sid, fromStatus);
    this.refreshOrder(sid, task.order_id);

    return { ok: true, refresh: true, toast: "Kitchen task updated" };
  },

  moveDeliveryOrder(event) {
    const sid = Pizza.util.sessionId(event.session);
    const orderId = Number(event.cardId);
    const toStatus = Pizza.util.validColumn(
      Pizza.columns.delivery,
      event.to && event.to.columnId,
      "waiting",
    );
    const toIndex = Number((event.to && event.to.index) || 0);

    const order = db.query(
      "SELECT * FROM pizza_orders WHERE session_id = ? AND id = ?",
      sid,
      orderId,
    )[0];

    if (!order) {
      return { ok: false, error: "Order not found" };
    }

    if (
      Pizza.util.deliveryRank(toStatus) >= Pizza.util.deliveryRank("quality")
    ) {
      if (!this.kitchenDone(sid, orderId)) {
        return {
          ok: false,
          error: "Kitchen tasks must all be done before quality or delivery.",
        };
      }
    }

    if (toStatus === "paid" && !Number(order.paid || 0)) {
      return {
        ok: false,
        error: "Use the payment form on the card to capture tip and mark paid.",
      };
    }

    const fromStatus = order.delivery_status;
    const destination = db.query(
      `SELECT * FROM pizza_orders
       WHERE session_id = ? AND delivery_status = ? AND id != ?
       ORDER BY position, id`,
      sid,
      toStatus,
      orderId,
    );

    destination.splice(
      Math.max(0, Math.min(toIndex, destination.length)),
      0,
      order,
    );

    db.exec(
      `UPDATE pizza_orders
       SET delivery_status = ?, updated_at = CURRENT_TIMESTAMP
       WHERE session_id = ? AND id = ?`,
      toStatus,
      sid,
      orderId,
    );

    destination.forEach((row, index) => {
      db.exec(
        "UPDATE pizza_orders SET position = ? WHERE session_id = ? AND id = ?",
        (index + 1) * 10,
        sid,
        row.id,
      );
    });

    if (fromStatus !== toStatus) this.normalizeOrders(sid, fromStatus);

    return { ok: true, refresh: true, toast: "Delivery status updated" };
  },

  payOrder(session, orderId, tipDollars) {
    const sid = Pizza.util.sessionId(session);
    const id = Number(orderId);

    if (!this.kitchenDone(sid, id)) {
      throw new Error("Kitchen is not done yet");
    }

    const tipCents = Math.max(0, Math.round(Number(tipDollars || 0) * 100));

    db.exec(
      `UPDATE pizza_orders
       SET paid = 1,
           tip_cents = ?,
           delivery_status = 'paid',
           updated_at = CURRENT_TIMESTAMP
       WHERE session_id = ? AND id = ?`,
      tipCents,
      sid,
      id,
    );
  },

  tally(session) {
    const orders = this.listOrders(session);

    return {
      orders: orders.length,
      delivered: orders.filter(
        (order) =>
          Pizza.util.deliveryRank(order.delivery_status) >=
          Pizza.util.deliveryRank("delivered"),
      ).length,
      paid: orders.filter((order) => Number(order.paid || 0)).length,
      revenueCents: orders
        .filter((order) => Number(order.paid || 0))
        .reduce((sum, order) => sum + Number(order.total_cents || 0), 0),
      tipCents: orders
        .filter((order) => Number(order.paid || 0))
        .reduce((sum, order) => sum + Number(order.tip_cents || 0), 0),
    };
  },

  dependenciesAreDone(task) {
    const deps = String(task.deps || "")
      .split(",")
      .filter(Boolean);
    if (deps.length === 0) return true;

    return deps.every((code) => {
      const row = db.query(
        "SELECT status FROM pizza_tasks WHERE session_id = ? AND order_id = ? AND code = ?",
        task.session_id,
        task.order_id,
        code,
      )[0];
      return row && row.status === "done";
    });
  },

  kitchenDone(sid, orderId) {
    const row = db.query(
      `SELECT COUNT(*) AS count
       FROM pizza_tasks
       WHERE session_id = ? AND order_id = ? AND status != 'done'`,
      sid,
      orderId,
    )[0];
    return Number((row && row.count) || 0) === 0;
  },

  refreshOrder(sid, orderId) {
    const tasks = db.query(
      "SELECT * FROM pizza_tasks WHERE session_id = ? AND order_id = ? ORDER BY position",
      sid,
      orderId,
    );

    tasks.forEach((task) => {
      if (task.status === "working" || task.status === "done") return;

      db.exec(
        "UPDATE pizza_tasks SET status = ? WHERE session_id = ? AND id = ?",
        this.dependenciesAreDone(task) ? "ready" : "blocked",
        sid,
        task.id,
      );
    });

    const updatedTasks = db.query(
      "SELECT * FROM pizza_tasks WHERE session_id = ? AND order_id = ?",
      sid,
      orderId,
    );
    const order = db.query(
      "SELECT * FROM pizza_orders WHERE session_id = ? AND id = ?",
      sid,
      orderId,
    )[0];

    if (!order) return;

    const allDone =
      updatedTasks.length > 0 &&
      updatedTasks.every((task) => task.status === "done");
    const anyStarted = updatedTasks.some(
      (task) => task.status === "working" || task.status === "done",
    );

    if (
      allDone &&
      Pizza.util.deliveryRank(order.delivery_status) <
        Pizza.util.deliveryRank("quality")
    ) {
      db.exec(
        `UPDATE pizza_orders
         SET delivery_status = 'quality', updated_at = CURRENT_TIMESTAMP
         WHERE session_id = ? AND id = ?`,
        sid,
        orderId,
      );
    } else if (
      anyStarted &&
      Pizza.util.deliveryRank(order.delivery_status) <
        Pizza.util.deliveryRank("cooking")
    ) {
      db.exec(
        `UPDATE pizza_orders
         SET delivery_status = 'cooking', updated_at = CURRENT_TIMESTAMP
         WHERE session_id = ? AND id = ?`,
        sid,
        orderId,
      );
    }
  },

  nextOrderPosition(sid, status) {
    const row = db.query(
      "SELECT COALESCE(MAX(position), 0) AS max_position FROM pizza_orders WHERE session_id = ? AND delivery_status = ?",
      sid,
      status,
    )[0];
    return Number((row && row.max_position) || 0) + 10;
  },

  normalizeTasks(sid, status) {
    db.query(
      "SELECT id FROM pizza_tasks WHERE session_id = ? AND status = ? ORDER BY position, id",
      sid,
      status,
    ).forEach((task, index) => {
      db.exec(
        "UPDATE pizza_tasks SET position = ? WHERE session_id = ? AND id = ?",
        (index + 1) * 10,
        sid,
        task.id,
      );
    });
  },

  normalizeOrders(sid, status) {
    db.query(
      "SELECT id FROM pizza_orders WHERE session_id = ? AND delivery_status = ? ORDER BY position, id",
      sid,
      status,
    ).forEach((order, index) => {
      db.exec(
        "UPDATE pizza_orders SET position = ? WHERE session_id = ? AND id = ?",
        (index + 1) * 10,
        sid,
        order.id,
      );
    });
  },
};

Pizza.store.migrate();
