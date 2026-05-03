package kanbanddsl

func ClientScript() string {
	return `(() => {
  if (window.__gojaKanbanRuntimeLoaded) return;
  window.__gojaKanbanRuntimeLoaded = true;

  function boardFor(element) {
    return element && element.closest('[data-kb-board-id]');
  }

  function actionBase(board) {
    return board.dataset.kbActionBase || ('/_kanban/' + encodeURIComponent(board.dataset.kbBoardId) + '/action');
  }

  function updateCounts(board) {
    board.querySelectorAll('[data-kb-column-id]').forEach(column => {
      const id = column.dataset.kbColumnId;
      const count = column.querySelectorAll('[data-kb-card-id]:not([hidden])').length;
      const badge = column.querySelector('[data-kb-count="' + CSS.escape(id) + '"]');
      if (badge) badge.textContent = String(count);
    });
  }

  function applySearch(input) {
    const board = boardFor(input);
    if (!board) return;
    const query = String(input.value || '').trim().toLowerCase();
    board.querySelectorAll('[data-kb-card-id]').forEach(card => {
      const text = card.dataset.kbSearchText || '';
      card.hidden = !!query && !text.includes(query);
    });
    updateCounts(board);
  }

  document.addEventListener('input', event => {
    const input = event.target.closest('[data-kb-search]');
    if (input) applySearch(input);
  });

  async function postAction(board, action, event) {
    const response = await fetch(actionBase(board) + '/' + encodeURIComponent(action), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'Accept': 'application/json' },
      body: JSON.stringify(event || {})
    });
    const payload = await response.json().catch(() => ({}));
    if (!response.ok || payload.ok === false) throw new Error(payload.error || response.statusText || 'Kanban action failed');
    if (payload.html) {
      const root = board.closest('[data-kb-root]') || board;
      const template = document.createElement('template');
      template.innerHTML = payload.html.trim();
      const replacement = template.content.firstElementChild;
      if (replacement) root.replaceWith(replacement);
    }
    if (payload.toast) console.info('[kanban]', payload.toast);
    return payload;
  }

  document.addEventListener('submit', async event => {
    const form = event.target.closest('[data-kb-move-form]');
    if (!form) return;
    event.preventDefault();
    const board = boardFor(form);
    const card = form.closest('[data-kb-card-id]');
    if (!board || !card) return;
    const data = new FormData(form);
    try {
      await postAction(board, 'cardMoved', {
        cardId: String(data.get('cardId') || card.dataset.kbCardId || ''),
        from: {
          columnId: String(data.get('fromColumnId') || card.dataset.kbCardColumn || ''),
          index: Number(data.get('fromIndex') || card.dataset.kbCardIndex || 0)
        },
        to: {
          columnId: String(data.get('toColumnId') || ''),
          index: Number(data.get('toIndex') || 0)
        }
      });
    } catch (error) {
      console.error('kanban move failed', error);
      alert(error.message || String(error));
    }
  });

  let dragged = null;

  document.addEventListener('dragstart', event => {
    const card = event.target.closest('[data-kb-card-id]');
    if (!card) return;
    dragged = card;
    card.classList.add('kb-dragging', 'dragging');
    event.dataTransfer.effectAllowed = 'move';
    event.dataTransfer.setData('text/plain', card.dataset.kbCardId || '');
  });

  document.addEventListener('dragend', () => {
    if (dragged) dragged.classList.remove('kb-dragging', 'dragging');
    document.querySelectorAll('.kb-drag-over').forEach(el => el.classList.remove('kb-drag-over', 'drag-over'));
    dragged = null;
  });

  function cardAfterPointer(list, y) {
    const cards = [...list.querySelectorAll('[data-kb-card-id]:not(.kb-dragging):not(.dragging)')];
    return cards.reduce((closest, child) => {
      const box = child.getBoundingClientRect();
      const offset = y - box.top - box.height / 2;
      if (offset < 0 && offset > closest.offset) return { offset, element: child };
      return closest;
    }, { offset: Number.NEGATIVE_INFINITY, element: null }).element;
  }

  document.addEventListener('dragover', event => {
    const list = event.target.closest('[data-kb-drop-column]');
    if (!list || !dragged) return;
    event.preventDefault();
    const column = list.closest('[data-kb-column-id]');
    if (column) column.classList.add('kb-drag-over', 'drag-over');
    const before = cardAfterPointer(list, event.clientY);
    if (before) list.insertBefore(dragged, before);
    else list.insertBefore(dragged, list.querySelector('[data-kb-drop-sentinel]'));
    updateCounts(boardFor(list));
  });

  document.addEventListener('drop', async event => {
    const list = event.target.closest('[data-kb-drop-column]');
    if (!list || !dragged) return;
    event.preventDefault();
    const board = boardFor(list);
    const card = dragged;
    const fromColumnId = card.dataset.kbCardColumn || '';
    const fromIndex = Number(card.dataset.kbCardIndex || 0);
    const toColumnId = list.dataset.kbDropColumn || '';
    const toCards = [...list.querySelectorAll('[data-kb-card-id]')];
    const toIndex = Math.max(0, toCards.indexOf(card));
    const visibleCardIds = toCards.map(el => el.dataset.kbCardId || '');
    try {
      await postAction(board, 'cardMoved', {
        cardId: card.dataset.kbCardId || '',
        from: { columnId: fromColumnId, index: fromIndex },
        to: { columnId: toColumnId, index: toIndex },
        visibleCardIds
      });
    } catch (error) {
      console.error('kanban drag/drop failed', error);
      window.location.reload();
    }
  });
})();`
}
