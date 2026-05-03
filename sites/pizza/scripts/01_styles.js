globalThis.Pizza = globalThis.Pizza || {};

Pizza.styles = `
:root {
  --ink: #0d0d0d;
  --paper: #fbfbfa;
  --panel: #ffffff;
  --soft: #f2f2ef;
  --muted: #666666;
  --line: #111111;
  font-family: "Courier New", "IBM Plex Mono", ui-monospace, monospace;
}

* {
  box-sizing: border-box;
}

body {
  margin: 0;
  color: var(--ink);
  background: var(--paper);
  font: 16px/1.45 "Courier New", ui-monospace, monospace;
}

[hidden] {
  display: none !important;
}

.page {
  max-width: 1780px;
  margin: 0 auto;
  padding: 22px 24px 30px;
}

.hero,
.pizza-builder,
.board-panel,
.api-links {
  background: var(--panel);
  border: 2px solid var(--line);
  border-radius: 4px;
}

.hero {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 28px;
  align-items: center;
  padding: 20px 28px;
  margin-bottom: 14px;
}

.hero-copy {
  display: flex;
  gap: 24px;
  align-items: center;
}

.mascot {
  width: 74px;
  height: 74px;
  display: grid;
  place-items: center;
  border: 3px solid var(--line);
  border-radius: 3px;
  box-shadow: 4px 4px 0 var(--line);
  background:
    radial-gradient(var(--line) 1px, transparent 1px),
    var(--panel);
  background-size: 4px 4px;
  font-size: 42px;
  line-height: 1;
}

.brand {
  margin: 0;
  font-size: 38px;
  line-height: 1;
  letter-spacing: -1px;
}

.subtitle {
  max-width: 700px;
  margin: 10px 0 0;
  font-size: 15px;
  line-height: 1.35;
}

.tally {
  display: grid;
  grid-template-columns: repeat(5, 150px);
  gap: 10px;
}

.tally-card {
  display: grid;
  grid-template-columns: 34px 1fr;
  gap: 10px;
  align-items: center;
  min-height: 62px;
  padding: 9px 12px;
  background: var(--panel);
  border: 2px solid var(--line);
  border-radius: 4px;
}

.tally-icon {
  display: grid;
  place-items: center;
  min-height: 36px;
  border-right: 1px solid var(--line);
  font-size: 22px;
}

.tally-card strong {
  display: block;
  font-size: 24px;
  line-height: 1;
  text-align: center;
}

.tally-card span:last-child {
  display: block;
  text-align: center;
}

.pizza-builder {
  padding: 16px 22px 20px;
  margin: 14px 0 16px;
}

.pizza-builder h2,
.board-heading h2 {
  margin: 0;
  font-size: 25px;
  line-height: 1.1;
}

.builder-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(180px, 1fr));
  gap: 14px 18px;
  margin-top: 12px;
}

.form-field {
  display: grid;
  gap: 6px;
  font-size: 15px;
  font-weight: 700;
}

.form-field.wide,
.topping-picker {
  grid-column: span 2;
}

.form-field input,
.form-field select,
.form-field textarea,
.pizza-builder button,
.payment-form input,
.payment-form button,
.kb-root input {
  min-height: 34px;
  padding: 6px 10px;
  color: var(--ink);
  font: inherit;
  background: var(--panel);
  border: 2px solid var(--line);
  border-radius: 3px;
}

.form-field input::placeholder,
.kb-root input::placeholder {
  color: #777777;
}

.form-field input:focus,
.form-field select:focus,
.kb-root input:focus {
  outline: 2px solid var(--line);
  outline-offset: 2px;
}

.topping-picker > strong {
  display: block;
  margin-bottom: 8px;
  font-size: 15px;
}

.topping-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(145px, 1fr));
  gap: 9px 20px;
}

.topping-option {
  display: flex;
  gap: 10px;
  align-items: center;
  font-size: 15px;
  font-weight: 700;
}

.topping-option input[type="checkbox"] {
  width: 17px;
  height: 17px;
  accent-color: var(--ink);
}

.primary-action {
  grid-column: 3 / span 2;
  justify-self: center;
  min-width: 520px;
  min-height: 64px !important;
  margin-top: 8px;
  color: var(--panel) !important;
  font-weight: 900;
  cursor: pointer;
  background:
    radial-gradient(#333333 0.8px, transparent 0.8px),
    var(--line) !important;
  background-size: 3px 3px !important;
  border: 3px solid var(--line) !important;
  border-radius: 3px !important;
  box-shadow: 3px 3px 0 var(--line) !important;
}

.primary-action::before {
  content: "♨  ";
}

.primary-action:active,
.payment-form button:active {
  transform: translate(1px, 1px);
  box-shadow: 1px 1px 0 var(--line) !important;
}

.boards {
  display: grid;
  grid-template-columns: 1fr;
  gap: 16px;
  align-items: start;
}

.board-panel {
  padding: 16px 22px 20px;
}

.board-heading {
  display: flex;
  gap: 18px;
  align-items: baseline;
  margin-bottom: 12px;
}

.hint {
  font-size: 15px;
  color: var(--ink);
}

.kb-root {
  padding: 0;
}

.kb-root > input,
.kb-root .search-form,
.kb-root [data-kb-search] {
  width: 390px;
  margin-bottom: 10px;
}

.kb-root [data-kb-search] {
  padding-left: 34px;
  background:
    linear-gradient(45deg, transparent 43%, var(--line) 43%, var(--line) 57%, transparent 57%) 16px 18px / 10px 10px no-repeat,
    radial-gradient(circle at 13px 13px, transparent 0, transparent 5px, var(--line) 5px, var(--line) 7px, transparent 7px) 3px 4px / 22px 22px no-repeat,
    var(--panel);
}

.board {
  display: grid;
  gap: 28px;
}

.kitchen-board {
  grid-template-columns: repeat(4, minmax(260px, 1fr));
}

.delivery-board {
  grid-template-columns: repeat(3, minmax(320px, 1fr));
}

.column {
  min-height: 560px;
  overflow: hidden;
  background: var(--soft);
  border: 2px solid var(--line);
  border-radius: 4px;
}

.column-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  min-height: 40px;
  padding: 8px 12px;
  background-image: radial-gradient(var(--line) 0.65px, transparent 0.65px);
  background-size: 4px 4px;
  border-bottom: 2px solid var(--line);
}

.column-header h2 {
  margin: 0;
  padding-right: 8px;
  background: var(--soft);
  font-size: 20px;
}

.count {
  min-width: 28px;
  padding: 1px 8px;
  text-align: center;
  font-weight: 700;
  background: var(--panel);
  border: 2px solid var(--line);
  border-radius: 3px;
  box-shadow: 2px 2px 0 var(--line);
}

.card-list {
  min-height: 480px;
  max-height: 660px;
  overflow-y: auto;
  padding: 12px;
  background: var(--soft);
}

.kanban-card {
  padding: 14px;
  margin-bottom: 12px;
  cursor: grab;
  background: var(--panel);
  border: 2px solid var(--line);
  border-radius: 5px;
  box-shadow: 3px 3px 0 var(--line);
}

.kanban-card.dragging {
  opacity: 0.45;
}

.card-topline {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  align-items: start;
  margin-bottom: 8px;
}

.ticket,
.price {
  color: var(--ink);
  font-size: 18px;
  font-weight: 900;
  line-height: 1.1;
}

.card-topline .chip {
  font-size: 13px;
  font-weight: 700;
}

.task-card h3,
.order-card h3 {
  margin: 8px 0 7px;
  font-size: 17px;
  line-height: 1.15;
}

.done-group-card {
  background:
    radial-gradient(var(--line) 0.45px, transparent 0.45px),
    var(--panel);
  background-size: 5px 5px;
}

.done-task-list {
  margin: 8px 0 10px 18px;
  padding: 0;
  font-size: 14px;
  line-height: 1.35;
}

.description,
.small-text {
  margin: 0 0 11px;
  font-size: 15px;
  line-height: 1.35;
  color: var(--ink);
}

.chip-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 10px;
}

.chip {
  padding: 3px 10px;
  font-size: 13px;
  line-height: 1.2;
  color: var(--ink);
  background: var(--panel);
  border: 2px solid var(--line);
  border-radius: 3px;
}

.chip.blocked {
  color: var(--panel);
  background: var(--line);
}

.chip.ready,
.paid-note {
  color: var(--ink);
  background: var(--panel);
}

.empty {
  display: grid;
  place-items: center;
  min-height: 300px;
  padding: 28px;
  color: var(--muted);
  text-align: center;
  font-size: 16px;
  font-style: normal;
}

.empty::before {
  display: block;
  margin-bottom: 14px;
  content: "♙";
  color: #999999;
  font-size: 58px;
  line-height: 1;
}

.payment-form {
  display: grid;
  grid-template-columns: 1fr;
  gap: 8px;
  margin-top: 12px;
  padding-top: 10px;
  border-top: 2px solid var(--line);
}

.payment-form button {
  min-height: 34px;
  font-weight: 900;
  cursor: pointer;
  background: var(--panel);
  border: 3px double var(--line);
  border-radius: 3px;
  box-shadow: 2px 2px 0 var(--line);
}

.payment-hint {
  padding-top: 9px;
  border-top: 2px solid var(--line);
  font-style: italic;
}

.api-links {
  display: flex;
  gap: 14px;
  margin-top: 18px;
  padding: 12px 16px;
}

.api-links a {
  color: var(--ink);
  font-weight: 900;
}

@media (max-width: 1300px) {
  .hero,
  .builder-grid,
  .tally {
    grid-template-columns: 1fr;
  }

  .primary-action,
  .form-field.wide,
  .topping-picker {
    grid-column: span 1;
  }

  .primary-action {
    min-width: 0;
    width: 100%;
  }

  .kitchen-board,
  .delivery-board {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 760px) {
  .page {
    padding: 14px;
  }

  .hero-copy {
    display: block;
  }

  .mascot {
    margin-bottom: 14px;
  }

  .kitchen-board,
  .delivery-board,
  .topping-grid {
    grid-template-columns: 1fr;
  }
}
`;
