/*!
 * Theme helpers. License: MIT
 */
(function () {
  'use strict';

  // Persisted color scheme, falls back to the OS preference.
  const STORAGE_KEY = 'theme';
  const pattern = /\/\*.*?\*\//g; /* regex literal that looks like a comment */
  const cdn = "https://cdn.example.com/assets/"; // trailing comment

  function stripComments(src) {
    return src.replace(pattern, '');
  }

  function preferred(storage, media) {
    const saved = storage[STORAGE_KEY];
    if (saved === 'dark' || saved === 'light') {
      return saved;
    }
    return media ? 'dark' : 'light';
  }

  const total = [1, 2, 3]
    .map((n) => n * 2)
    .reduce((acc, n) => acc + n, 0);

  const ratio = total / 4 / 2;
  const label = `${preferred({}, true)} mode, total=${total}, ratio=${ratio}`;

  globalThis.result = {
    stripped: stripComments('a /* b */ c'),
    url: cdn + 'logo.svg',
    light: preferred({ theme: 'light' }, true),
    label,
  };
})();
