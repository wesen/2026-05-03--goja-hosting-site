globalThis.Herald = globalThis.Herald || {};

Herald.styles = `
:root {
  --paper: #f5f0e7;
  --panel: #fbf7ef;
  --ink: #16110c;
  --muted: #6d6254;
  --line: #d8cec0;
  --accent: #315f7f;
  --button: #17130f;
  font-family: Georgia, "Times New Roman", serif;
}
* { box-sizing: border-box; }
body { margin: 0; color: var(--ink); background: radial-gradient(circle at top left, #fffdf8, var(--paper)); }
a { color: inherit; text-decoration: none; }
.page { min-height: 100vh; display: grid; grid-template-columns: 245px minmax(780px, 1fr) 360px; }
.sidebar { border-right: 1px solid var(--line); padding: 26px 18px; background: rgba(244, 238, 228, .78); display: flex; flex-direction: column; gap: 24px; }
.building { height: 92px; border: 1px solid var(--line); display: grid; place-items: center; font-size: 48px; background: #fffaf1; }
.nav-group { display: grid; gap: 6px; }
.nav-title { margin: 14px 0 8px; color: var(--muted); font-size: 12px; letter-spacing: .18em; text-transform: uppercase; }
.nav-item { display: flex; gap: 12px; align-items: center; padding: 11px 12px; border-radius: 8px; font-size: 16px; }
.nav-item.active { background: #ebe3d7; }
.editor-footer { margin-top: auto; border-top: 1px solid var(--line); padding-top: 18px; display: grid; gap: 14px; }
.profile { display: flex; align-items: center; gap: 12px; }
.avatar { width: 38px; height: 38px; border-radius: 50%; background: #30261d; color: #fff7ec; display: grid; place-items: center; font-size: 18px; border: 2px solid #bfb09e; }
.main { min-width: 0; }
.topbar { height: 138px; border-bottom: 1px solid var(--line); display: flex; align-items: center; justify-content: space-between; padding: 0 40px; }
.masthead h1 { margin: 0; font-size: 33px; line-height: 1; }
.masthead p { margin: 9px 0 0; color: var(--muted); font-size: 17px; }
.global-actions { display: flex; align-items: center; gap: 22px; }
.search { width: 440px; min-height: 48px; border: 1px solid var(--line); border-radius: 8px; background: #fffaf4; padding: 0 16px 0 42px; font: 16px Georgia, serif; }
.search-wrap { position: relative; }
.search-wrap::before { content: "⌕"; position: absolute; left: 16px; top: 9px; font-size: 24px; color: var(--muted); }
.icon-button { border: 0; background: transparent; font-size: 24px; }
.content { padding: 30px 32px 26px; }
.workflow-head { display: flex; justify-content: space-between; align-items: end; gap: 20px; margin-bottom: 22px; }
.workflow-head h2 { margin: 0; font-size: 36px; line-height: 1; }
.tabs { display: flex; gap: 28px; margin-top: 28px; }
.tab { display: inline-flex; gap: 9px; align-items: center; padding-bottom: 12px; border-bottom: 2px solid transparent; color: var(--muted); }
.tab.active { color: var(--ink); border-color: var(--ink); }
.board-tools { display: flex; align-items: center; gap: 14px; color: var(--muted); }
.new-story { background: var(--button); color: #fffaf4; border: 0; border-radius: 7px; min-height: 48px; padding: 0 20px; font: 16px Georgia, serif; }
.new-story-form { display: grid; grid-template-columns: 1fr 1.2fr 130px 130px auto; gap: 10px; border: 1px solid var(--line); padding: 12px; margin-bottom: 18px; background: rgba(255,250,244,.72); }
.new-story-form input, .new-story-form select, .new-story-form button { min-height: 38px; border: 1px solid var(--line); background: #fffaf4; border-radius: 6px; padding: 0 10px; font: 15px Georgia, serif; }
.new-story-form button { background: var(--button); color: white; }
.board { display: grid; grid-template-columns: repeat(5, minmax(180px, 1fr)); gap: 10px; }
.column { min-height: 740px; border-left: 1px solid var(--line); border-right: 1px solid #eee5d9; background: rgba(255,250,244,.45); }
.column-header { display: flex; justify-content: space-between; align-items: center; padding: 17px 16px; }
.column-header h2 { margin: 0; font-size: 13px; letter-spacing: .14em; text-transform: uppercase; }
.count { min-width: 24px; min-height: 24px; border-radius: 50%; background: #e9e1d4; display: inline-grid; place-items: center; font-size: 13px; }
.card-list { min-height: 650px; padding: 0 10px 20px; }
.kanban-card { background: #fffaf4; border: 1px solid var(--line); border-radius: 7px; padding: 16px 14px; margin-bottom: 10px; cursor: grab; }
.kanban-card.selected { border-color: var(--accent); box-shadow: 0 0 0 1px var(--accent) inset; }
.story-card h3 { margin: 0 0 14px; font-size: 17px; line-height: 1.35; font-weight: 500; }
.story-meta { color: var(--muted); font-size: 14px; line-height: 1.55; }
.card-foot { display: flex; justify-content: space-between; align-items: end; margin-top: 10px; }
.progress { display: flex; gap: 4px; align-items: center; color: var(--muted); font-size: 13px; }
.ready-mark { width: 18px; height: 18px; border: 1px solid #9d927f; border-radius: 50%; display: inline-grid; place-items: center; }
.add-story-link { color: var(--muted); text-align: center; padding: 16px; }
.empty { color: var(--muted); padding: 12px; }
.dossier { border-left: 1px solid var(--line); padding: 34px 32px; background: rgba(250,246,238,.82); transition: opacity .12s ease; }
.dossier[aria-busy="true"] { opacity: .58; }
.close { float: right; font-size: 28px; color: var(--muted); }
.status-label { color: var(--muted); letter-spacing: .18em; text-transform: uppercase; font-size: 13px; }
.dossier h2 { font-size: 27px; line-height: 1.35; margin: 16px 0 8px; font-weight: 500; }
.desk-line { color: var(--muted); font-size: 17px; }
.dossier-section { border-top: 1px solid var(--line); margin-top: 26px; padding-top: 24px; }
.author-row, .fact-row { display: flex; align-items: center; gap: 13px; margin-top: 22px; }
.section-title { color: var(--muted); letter-spacing: .18em; text-transform: uppercase; font-size: 12px; margin: 0 0 16px; }
.description { font-size: 17px; line-height: 1.55; }
.tags { display: flex; gap: 10px; flex-wrap: wrap; }
.tag { border: 1px solid var(--line); border-radius: 7px; background: #eee6dc; padding: 9px 13px; }
a.tag:hover { background: #ded1c0; }
.active-filter { display: inline-flex; align-items: center; gap: 10px; margin: 0 0 18px; border: 1px solid var(--line); border-radius: 999px; background: #fffaf4; padding: 8px 13px; color: var(--muted); }
.table-wrap { border: 1px solid var(--line); background: rgba(255,250,244,.65); overflow: auto; }
.story-table { width: 100%; border-collapse: collapse; font-size: 15px; }
.story-table th { text-align: left; letter-spacing: .12em; text-transform: uppercase; font-size: 12px; color: var(--muted); background: #eee6dc; }
.story-table th, .story-table td { border-bottom: 1px solid var(--line); padding: 13px 14px; vertical-align: top; }
.story-table tr:hover td { background: #fffaf4; }
.checklist { display: grid; gap: 13px; }
.check-item { display: flex; gap: 11px; align-items: center; border: 0; background: transparent; font: 15px Georgia, serif; text-align: left; padding: 0; }
.box { width: 18px; height: 18px; border: 1px solid #9d927f; display: inline-grid; place-items: center; border-radius: 3px; }
.box.done { background: var(--ink); color: white; }
.assignment-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; font-size: 14px; }
.assignment-grid strong { display: block; font-size: 18px; margin-top: 2px; }
.open-full { width: 100%; min-height: 58px; margin-top: 90px; border: 1px solid var(--line); border-radius: 7px; background: #fffaf4; font: 17px Georgia, serif; }
@media (max-width: 1300px) { .page { grid-template-columns: 1fr; } .sidebar, .dossier { display:none; } .board { grid-template-columns: repeat(2, minmax(220px, 1fr)); } }
`;
