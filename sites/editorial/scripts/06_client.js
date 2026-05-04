globalThis.Herald = globalThis.Herald || {};

Herald.clientScript = `
(function () {
  function closest(element, selector) {
    while (element && element !== document) {
      if (element.matches && element.matches(selector)) return element;
      element = element.parentNode;
    }
    return null;
  }

  function replaceDossier(html) {
    var current = document.querySelector(".dossier");
    if (!current) return;

    var template = document.createElement("template");
    template.innerHTML = html.trim();
    var next = template.content.firstElementChild;
    if (!next) return;

    current.replaceWith(next);
  }

  async function fetchPanel(url, options) {
    var dossier = document.querySelector(".dossier");
    if (dossier) dossier.setAttribute("aria-busy", "true");

    var response = await fetch(url, Object.assign({
      credentials: "same-origin",
      headers: { "X-Herald-Panel": "1" }
    }, options || {}));

    if (!response.ok) throw new Error("Panel request failed: " + response.status);
    replaceDossier(await response.text());
  }

  document.addEventListener("click", function (event) {
    var link = closest(event.target, "[data-herald-story-link]");
    if (!link) return;
    if (event.defaultPrevented || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;

    event.preventDefault();
    fetchPanel(link.getAttribute("data-herald-panel-url") || link.href)
      .then(function () {
        history.pushState({}, "", link.href);
      })
      .catch(function (error) {
        console.error(error);
        window.location.href = link.href;
      });
  });

  document.addEventListener("submit", function (event) {
    var form = closest(event.target, "[data-herald-panel-form]");
    if (!form) return;

    event.preventDefault();
    fetchPanel(form.action, {
      method: form.method || "POST",
      headers: {
        "X-Herald-Panel": "1",
        "Content-Type": "application/x-www-form-urlencoded;charset=UTF-8"
      },
      body: new URLSearchParams(new FormData(form)).toString()
    }).catch(function (error) {
      console.error(error);
      form.submit();
    });
  });

  window.addEventListener("popstate", function () {
    var url = new URL(window.location.href);
    var storyId = url.searchParams.get("story");
    if (storyId) fetchPanel("/stories/" + encodeURIComponent(storyId) + "/panel");
  });
})();
`;
