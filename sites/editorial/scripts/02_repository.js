const db = require("database");

globalThis.Herald = globalThis.Herald || {};

Herald.repo = {
  migrate() {
    db.exec(`CREATE TABLE IF NOT EXISTS herald_staff (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      session_id TEXT NOT NULL DEFAULT 'default',
      staff_key TEXT NOT NULL,
      name TEXT NOT NULL,
      role TEXT NOT NULL,
      initials TEXT NOT NULL,
      portrait TEXT NOT NULL
    )`);

    db.exec(`CREATE TABLE IF NOT EXISTS herald_stories (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      session_id TEXT NOT NULL DEFAULT 'default',
      title TEXT NOT NULL,
      dek TEXT NOT NULL DEFAULT '',
      description TEXT NOT NULL DEFAULT '',
      workflow_status TEXT NOT NULL DEFAULT 'pitches',
      position INTEGER NOT NULL DEFAULT 0,
      desk TEXT NOT NULL DEFAULT 'city',
      author_key TEXT NOT NULL DEFAULT 'blackwell',
      priority TEXT NOT NULL DEFAULT 'Normal',
      due_date TEXT NOT NULL DEFAULT '',
      tags TEXT NOT NULL DEFAULT '',
      source_notes TEXT NOT NULL DEFAULT '',
      word_target INTEGER NOT NULL DEFAULT 800,
      word_count INTEGER NOT NULL DEFAULT 0,
      created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
    )`);

    db.exec(`CREATE TABLE IF NOT EXISTS herald_checklist (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      session_id TEXT NOT NULL DEFAULT 'default',
      story_id INTEGER NOT NULL,
      label TEXT NOT NULL,
      done INTEGER NOT NULL DEFAULT 0,
      position INTEGER NOT NULL DEFAULT 0
    )`);
  },

  countStories(sid) {
    const row = db.query(
      "SELECT COUNT(*) AS count FROM herald_stories WHERE session_id = ?",
      sid,
    )[0];
    return Number((row && row.count) || 0);
  },

  insertStaff(sid, staff) {
    db.exec(
      "INSERT INTO herald_staff(session_id, staff_key, name, role, initials, portrait) VALUES (?, ?, ?, ?, ?, ?)",
      sid,
      staff.key,
      staff.name,
      staff.role,
      staff.initials,
      staff.portrait,
    );
  },

  insertStory(sid, story) {
    db.exec(
      `INSERT INTO herald_stories(
        session_id, title, dek, description, workflow_status, position, desk,
        author_key, priority, due_date, tags, source_notes, word_target, word_count
      ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
      sid,
      story.title,
      story.dek,
      story.description,
      story.status,
      story.position,
      story.desk,
      story.authorKey,
      story.priority,
      story.dueDate,
      story.tags,
      story.sourceNotes,
      story.wordTarget,
      story.wordCount,
    );
    return db.query(
      "SELECT * FROM herald_stories WHERE session_id = ? ORDER BY id DESC LIMIT 1",
      sid,
    )[0];
  },

  insertChecklist(sid, storyId, item, index) {
    db.exec(
      "INSERT INTO herald_checklist(session_id, story_id, label, done, position) VALUES (?, ?, ?, ?, ?)",
      sid,
      storyId,
      item.label,
      item.done ? 1 : 0,
      (index + 1) * 10,
    );
  },

  listStories(sid) {
    return db.query(
      `SELECT stories.*, staff.name AS author_name, staff.role AS author_role,
              staff.initials AS author_initials, staff.portrait AS author_portrait
       FROM herald_stories stories
       LEFT JOIN herald_staff staff
         ON staff.session_id = stories.session_id
        AND staff.staff_key = stories.author_key
       WHERE stories.session_id = ?
       ORDER BY stories.position, stories.id`,
      sid,
    );
  },

  listStoriesByStatus(sid, status, exceptId) {
    return db.query(
      `SELECT * FROM herald_stories
       WHERE session_id = ? AND workflow_status = ? AND id != ?
       ORDER BY position, id`,
      sid,
      status,
      Number(exceptId || 0),
    );
  },

  getStory(sid, id) {
    return db.query(
      `SELECT stories.*, staff.name AS author_name, staff.role AS author_role,
              staff.initials AS author_initials, staff.portrait AS author_portrait
       FROM herald_stories stories
       LEFT JOIN herald_staff staff
         ON staff.session_id = stories.session_id
        AND staff.staff_key = stories.author_key
       WHERE stories.session_id = ? AND stories.id = ?`,
      sid,
      Number(id),
    )[0];
  },

  firstStory(sid) {
    return db.query(
      `SELECT * FROM herald_stories
       WHERE session_id = ?
       ORDER BY CASE workflow_status WHEN 'writing' THEN 0 WHEN 'editing' THEN 1 ELSE 2 END, position, id
       LIMIT 1`,
      sid,
    )[0];
  },

  listChecklist(sid, storyId) {
    return db.query(
      "SELECT * FROM herald_checklist WHERE session_id = ? AND story_id = ? ORDER BY position, id",
      sid,
      Number(storyId),
    );
  },

  toggleChecklist(sid, storyId, itemId) {
    const row = db.query(
      "SELECT done FROM herald_checklist WHERE session_id = ? AND story_id = ? AND id = ?",
      sid,
      Number(storyId),
      Number(itemId),
    )[0];
    if (!row) return;
    db.exec(
      "UPDATE herald_checklist SET done = ? WHERE session_id = ? AND story_id = ? AND id = ?",
      Number(row.done || 0) ? 0 : 1,
      sid,
      Number(storyId),
      Number(itemId),
    );
  },

  setStoryStatus(sid, storyId, status) {
    db.exec(
      "UPDATE herald_stories SET workflow_status = ?, updated_at = CURRENT_TIMESTAMP WHERE session_id = ? AND id = ?",
      status,
      sid,
      Number(storyId),
    );
  },

  setStoryPosition(sid, storyId, position) {
    db.exec(
      "UPDATE herald_stories SET position = ? WHERE session_id = ? AND id = ?",
      Number(position),
      sid,
      Number(storyId),
    );
  },
};

Herald.repo.migrate();
