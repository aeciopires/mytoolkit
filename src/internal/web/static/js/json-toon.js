// Client-side TOON encoder for the JSON to TOON Converter tool page.
// Mirrors internal/tools/jsontoon/jsontoon.go rule-for-rule (see that file's
// fixture-based tests) so this file and the Go implementation must be kept
// in sync — see .skills/json-toon/SKILL.md. Runs entirely in the browser
// with zero network activity of any kind.
(function () {
  var NUMERIC_PATTERN = /^-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?$/;

  function delimChar(d) {
    if (d === 'tab') return '\t';
    if (d === 'pipe') return '|';
    return ',';
  }

  function delimSymbol(d) {
    if (d === 'tab') return '~';
    if (d === 'pipe') return '|';
    return '';
  }

  function isPlainObject(v) {
    return v !== null && typeof v === 'object' && !Array.isArray(v);
  }

  function isPrimitive(v) {
    return v === null || typeof v === 'string' || typeof v === 'number' || typeof v === 'boolean';
  }

  function needsQuoting(s, delim, inArray) {
    if (s === '') return true;
    if (s.trim() !== s) return true;
    if (s === 'true' || s === 'false' || s === 'null') return true;
    if (NUMERIC_PATTERN.test(s)) return true;
    if (s.charAt(0) === '-') return true;
    for (var i = 0; i < s.length; i++) {
      var c = s[i];
      if (c === ':' || c === '"' || c === '\\' || c === '[' || c === ']' || c === '{' || c === '}' || s.charCodeAt(i) < 0x20) {
        return true;
      }
      if (inArray && c === delim) return true;
    }
    return false;
  }

  function quoteString(s) {
    var out = '"';
    for (var i = 0; i < s.length; i++) {
      var c = s[i];
      if (c === '"') out += '\\"';
      else if (c === '\\') out += '\\\\';
      else if (c === '\n') out += '\\n';
      else if (c === '\t') out += '\\t';
      else out += c;
    }
    return out + '"';
  }

  function formatString(s, delim, inArray) {
    return needsQuoting(s, delim, inArray) ? quoteString(s) : s;
  }

  function formatNumber(n) {
    if (n === 0) return '0'; // also normalizes -0
    var abs = Math.abs(n);
    if (abs >= 1e-6 && abs < 1e21) {
      // Avoid exponent notation and trailing zeros in the normal range.
      return String(n);
    }
    return n.toExponential().replace(/e\+?/, 'e');
  }

  function formatPrimitive(v, delim, inArray) {
    if (v === null) return 'null';
    if (typeof v === 'boolean') return v ? 'true' : 'false';
    if (typeof v === 'number') return formatNumber(v);
    return formatString(v, delim, inArray);
  }

  function uniformObjectFields(items) {
    if (items.length === 0 || !isPlainObject(items[0])) return null;
    var fields = Object.keys(items[0]);
    for (var i = 0; i < fields.length; i++) {
      if (!isPrimitive(items[0][fields[i]])) return null;
    }
    for (var j = 0; j < items.length; j++) {
      var item = items[j];
      if (!isPlainObject(item)) return null;
      var keys = Object.keys(item);
      if (keys.length !== fields.length) return null;
      for (var k = 0; k < fields.length; k++) {
        if (keys[k] !== fields[k] || !isPrimitive(item[fields[k]])) return null;
      }
    }
    return fields;
  }

  function allPrimitive(items) {
    return items.every(isPrimitive);
  }

  function indentStr(depth, indentSize) {
    return new Array(depth * indentSize + 1).join(' ');
  }

  function emitField(buf, key, v, depth, opts) {
    var ind = indentStr(depth, opts.indentSize);
    if (isPlainObject(v)) {
      buf.push(ind + key + ':');
      Object.keys(v).forEach(function (k) {
        emitField(buf, k, v[k], depth + 1, opts);
      });
    } else if (Array.isArray(v)) {
      emitArrayField(buf, key, v, depth, opts);
    } else {
      buf.push(ind + key + ': ' + formatPrimitive(v, opts.delim, false));
    }
  }

  function emitArrayField(buf, key, items, depth, opts) {
    var ind = indentStr(depth, opts.indentSize);
    var n = items.length;

    if (n === 0) {
      buf.push(ind + key + '[0]:');
      return;
    }

    var fields = uniformObjectFields(items);
    if (fields) {
      buf.push(ind + key + '[' + n + opts.delimSym + ']{' + fields.join(opts.delim) + '}:');
      var rowIndent = indentStr(depth + 1, opts.indentSize);
      items.forEach(function (item) {
        var row = fields.map(function (f) { return formatPrimitive(item[f], opts.delim, true); });
        buf.push(rowIndent + row.join(opts.delim));
      });
      return;
    }

    if (allPrimitive(items)) {
      var parts = items.map(function (it) { return formatPrimitive(it, opts.delim, true); });
      buf.push(ind + key + '[' + n + opts.delimSym + ']: ' + parts.join(opts.delim));
      return;
    }

    buf.push(ind + key + '[' + n + ']:');
    var elemIndent = indentStr(depth + 1, opts.indentSize);
    items.forEach(function (item) {
      emitListElement(buf, item, elemIndent, depth + 1, opts);
    });
  }

  function emitListElement(buf, v, ind, depth, opts) {
    if (Array.isArray(v)) {
      var parts = v.map(function (it) { return formatPrimitive(it, opts.delim, true); });
      buf.push(ind + '- [' + v.length + ']: ' + parts.join(opts.delim));
    } else if (isPlainObject(v)) {
      buf.push(ind + '-:');
      Object.keys(v).forEach(function (k) {
        emitField(buf, k, v[k], depth + 1, opts);
      });
    } else {
      buf.push(ind + '- ' + formatPrimitive(v, opts.delim, false));
    }
  }

  function emitRoot(buf, v, opts) {
    if (isPlainObject(v)) {
      Object.keys(v).forEach(function (k) { emitField(buf, k, v[k], 0, opts); });
    } else if (Array.isArray(v)) {
      emitArrayField(buf, '', v, 0, opts);
    } else {
      buf.push(formatPrimitive(v, opts.delim, false));
    }
  }

  // window.jsonToToon(text, { delimiter, indentSize }) -> string
  // Throws on empty input or invalid JSON (mirrors apperr.ErrEmptyInput /
  // INVALID_JSON from the Go implementation, as plain Error objects).
  window.jsonToToon = function (text, options) {
    options = options || {};
    var delimiter = options.delimiter || 'comma';
    var indentSize = options.indentSize > 0 ? options.indentSize : 2;
    if (!text || !text.trim()) {
      throw new Error('input must not be empty');
    }
    var parsed;
    try {
      parsed = JSON.parse(text);
    } catch (e) {
      throw new Error('invalid JSON: ' + e.message);
    }
    var buf = [];
    emitRoot(buf, parsed, {
      delim: delimChar(delimiter),
      delimSym: delimSymbol(delimiter),
      indentSize: indentSize,
    });
    return buf.join('\n') + '\n';
  };
})();
