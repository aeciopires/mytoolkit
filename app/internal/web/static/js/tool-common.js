(function () {
  var panel = document.querySelector('.tool-panel');
  if (!panel) return;
  var slug = panel.dataset.slug;
  var input = document.getElementById('tool-input');
  var output = document.getElementById('tool-output');
  var errorBanner = document.getElementById('tool-error');
  var timer = null;

  // A JSON number string (e.g. a <select> populated with "2"/"4") — used to
  // decide whether a <select>'s value should be sent as a number, since a
  // Go Options struct field like `Indent int` rejects a JSON string.
  var NUMERIC_VALUE = /^-?\d+(\.\d+)?$/;

  function collectOptions() {
    var opts = {};
    panel.querySelectorAll('[data-option]').forEach(function (el) {
      var name = el.name;
      if (!name) return;
      if (el.type === 'checkbox') {
        opts[name] = el.checked;
      } else if (el.type === 'number') {
        opts[name] = Number(el.value);
      } else if (el.tagName === 'SELECT' && NUMERIC_VALUE.test(el.value)) {
        // <select> option values are always strings in the DOM. If every
        // option happens to be numeric (e.g. an indent/size picker backed
        // by a Go `int` field), send it as a JSON number — the backend's
        // json.Unmarshal has no coercion and 400s on a numeric string.
        opts[name] = Number(el.value);
      } else {
        opts[name] = el.value;
      }
    });
    return opts;
  }

  function showError(msg) {
    if (!errorBanner) return;
    errorBanner.textContent = msg;
    errorBanner.hidden = false;
  }
  function clearError() {
    if (!errorBanner) return;
    errorBanner.hidden = true;
    errorBanner.textContent = '';
  }

  function run() {
    clearError();
    var body = JSON.stringify({ input: input ? input.value : '', options: collectOptions() });
    fetch('/api/v1/tools/' + slug, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: body })
      .then(function (res) {
        var contentType = res.headers.get('Content-Type') || '';
        if (contentType.indexOf('image/') === 0) {
          return res.blob().then(function (blob) {
            var url = URL.createObjectURL(blob);
            var img = panel.querySelector('.tool-image-output');
            if (img) { img.src = url; img.hidden = false; }
            window.__mytoolkitLastBlobUrl = url;
          });
        }
        return res.json().then(function (json) {
          if (!json.success) {
            showError((json.error && json.error.message) || 'request failed');
            if (output) output.value = '';
            return;
          }
          if (output) {
            output.value = (json.data && json.data.output !== undefined) ? json.data.output : JSON.stringify(json.data, null, 2);
          }
        });
      })
      .catch(function (err) { showError(String(err)); });
  }

  function debounce() {
    clearTimeout(timer);
    timer = setTimeout(run, 400);
  }

  // data-client-side opts a page out of this file's automatic
  // input/option-driven fetch() wiring, so the page's own inline script
  // fully owns when/how conversion happens. Two different reasons a tool
  // uses it: JSON to TOON Converter never calls the REST API at all (true
  // client-side conversion); JSON Tree Viewer still calls the REST API, but
  // only on an explicit "Generate Tree View" click, not on every keystroke.
  // Copy/reset/download button wiring below still applies either way; only
  // the fetch-based run() wiring is skipped here.
  var clientSide = panel.hasAttribute('data-client-side');

  if (!clientSide) {
    if (input) input.addEventListener('input', debounce);
    panel.querySelectorAll('[data-option]').forEach(function (el) {
      el.addEventListener('change', debounce);
    });
  }

  panel.querySelectorAll('[data-action]').forEach(function (btn) {
    btn.addEventListener('click', function () {
      var action = btn.dataset.action;
      if (action === 'copy' && output) {
        navigator.clipboard && navigator.clipboard.writeText(output.value);
      } else if (action === 'reset') {
        if (input) input.value = '';
        if (output) output.value = '';
        clearError();
      } else if (action === 'generate') {
        run();
      } else if (action === 'download') {
        var url = window.__mytoolkitLastBlobUrl;
        if (url) {
          var a = document.createElement('a');
          a.href = url;
          a.download = slug + '.png';
          document.body.appendChild(a);
          a.click();
          a.remove();
        }
      }
    });
  });

  if (!clientSide) {
    window.mytoolkitRun = run;

    if (panel.hasAttribute('data-autorun')) {
      run();
    }
  }
})();
