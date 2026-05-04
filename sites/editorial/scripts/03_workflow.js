globalThis.Herald = globalThis.Herald || {};

Herald.workflow = {
  seedIfEmpty(session) {
    const sid = Herald.util.sessionId(session);
    if (Herald.repo.countStories(sid) > 0) return;

    Herald.staffSeed.forEach((staff) => Herald.repo.insertStaff(sid, staff));

    const stories = [
      [
        "Gas streetlights to be phased out by 1890",
        "Environment",
        "The municipal lighting committee prepares a transition plan for electric lamps.",
        "pitches",
        "environment",
        "whitmore",
        "Normal",
        "May 24, 1887",
        "Environment,Utilities",
        700,
        120,
        [
          "Confirm committee minutes",
          "Find street map",
          "Interview lamp inspector",
        ],
      ],
      [
        "The rise of women in medicine",
        "Society",
        "A reported feature on the women entering hospital training programs.",
        "pitches",
        "society",
        "davenport",
        "High",
        "May 24, 1887",
        "Medicine,Society",
        1100,
        180,
        [
          "Interview Dr. Bell",
          "Request hospital figures",
          "Collect portrait engraving",
        ],
      ],
      [
        "New telegraph line connects the coast",
        "Business",
        "Investors celebrate a faster wire between the city and western exchanges.",
        "pitches",
        "business",
        "fletcher",
        "Low",
        "May 23, 1887",
        "Telegraph,Business",
        650,
        90,
        ["Confirm route", "Check tariffs", "Wire coastal office"],
      ],
      [
        "Dockworkers demand better conditions",
        "City News",
        "Longshoremen gather at the east pier to demand safer night shifts.",
        "reporting",
        "city",
        "meredith",
        "High",
        "May 25, 1887",
        "Labor,Docks",
        900,
        430,
        [
          "Interview union steward",
          "Visit pier office",
          "Confirm injury records",
        ],
      ],
      [
        "Fire breaks out in warehouse district",
        "City News",
        "A late-night fire damages three warehouses near Canal Street.",
        "reporting",
        "city",
        "collins",
        "High",
        "May 24, 1887",
        "Fire,City News",
        800,
        510,
        [
          "Speak with fire marshal",
          "Count damaged buildings",
          "Confirm insurance estimate",
        ],
      ],
      [
        "City Council debates new housing plan",
        "City Hall",
        "The City Council met today to discuss the proposed housing plan aimed at addressing overcrowding in the East End.",
        "writing",
        "city",
        "blackwell",
        "High",
        "May 25, 1887",
        "City Hall,Policy",
        950,
        680,
        [
          "Interview Alderman Pierce",
          "Review housing report",
          "Verify budget figures",
        ],
      ],
      [
        "The future of electric lighting in our city",
        "Technology",
        "A Sunday analysis of dynamos, arc lamps, and the promise of safer streets.",
        "writing",
        "technology",
        "whitmore",
        "Normal",
        "May 26, 1887",
        "Technology,Utilities",
        1000,
        520,
        ["Visit power station", "Check patent names", "Add explainer box"],
      ],
      [
        "School board approves new curriculum",
        "Education",
        "The school board passes a revised arithmetic and natural philosophy program.",
        "editing",
        "education",
        "davenport",
        "Normal",
        "May 24, 1887",
        "Education,Schools",
        780,
        760,
        ["Copy edit quotations", "Confirm board vote", "Typeset headline"],
      ],
      [
        "Mayor outlines plan for public parks",
        "City Hall",
        "The mayor promises new green spaces for crowded wards.",
        "ready",
        "city",
        "blackwell",
        "Normal",
        "May 22, 1887",
        "City Hall,Parks",
        720,
        720,
        ["Final proof", "Engraving placed", "Ready for forme"],
      ],
      [
        "Summer fashion arrives in the city",
        "Society",
        "Dressmakers report a rush for lighter fabrics and imported ribbons.",
        "ready",
        "society",
        "beaumont",
        "Low",
        "May 22, 1887",
        "Society,Fashion",
        620,
        620,
        ["Final proof", "Caption checked", "Ready for forme"],
      ],
    ];

    stories.forEach((row, index) => {
      const story = Herald.repo.insertStory(sid, {
        title: row[0],
        dek: row[1],
        description: row[2],
        status: row[3],
        desk: row[4],
        authorKey: row[5],
        priority: row[6],
        dueDate: row[7],
        tags: row[8],
        sourceNotes: "Filed from the city room ledger, edition 1887.",
        wordTarget: row[9],
        wordCount: row[10],
        position: (index + 1) * 10,
      });
      row[11].forEach((label, checklistIndex) =>
        Herald.repo.insertChecklist(
          sid,
          story.id,
          { label, done: checklistIndex === 0 && row[3] !== "pitches" },
          checklistIndex,
        ),
      );
    });
  },

  listStories(session, query) {
    const sid = Herald.util.sessionId(session);
    this.seedIfEmpty(sid);
    return Herald.repo
      .listStories(sid)
      .map((story) => this.decorateStory(sid, story))
      .filter((story) => this.matchesQuery(story, query || {}));
  },

  matchesQuery(story, query) {
    const tag = String(query.tag || "")
      .trim()
      .toLowerCase();
    if (tag) {
      const tags = Herald.util
        .splitTags(story.tags)
        .map((item) => item.toLowerCase());
      if (!tags.includes(tag)) return false;
    }

    const desk = String(query.desk || "").trim();
    if (desk && story.desk !== desk) return false;

    return true;
  },

  activeStory(session, query) {
    const sid = Herald.util.sessionId(session);
    this.seedIfEmpty(sid);
    const id = query && query.story ? Number(query.story) : 0;
    if (id) {
      const row = Herald.repo.getStory(sid, id);
      return row ? this.decorateStory(sid, row) : null;
    }
    const filtered = this.listStories(sid, query || {});
    if (filtered.length > 0) return filtered[0];
    const row = Herald.repo.firstStory(sid);
    return row ? this.decorateStory(sid, row) : null;
  },

  decorateStory(sid, story) {
    const checklist = Herald.repo.listChecklist(sid, story.id);
    const done = checklist.filter((item) => Number(item.done || 0)).length;
    return {
      ...story,
      desk_label: Herald.util.deskLabel(story.desk),
      checklist,
      checklist_done: done,
      checklist_total: checklist.length,
      progress_label: `${done}/${checklist.length}`,
    };
  },

  moveStory(event) {
    const sid = Herald.util.sessionId(event.session);
    const storyId = Number(event.cardId);
    const toStatus = Herald.util.validStatus(event.to && event.to.columnId);
    const toIndex = Number((event.to && event.to.index) || 0);
    const story = Herald.repo.getStory(sid, storyId);
    if (!story) return { ok: false, error: "Story not found" };

    const fromStatus = story.workflow_status;
    const destination = Herald.repo.listStoriesByStatus(sid, toStatus, storyId);
    destination.splice(
      Math.max(0, Math.min(toIndex, destination.length)),
      0,
      story,
    );
    Herald.repo.setStoryStatus(sid, storyId, toStatus);
    destination.forEach((row, index) =>
      Herald.repo.setStoryPosition(sid, row.id, (index + 1) * 10),
    );
    if (fromStatus !== toStatus) this.normalizeStatus(sid, fromStatus);
    return { ok: true, refresh: true, toast: "Story moved" };
  },

  normalizeStatus(sid, status) {
    Herald.repo
      .listStoriesByStatus(sid, status, 0)
      .forEach((story, index) =>
        Herald.repo.setStoryPosition(sid, story.id, (index + 1) * 10),
      );
  },

  toggleChecklist(session, storyId, itemId) {
    const sid = Herald.util.sessionId(session);
    Herald.repo.toggleChecklist(sid, storyId, itemId);
  },

  deskSummary(session) {
    const stories = this.listStories(session, {});
    return Herald.desks.map((desk) => ({
      ...desk,
      total: stories.filter((story) => story.desk === desk.id).length,
      ready: stories.filter(
        (story) => story.desk === desk.id && story.workflow_status === "ready",
      ).length,
    }));
  },

  metrics(session) {
    const stories = this.listStories(session, {});
    const checklistDone = stories.reduce(
      (sum, story) => sum + story.checklist_done,
      0,
    );
    const checklistTotal = stories.reduce(
      (sum, story) => sum + story.checklist_total,
      0,
    );
    return {
      active: stories.filter((story) => story.workflow_status !== "ready")
        .length,
      ready: stories.filter((story) => story.workflow_status === "ready")
        .length,
      highPriority: stories.filter((story) => story.priority === "High").length,
      checklistPercent: checklistTotal
        ? Math.round((checklistDone / checklistTotal) * 100)
        : 0,
    };
  },
};
