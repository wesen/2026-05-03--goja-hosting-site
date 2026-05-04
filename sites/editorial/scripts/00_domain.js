globalThis.Herald = globalThis.Herald || {};

Herald.columns = [
  ["pitches", "Pitches"],
  ["reporting", "Reporting"],
  ["writing", "Writing"],
  ["editing", "Editing"],
  ["ready", "Ready for Print"],
];

Herald.desks = [
  { id: "city", label: "City Hall", nav: "News Desk" },
  { id: "society", label: "Society", nav: "Society Desk" },
  { id: "business", label: "Business", nav: "Business Desk" },
  { id: "education", label: "Education", nav: "Education Desk" },
  { id: "technology", label: "Technology", nav: "Technology Desk" },
  { id: "environment", label: "Environment", nav: "Environment Desk" },
];

Herald.staffSeed = [
  {
    key: "blackwell",
    name: "E. Blackwell",
    role: "Editor in Chief",
    initials: "EB",
    portrait: "♟",
  },
  {
    key: "whitmore",
    name: "J. Whitmore",
    role: "City Correspondent",
    initials: "JW",
    portrait: "♜",
  },
  {
    key: "davenport",
    name: "M. Davenport",
    role: "Features Editor",
    initials: "MD",
    portrait: "♞",
  },
  {
    key: "meredith",
    name: "T. Meredith",
    role: "Labor Reporter",
    initials: "TM",
    portrait: "♝",
  },
  {
    key: "collins",
    name: "A. Collins",
    role: "Night Desk",
    initials: "AC",
    portrait: "♛",
  },
  {
    key: "beaumont",
    name: "L. Beaumont",
    role: "Society Columnist",
    initials: "LB",
    portrait: "♚",
  },
  {
    key: "fletcher",
    name: "R. Fletcher",
    role: "Business Desk",
    initials: "RF",
    portrait: "♙",
  },
];

Herald.util = {
  sessionId(session) {
    if (typeof session === "string") return session;
    return String((session && session.id) || "default");
  },

  validStatus(status) {
    return Herald.columns.some(([id]) => id === status) ? status : "pitches";
  },

  statusLabel(status) {
    const found = Herald.columns.find(([id]) => id === status);
    return found ? found[1] : status;
  },

  deskLabel(desk) {
    const found = Herald.desks.find(
      (item) => item.id === desk || item.label === desk,
    );
    return found ? found.label : desk;
  },

  splitTags(value) {
    return String(value || "")
      .split(",")
      .map((tag) => tag.trim())
      .filter(Boolean);
  },

  storySearchText(story) {
    return [
      story.title,
      story.dek,
      story.description,
      story.desk_label || story.desk,
      story.author_name,
      story.priority,
      story.tags,
      story.workflow_status,
    ]
      .join(" ")
      .toLowerCase();
  },
};
