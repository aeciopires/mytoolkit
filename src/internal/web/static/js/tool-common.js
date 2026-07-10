(function () {
  var panel = document.querySelector('.tool-panel');
  if (!panel) return;
  var slug = panel.dataset.slug;
  var input = document.getElementById('tool-input');
  var output = document.getElementById('tool-output');
  var errorBanner = document.getElementById('tool-error');
  var timer = null;

  function collectOptions() {
    var opts = {};
    panel.querySelectorAll('[data-option]').forEach(function (el) {
      var name = el.name;
      if (!name) return;
      if (el.type === 'checkbox') {
        opts[name] = el.checked;
      } else if (el.type === 'number') {
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

  // Tools carrying data-client-side (e.g. JSON to TOON Converter) never call
  // the REST API from their web page — the page's own inline script owns
  // input -> output conversion entirely in the browser. Copy/reset/download
  // button wiring below still applies; only the fetch-based run() wiring is
  // skipped here.
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
