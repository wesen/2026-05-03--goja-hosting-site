const db = require("database");

globalThis.Pizza = globalThis.Pizza || {};

Pizza.repo = {
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

  countOrders(sid) {
    const row = db.query(
      "SELECT COUNT(*) AS count FROM pizza_orders WHERE session_id = ?",
      sid,
    )[0];
    return Number((row && row.count) || 0);
  },

  insertOrder(sid, order) {
    db.exec(
      `INSERT INTO pizza_orders(
        session_id, customer, address, size, crust, sauce, toppings, notes,
        delivery_status, total_cents, position
      ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'waiting', ?, ?)`,
      sid,
      order.customer,
      order.address,
      order.size,
      order.crust,
      order.sauce,
      order.toppings,
      order.notes,
      order.totalCents,
      order.position,
    );

    return db.query(
      "SELECT * FROM pizza_orders WHERE session_id = ? ORDER BY id DESC LIMIT 1",
      sid,
    )[0];
  },

  insertTask(sid, task) {
    db.exec(
      `INSERT INTO pizza_tasks(
        session_id, order_id, code, title, description, station, deps, status, position
      ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
      sid,
      task.orderId,
      task.code,
      task.title,
      task.description,
      task.station,
      task.deps,
      task.status,
      task.position,
    );
  },

  getOrder(sid, orderId) {
    return db.query(
      "SELECT * FROM pizza_orders WHERE session_id = ? AND id = ?",
      sid,
      Number(orderId),
    )[0];
  },

  listOrders(sid) {
    return db.query(
      "SELECT * FROM pizza_orders WHERE session_id = ? ORDER BY position, id",
      sid,
    );
  },

  listOrdersByStatus(sid, status, exceptId) {
    return db.query(
      `SELECT * FROM pizza_orders
       WHERE session_id = ? AND delivery_status = ? AND id != ?
       ORDER BY position, id`,
      sid,
      status,
      Number(exceptId || 0),
    );
  },

  getTask(sid, taskId) {
    return db.query(
      "SELECT * FROM pizza_tasks WHERE session_id = ? AND id = ?",
      sid,
      Number(taskId),
    )[0];
  },

  getTaskByCode(sid, orderId, code) {
    return db.query(
      "SELECT * FROM pizza_tasks WHERE session_id = ? AND order_id = ? AND code = ?",
      sid,
      Number(orderId),
      code,
    )[0];
  },

  listTasksForOrder(sid, orderId) {
    return db.query(
      "SELECT * FROM pizza_tasks WHERE session_id = ? AND order_id = ? ORDER BY position, id",
      sid,
      Number(orderId),
    );
  },

  listTasksWithOrders(sid) {
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

  listTasksByStatus(sid, status, exceptId) {
    return db.query(
      `SELECT * FROM pizza_tasks
       WHERE session_id = ? AND status = ? AND id != ?
       ORDER BY position, id`,
      sid,
      status,
      Number(exceptId || 0),
    );
  },

  setTaskStatus(sid, taskId, status) {
    db.exec(
      "UPDATE pizza_tasks SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE session_id = ? AND id = ?",
      status,
      sid,
      Number(taskId),
    );
  },

  setTaskPosition(sid, taskId, position) {
    db.exec(
      "UPDATE pizza_tasks SET position = ? WHERE session_id = ? AND id = ?",
      Number(position),
      sid,
      Number(taskId),
    );
  },

  setOrderDeliveryStatus(sid, orderId, status) {
    db.exec(
      `UPDATE pizza_orders
       SET delivery_status = ?, updated_at = CURRENT_TIMESTAMP
       WHERE session_id = ? AND id = ?`,
      status,
      sid,
      Number(orderId),
    );
  },

  setOrderPosition(sid, orderId, position) {
    db.exec(
      "UPDATE pizza_orders SET position = ? WHERE session_id = ? AND id = ?",
      Number(position),
      sid,
      Number(orderId),
    );
  },

  markOrderPaid(sid, orderId, tipCents) {
    db.exec(
      `UPDATE pizza_orders
       SET paid = 1,
           tip_cents = ?,
           delivery_status = 'paid',
           updated_at = CURRENT_TIMESTAMP
       WHERE session_id = ? AND id = ?`,
      Number(tipCents),
      sid,
      Number(orderId),
    );
  },

  nextOrderPosition(sid, status) {
    const row = db.query(
      "SELECT COALESCE(MAX(position), 0) AS max_position FROM pizza_orders WHERE session_id = ? AND delivery_status = ?",
      sid,
      status,
    )[0];
    return Number((row && row.max_position) || 0) + 10;
  },
};

Pizza.repo.migrate();
