package kanbanddsl

func ClientScript() string {
	return `(() => {
  if (window.__gojaKanbanRuntimeLoaded) return;
  window.__gojaKanbanRuntimeLoaded = true;

  function debug(...args) {
    try {
      if (window.localStorage && window.localStorage.getItem('gojaKanbanDebug') === '1') {
        console.debug('[kanban.debug]', ...args);
      }
    } catch (_) {}
  }

  function ensureRuntimeStyles() {
    if (document.getElementById('goja-kanban-runtime-styles')) return;
    const style = document.createElement('style');
    style.id = 'goja-kanban-runtime-styles';
    style.textContent = '\n' +
      '  [data-kb-card-id][draggable="true"] { cursor: grab; }\n' +
      '  [data-kb-card-id][draggable="true"]:active { cursor: grabbing; }\n' +
      '  [data-kb-card-id][draggable="true"],\n' +
      '  [data-kb-card-id][draggable="true"] * { user-select: none; -webkit-user-select: none; }\n' +
      '  [data-kb-card-id].kb-dragging { opacity: .45; }\n' +
      '  [data-kb-column-id].kb-drag-over { outline: 3px dashed currentColor; outline-offset: 4px; }\n' +
      '  .kb-action-menu { display: inline-flex; flex-direction: column; gap: .25rem; margin-top: .5rem; padding: .5rem; border: 1px solid currentColor; background: Canvas; color: CanvasText; z-index: 20; }\n' +
      '  .kb-action-menu[hidden] { display: none; }\n' +
      '  .kb-sr-only { position: absolute !important; width: 1px; height: 1px; padding: 0; margin: -1px; overflow: hidden; clip: rect(0,0,0,0); white-space: nowrap; border: 0; }\n';
    const parent = document.head || document.body || document.documentElement;
    if (parent) parent.appendChild(style);
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', ensureRuntimeStyles, { once: true });
  } else {
    ensureRuntimeStyles();
  }

  function boardFor(element) {
    if (!element) return null;
    const board = element.closest('[data-kb-board-id]');
    if (board) return board;
    const root = element.closest('[data-kb-root]');
    return root ? root.querySelector('[data-kb-board-id]') : null;
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

  function liveRegion() {
    let region = document.querySelector('[data-kb-live-region]');
    if (region) return region;
    region = document.createElement('div');
    region.className = 'kb-sr-only';
    region.setAttribute('data-kb-live-region', '');
    region.setAttribute('aria-live', 'polite');
    region.setAttribute('aria-atomic', 'true');
    (document.body || document.documentElement).appendChild(region);
    return region;
  }

  function announce(message) {
    const region = liveRegion();
    region.textContent = '';
    window.setTimeout(() => { region.textContent = String(message || ''); }, 0);
  }

  function cardsInList(list) {
    return [...list.querySelectorAll('[data-kb-card-id]')].filter(card => !card.hidden && !card.matches('[data-kb-drop-sentinel]'));
  }

  function cardListForColumn(board, columnId) {
    if (!board || !columnId) return null;
    return board.querySelector('[data-kb-drop-column="' + CSS.escape(columnId) + '"]');
  }

  function columnsFor(board) {
    if (!board) return [];
    return [...board.querySelectorAll('[data-kb-column-id]')].map(column => ({
      id: column.dataset.kbColumnId || '',
      title: column.dataset.kbColumnTitle || column.dataset.kbColumnId || '',
      element: column,
      list: cardListForColumn(board, column.dataset.kbColumnId || '')
    })).filter(column => column.id && column.list);
  }

  async function postAction(board, action, event) {
    const url = actionBase(board) + '/' + encodeURIComponent(action);
    debug('postAction', { boardId: board && board.dataset.kbBoardId, action, url, event });
    const response = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'Accept': 'application/json' },
      body: JSON.stringify(event || {})
    });
    const payload = await response.json().catch(() => ({}));
    if (!response.ok || payload.ok === false) throw new Error(payload.error || response.statusText || 'Kanban action failed');
    debug('postAction response', { action, status: response.status, hasHtml: !!payload.html, payload });
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

  let openMenu = null;

  function closeActionMenu({ restoreFocus = false } = {}) {
    if (!openMenu) return;
    const { menu, trigger } = openMenu;
    if (trigger) trigger.setAttribute('aria-expanded', 'false');
    if (menu && menu.parentNode) menu.parentNode.removeChild(menu);
    openMenu = null;
    if (restoreFocus && trigger) trigger.focus();
  }

  async function moveCard(card, toColumnId, toIndex) {
    const board = boardFor(card);
    if (!board || !card || !toColumnId) return;
    const cardId = card.dataset.kbCardId || '';
    const fromColumnId = card.dataset.kbCardColumn || '';
    const fromIndex = Number(card.dataset.kbCardIndex || 0);
    try {
      const payload = await postAction(board, 'cardMoved', {
        cardId,
        from: { columnId: fromColumnId, index: fromIndex },
        to: { columnId: toColumnId, index: toIndex }
      });
      announce(payload.announcement || ('Moved card ' + cardId));
      window.setTimeout(() => {
        const moved = document.querySelector('[data-kb-card-id="' + CSS.escape(cardId) + '"]');
        if (moved) moved.focus();
      }, 0);
    } catch (error) {
      console.error('kanban move failed', error);
      announce(error.message || String(error));
      alert(error.message || String(error));
    }
  }

  function menuButton(label, onClick, disabled) {
    const button = document.createElement('button');
    button.type = 'button';
    button.setAttribute('role', 'menuitem');
    button.textContent = label;
    button.disabled = !!disabled;
    button.addEventListener('click', async () => {
      closeActionMenu();
      await onClick();
    });
    return button;
  }

  function openActionMenu(trigger) {
    const card = trigger && trigger.closest('[data-kb-card-id]');
    const board = boardFor(card);
    if (!card || !board) return;
    closeActionMenu();
    const currentColumnId = card.dataset.kbCardColumn || '';
    const currentList = cardListForColumn(board, currentColumnId);
    const currentCards = currentList ? cardsInList(currentList) : [];
    const currentIndex = Math.max(0, currentCards.indexOf(card));
    const menu = document.createElement('div');
    menu.className = 'kb-action-menu';
    menu.setAttribute('role', 'menu');
    menu.setAttribute('aria-label', 'Card actions');
    menu.setAttribute('data-kb-action-menu', '');

    menu.appendChild(menuButton('Move up', () => moveCard(card, currentColumnId, Math.max(0, currentIndex - 1)), currentIndex <= 0));
    menu.appendChild(menuButton('Move down', () => moveCard(card, currentColumnId, Math.min(currentCards.length - 1, currentIndex + 1)), currentIndex >= currentCards.length - 1));
    menu.appendChild(menuButton('Move to top', () => moveCard(card, currentColumnId, 0), currentIndex <= 0));
    menu.appendChild(menuButton('Move to bottom', () => moveCard(card, currentColumnId, Math.max(0, currentCards.length - 1)), currentIndex >= currentCards.length - 1));

    columnsFor(board).forEach(column => {
      if (column.id === currentColumnId) return;
      const targetIndex = cardsInList(column.list).length;
      menu.appendChild(menuButton('Move to ' + column.title, () => moveCard(card, column.id, targetIndex), false));
    });

    trigger.setAttribute('aria-expanded', 'true');
    card.appendChild(menu);
    openMenu = { menu, trigger };
    const first = menu.querySelector('button:not([disabled])') || menu.querySelector('button');
    if (first) first.focus();
  }

  document.addEventListener('click', event => {
    const trigger = event.target.closest('[data-kb-card-actions]');
    if (trigger) {
      event.preventDefault();
      openActionMenu(trigger);
      return;
    }
    if (openMenu && !event.target.closest('[data-kb-action-menu]')) closeActionMenu();
  });

  document.addEventListener('keydown', event => {
    if (openMenu && event.target.closest('[data-kb-action-menu]')) {
      const items = [...openMenu.menu.querySelectorAll('button:not([disabled])')];
      const index = items.indexOf(document.activeElement);
      if (event.key === 'Escape') {
        event.preventDefault();
        closeActionMenu({ restoreFocus: true });
      } else if (event.key === 'ArrowDown') {
        event.preventDefault();
        items[(index + 1 + items.length) % items.length]?.focus();
      } else if (event.key === 'ArrowUp') {
        event.preventDefault();
        items[(index - 1 + items.length) % items.length]?.focus();
      }
      return;
    }
    const card = event.target.closest('[data-kb-card-id]');
    if (!card || event.target.closest('button, a, input, select, textarea')) return;
    if (event.key === 'Enter' || event.key === ' ') {
      const trigger = card.querySelector('[data-kb-card-actions]');
      if (trigger) {
        event.preventDefault();
        openActionMenu(trigger);
      }
    }
  });

  let dragged = null;

  document.addEventListener('dragstart', event => {
    const card = event.target.closest('[data-kb-card-id]');
    if (!card) return;
    dragged = card;
    debug('dragstart', { cardId: card.dataset.kbCardId, columnId: card.dataset.kbCardColumn, index: card.dataset.kbCardIndex });
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
    debug('drop', { cardId: card.dataset.kbCardId, fromColumnId, fromIndex, toColumnId, toIndex, visibleCardIds });
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
