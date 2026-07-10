(function () {
  // --- Navigation drawer (M3 modal navigation drawer pattern) ---
  var drawer = document.getElementById('nav-drawer');
  var scrim = document.getElementById('nav-scrim');
  var toggleBtn = document.getElementById('nav-drawer-toggle');
  var closeBtn = document.getElementById('nav-drawer-close');

  function openDrawer() {
    drawer.classList.add('open');
    scrim.hidden = false;
    requestAnimationFrame(function () { scrim.classList.add('open'); });
    drawer.setAttribute('aria-hidden', 'false');
    toggleBtn.setAttribute('aria-expanded', 'true');
  }

  function closeDrawer() {
    drawer.classList.remove('open');
    scrim.classList.remove('open');
    drawer.setAttribute('aria-hidden', 'true');
    toggleBtn.setAttribute('aria-expanded', 'false');
    setTimeout(function () { if (!drawer.classList.contains('open')) scrim.hidden = true; }, 200);
  }

  if (toggleBtn && drawer && scrim) {
    toggleBtn.addEventListener('click', function () {
      if (drawer.classList.contains('open')) closeDrawer(); else openDrawer();
    });
    if (closeBtn) closeBtn.addEventListener('click', closeDrawer);
    scrim.addEventListener('click', closeDrawer);
    document.addEventListener('keydown', function (e) {
      if (e.key === 'Escape' && drawer.classList.contains('open')) closeDrawer();
    });
  }

  // --- Search bar: indexes each tool's name + description (the same
  // content shown in its homepage card / tool-page hero card), entirely
  // client-side against window.MYTOOLKIT_TOOLS (no network call). ---
  var searchInput = document.getElementById('tool-search');
  var resultsList = document.getElementById('search-results');
  var tools = window.MYTOOLKIT_TOOLS || [];

  function renderResults(matches) {
    resultsList.innerHTML = '';
    if (matches.length === 0) {
      resultsList.hidden = true;
      return;
    }
    matches.forEach(function (t) {
      var li = document.createElement('li');
      var a = document.createElement('a');
      a.href = '/tools/' + t.slug;
      a.className = 'nav-list-item';
      a.innerHTML = '<span class="nav-list-icon" aria-hidden="true">' + t.emoji + '</span><span>' + t.name + '</span>';
      li.appendChild(a);
      resultsList.appendChild(li);
    });
    resultsList.hidden = false;
  }

  function search(query) {
    var q = query.trim().toLowerCase();
    if (!q) {
      renderResults([]);
      return;
    }
    var matches = tools.filter(function (t) {
      return t.name.toLowerCase().indexOf(q) !== -1 || t.description.toLowerCase().indexOf(q) !== -1;
    });
    renderResults(matches);
  }

  if (searchInput && resultsList) {
    searchInput.addEventListener('input', function () { search(searchInput.value); });
    searchInput.addEventListener('keydown', function (e) {
      if (e.key === 'Enter') {
        var first = resultsList.querySelector('a');
        if (first) { window.location.href = first.getAttribute('href'); }
      } else if (e.key === 'Escape') {
        searchInput.value = '';
        renderResults([]);
        searchInput.blur();
      }
    });
    document.addEventListener('click', function (e) {
      if (!e.target.closest('.search-field')) {
        resultsList.hidden = true;
      }
    });
    searchInput.addEventListener('focus', function () {
      if (searchInput.value.trim()) search(searchInput.value);
    });
  }
})();
