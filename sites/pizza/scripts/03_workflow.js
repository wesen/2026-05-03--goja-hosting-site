globalThis.Pizza = globalThis.Pizza || {};

Pizza.store = {
  seedIfEmpty(session) {
    const sid = Pizza.util.sessionId(session);
    if (Pizza.repo.countOrders(sid) > 0) return;

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

    ["stretch", "sauce", "cheese"].forEach((code) => {
      const task = Pizza.repo.getTaskByCode(sid, first.id, code);
      if (task) Pizza.repo.setTaskStatus(sid, task.id, "done");
    });

    this.refreshOrder(sid, first.id);
  },

  createOrder(session, body) {
    const sid = Pizza.util.sessionId(session);
    const toppings = Pizza.util
      .selectedList(body.toppings || body.topping)
      .filter(Boolean);
    const size = String(body.size || "medium");

    const order = Pizza.repo.insertOrder(sid, {
      customer: String(body.customer || "Guest"),
      address: String(body.address || "Pickup counter"),
      size,
      crust: String(body.crust || "neapolitan"),
      sauce: String(body.sauce || "tomato"),
      toppings: toppings.join(","),
      notes: String(body.notes || ""),
      totalCents: Pizza.util.priceFor(size, toppings),
      position: Pizza.repo.nextOrderPosition(sid, "waiting"),
    });

    Pizza.taskTemplate.forEach((task, index) => {
      Pizza.repo.insertTask(sid, {
        orderId: order.id,
        code: task.code,
        title: task.title,
        description: task.description,
        station: task.station,
        deps: task.dependencies.join(","),
        status: task.dependencies.length === 0 ? "ready" : "blocked",
        position: (index + 1) * 10,
      });
    });

    return order;
  },

  listOrders(session) {
    const sid = Pizza.util.sessionId(session);
    this.seedIfEmpty(sid);
    this.refreshAllOrders(sid);
    return Pizza.repo.listOrders(sid);
  },

  listTasks(session) {
    const sid = Pizza.util.sessionId(session);
    this.seedIfEmpty(sid);
    this.refreshAllOrders(sid);
    return Pizza.repo.listTasksWithOrders(sid);
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
    const task = Pizza.repo.getTask(sid, taskId);

    if (!task) return { ok: false, error: "Task not found" };

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

    this.repositionTask(sid, task, toStatus, toIndex);
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
    const order = Pizza.repo.getOrder(sid, orderId);

    if (!order) return { ok: false, error: "Order not found" };

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

    this.repositionOrder(sid, order, toStatus, toIndex);

    return { ok: true, refresh: true, toast: "Delivery status updated" };
  },

  payOrder(session, orderId, tipDollars) {
    const sid = Pizza.util.sessionId(session);
    const id = Number(orderId);

    if (!this.kitchenDone(sid, id)) {
      throw new Error("Kitchen is not done yet");
    }

    const tipCents = Math.max(0, Math.round(Number(tipDollars || 0) * 100));
    Pizza.repo.markOrderPaid(sid, id, tipCents);
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
      const dependency = Pizza.repo.getTaskByCode(
        task.session_id,
        task.order_id,
        code,
      );
      return dependency && dependency.status === "done";
    });
  },

  kitchenDone(sid, orderId) {
    const tasks = Pizza.repo.listTasksForOrder(sid, orderId);
    return tasks.length > 0 && tasks.every((task) => task.status === "done");
  },

  refreshAllOrders(sid) {
    Pizza.repo
      .listOrders(sid)
      .forEach((order) => this.refreshOrder(sid, order.id));
  },

  refreshOrder(sid, orderId) {
    const tasks = Pizza.repo.listTasksForOrder(sid, orderId);

    tasks.forEach((task) => {
      if (task.status === "working" || task.status === "done") return;
      Pizza.repo.setTaskStatus(
        sid,
        task.id,
        this.dependenciesAreDone(task) ? "ready" : "blocked",
      );
    });

    const refreshedTasks = Pizza.repo.listTasksForOrder(sid, orderId);
    const order = Pizza.repo.getOrder(sid, orderId);
    if (!order) return;

    const allDone =
      refreshedTasks.length > 0 &&
      refreshedTasks.every((task) => task.status === "done");
    const anyLeftReady = refreshedTasks.some((task) => task.status === "ready");
    const anyStarted = refreshedTasks.some(
      (task) => task.status === "working" || task.status === "done",
    );
    const anyMovedOutOfReady =
      anyStarted || (refreshedTasks.length > 0 && !anyLeftReady);

    if (
      allDone &&
      Pizza.util.deliveryRank(order.delivery_status) <
        Pizza.util.deliveryRank("quality")
    ) {
      this.moveOrderStatusPreservingRelativePosition(sid, order, "quality");
      return;
    }

    if (
      anyMovedOutOfReady &&
      Pizza.util.deliveryRank(order.delivery_status) <
        Pizza.util.deliveryRank("cooking")
    ) {
      this.moveOrderStatusPreservingRelativePosition(sid, order, "cooking");
    }
  },

  moveOrderStatusPreservingRelativePosition(sid, order, toStatus) {
    const destination = Pizza.repo.listOrdersByStatus(sid, toStatus, order.id);
    destination.push(order);
    Pizza.repo.setOrderDeliveryStatus(sid, order.id, toStatus);
    destination.forEach((row, index) => {
      Pizza.repo.setOrderPosition(sid, row.id, (index + 1) * 10);
    });
    this.normalizeOrders(sid, order.delivery_status);
  },

  repositionTask(sid, task, toStatus, toIndex) {
    const fromStatus = task.status;
    const destination = Pizza.repo.listTasksByStatus(sid, toStatus, task.id);
    destination.splice(
      Math.max(0, Math.min(toIndex, destination.length)),
      0,
      task,
    );

    Pizza.repo.setTaskStatus(sid, task.id, toStatus);
    destination.forEach((row, index) => {
      Pizza.repo.setTaskPosition(sid, row.id, (index + 1) * 10);
    });

    if (fromStatus !== toStatus) this.normalizeTasks(sid, fromStatus);
  },

  repositionOrder(sid, order, toStatus, toIndex) {
    const fromStatus = order.delivery_status;
    const destination = Pizza.repo.listOrdersByStatus(sid, toStatus, order.id);
    destination.splice(
      Math.max(0, Math.min(toIndex, destination.length)),
      0,
      order,
    );

    Pizza.repo.setOrderDeliveryStatus(sid, order.id, toStatus);
    destination.forEach((row, index) => {
      Pizza.repo.setOrderPosition(sid, row.id, (index + 1) * 10);
    });

    if (fromStatus !== toStatus) this.normalizeOrders(sid, fromStatus);
  },

  normalizeTasks(sid, status) {
    Pizza.repo.listTasksByStatus(sid, status, 0).forEach((task, index) => {
      Pizza.repo.setTaskPosition(sid, task.id, (index + 1) * 10);
    });
  },

  normalizeOrders(sid, status) {
    Pizza.repo.listOrdersByStatus(sid, status, 0).forEach((order, index) => {
      Pizza.repo.setOrderPosition(sid, order.id, (index + 1) * 10);
    });
  },
};
